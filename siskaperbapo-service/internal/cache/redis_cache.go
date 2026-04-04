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
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (r *RedisCache) contextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func (r *RedisCache) Get(key string) (interface{}, bool) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil || err != nil {
		return nil, false
	}
	
	return val, true
}

func (r *RedisCache) Set(key string, val interface{}, ttl time.Duration) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	data, err := json.Marshal(val)
	if err != nil {
		return
	}

	r.client.Set(ctx, key, data, ttl)
}

func (r *RedisCache) GetImmutable(key string) (interface{}, bool) {
	return r.Get(key)
}

func (r *RedisCache) SetImmutable(key string, val interface{}) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	data, err := json.Marshal(val)
	if err != nil {
		return
	}

	r.client.Set(ctx, key, data, 0) 
}

func (r *RedisCache) InvalidatePattern(pattern string) {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "*"+pattern+"*", 100).Result()
		if err != nil {
			fmt.Printf("[REDIS ERROR] Gangguan saat InvalidatePattern: %v\n", err)
			return
		}

		if len(keys) > 0 {
			r.client.Del(ctx, keys...)
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
			r.client.Del(ctx, keys...)
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

	r.client.Del(ctx, key)
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) Flush() error {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	return r.client.FlushAll(ctx).Err()
}

func (r *RedisCache) Stats() map[string]string {
	ctx, cancel := r.contextWithTimeout()
	defer cancel()

	info := r.client.Info(ctx, "stats").Val()
	return map[string]string{
		"info": info,
	}
}