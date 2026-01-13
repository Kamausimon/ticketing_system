package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// EventsCache provides caching for event-related data
type EventsCache struct {
	sm *SessionManager
}

// NewEventsCache creates a new events cache
func NewEventsCache(sm *SessionManager) *EventsCache {
	return &EventsCache{sm: sm}
}

// GetEventsList retrieves cached events list
func (ec *EventsCache) GetEventsList(key string) ([]byte, error) {
	ctx := context.Background()
	var data []byte
	err := ec.sm.Get(ctx, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetEventsList caches events list
func (ec *EventsCache) SetEventsList(key string, data []byte, ttl time.Duration) error {
	ctx := context.Background()
	return ec.sm.Set(ctx, key, data, ttl)
}

// GetSearchResults retrieves cached search results
func (ec *EventsCache) GetSearchResults(query string) ([]byte, error) {
	key := fmt.Sprintf("search:%s", query)
	return ec.GetEventsList(key)
}

// SetSearchResults caches search results
func (ec *EventsCache) SetSearchResults(query string, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("search:%s", query)
	return ec.SetEventsList(key, data, ttl)
}

// InvalidateEventsList clears the events list cache
func (ec *EventsCache) InvalidateEventsList() error {
	ctx := context.Background()
	// Delete common cache keys
	keys := []string{
		"events:list",
		"events:list:page:*",
	}
	
	for _, key := range keys {
		ec.sm.Delete(ctx, key)
	}
	return nil
}

// InvalidateEvent clears cache for a specific event
func (ec *EventsCache) InvalidateEvent(eventID int) error {
	ctx := context.Background()
	key := fmt.Sprintf("event:%d", eventID)
	return ec.sm.Delete(ctx, key)
}

// GetMetrics returns cache statistics
func (ec *EventsCache) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"status": "active",
		"backend": func() string {
			ec.sm.mu.RLock()
			defer ec.sm.mu.RUnlock()
			if ec.sm.redisReady {
				return "redis"
			}
			return "memory"
		}(),
	}
}

// WarmUp pre-loads frequently accessed data into cache
func (ec *EventsCache) WarmUp(dataLoader func() ([]byte, error)) error {
	data, err := dataLoader()
	if err != nil {
		return err
	}
	return ec.SetEventsList("events:list", data, 5*time.Minute)
}

// BatchInvalidate clears multiple cache keys at once
func (ec *EventsCache) BatchInvalidate(keys []string) error {
	ctx := context.Background()
	for _, key := range keys {
		ec.sm.Delete(ctx, key)
	}
	return nil
}

// GetOrSet retrieves from cache or computes and stores the value
func (ec *EventsCache) GetOrSet(key string, ttl time.Duration, compute func() (interface{}, error)) ([]byte, error) {
	ctx := context.Background()
	
	// Try to get from cache
	var data []byte
	err := ec.sm.Get(ctx, key, &data)
	if err == nil {
		return data, nil
	}
	
	// Cache miss - compute value
	result, err := compute()
	if err != nil {
		return nil, err
	}
	
	// Serialize result
	data, err = json.Marshal(result)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	ec.sm.Set(ctx, key, data, ttl)
	
	return data, nil
}
