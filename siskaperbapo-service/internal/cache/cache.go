package cache

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, val interface{}, ttl time.Duration)
	GetImmutable(key string) (interface{}, bool)
	SetImmutable(key string, val interface{})
	InvalidatePattern(pattern string)
	DeleteByPrefix(prefix string)
	Delete(key string)
}

type CacheEntry struct {
	Data      []byte 
	ExpiresAt time.Time
}

type SimpleCache struct {
	mu    sync.RWMutex
	store map[string]CacheEntry
}

var GlobalCache Cache = &SimpleCache{
	store: make(map[string]CacheEntry),
}

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if simpleCache, ok := GlobalCache.(*SimpleCache); ok {
				simpleCache.cleanup()
			}
		}
	}()
}

func (c *SimpleCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, entry := range c.store {
		if now.After(entry.ExpiresAt) {
			delete(c.store, key)
		}
	}
}

func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Data, true 
}

func (c *SimpleCache) Set(key string, val interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	dataBytes, _ := json.Marshal(val)
	c.store[key] = CacheEntry{
		Data:      dataBytes,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (c *SimpleCache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.store {
		if strings.HasPrefix(key, prefix) {
			delete(c.store, key)
		}
	}
}

func (c *SimpleCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

func (c *SimpleCache) SetImmutable(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	dataBytes, _ := json.Marshal(val)
	c.store[key] = CacheEntry{
		Data:      dataBytes,
		ExpiresAt: time.Unix(1<<63-1, 0),
	}
}

func (c *SimpleCache) GetImmutable(key string) (interface{}, bool) {
	return c.Get(key)
}

func (c *SimpleCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.store {
		if strings.Contains(key, pattern) {
			delete(c.store, key)
		}
	}
}