package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultRedisOpTimeout = 2 * time.Second

type RedisCache struct {
	client *redis.Client

	hits   atomic.Uint64
	misses atomic.Uint64
	errs   atomic.Uint64
}

// NewRedisCache — bikin Redis cache dengan tuning pool yang aman buat production
func NewRedisCache(redisURL string) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing redis URL: %w", err)
	}

	// Tuning pool — penting buat production
	opt.PoolSize = 50              // max connections
	opt.MinIdleConns = 5           // pre-warmed connections
	opt.MaxRetries = 3             // retry transient errors
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 2 * time.Second
	opt.WriteTimeout = 2 * time.Second

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("gagal koneksi ke Redis: %w", err)
	}
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultRedisOpTimeout)
}

// -----------------------------------------------------------------------------
// Public methods — sama interface dengan SimpleCache
// -----------------------------------------------------------------------------

func (r *RedisCache) Get(key string) ([]byte, bool) {
	ctx, cancel := r.opCtx()
	defer cancel()

	val, err := r.client.Get(ctx, key).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		r.misses.Add(1)
		return nil, false
	case err != nil:
		r.errs.Add(1)
		slog.Warn("cache.redis_get_error",
			slog.String("key", key),
			slog.String("err", err.Error()),
		)
		return nil, false
	}
	r.hits.Add(1)
	return val, true
}

func (r *RedisCache) Set(key string, val interface{}) {
	r.SetWithTTL(key, val, 0)
}

func (r *RedisCache) SetWithTTL(key string, val interface{}, ttl time.Duration) {
	data, ok := serialize(key, val)
	if !ok {
		r.errs.Add(1)
		return
	}

	ctx, cancel := r.opCtx()
	defer cancel()

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		r.errs.Add(1)
		slog.Warn("cache.redis_set_error",
			slog.String("key", key),
			slog.String("err", err.Error()),
		)
	}
}

func (r *RedisCache) Has(key string) bool {
	ctx, cancel := r.opCtx()
	defer cancel()
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.errs.Add(1)
		return false
	}
	return n > 0
}

func (r *RedisCache) Delete(key string) {
	ctx, cancel := r.opCtx()
	defer cancel()
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.errs.Add(1)
		slog.Warn("cache.redis_del_error",
			slog.String("key", key),
			slog.String("err", err.Error()),
		)
	}
}

func (r *RedisCache) DeleteByPrefix(prefix string) {
	r.deleteByScan(prefix + "*")
}

func (r *RedisCache) InvalidatePattern(pattern string) {
	r.deleteByScan("*" + pattern + "*")
}

func (r *RedisCache) Stats() Stats {
	return Stats{
		Hits:   r.hits.Load(),
		Misses: r.misses.Load(),
		// Size: -1 — bisa diisi via DBSIZE tapi ekspensif, biarin
	}
}

// ErrorCount — buat health check / monitoring
func (r *RedisCache) ErrorCount() uint64 {
	return r.errs.Load()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) Flush() error {
	ctx, cancel := r.opCtx()
	defer cancel()
	return r.client.FlushAll(ctx).Err()
}

// -----------------------------------------------------------------------------
// Private helpers
// -----------------------------------------------------------------------------

func (r *RedisCache) deleteByScan(match string) {
	ctx, cancel := r.opCtx()
	defer cancel()

	var (
		cursor       uint64
		totalDeleted int
	)
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, match, 100).Result()
		if err != nil {
			r.errs.Add(1)
			slog.Warn("cache.redis_scan_error",
				slog.String("match", match),
				slog.String("err", err.Error()),
			)
			return
		}
		if len(keys) > 0 {
			if delErr := r.client.Del(ctx, keys...).Err(); delErr != nil {
				r.errs.Add(1)
				slog.Warn("cache.redis_batch_del_error",
					slog.String("err", delErr.Error()),
				)
			} else {
				totalDeleted += len(keys)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	if totalDeleted > 0 {
		slog.Debug("cache.redis_batch_delete",
			slog.String("match", match),
			slog.Int("deleted", totalDeleted),
		)
	}
}