package cache

import (
	"strings"
	"sync"
)

// Cache interface sekarang jauh lebih sederhana karena semuanya Immutable
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, val interface{})
	InvalidatePattern(pattern string)
	DeleteByPrefix(prefix string)
	Delete(key string)
}

type SimpleCache struct {
	mu    sync.RWMutex
	// Langsung menyimpan data tanpa perlu struct CacheEntry (karena tidak ada ExpiresAt)
	store map[string]interface{} 
}

// GlobalCache default menggunakan Memory Cache sebelum Redis terkoneksi di main.go
var GlobalCache Cache = &SimpleCache{
	store: make(map[string]interface{}),
}

func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.store[key]
	return data, ok
}

func (c *SimpleCache) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = val
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

func (c *SimpleCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.store {
		if strings.Contains(key, pattern) {
			delete(c.store, key)
		}
	}
}