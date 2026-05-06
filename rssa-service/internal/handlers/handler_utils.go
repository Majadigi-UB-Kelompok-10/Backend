package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/farildzaky/rssa-service/internal/cache"
	"github.com/farildzaky/rssa-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// CONTEXT TIMEOUTS
// =============================================================================

const (
	ContextQueryTimeout = 5 * time.Second
	ContextDBTimeout    = 5 * time.Second
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
	// TTL Khusus RSSA: Data ketersediaan kamar harus cukup update
	CacheTTLPublic = getEnvDuration("CACHE_TTL_PUBLIC", 2*time.Minute)
	CacheTTLKelas  = getEnvDuration("CACHE_TTL_KELAS", 24*time.Hour) // Master data jarang berubah
)

// =============================================================================
// CACHE HELPERS
// =============================================================================

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
		_ = c.Send(cached)
		return true
	}
	c.Set("X-Cache", "MISS")
	return false
}

func cacheJSON(c fiber.Ctx, key string, ttl time.Duration, body interface{}) error {
	cache.GlobalCache.Set(key, body, ttl)
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

func validationErrorResponse(c fiber.Ctx, ve *utils.ValidationError) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Code:    "ERR_VALIDATION_FAILED",
		Message: "Input tidak valid",
		Action:  "Harap perbaiki data input",
		Errors: []FieldError{
			{Field: ve.Field, Message: ve.Message},
		},
	})
}

// =============================================================================
// PAGINATION HELPERS
// =============================================================================

func parsePagination(c fiber.Ctx) (page, limit, offset int) {
	return utils.ParsePagination(c.Query("page"), c.Query("limit"))
}

func buildPaginationMeta(page, limit int, totalData int64) *PaginationMeta {
	return &PaginationMeta{
		Page:  page,
		Limit: limit,
		Total: totalData,
	}
}