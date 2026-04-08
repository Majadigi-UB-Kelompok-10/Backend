package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing redis URL: %w", err)
	}

	opt.PoolSize = 100       
	opt.MinIdleConns = 10     
	opt.ConnMaxLifetime = 5 * time.Minute 

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("gagal koneksi ke Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (r *RedisCache) contextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func (r *RedisCache) Get(key string) ([]byte, bool) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	return val, true
}

// Set sekarang menerima TTL!
func (r *RedisCache) Set(key string, val interface{}, ttl time.Duration) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	var data []byte
	var err error

	if v, ok := val.([]byte); ok {
		data = v
	} else {
		data, err = json.Marshal(val)
		if err != nil {
			fmt.Printf("[REDIS ERROR] Gagal marshal val untuk key %s: %v\n", key, err)
			return
		}
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		fmt.Printf("[REDIS ERROR] Gagal simpan key %s: %v\n", key, err)
	}
}

func (r *RedisCache) InvalidatePattern(pattern string) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "*"+pattern+"*", 100).Result()
		if err != nil {
			fmt.Printf("[REDIS ERROR] Scan InvalidatePattern gagal: %v\n", err)
			return
		}

		if len(keys) > 0 {
			r.client.Unlink(ctx, keys...)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

func (r *RedisCache) DeleteByPrefix(prefix string) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return
		}

		if len(keys) > 0 {
			r.client.Unlink(ctx, keys...) 
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

func (r *RedisCache) Delete(key string) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	r.client.Unlink(ctx, key) 
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) Flush() error {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()
	return r.client.FlushAll(ctx).Err()
}