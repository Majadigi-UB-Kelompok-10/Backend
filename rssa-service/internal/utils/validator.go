package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// =============================================================================
// KUSTOM ERROR
// =============================================================================
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// =============================================================================
// SANITIZE & VALIDATE STRINGS
// =============================================================================
var dangerousChars = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)

func SanitizeQueryString(input string, maxLength int, fieldName string) (string, *ValidationError) {
	if input == "" {
		return "", nil
	}

	trimmed := strings.TrimSpace(input)

	if utf8.RuneCountInString(trimmed) > maxLength {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Input terlalu panjang (maks %d karakter)", maxLength),
		}
	}

	sanitized := dangerousChars.ReplaceAllString(trimmed, "")
	sanitized = strings.TrimSpace(sanitized)

	if sanitized == "" && trimmed != "" {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Input mengandung karakter yang tidak diizinkan",
		}
	}

	return sanitized, nil
}

// Validasi Slug (Sangat berguna untuk Master Kelas & Ruangan RSSA)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func ValidateSlug(slug string, fieldName string) (string, *ValidationError) {
	cleaned := strings.TrimSpace(slug)
	if cleaned == "" {
		return "", &ValidationError{Field: fieldName, Message: fieldName + " wajib diisi"}
	}
	if !slugPattern.MatchString(cleaned) {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fieldName + " hanya boleh berisi huruf kecil, angka, dan strip (-)",
		}
	}
	return cleaned, nil
}

// =============================================================================
// PAGINATION HELPERS
// =============================================================================
const (
	DefaultPage    = 1
	DefaultLimit   = 10
	MaxLimit       = 100
	maxPageDigits  = 6
	maxLimitDigits = 3
)

func ParsePagination(pageStr, limitStr string) (page, limit, offset int) {
	page = parseIntBounded(pageStr, maxPageDigits, DefaultPage)
	if page < 1 {
		page = DefaultPage
	}

	limit = parseIntBounded(limitStr, maxLimitDigits, DefaultLimit)
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset = (page - 1) * limit
	return
}

func parseIntBounded(s string, maxDigits, fallback int) int {
	if s == "" || len(s) > maxDigits {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}