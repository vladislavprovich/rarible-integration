package cache

import (
	"context"
	"time"
)

// CacheEntry represents a cache entry with metadata
type CacheEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	Hits      int64       `json:"hits"`
}

// Service defines the interface for cache operations
type Service interface {
	// Set stores a value with the given key and TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Get retrieves a value by key
	Get(ctx context.Context, key string) (interface{}, error)
	
	// Delete removes a value by key
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
	
	// Clear removes all entries from cache
	Clear(ctx context.Context) error
	
	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*Stats, error)
	
	// GetMultiple retrieves multiple values by keys
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	
	// SetMultiple stores multiple key-value pairs
	SetMultiple(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error
}

// Stats represents cache statistics
type Stats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Keys        int64     `json:"keys"`
	Memory      int64     `json:"memory_bytes"`
	Connections int       `json:"connections"`
	Uptime      time.Time `json:"uptime"`
}