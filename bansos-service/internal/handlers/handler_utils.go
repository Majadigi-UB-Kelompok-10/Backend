package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/farildzaky/bansos-service/internal/cache"
	"github.com/farildzaky/bansos-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// CACHE TTL — Diatur via env, fallback ke default (Khusus Bansos)
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
	CacheTTLProgram    = getEnvDuration("CACHE_TTL_PROGRAM", 24*time.Hour)   
	CacheTTLInfoBansos = getEnvDuration("CACHE_TTL_INFO_BANSOS", 5*time.Minute) 
)

// =============================================================================
// CACHE HELPERS
// =============================================================================

func normalizeKey(parts ...string) string {
	clean := make([]string, len(parts))
	for i, p := range parts {
		clean[i] = strings.ToLower(strings.ReplaceAll(p, " ", "_"))
	}
	return strings.Join(clean, ":")
}

func respondCached(c fiber.Ctx, cc cache.Cache, key string) bool {
	if cached, ok := cc.Get(key); ok {
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

func cacheJSON(c fiber.Ctx, cc cache.Cache, key string, ttl time.Duration, body interface{}) error {
	data, err := sonic.Marshal(body)
	if err != nil {
		slog.Error("cacheJSON: gagal marshal body", slog.String("key", key), slog.String("error", err.Error()))
		return c.JSON(body)
	}
	cc.Set(key, data, ttl)

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

// =============================================================================
// PAGINATION
// =============================================================================

func parsePagination(c fiber.Ctx) (page, limit, offset int) {
	return utils.ValidatePaginationParams(c.Query("page"), c.Query("limit"))
}

// =============================================================================
// FORMATTER HELPERS (Sesuai UI Sapa Bansos)
// =============================================================================

func FormatRupiah(amount int) string {
	s := strconv.Itoa(amount)
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return "Rp. " + string(result)
}

var bulanIndo = []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
var bulanSingkat = []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Ags", "Sep", "Okt", "Nov", "Des"}

func FormatPeriodeTunggal(t time.Time) string {
	return fmt.Sprintf("%s %d", bulanIndo[t.Month()], t.Year())
}

func FormatPeriodeRentang(mulai, selesai time.Time) string {
	if mulai.Year() == selesai.Year() {
		if mulai.Month() == selesai.Month() {
			return fmt.Sprintf("%s %d", bulanSingkat[mulai.Month()], mulai.Year())
		}
		return fmt.Sprintf("%s - %s %d", bulanSingkat[mulai.Month()], bulanSingkat[selesai.Month()], mulai.Year())
	}
	
	return fmt.Sprintf("%s %d - %s %d", bulanSingkat[mulai.Month()], mulai.Year(), bulanSingkat[selesai.Month()], selesai.Year())
}

func FormatMetode(metode string) string {
	if metode == "tunai" {
		return "Tunai"
	}
	parts := strings.Split(metode, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			if parts[i] == "bri" || parts[i] == "bni" || parts[i] == "bca" {
				parts[i] = strings.ToUpper(parts[i])
			} else {
				parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
			}
		}
	}
	return strings.Join(parts, " ")
}