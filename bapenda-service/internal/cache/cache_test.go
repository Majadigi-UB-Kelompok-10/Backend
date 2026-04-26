package cache

import (
	"testing"
)

// TestSimpleCache tests the in-memory cache implementation
func TestSimpleCache_Set_Get(t *testing.T) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	tests := []struct {
		name  string
		key   string
		value map[string]interface{}
	}{
		{"Simple value", "key1", map[string]interface{}{"test": "value1"}},
		{"Another value", "key2", map[string]interface{}{"id": 12345}},
		{"Complex value", "key3", map[string]interface{}{"nested": map[string]string{"a": "b"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Set(tt.key, tt.value)
			got, found := c.Get(tt.key)

			if !found {
				t.Errorf("Get() found = false, want true")
				return
			}

			if len(got) == 0 {
				t.Errorf("Get() returned empty bytes")
			}
		})
	}
}

func TestSimpleCache_Get_NotFound(t *testing.T) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	_, found := c.Get("nonexistent")
	if found {
		t.Errorf("Get() found = true, want false for nonexistent key")
	}
}

func TestSimpleCache_Set_Overwrite(t *testing.T) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	c.Set("key", map[string]interface{}{"value": "1"})
	c.Set("key", map[string]interface{}{"value": "2"})

	got, found := c.Get("key")
	if !found {
		t.Errorf("Get() found = false after overwrite")
		return
	}

	if len(got) == 0 {
		t.Errorf("Get() returned empty after overwrite")
	}
}

func TestSimpleCache_Delete(t *testing.T) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	c.Set("key", map[string]interface{}{"value": "data"})
	c.Delete("key")

	_, found := c.Get("key")
	if found {
		t.Errorf("Get() found = true after Delete, want false")
	}
}

func TestSimpleCache_DeleteByPrefix(t *testing.T) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	c.Set("prefix:key1", map[string]interface{}{"v": "1"})
	c.Set("prefix:key2", map[string]interface{}{"v": "2"})
	c.Set("other:key3", map[string]interface{}{"v": "3"})

	c.DeleteByPrefix("prefix:")

	// Prefix keys should be deleted
	_, found1 := c.Get("prefix:key1")
	_, found2 := c.Get("prefix:key2")
	if found1 || found2 {
		t.Errorf("Keys with prefix not deleted")
	}

	// Other key should still exist
	_, found3 := c.Get("other:key3")
	if !found3 {
		t.Errorf("Unrelated key was deleted")
	}
}

func TestSimpleCache_InvalidatePattern(t *testing.T) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	c.Set("cache:data:1", map[string]interface{}{"v": "1"})
	c.Set("cache:info:1", map[string]interface{}{"v": "2"})
	c.Set("other:1", map[string]interface{}{"v": "3"})

	c.InvalidatePattern("cache:data")

	// Pattern matching keys should be deleted
	_, found1 := c.Get("cache:data:1")
	if found1 {
		t.Errorf("Key with pattern not invalidated")
	}

	// Other pattern keys should exist
	_, found2 := c.Get("cache:info:1")
	_, found3 := c.Get("other:1")
	if !found2 || !found3 {
		t.Errorf("Unrelated keys were invalidated")
	}
}

// Benchmark tests
func BenchmarkSimpleCache_Set(b *testing.B) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}
	for i := 0; i < b.N; i++ {
		c.Set("key", map[string]interface{}{"value": "data"})
	}
}

func BenchmarkSimpleCache_Get(b *testing.B) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}
	c.Set("key", map[string]interface{}{"value": "data"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}

func BenchmarkSimpleCache_Set_Get(b *testing.B) {
	c := &SimpleCache{
		store: make(map[string][]byte),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", map[string]interface{}{"value": "data"})
		c.Get("key")
	}
}
