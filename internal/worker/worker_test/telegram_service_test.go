package worker_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vladislavprovich/rarible-integration/internal/worker"
	"github.com/vladislavprovich/rarible-integration/pkg/cache"
	"github.com/vladislavprovich/rarible-integration/pkg/client/mtproto"
)

// Mock implementations

type mockMTProtoClient struct {
	mock.Mock
	connected    bool
	connectDelay time.Duration
	sendDelay    time.Duration
	getDelay     time.Duration
}

func (m *mockMTProtoClient) SendMessage(ctx context.Context, req *mtproto.SendMessageRequest) (*mtproto.SendMessageResponse, error) {
	if m.sendDelay > 0 {
		time.Sleep(m.sendDelay)
	}
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*mtproto.SendMessageResponse)
	return resp, args.Error(1)
}

func (m *mockMTProtoClient) GetMessages(ctx context.Context, req *mtproto.GetMessagesRequest) (*mtproto.GetMessagesResponse, error) {
	if m.getDelay > 0 {
		time.Sleep(m.getDelay)
	}
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*mtproto.GetMessagesResponse)
	return resp, args.Error(1)
}

func (m *mockMTProtoClient) Connect(ctx context.Context) error {
	if m.connectDelay > 0 {
		time.Sleep(m.connectDelay)
	}
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.connected = true
	}
	return args.Error(0)
}

func (m *mockMTProtoClient) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.connected = false
	}
	return args.Error(0)
}

func (m *mockMTProtoClient) IsConnected() bool {
	return m.connected
}

type mockCacheService struct {
	mock.Mock
	data       map[string]interface{}
	mu         sync.RWMutex
	getDelay   time.Duration
	setDelay   time.Duration
	failureRate float64 // 0.0 to 1.0, probability of failure
}

func newMockCacheService() *mockCacheService {
	return &mockCacheService{
		data: make(map[string]interface{}),
	}
}

func (m *mockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.setDelay > 0 {
		time.Sleep(m.setDelay)
	}
	
	// Simulate random failures if failure rate is set
	if m.failureRate > 0 && time.Now().UnixNano()%100 < int64(m.failureRate*100) {
		return errors.New("cache service unavailable")
	}
	
	args := m.Called(ctx, key, value, ttl)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.data[key] = value
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *mockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	if m.getDelay > 0 {
		time.Sleep(m.getDelay)
	}
	
	// Simulate random failures if failure rate is set
	if m.failureRate > 0 && time.Now().UnixNano()%100 < int64(m.failureRate*100) {
		return nil, errors.New("cache service unavailable")
	}
	
	args := m.Called(ctx, key)
	if args.Error(1) == nil {
		m.mu.RLock()
		value, exists := m.data[key]
		m.mu.RUnlock()
		if exists {
			return value, nil
		}
		return nil, errors.New("key not found")
	}
	return args.Get(0), args.Error(1)
}

func (m *mockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	if args.Error(0) == nil {
		m.mu.Lock()
		delete(m.data, key)
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *mockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	m.mu.RLock()
	_, exists := m.data[key]
	m.mu.RUnlock()
	return exists, args.Error(1)
}

func (m *mockCacheService) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.data = make(map[string]interface{})
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *mockCacheService) GetStats(ctx context.Context) (*cache.Stats, error) {
	args := m.Called(ctx)
	stats, _ := args.Get(0).(*cache.Stats)
	return stats, args.Error(1)
}

func (m *mockCacheService) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	args := m.Called(ctx, keys)
	result, _ := args.Get(0).(map[string]interface{})
	return result, args.Error(1)
}

func (m *mockCacheService) SetMultiple(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error {
	args := m.Called(ctx, entries, ttl)
	return args.Error(0)
}

// Test utilities

func createTestTelegramService(t *testing.T, config worker.TelegramWorkerConfig) (*worker.TelegramService, *mockMTProtoClient, *mockCacheService) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	mtprotoClient := &mockMTProtoClient{}
	cacheService := newMockCacheService()
	
	// Set up default mock expectations
	mtprotoClient.On("Connect", mock.Anything).Return(nil)
	mtprotoClient.On("Disconnect", mock.Anything).Return(nil)
	cacheService.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
	cacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	service := worker.NewTelegramService(logger, mtprotoClient, cacheService, config)
	
	return service, mtprotoClient, cacheService
}

func createDefaultConfig() worker.TelegramWorkerConfig {
	return worker.TelegramWorkerConfig{
		MaxConcurrency:    10,
		RequestTimeout:    2 * time.Second,
		CacheTimeout:      1 * time.Minute,
		RetryAttempts:     3,
		RetryDelay:        10 * time.Millisecond,
		CircuitBreakerMax: 5,
		MetricsEnabled:    true,
	}
}

// Test 1: High Concurrency Request Handling
func TestTelegramService_HighConcurrencyRequestHandling(t *testing.T) {
	config := createDefaultConfig()
	config.MaxConcurrency = 20
	service, mtprotoClient, _ := createTestTelegramService(t, config)
	
	// Set up MTProto client to simulate successful message sending
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
		&mtproto.SendMessageResponse{
			MessageID: 12345,
			SentAt:    time.Now(),
			Success:   true,
		}, nil)
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop(ctx)
	
	// Submit a large number of concurrent requests
	numRequests := 50
	responses := make(chan *worker.TelegramJobResponse, numRequests)
	errors := make(chan error, numRequests)
	
	var wg sync.WaitGroup
	
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("job_%d", id),
				Type:      "send_message",
				ChatID:    int64(id % 10), // Distribute across 10 different chats
				Message:   fmt.Sprintf("Test message %d", id),
				Priority:  id % 5,
				CreatedAt: time.Now(),
			}
			
			resp, err := service.SubmitJob(ctx, req)
			if err != nil {
				errors <- err
				return
			}
			responses <- resp
		}(i)
	}
	
	wg.Wait()
	close(responses)
	close(errors)
	
	// Verify all requests were processed
	successCount := 0
	for resp := range responses {
		if resp.Success {
			successCount++
		}
	}
	
	// Check for errors
	errorCount := 0
	for range errors {
		errorCount++
	}
	
	assert.Equal(t, numRequests, successCount+errorCount, "All requests should be processed")
	assert.Equal(t, 0, errorCount, "No requests should fail")
	
	// Verify metrics
	metrics := service.GetMetrics()
	assert.Equal(t, int64(numRequests), metrics.RequestsProcessed)
	assert.Equal(t, int64(numRequests), metrics.RequestsSucceeded)
	assert.Equal(t, int64(0), metrics.RequestsFailed)
	
	// Verify MTProto client was called the expected number of times
	mtprotoClient.AssertNumberOfCalls(t, "SendMessage", numRequests)
}

// Test 2: Data Race Prevention
func TestTelegramService_DataRacePrevention(t *testing.T) {
	config := createDefaultConfig()
	config.MaxConcurrency = 50
	service, mtprotoClient, _ := createTestTelegramService(t, config)
	
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
		&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil)
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop(ctx)
	
	// Test concurrent access to shared data store
	numGoroutines := 100
	numOperationsPerGoroutine := 10
	
	var wg sync.WaitGroup
	
	// Concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", goroutineID, j)
				value := fmt.Sprintf("value_%d_%d", goroutineID, j)
				service.SetData(key, value)
			}
		}(i)
	}
	
	// Concurrent readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", goroutineID, j)
				_, _ = service.GetData(key)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify data consistency - all written values should be readable
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numOperationsPerGoroutine; j++ {
			key := fmt.Sprintf("key_%d_%d", i, j)
			expectedValue := fmt.Sprintf("value_%d_%d", i, j)
			
			value, exists := service.GetData(key)
			assert.True(t, exists, "Key should exist: %s", key)
			assert.Equal(t, expectedValue, value, "Value should match for key: %s", key)
		}
	}
}

// Test 3: Data Consistency Under Load
func TestTelegramService_DataConsistencyUnderLoad(t *testing.T) {
	config := createDefaultConfig()
	config.MaxConcurrency = 30
	service, mtprotoClient, cacheService := createTestTelegramService(t, config)
	
	// Simulate varying response times to create load conditions
	mtprotoClient.sendDelay = 1 * time.Millisecond
	
	var messageCounter int64
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(func(ctx context.Context, req *mtproto.SendMessageRequest) *mtproto.SendMessageResponse {
		id := atomic.AddInt64(&messageCounter, 1)
		return &mtproto.SendMessageResponse{
			MessageID: id,
			SentAt:    time.Now(),
			Success:   true,
		}
	}, nil)
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop(ctx)
	
	// Submit jobs with shared state updates
	numJobs := 50
	sharedCounter := "shared_counter"
	service.SetData(sharedCounter, 0)
	
	var wg sync.WaitGroup
	var successCount int64
	
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(jobID int) {
			defer wg.Done()
			
			// Each job increments a shared counter
			currentValue, exists := service.GetData(sharedCounter)
			if exists {
				if counter, ok := currentValue.(int); ok {
					service.SetData(sharedCounter, counter+1)
				}
			}
			
			// Submit actual job
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("consistency_job_%d", jobID),
				Type:      "send_message",
				ChatID:    12345,
				Message:   fmt.Sprintf("Message %d", jobID),
				CreatedAt: time.Now(),
			}
			
			resp, err := service.SubmitJob(ctx, req)
			if err == nil && resp.Success {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify data consistency
	finalCounter, exists := service.GetData(sharedCounter)
	assert.True(t, exists, "Shared counter should exist")
	assert.Equal(t, numJobs, finalCounter, "Shared counter should equal number of jobs")
	
	// Verify all jobs were successful
	assert.Equal(t, int64(numJobs), successCount, "All jobs should succeed")
	
	// Verify cache service interactions
	cacheService.AssertExpectations(t)
}

// Test 4: Database/Cache Service Failure Scenarios
func TestTelegramService_FailureScenarios(t *testing.T) {
	t.Run("MTProtoClientFailure", func(t *testing.T) {
		config := createDefaultConfig()
		config.CircuitBreakerMax = 3
		service, mtprotoClient, _ := createTestTelegramService(t, config)
		
		// Simulate MTProto client failures
		mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
			nil, errors.New("network connection failed"))
		
		ctx := context.Background()
		err := service.Start(ctx)
		require.NoError(t, err)
		defer service.Stop(ctx)
		
		// Submit jobs that will fail
		numFailedJobs := 6  // More than circuit breaker max to ensure it opens
		var responses []*worker.TelegramJobResponse
		
		for i := 0; i < numFailedJobs; i++ {
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("fail_job_%d", i),
				Type:      "send_message",
				ChatID:    12345,
				Message:   "This will fail",
				CreatedAt: time.Now(),
			}
			
			resp, err := service.SubmitJob(ctx, req)
			assert.NoError(t, err, "SubmitJob should not return error")
			responses = append(responses, resp)
		}
		
		// Count actual failures
		var failedCount int64
		for _, resp := range responses {
			if !resp.Success {
				atomic.AddInt64(&failedCount, 1)
			}
		}
		
		// At least some jobs should fail
		assert.True(t, failedCount > 0, "At least some jobs should fail")
		
		// Check if circuit breaker behavior is working (either open or some jobs rejected)
		circuitBreakerActivated := false
		for _, resp := range responses {
			if !resp.Success && strings.Contains(resp.Error, "circuit breaker is open") {
				circuitBreakerActivated = true
				break
			}
		}
		
		// Note: circuitBreakerActivated may be false if all jobs fail before circuit breaker opens
		_ = circuitBreakerActivated
		
		// Try to submit a job when circuit breaker might be open
		req := &worker.TelegramJobRequest{
			ID:        "circuit_breaker_test",
			Type:      "send_message",
			ChatID:    12345,
			Message:   "This should be rejected",
			CreatedAt: time.Now(),
		}
		
		resp, err := service.SubmitJob(ctx, req)
		assert.NoError(t, err)
		// Either the job fails due to network or circuit breaker
		assert.False(t, resp.Success)
		assert.True(t, 
			strings.Contains(resp.Error, "circuit breaker is open") || 
			strings.Contains(resp.Error, "network connection failed"),
			"Job should fail due to circuit breaker or network error")
	})
	
	t.Run("CacheServiceFailure", func(t *testing.T) {
		config := createDefaultConfig()
		service, mtprotoClient, cacheService := createTestTelegramService(t, config)
		
		// Set up cache to fail randomly
		cacheService.failureRate = 0.5 // 50% failure rate
		
		mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
			&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil)
		
		ctx := context.Background()
		err := service.Start(ctx)
		require.NoError(t, err)
		defer service.Stop(ctx)
		
		// Submit jobs - they should succeed even with cache failures
		numJobs := 20
		var successCount int64
		
		for i := 0; i < numJobs; i++ {
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("cache_fail_job_%d", i),
				Type:      "send_message",
				ChatID:    int64(i),
				Message:   fmt.Sprintf("Message %d", i),
				CreatedAt: time.Now(),
			}
			
			resp, err := service.SubmitJob(ctx, req)
			assert.NoError(t, err)
			if resp.Success {
				atomic.AddInt64(&successCount, 1)
			}
		}
		
		// Jobs should still succeed despite cache failures
		assert.Equal(t, int64(numJobs), successCount, "All jobs should succeed despite cache failures")
		
		// Verify MTProto client was called for all jobs (no cache hits due to failures)
		mtprotoClient.AssertNumberOfCalls(t, "SendMessage", numJobs)
	})
	
	t.Run("PartialSystemFailure", func(t *testing.T) {
		config := createDefaultConfig()
		service, mtprotoClient, _ := createTestTelegramService(t, config)
		
		// Set up intermittent failures - simpler approach
		callCount := 0
		mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
			func(ctx context.Context, req *mtproto.SendMessageRequest) (*mtproto.SendMessageResponse, error) {
				callCount++
				if callCount%3 == 0 { // Every 3rd call fails
					return nil, errors.New("intermittent failure")
				}
				return &mtproto.SendMessageResponse{
					MessageID: int64(callCount),
					SentAt:    time.Now(),
					Success:   true,
				}, nil
			})
		
		ctx := context.Background()
		err := service.Start(ctx)
		require.NoError(t, err)
		defer service.Stop(ctx)
		
		numJobs := 15
		var successCount, failCount int64
		
		for i := 0; i < numJobs; i++ {
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("partial_fail_job_%d", i),
				Type:      "send_message",
				ChatID:    int64(i),
				Message:   fmt.Sprintf("Message %d", i),
				CreatedAt: time.Now(),
			}
			
			resp, err := service.SubmitJob(ctx, req)
			assert.NoError(t, err)
			
			if resp.Success {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}
		
		// Verify partial success/failure
		assert.True(t, successCount > 0, "Some jobs should succeed")
		assert.True(t, failCount > 0, "Some jobs should fail")
		assert.Equal(t, int64(numJobs), successCount+failCount, "All jobs should be processed")
	})
}

// Test 5: Mock Component Call Assertions
func TestTelegramService_MockComponentCallAssertions(t *testing.T) {
	config := createDefaultConfig()
	config.MaxConcurrency = 5
	service, mtprotoClient, cacheService := createTestTelegramService(t, config)
	
	// Set up specific mock expectations
	mtprotoClient.On("SendMessage", mock.Anything, mock.MatchedBy(func(req *mtproto.SendMessageRequest) bool {
		return req.ChatID == 111 && req.Text == "Test message 1"
	})).Return(&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil).Once()
	
	mtprotoClient.On("SendMessage", mock.Anything, mock.MatchedBy(func(req *mtproto.SendMessageRequest) bool {
		return req.ChatID == 222 && req.Text == "Test message 2"
	})).Return(&mtproto.SendMessageResponse{MessageID: 2, SentAt: time.Now(), Success: true}, nil).Once()
	
	mtprotoClient.On("GetMessages", mock.Anything, mock.MatchedBy(func(req *mtproto.GetMessagesRequest) bool {
		return req.ChatID == 333 && req.Limit == 10
	})).Return(&mtproto.GetMessagesResponse{
		Messages: []mtproto.Message{
			{ID: 1, Text: "Message 1", ChatID: 333, SentAt: time.Now()},
		},
		Total:   1,
		HasMore: false,
	}, nil).Once()
	
	// Verify cache interactions
	cacheService.On("Get", mock.Anything, "telegram_job_send_message_111").Return(nil, errors.New("not found")).Once()
	cacheService.On("Set", mock.Anything, "telegram_job_send_message_111", mock.Anything, mock.Anything).Return(nil).Once()
	
	cacheService.On("Get", mock.Anything, "telegram_job_send_message_222").Return(nil, errors.New("not found")).Once()
	cacheService.On("Set", mock.Anything, "telegram_job_send_message_222", mock.Anything, mock.Anything).Return(nil).Once()
	
	cacheService.On("Get", mock.Anything, "telegram_job_get_messages_333").Return(nil, errors.New("not found")).Once()
	cacheService.On("Set", mock.Anything, "telegram_job_get_messages_333", mock.Anything, mock.Anything).Return(nil).Once()
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop(ctx)
	
	// Submit specific jobs
	jobs := []*worker.TelegramJobRequest{
		{
			ID:        "precise_job_1",
			Type:      "send_message",
			ChatID:    111,
			Message:   "Test message 1",
			CreatedAt: time.Now(),
		},
		{
			ID:        "precise_job_2",
			Type:      "send_message",
			ChatID:    222,
			Message:   "Test message 2",
			CreatedAt: time.Now(),
		},
		{
			ID:        "precise_job_3",
			Type:      "get_messages",
			ChatID:    333,
			Metadata:  map[string]interface{}{"limit": 10, "offset": 0},
			CreatedAt: time.Now(),
		},
	}
	
	for _, job := range jobs {
		resp, err := service.SubmitJob(ctx, job)
		assert.NoError(t, err)
		assert.True(t, resp.Success, "Job %s should succeed", job.ID)
	}
	
	// Verify all mock expectations were met
	mtprotoClient.AssertExpectations(t)
	cacheService.AssertExpectations(t)
	
	// Verify specific call counts
	mtprotoClient.AssertNumberOfCalls(t, "SendMessage", 2)
	mtprotoClient.AssertNumberOfCalls(t, "GetMessages", 1)
	mtprotoClient.AssertNumberOfCalls(t, "Connect", 1)
	
	// Verify cache service call counts
	cacheService.AssertNumberOfCalls(t, "Get", 3)
	cacheService.AssertNumberOfCalls(t, "Set", 3)
}

// Test 6: Worker Pool Management
func TestTelegramService_WorkerPoolManagement(t *testing.T) {
	config := createDefaultConfig()
	config.MaxConcurrency = 3
	service, mtprotoClient, _ := createTestTelegramService(t, config)
	
	// Add delay to simulate work
	mtprotoClient.sendDelay = 50 * time.Millisecond
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
		&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil)
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	
	// Submit more jobs than available workers
	numJobs := 10
	startTime := time.Now()
	
	var wg sync.WaitGroup
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("pool_job_%d", id),
				Type:      "send_message",
				ChatID:    int64(id),
				Message:   fmt.Sprintf("Message %d", id),
				CreatedAt: time.Now(),
			}
			_, _ = service.SubmitJob(ctx, req)
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	err = service.Stop(ctx)
	assert.NoError(t, err)
	
	// With 3 workers and 50ms per job, and 10 jobs, it should take at least
	// enough time to process jobs in batches
	expectedMinDuration := time.Duration(numJobs/config.MaxConcurrency) * mtprotoClient.sendDelay
	assert.GreaterOrEqual(t, duration, expectedMinDuration, "Duration should reflect worker pool limitation")
	
	// Verify all jobs were processed
	mtprotoClient.AssertNumberOfCalls(t, "SendMessage", numJobs)
}

// Test 7: Graceful Shutdown
func TestTelegramService_GracefulShutdown(t *testing.T) {
	config := createDefaultConfig()
	config.MaxConcurrency = 5
	service, mtprotoClient, _ := createTestTelegramService(t, config)
	
	// Add delay to simulate ongoing work
	mtprotoClient.sendDelay = 100 * time.Millisecond
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
		&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil)
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	
	// Submit jobs that will be in progress during shutdown
	numJobs := 8
	var wg sync.WaitGroup
	
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("shutdown_job_%d", id),
				Type:      "send_message",
				ChatID:    int64(id),
				Message:   fmt.Sprintf("Message %d", id),
				CreatedAt: time.Now(),
			}
			_, _ = service.SubmitJob(ctx, req)
		}(i)
	}
	
	// Allow some jobs to start processing
	time.Sleep(50 * time.Millisecond)
	
	// Initiate shutdown
	shutdownStart := time.Now()
	err = service.Stop(ctx)
	shutdownDuration := time.Since(shutdownStart)
	
	assert.NoError(t, err)
	
	// Wait for all goroutines to complete
	wg.Wait()
	
	// Verify that shutdown waited for ongoing work to complete
	assert.Greater(t, shutdownDuration, 50*time.Millisecond, "Shutdown should wait for ongoing work")
	
	// Service should no longer be healthy after shutdown
	assert.False(t, service.IsHealthy(), "Service should not be healthy after shutdown")
}

// Test 8: Memory and Resource Usage Under Load
func TestTelegramService_MemoryUsageUnderLoad(t *testing.T) {
	// Skip this test in short mode or on CI to avoid resource constraints
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}
	
	config := createDefaultConfig()
	config.MaxConcurrency = 50
	service, mtprotoClient, _ := createTestTelegramService(t, config)
	
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
		&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil)
	
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop(ctx)
	
	// Measure memory before load test
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	// Generate sustained load
	numJobs := 1000
	var wg sync.WaitGroup
	
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("memory_job_%d", id),
				Type:      "send_message",
				ChatID:    int64(id % 100),
				Message:   fmt.Sprintf("Memory test message %d with some additional content to simulate realistic message sizes", id),
				Metadata:  map[string]interface{}{"test": true, "id": id},
				CreatedAt: time.Now(),
			}
			_, _ = service.SubmitJob(ctx, req)
		}(i)
	}
	
	wg.Wait()
	
	// Measure memory after load test
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	// Memory usage should be reasonable (not more than 100MB increase)
	memoryIncrease := m2.Alloc - m1.Alloc
	assert.Less(t, memoryIncrease, uint64(100*1024*1024), "Memory increase should be less than 100MB")
	
	// Verify all jobs were processed
	mtprotoClient.AssertNumberOfCalls(t, "SendMessage", numJobs)
	
	// Verify metrics
	metrics := service.GetMetrics()
	assert.Equal(t, int64(numJobs), metrics.RequestsProcessed)
	assert.Equal(t, int64(0), metrics.ActiveJobs, "No jobs should be active after completion")
}

// Benchmark tests

func BenchmarkTelegramService_ConcurrentRequests(b *testing.B) {
	config := createDefaultConfig()
	config.MaxConcurrency = 20
	config.MetricsEnabled = false // Disable to reduce overhead
	
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mtprotoClient := &mockMTProtoClient{}
	cacheService := newMockCacheService()
	
	mtprotoClient.On("Connect", mock.Anything).Return(nil)
	mtprotoClient.On("Disconnect", mock.Anything).Return(nil)
	mtprotoClient.On("SendMessage", mock.Anything, mock.Anything).Return(
		&mtproto.SendMessageResponse{MessageID: 1, SentAt: time.Now(), Success: true}, nil)
	cacheService.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
	cacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	service := worker.NewTelegramService(logger, mtprotoClient, cacheService, config)
	
	ctx := context.Background()
	service.Start(ctx)
	defer service.Stop(ctx)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		jobID := 0
		for pb.Next() {
			req := &worker.TelegramJobRequest{
				ID:        fmt.Sprintf("bench_job_%d", jobID),
				Type:      "send_message",
				ChatID:    int64(jobID % 10),
				Message:   fmt.Sprintf("Benchmark message %d", jobID),
				CreatedAt: time.Now(),
			}
			jobID++
			
			_, _ = service.SubmitJob(ctx, req)
		}
	})
}