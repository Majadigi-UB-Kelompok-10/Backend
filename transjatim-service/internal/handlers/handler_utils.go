package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/farildzaky/transjatim-service/internal/cache"
	"github.com/farildzaky/transjatim-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

const (
	ContextQueryTimeout = 5 * time.Second
	ContextDBTimeout    = 5 * time.Second
)

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

var (
	CacheTTLTerminals = getEnvDuration("CACHE_TTL_TERMINALS", 24*time.Hour)
	CacheTTLSearch    = getEnvDuration("CACHE_TTL_SEARCH", 5*time.Minute)
	CacheTTLDetail    = getEnvDuration("CACHE_TTL_DETAIL", 1*time.Hour)
	CacheTTLHargaInfo = getEnvDuration("CACHE_TTL_HARGA_INFO", 24*time.Hour)
	CacheTTLStops     = getEnvDuration("CACHE_TTL_STOPS", 2*time.Hour)
	CacheTTLList      = getEnvDuration("CACHE_TTL_LIST", 1*time.Hour)
)

func normalizeKey(parts ...string) string {
	clean := make([]string, len(parts))
	for i, p := range parts {
		clean[i] = strings.ToLower(strings.ReplaceAll(p, " ", ""))
	}
	return strings.Join(clean, ":")
}

func respondCached(c fiber.Ctx, key string) bool {
	if cached, ok := cache.GlobalCache.Get(key); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		if err := c.Send(cached); err != nil {
			slog.Error("respondCached: gagal kirim response", "key", key, "error", err)
		}
		return true
	}
	c.Set("X-Cache", "MISS")
	return false
}

func cacheJSON(c fiber.Ctx, key string, ttl time.Duration, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		slog.Error("cacheJSON: gagal marshal", "key", key, "error", err)
		return c.JSON(body)
	}
	cache.GlobalCache.Set(key, data, ttl)
	return c.JSON(body)
}

func requireFields(c fiber.Ctx, fields map[string]string) error {
	var errs []FieldError
	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, FieldError{
				Field:   field,
				Message: fmt.Sprintf("%s wajib diisi", field),
			})
		}
	}
	if len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION_FAILED",
			Message: "Input tidak valid",
			Errors:  errs,
		})
	}
	return nil
}

func parsePagination(c fiber.Ctx) (page, limit, offset int) {
	return utils.ValidatePaginationParams(c.Query("page"), c.Query("limit"))
}

func buildPaginationMeta(page, limit int, total int64) *PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	_ = totalPages
	return &PaginationMeta{Page: page, Limit: limit, Total: total}
}

// dayBitmask converts weekday to hari_operasi bitmask value
// 1=Sen, 2=Sel, 4=Rab, 8=Kam, 16=Jum, 32=Sab, 64=Min
func dayBitmask(t time.Time) int16 {
	switch t.Weekday() {
	case time.Monday:
		return 1
	case time.Tuesday:
		return 2
	case time.Wednesday:
		return 4
	case time.Thursday:
		return 8
	case time.Friday:
		return 16
	case time.Saturday:
		return 32
	case time.Sunday:
		return 64
	}
	return 1
}
