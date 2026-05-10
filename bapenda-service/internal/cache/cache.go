package cache

import (
	"container/list"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
)

// =============================================================================
// INTERFACE
// =============================================================================

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, val interface{}, ttl time.Duration)
	Has(key string) bool
	Delete(key string)
	DeleteByPrefix(prefix string)
	InvalidatePattern(pattern string)
	Stats() Stats
}

// Stats — metrics buat observability (bisa di-expose di /metrics atau /health)
type Stats struct {
	Hits      uint64 `json:"hits"`
	Misses    uint64 `json:"misses"`
	Evictions uint64 `json:"evictions"`
	Size      int    `json:"size"`
}

// HitRate menghitung rasio hit dari total request
func (s Stats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// =============================================================================
// SIMPLE CACHE — in-memory dengan LRU eviction + TTL background cleanup
// =============================================================================

const (
	defaultMaxItems        = 10_000
	defaultCleanupInterval = 5 * time.Minute
)

type cacheItem struct {
	key       string
	data      []byte
	expiresAt time.Time
	elem      *list.Element // pointer ke posisi di LRU list
}

type SimpleCache struct {
	mu       sync.RWMutex
	store    map[string]*cacheItem
	lruList  *list.List // front = most recent, back = LRU candidate
	maxItems int

	// Atomic counters — gak perlu lock buat baca/tulis
	hits      atomic.Uint64
	misses    atomic.Uint64
	evictions atomic.Uint64

	// Background cleanup
	stopCleanup chan struct{}
	closeOnce   sync.Once
}

// NewSimpleCache bikin in-memory cache.
//   maxItems         : 0 = pakai default (10k)
//   cleanupInterval  : 0 = pakai default (5 menit)
func NewSimpleCache(maxItems int, cleanupInterval time.Duration) *SimpleCache {
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}
	if cleanupInterval <= 0 {
		cleanupInterval = defaultCleanupInterval
	}
	sc := &SimpleCache{
		store:       make(map[string]*cacheItem),
		lruList:     list.New(),
		maxItems:    maxItems,
		stopCleanup: make(chan struct{}),
	}
	go sc.cleanupLoop(cleanupInterval)
	return sc
}

// GlobalCache — backward compat dengan kode lama
var GlobalCache Cache = NewSimpleCache(defaultMaxItems, defaultCleanupInterval)

// -----------------------------------------------------------------------------
// Public methods
// -----------------------------------------------------------------------------

func (c *SimpleCache) Get(key string) ([]byte, bool) {
	c.mu.Lock() // Lock (bukan RLock) karena update LRU position
	defer c.mu.Unlock()

	item, ok := c.store[key]
	if !ok {
		c.misses.Add(1)
		return nil, false
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.removeLocked(key)
		c.misses.Add(1)
		return nil, false
	}

	c.lruList.MoveToFront(item.elem) // refresh posisi LRU
	c.hits.Add(1)
	return item.data, true
}

func (c *SimpleCache) Set(key string, val interface{}, ttl time.Duration) {
	data, ok := serialize(key, val)
	if !ok {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// Update existing
	if existing, ok := c.store[key]; ok {
		existing.data = data
		existing.expiresAt = expiresAt
		c.lruList.MoveToFront(existing.elem)
		return
	}

	// Insert new
	item := &cacheItem{key: key, data: data, expiresAt: expiresAt}
	item.elem = c.lruList.PushFront(item)
	c.store[key] = item

	// LRU eviction kalau melebihi max
	for c.lruList.Len() > c.maxItems {
		oldest := c.lruList.Back()
		if oldest == nil {
			break
		}
		evicted := oldest.Value.(*cacheItem)
		c.removeLocked(evicted.key)
		c.evictions.Add(1)
	}
}

func (c *SimpleCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.store[key]
	if !ok {
		return false
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false
	}
	return true
}

func (c *SimpleCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removeLocked(key)
}

func (c *SimpleCache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.store {
		if strings.HasPrefix(key, prefix) {
			c.removeLocked(key)
		}
	}
}

func (c *SimpleCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.store {
		if strings.Contains(key, pattern) {
			c.removeLocked(key)
		}
	}
}

func (c *SimpleCache) Stats() Stats {
	c.mu.RLock()
	size := len(c.store)
	c.mu.RUnlock()
	return Stats{
		Hits:      c.hits.Load(),
		Misses:    c.misses.Load(),
		Evictions: c.evictions.Load(),
		Size:      size,
	}
}

// Close hentiin background goroutine — panggil saat shutdown (opsional)
func (c *SimpleCache) Close() {
	c.closeOnce.Do(func() {
		close(c.stopCleanup)
	})
}

// -----------------------------------------------------------------------------
// Private helpers
// -----------------------------------------------------------------------------

// removeLocked — WAJIB dipanggil sambil pegang c.mu.Lock()
func (c *SimpleCache) removeLocked(key string) {
	if item, ok := c.store[key]; ok {
		c.lruList.Remove(item.elem)
		delete(c.store, key)
	}
}

// cleanupLoop background goroutine — hapus item kadaluarsa biar gak memory leak
func (c *SimpleCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *SimpleCache) cleanupExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	var removed int
	for key, item := range c.store {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			c.removeLocked(key)
			removed++
		}
	}
	if removed > 0 {
		slog.Debug("cache.cleanup",
			slog.Int("removed", removed),
			slog.Int("remaining", len(c.store)),
		)
	}
}

func serialize(key string, val interface{}) ([]byte, bool) {
	switch v := val.(type) {
	case []byte:
		return v, true
	case string:
		return []byte(v), true
	default:
		b, err := sonic.Marshal(val)
		if err != nil {
			slog.Warn("cache.serialize_failed",
				slog.String("key", key),
				slog.String("err", err.Error()),
			)
			return nil, false
		}
		return b, true
	}
}