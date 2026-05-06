package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/farildzaky/sidita-service/internal/cache"
	"github.com/farildzaky/sidita-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// CONTEXT TIMEOUTS
// =============================================================================

const (
	ContextQueryTimeout  = 5 * time.Second
	ContextUploadTimeout = 15 * time.Second
	ContextDBTimeout     = 5 * time.Second
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
	CacheTTLList = getEnvDuration("CACHE_TTL_LIST", 1*time.Hour)

	CacheTTLDetail = getEnvDuration("CACHE_TTL_DETAIL", 2*time.Hour)

	CacheTTLMaps = getEnvDuration("CACHE_TTL_MAPS", 2*time.Hour)

	CacheTTLArea = getEnvDuration("CACHE_TTL_AREA", 24*time.Hour)

	CacheTTLEventList = getEnvDuration("CACHE_TTL_EVENT_LIST", 30*time.Minute)

	CacheTTLRecommendation = getEnvDuration("CACHE_TTL_RECOMMENDATION", 2*time.Hour)
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
		if err := c.Send(cached); err != nil {
			slog.Error("respondCached: gagal mengirim response", "key", key, "error", err)
		}
		return true
	}
	c.Set("X-Cache", "MISS")
	return false
}

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
	return utils.ValidatePaginationParams(c.Query("page"), c.Query("limit"))
}

func buildPaginationMeta(page, limit int, totalData int64) *PaginationMeta {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(totalData) / float64(limit)))
	}
	return &PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalData:  totalData,
		TotalPages: totalPages,
	}
}

// =============================================================================
// CLOUDINARY UPLOAD HELPERS 
// =============================================================================

const maxImageSizeBytes = 2 * 1024 * 1024 
func uploadImage(
	ctx context.Context,
	cld *cloudinary.Cloudinary,
	fileHeader *multipart.FileHeader,
	folderName string,
) (string, string, error) {
	if cld == nil {
		return "", "", fmt.Errorf("layanan penyimpanan gambar belum terkonfigurasi")
	}
	if fileHeader.Size > maxImageSizeBytes {
		return "", "", fmt.Errorf("ukuran gambar maksimal 2MB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("gagal membuka file gambar")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", "", fmt.Errorf("gagal membaca file gambar")
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", "", fmt.Errorf("gagal reset pointer file")
	}

	mimeType := http.DetectContentType(buffer)
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return "", "", fmt.Errorf("format gambar harus JPG/PNG/WEBP")
	}

	dangerousExts := map[string]bool{
		"php": true, "jsp": true, "asp": true, "aspx": true,
		"sh": true, "exe": true, "py": true, "rb": true, "cgi": true,
	}
	nameParts := strings.Split(strings.ToLower(fileHeader.Filename), ".")
	for _, part := range nameParts[:len(nameParts)-1] {
		if dangerousExts[part] {
			return "", "", fmt.Errorf("nama file mengandung ekstensi berbahaya")
		}
	}

	resCld, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folderName,
	})
	if err != nil {
		return "", "", fmt.Errorf("gagal upload ke Cloudinary: %v", err)
	}
	return resCld.SecureURL, resCld.PublicID, nil
}

func destroyImageAsync(cld *cloudinary.Cloudinary, publicID string) {
	if cld == nil || publicID == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		if err != nil {
			slog.Error("destroyImageAsync: gagal hapus gambar Cloudinary", "publicID", publicID, "error", err)
		}
	}()
}