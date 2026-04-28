package cache

import (
	"context"
	"fmt"
	"time"
	"log/slog" 

	"github.com/bytedance/sonic" 
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

func (r *RedisCache) Set(key string, val interface{}, ttl time.Duration) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	var data []byte
	var err error

	if v, ok := val.([]byte); ok {
		data = v
	} else {
		data, err = sonic.Marshal(val)
		if err != nil {
			slog.Error("Gagal marshal val Redis", slog.String("key", key), slog.Any("error", err))
			return
		}
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		slog.Error("Gagal simpan key Redis", slog.String("key", key), slog.Any("error", err))
	}
}

func (r *RedisCache) InvalidatePattern(pattern string) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "*"+pattern+"*", 100).Result()
		if err != nil {
			slog.Error("Scan InvalidatePattern gagal", slog.Any("error", err))
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
			r.client.Unlink(ctx, keys...) // 🚀 Mempertahankan Unlink
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