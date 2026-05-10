package handlers

import (
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/farildzaky/bansos-service/internal/cache"
	"github.com/farildzaky/bansos-service/internal/db"
)

const (
	ContextQueryTimeout = 5 * time.Second
)



type BansosHandler struct {
	Queries *db.Queries
	DB      *pgxpool.Pool
	Cache   cache.Cache 
}

func NewBansosHandler(q *db.Queries, dbPool *pgxpool.Pool, c cache.Cache) *BansosHandler {
	return &BansosHandler{
		Queries: q,
		DB:      dbPool,
		Cache:   c,
	}
}

func (h *BansosHandler) CacheWarmup() {
	slog.Info("bansos.cache.warmup.start")
	startTime := time.Now()

	slog.Info("bansos.cache.warmup.success",
		slog.String("duration", time.Since(startTime).String()),
	)
}

func (h *BansosHandler) cacheJSONToWarmup(key string, ttl time.Duration, data interface{}) {
	if b, err := sonic.Marshal(data); err == nil {
		h.Cache.Set(key, b, ttl)
	}
}