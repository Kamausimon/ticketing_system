package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheMiss     = errors.New("cache miss")
	ErrInvalidValue  = errors.New("invalid value")
	ErrCacheNotReady = errors.New("cache not ready")
)

// SessionManager manages sessions with Redis primary and in-memory fallback
type SessionManager struct {
	redis      *redis.Client
	redisReady bool
	fallback   *MemoryCache
	mu         sync.RWMutex
	ctx        context.Context
}

// MemoryCache provides in-memory fallback storage
type MemoryCache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// NewSessionManager creates a new session manager with Redis and fallback
func NewSessionManager(redisAddr, password string, db int) *SessionManager {
	ctx := context.Background()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	sm := &SessionManager{
		redis:      rdb,
		redisReady: false,
		fallback:   NewMemoryCache(),
		ctx:        ctx,
	}

	// Test Redis connection
	go sm.healthCheck()

	return sm
}

// healthCheck continuously monitors Redis health
func (sm *SessionManager) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		err := sm.redis.Ping(sm.ctx).Err()
		sm.mu.Lock()
		sm.redisReady = (err == nil)
		sm.mu.Unlock()

		if err != nil {
			fmt.Printf("⚠️  Redis health check failed: %v (using fallback cache)\n", err)
		}

		<-ticker.C
	}
}

// Set stores a value with expiration
func (sm *SessionManager) Set(key string, value interface{}, expiration time.Duration) error {
	sm.mu.RLock()
	redisReady := sm.redisReady
	sm.mu.RUnlock()

	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Try Redis first if available
	if redisReady {
		err := sm.redis.Set(sm.ctx, key, data, expiration).Err()
		if err == nil {
			return nil
		}
		fmt.Printf("⚠️  Redis set failed: %v (using fallback)\n", err)
	}

	// Fallback to memory cache
	return sm.fallback.Set(key, value, expiration)
}

// Get retrieves a value
func (sm *SessionManager) Get(key string, dest interface{}) error {
	sm.mu.RLock()
	redisReady := sm.redisReady
	sm.mu.RUnlock()

	// Try Redis first if available
	if redisReady {
		data, err := sm.redis.Get(sm.ctx, key).Bytes()
		if err == nil {
			return json.Unmarshal(data, dest)
		}
		if err != redis.Nil {
			fmt.Printf("⚠️  Redis get failed: %v (trying fallback)\n", err)
		}
	}

	// Fallback to memory cache
	return sm.fallback.Get(key, dest)
}

// Delete removes a value
func (sm *SessionManager) Delete(key string) error {
	sm.mu.RLock()
	redisReady := sm.redisReady
	sm.mu.RUnlock()

	var redisErr error
	if redisReady {
		redisErr = sm.redis.Del(sm.ctx, key).Err()
	}

	fallbackErr := sm.fallback.Delete(key)

	// Return error only if both failed
	if redisErr != nil && fallbackErr != nil {
		return fmt.Errorf("both redis and fallback delete failed: redis=%v, fallback=%v", redisErr, fallbackErr)
	}

	return nil
}

// Exists checks if a key exists
func (sm *SessionManager) Exists(key string) bool {
	sm.mu.RLock()
	redisReady := sm.redisReady
	sm.mu.RUnlock()

	if redisReady {
		count, err := sm.redis.Exists(sm.ctx, key).Result()
		if err == nil {
			return count > 0
		}
	}

	return sm.fallback.Exists(key)
}

// Expire sets expiration on a key
func (sm *SessionManager) Expire(key string, expiration time.Duration) error {
	sm.mu.RLock()
	redisReady := sm.redisReady
	sm.mu.RUnlock()

	if redisReady {
		err := sm.redis.Expire(sm.ctx, key, expiration).Err()
		if err == nil {
			return nil
		}
	}

	return sm.fallback.Expire(key, expiration)
}

// IsRedisHealthy returns true if Redis is available
func (sm *SessionManager) IsRedisHealthy() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.redisReady
}

// Close closes the Redis connection
func (sm *SessionManager) Close() error {
	return sm.redis.Close()
}

// MemoryCache implementation

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		data: make(map[string]cacheEntry),
	}

	// Start cleanup goroutine
	go mc.cleanup()

	return mc
}

// Set stores a value in memory
func (mc *MemoryCache) Set(key string, value interface{}, expiration time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(expiration),
	}

	return nil
}

// Get retrieves a value from memory
func (mc *MemoryCache) Get(key string, dest interface{}) error {
	mc.mu.RLock()
	entry, exists := mc.data[key]
	mc.mu.RUnlock()

	if !exists {
		return ErrCacheMiss
	}

	if time.Now().After(entry.expiration) {
		mc.Delete(key)
		return ErrCacheMiss
	}

	// Use JSON marshal/unmarshal to copy data
	data, err := json.Marshal(entry.value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a value from memory
func (mc *MemoryCache) Delete(key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.data, key)
	return nil
}

// Exists checks if a key exists in memory
func (mc *MemoryCache) Exists(key string) bool {
	mc.mu.RLock()
	entry, exists := mc.data[key]
	mc.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(entry.expiration) {
		mc.Delete(key)
		return false
	}

	return true
}

// Expire sets expiration on a key
func (mc *MemoryCache) Expire(key string, expiration time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	entry, exists := mc.data[key]
	if !exists {
		return ErrCacheMiss
	}

	entry.expiration = time.Now().Add(expiration)
	mc.data[key] = entry

	return nil
}

// cleanup removes expired entries periodically
func (mc *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()
		now := time.Now()
		for key, entry := range mc.data {
			if now.After(entry.expiration) {
				delete(mc.data, key)
			}
		}
		mc.mu.Unlock()
	}
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"entries": len(mc.data),
	}
}
