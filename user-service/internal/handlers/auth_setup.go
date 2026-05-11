package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/farildzaky/user-service/internal/cache"
	"github.com/farildzaky/user-service/internal/db"
)

const ContextQueryTimeout = 5 * time.Second

type AuthHandler struct {
	Queries      *db.Queries
	JWTSecret    string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

func NewAuthHandler(q *db.Queries, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		Queries:    q,
		JWTSecret:  jwtSecret,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

func (h *AuthHandler) CacheWarmup() {
	slog.Info("cache.warmup.start")
}

func (h *AuthHandler) StartTokenCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cleanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := h.Queries.DeleteExpiredTokens(cleanCtx); err != nil {
				slog.Warn("token.cleanup.error", slog.String("err", err.Error()))
			} else {
				slog.Debug("token.cleanup.done")
			}
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

func invalidateUserCache(userID string) {
	cache.GlobalCache.Delete("user:me:" + userID)
	cache.GlobalCache.Delete("user:prefs:" + userID)
	cache.GlobalCache.Delete("user:favorites:" + userID)
}
