package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vladislavprovich/rarible-integration/pkg/cache"
	"github.com/vladislavprovich/rarible-integration/pkg/client/mtproto"
)

// TelegramWorkerConfig contains configuration for the TelegramService worker
type TelegramWorkerConfig struct {
	MaxConcurrency    int           `json:"max_concurrency"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	CacheTimeout      time.Duration `json:"cache_timeout"`
	RetryAttempts     int           `json:"retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
	CircuitBreakerMax int           `json:"circuit_breaker_max"`
	MetricsEnabled    bool          `json:"metrics_enabled"`
}

// TelegramJobRequest represents a job request for the worker
type TelegramJobRequest struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	ChatID    int64                  `json:"chat_id"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Priority  int                    `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
	ResultChan chan *TelegramJobResponse `json:"-"` // Channel for returning result
}

// TelegramJobResponse represents a job response from the worker
type TelegramJobResponse struct {
	ID          string      `json:"id"`
	Success     bool        `json:"success"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	ProcessedAt time.Time   `json:"processed_at"`
	Duration    time.Duration `json:"duration"`
}

// TelegramWorkerMetrics contains worker performance metrics
type TelegramWorkerMetrics struct {
	RequestsProcessed int64         `json:"requests_processed"`
	RequestsSucceeded int64         `json:"requests_succeeded"`
	RequestsFailed    int64         `json:"requests_failed"`
	AverageLatency    time.Duration `json:"average_latency"`
	ActiveJobs        int64         `json:"active_jobs"`
	QueueSize         int64         `json:"queue_size"`
	CircuitBreakerOpen bool         `json:"circuit_breaker_open"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// TelegramService worker handles Telegram operations with concurrency control
type TelegramService struct {
	logger       *slog.Logger
	mtprotoClient mtproto.Client
	cacheService cache.Service
	config       TelegramWorkerConfig
	
	// Concurrency control
	jobQueue    chan *TelegramJobRequest
	workers     []*worker
	wg          sync.WaitGroup
	mu          sync.RWMutex
	
	// Circuit breaker
	failureCount    int64
	circuitOpen     bool
	circuitOpenTime time.Time
	
	// Metrics
	metrics atomic.Value // holds *TelegramWorkerMetrics
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	
	// Data consistency
	dataStore map[string]interface{}
	dataMutex sync.RWMutex
}

// worker represents an individual worker goroutine
type worker struct {
	id      int
	service *TelegramService
	active  bool
	mu      sync.Mutex
}

// TelegramWorker defines the interface for the telegram worker
type TelegramWorker interface {
	// Start initializes and starts the worker pool
	Start(ctx context.Context) error
	
	// Stop gracefully stops the worker pool
	Stop(ctx context.Context) error
	
	// SubmitJob submits a job to the worker pool
	SubmitJob(ctx context.Context, req *TelegramJobRequest) (*TelegramJobResponse, error)
	
	// GetMetrics returns current worker metrics
	GetMetrics() *TelegramWorkerMetrics
	
	// IsHealthy returns the health status of the worker
	IsHealthy() bool
	
	// SetData stores data in the worker's data store (for testing data consistency)
	SetData(key string, value interface{})
	
	// GetData retrieves data from the worker's data store
	GetData(key string) (interface{}, bool)
}

// NewTelegramService creates a new TelegramService worker
func NewTelegramService(
	logger *slog.Logger,
	mtprotoClient mtproto.Client,
	cacheService cache.Service,
	config TelegramWorkerConfig,
) *TelegramService {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 10
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.CacheTimeout == 0 {
		config.CacheTimeout = 5 * time.Minute
	}
	if config.RetryAttempts <= 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	if config.CircuitBreakerMax <= 0 {
		config.CircuitBreakerMax = 10
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	ts := &TelegramService{
		logger:        logger,
		mtprotoClient: mtprotoClient,
		cacheService:  cacheService,
		config:        config,
		jobQueue:      make(chan *TelegramJobRequest, config.MaxConcurrency*2),
		workers:       make([]*worker, config.MaxConcurrency),
		ctx:           ctx,
		cancel:        cancel,
		dataStore:     make(map[string]interface{}),
	}
	
	// Initialize metrics
	ts.metrics.Store(&TelegramWorkerMetrics{
		LastUpdated: time.Now(),
	})
	
	return ts
}

// Start implements TelegramWorker interface
func (ts *TelegramService) Start(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	ts.logger.Info("Starting TelegramService worker", "max_concurrency", ts.config.MaxConcurrency)
	
	// Connect to MTProto client
	if err := ts.mtprotoClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect MTProto client: %w", err)
	}
	
	// Start worker goroutines
	for i := 0; i < ts.config.MaxConcurrency; i++ {
		w := &worker{
			id:      i,
			service: ts,
			active:  false,
		}
		ts.workers[i] = w
		
		ts.wg.Add(1)
		go ts.workerLoop(w)
	}
	
	// Start metrics updater if enabled
	if ts.config.MetricsEnabled {
		ts.wg.Add(1)
		go ts.metricsLoop()
	}
	
	ts.logger.Info("TelegramService worker started successfully")
	return nil
}

// Stop implements TelegramWorker interface
func (ts *TelegramService) Stop(ctx context.Context) error {
	ts.logger.Info("Stopping TelegramService worker")
	
	// Cancel context to signal shutdown
	ts.cancel()
	
	// Close job queue
	close(ts.jobQueue)
	
	// Wait for all workers to finish
	ts.wg.Wait()
	
	// Disconnect MTProto client
	if err := ts.mtprotoClient.Disconnect(ctx); err != nil {
		ts.logger.Error("Failed to disconnect MTProto client", "error", err)
	}
	
	ts.logger.Info("TelegramService worker stopped")
	return nil
}

// SubmitJob implements TelegramWorker interface
func (ts *TelegramService) SubmitJob(ctx context.Context, req *TelegramJobRequest) (*TelegramJobResponse, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("job ID is required")
	}
	
	// Check circuit breaker
	if ts.isCircuitOpen() {
		return &TelegramJobResponse{
			ID:          req.ID,
			Success:     false,
			Error:       "circuit breaker is open",
			ProcessedAt: time.Now(),
		}, nil
	}
	
	// Create result channel for this job
	req.ResultChan = make(chan *TelegramJobResponse, 1)
	defer close(req.ResultChan)
	
	// Add to queue with timeout
	select {
	case ts.jobQueue <- req:
		// Job queued successfully
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(ts.config.RequestTimeout):
		return nil, fmt.Errorf("timeout queuing job")
	}
	
	// Wait for result with timeout
	select {
	case result := <-req.ResultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(ts.config.RequestTimeout):
		return nil, fmt.Errorf("timeout waiting for result")
	}
}

// GetMetrics implements TelegramWorker interface
func (ts *TelegramService) GetMetrics() *TelegramWorkerMetrics {
	if metrics, ok := ts.metrics.Load().(*TelegramWorkerMetrics); ok {
		return metrics
	}
	return &TelegramWorkerMetrics{}
}

// IsHealthy implements TelegramWorker interface
func (ts *TelegramService) IsHealthy() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	// Check if circuit breaker is open
	if ts.isCircuitOpen() {
		return false
	}
	
	// Check MTProto connection
	if !ts.mtprotoClient.IsConnected() {
		return false
	}
	
	// Check if workers are running
	activeWorkers := 0
	for _, w := range ts.workers {
		w.mu.Lock()
		if w.active {
			activeWorkers++
		}
		w.mu.Unlock()
	}
	
	return activeWorkers > 0
}

// SetData implements TelegramWorker interface
func (ts *TelegramService) SetData(key string, value interface{}) {
	ts.dataMutex.Lock()
	defer ts.dataMutex.Unlock()
	ts.dataStore[key] = value
}

// GetData implements TelegramWorker interface
func (ts *TelegramService) GetData(key string) (interface{}, bool) {
	ts.dataMutex.RLock()
	defer ts.dataMutex.RUnlock()
	value, exists := ts.dataStore[key]
	return value, exists
}

// workerLoop is the main loop for each worker goroutine
func (ts *TelegramService) workerLoop(w *worker) {
	defer ts.wg.Done()
	
	ts.logger.Debug("Worker started", "worker_id", w.id)
	
	for {
		select {
		case job, ok := <-ts.jobQueue:
			if !ok {
				ts.logger.Debug("Worker stopping - job queue closed", "worker_id", w.id)
				return
			}
			
			w.mu.Lock()
			w.active = true
			w.mu.Unlock()
			
			ts.processJob(w, job)
			
			w.mu.Lock()
			w.active = false
			w.mu.Unlock()
			
		case <-ts.ctx.Done():
			ts.logger.Debug("Worker stopping - context cancelled", "worker_id", w.id)
			return
		}
	}
}

// processJob processes a single job
func (ts *TelegramService) processJob(w *worker, job *TelegramJobRequest) {
	startTime := time.Now()
	
	ts.logger.Debug("Processing job", "worker_id", w.id, "job_id", job.ID, "type", job.Type)
	
	result := &TelegramJobResponse{
		ID:          job.ID,
		ProcessedAt: time.Now(),
	}
	
	// Update metrics - increment active jobs
	ts.updateMetrics(func(m *TelegramWorkerMetrics) {
		atomic.AddInt64(&m.ActiveJobs, 1)
		atomic.AddInt64(&m.RequestsProcessed, 1)
	})
	
	defer func() {
		result.Duration = time.Since(startTime)
		
		// Update metrics - decrement active jobs
		ts.updateMetrics(func(m *TelegramWorkerMetrics) {
			atomic.AddInt64(&m.ActiveJobs, -1)
			if result.Success {
				atomic.AddInt64(&m.RequestsSucceeded, 1)
			} else {
				atomic.AddInt64(&m.RequestsFailed, 1)
				atomic.AddInt64(&ts.failureCount, 1)
			}
		})
		
		// Send result to job's result channel
		select {
		case job.ResultChan <- result:
		case <-ts.ctx.Done():
		case <-time.After(time.Second):
			ts.logger.Error("Failed to send result - channel timeout", "job_id", job.ID)
		}
	}()
	
	// Try to get from cache first
	cacheKey := fmt.Sprintf("telegram_job_%s_%d", job.Type, job.ChatID)
	if cachedResult, err := ts.cacheService.Get(context.Background(), cacheKey); err == nil {
		ts.logger.Debug("Cache hit", "job_id", job.ID, "cache_key", cacheKey)
		result.Success = true
		result.Result = cachedResult
		return
	}
	
	// Process based on job type
	var err error
	switch job.Type {
	case "send_message":
		result.Result, err = ts.processSendMessage(job)
	case "get_messages":
		result.Result, err = ts.processGetMessages(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		ts.logger.Error("Job processing failed", "job_id", job.ID, "error", err)
		
		// Check circuit breaker
		if atomic.LoadInt64(&ts.failureCount) >= int64(ts.config.CircuitBreakerMax) {
			ts.openCircuitBreaker()
		}
		return
	}
	
	result.Success = true
	
	// Cache the result
	_ = ts.cacheService.Set(context.Background(), cacheKey, result.Result, ts.config.CacheTimeout)
	
	// Reset failure count on success
	atomic.StoreInt64(&ts.failureCount, 0)
	ts.closeCircuitBreaker()
}

// processSendMessage handles send message jobs
func (ts *TelegramService) processSendMessage(job *TelegramJobRequest) (interface{}, error) {
	req := &mtproto.SendMessageRequest{
		ChatID:   job.ChatID,
		Text:     job.Message,
		Metadata: job.Metadata,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), ts.config.RequestTimeout)
	defer cancel()
	
	return ts.mtprotoClient.SendMessage(ctx, req)
}

// processGetMessages handles get messages jobs
func (ts *TelegramService) processGetMessages(job *TelegramJobRequest) (interface{}, error) {
	limit := 10
	offset := 0
	
	if limitVal, ok := job.Metadata["limit"].(int); ok {
		limit = limitVal
	}
	if offsetVal, ok := job.Metadata["offset"].(int); ok {
		offset = offsetVal
	}
	
	req := &mtproto.GetMessagesRequest{
		ChatID: job.ChatID,
		Limit:  limit,
		Offset: offset,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), ts.config.RequestTimeout)
	defer cancel()
	
	return ts.mtprotoClient.GetMessages(ctx, req)
}

// metricsLoop updates metrics periodically
func (ts *TelegramService) metricsLoop() {
	defer ts.wg.Done()
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ts.updateMetrics(func(m *TelegramWorkerMetrics) {
				m.LastUpdated = time.Now()
				m.QueueSize = int64(len(ts.jobQueue))
				m.CircuitBreakerOpen = ts.isCircuitOpen()
			})
		case <-ts.ctx.Done():
			return
		}
	}
}

// updateMetrics safely updates metrics
func (ts *TelegramService) updateMetrics(updateFn func(*TelegramWorkerMetrics)) {
	current := ts.GetMetrics()
	updated := *current
	updateFn(&updated)
	ts.metrics.Store(&updated)
}

// isCircuitOpen checks if circuit breaker is open
func (ts *TelegramService) isCircuitOpen() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	if !ts.circuitOpen {
		return false
	}
	
	// Check if circuit should be reset (after timeout)
	if time.Since(ts.circuitOpenTime) > time.Minute {
		return false
	}
	
	return true
}

// openCircuitBreaker opens the circuit breaker
func (ts *TelegramService) openCircuitBreaker() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if !ts.circuitOpen {
		ts.circuitOpen = true
		ts.circuitOpenTime = time.Now()
		ts.logger.Warn("Circuit breaker opened due to failures")
	}
}

// closeCircuitBreaker closes the circuit breaker
func (ts *TelegramService) closeCircuitBreaker() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if ts.circuitOpen {
		ts.circuitOpen = false
		ts.logger.Info("Circuit breaker closed")
	}
}