package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// CACHE TTL — diatur via env, fallback ke default
// =============================================================================

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
	CacheTTLList = getEnvDuration("CACHE_TTL_LIST", 10*time.Minute)

	CacheTTLDetail = getEnvDuration("CACHE_TTL_DETAIL", 30*time.Minute)

	CacheTTLMaps = getEnvDuration("CACHE_TTL_MAPS", 1*time.Hour)

	CacheTTLStatic = getEnvDuration("CACHE_TTL_AREA", 24*time.Hour)
)

// =============================================================================
// CACHE HELPERS — konsisten di semua handler
// =============================================================================


func normalizeKey(parts ...string) string {
	clean := make([]string, len(parts))
	for i, p := range parts {
		clean[i] = strings.ToLower(strings.ReplaceAll(p, " ", ""))
	}
	return strings.Join(clean, ":")
}

// respondCached: kalau key ada di cache, langsung return. True kalau cache hit.
func respondCached(c fiber.Ctx, key string) bool {
	if cached, ok := cache.GlobalCache.Get(key); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		if err := c.Send(cached); err != nil {
			slog.Error("respondCached: gagal mengirim response", "key", key, "error", err)
		}
		return true
	}
	c.Set("X-Cache", "MISS")
	return false
}

// cacheJSON: simpan response ke cache dengan TTL, lalu return JSON ke client.
func cacheJSON(c fiber.Ctx, key string, ttl time.Duration, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		slog.Error("cacheJSON: gagal marshal body", "key", key, "error", err)
		return c.JSON(body)
	}
	cache.GlobalCache.Set(key, data, ttl)
	return c.JSON(body)
}

// =============================================================================
// VALIDATION HELPERS
// =============================================================================


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
			Action:  "Harap perbaiki data input",
			Errors:  errs,
		})
	}
	return nil
}


func parsePagination(c fiber.Ctx) (page, limit, offset int) {
	return utils.ParsePagination(c.Query("page"), c.Query("limit"))
}