package cache

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, val interface{})
	SetWithTTL(key string, val interface{}, ttl time.Duration)
	InvalidatePattern(pattern string)
	DeleteByPrefix(prefix string)
	Delete(key string)
}

type cacheItem struct {
	data      []byte
	expiresAt time.Time
}

type SimpleCache struct {
	mu    sync.RWMutex
	store map[string]cacheItem
}

var GlobalCache Cache = &SimpleCache{
	store: make(map[string]cacheItem),
}

func (c *SimpleCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, ok := c.store[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return nil, false
	}

	return item.data, true
}

func (c *SimpleCache) Set(key string, val interface{}) {
	c.SetWithTTL(key, val, 0)
}

func (c *SimpleCache) SetWithTTL(key string, val interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var data []byte
	switch v := val.(type) {
	case []byte:
		data = v
	default:
		if jsonData, err := json.Marshal(val); err == nil {
			data = jsonData
		} else {
			return
		}
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.store[key] = cacheItem{
		data:      data,
		expiresAt: expiresAt,
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