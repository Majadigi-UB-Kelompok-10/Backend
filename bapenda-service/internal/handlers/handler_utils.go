package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/gofiber/fiber/v3"
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
	CacheTTLList = getEnvDuration("CACHE_TTL_LIST", 10*time.Minute)

	CacheTTLDetail = getEnvDuration("CACHE_TTL_DETAIL", 30*time.Minute)

	CacheTTLMaps = getEnvDuration("CACHE_TTL_MAPS", 1*time.Hour)

	CacheTTLStatic = getEnvDuration("CACHE_TTL_AREA", 24*time.Hour)
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
		_ = c.Send(cached)
		return true
	}
	c.Set("X-Cache", "MISS")
	return false
}


func cacheJSON(c fiber.Ctx, key string, ttl time.Duration, body interface{}) error {
	cache.GlobalCache.SetWithTTL(key, body, ttl)
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
			Action:  "Harap perbaiki data input",
			Errors:  errs,
		})
	}
	return nil
}



const (
	defaultPageLimit = 10
	maxPageLimit     = 100
)


func parsePagination(c fiber.Ctx) (page, limit, offset int) {
	page = atoiSafe(c.Query("page", "1"), 1)
	if page < 1 {
		page = 1
	}
	limit = atoiSafe(c.Query("limit", fmt.Sprintf("%d", defaultPageLimit)), defaultPageLimit)
	if limit < 1 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	offset = (page - 1) * limit
	return
}

func atoiSafe(s string, fallback int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return fallback
	}
	return n
}