package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/farildzaky/sinaker-service/internal/cache"
	"github.com/farildzaky/sinaker-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// CACHE TTL — Diatur via env, fallback ke default
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
	CacheTTLList   = getEnvDuration("CACHE_TTL_LIST", 2*time.Hour)
	CacheTTLDetail = getEnvDuration("CACHE_TTL_DETAIL", 1*time.Hour)
	CacheTTLStatic = getEnvDuration("CACHE_TTL_STATIC", 24*time.Hour)
)

// =============================================================================
// CACHE HELPERS — Konsisten di semua handler
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
		if err := c.Send(cached); err != nil {
			slog.Error("respondCached: gagal mengirim response", slog.String("key", key), slog.String("error", err.Error()))
		}
		return true
	}
	c.Set("X-Cache", "MISS")
	return false
}

func cacheJSON(c fiber.Ctx, key string, ttl time.Duration, body interface{}) error {
	data, err := sonic.Marshal(body)
	if err != nil {
		slog.Error("cacheJSON: gagal marshal body", slog.String("key", key), slog.String("error", err.Error()))
		return c.JSON(body)
	}
	cache.GlobalCache.Set(key, data, ttl)
	
	c.Set("Content-Type", "application/json")
	return c.Send(data)
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


func parsePagination(c fiber.Ctx) (page, limit, offset int) {
	return utils.ValidatePaginationParams(c.Query("page"), c.Query("limit"))
}