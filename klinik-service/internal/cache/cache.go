package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic" 
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, val interface{}, ttl time.Duration)
	InvalidatePattern(pattern string)
	DeleteByPrefix(prefix string)
	Delete(key string)
}

type SimpleCache struct {
	mu    sync.RWMutex
	store map[string]cacheItem
}

type cacheItem struct {
	Data      []byte
	ExpiresAt int64
}

var GlobalCache Cache = &SimpleCache{
	store: make(map[string]cacheItem),
}

func (c *SimpleCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.store[key]
	if !ok {
		return nil, false
	}
	if item.ExpiresAt > 0 && time.Now().UnixNano() > item.ExpiresAt {
		return nil, false
	}
	return item.Data, true
}

func (c *SimpleCache) Set(key string, val interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var data []byte
	if v, ok := val.([]byte); ok {
		data = v
	} else {
		if jsonData, err := sonic.Marshal(val); err == nil {
			data = jsonData
		} else {
			return
		}
	}

	exp := int64(0)
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.store[key] = cacheItem{
		Data:      data,
		ExpiresAt: exp,
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