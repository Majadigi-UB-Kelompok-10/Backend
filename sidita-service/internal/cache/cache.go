package cache

import (
	"encoding/json"
	"strings"
	"sync"
)

type Cache interface {
	Get(key string) ([]byte, bool)
    Set(key string, val interface{})
    InvalidatePattern(pattern string)
    DeleteByPrefix(prefix string)
    Delete(key string)
}

type SimpleCache struct {
    mu    sync.RWMutex
    store map[string][]byte 
}

var GlobalCache Cache = &SimpleCache{
	store: make(map[string][]byte),
}

func (c *SimpleCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.store[key]
	return data, ok
}

func (c *SimpleCache) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch v := val.(type) {
	case []byte:
		c.store[key] = v
	default:
		if jsonData, err := json.Marshal(val); err == nil {
			c.store[key] = jsonData
		}
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

func (c *SimpleCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.store {
		if strings.Contains(key, pattern) {
			delete(c.store, key)
		}
	}
}