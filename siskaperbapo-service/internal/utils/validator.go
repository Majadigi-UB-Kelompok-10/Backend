package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// =============================================================================
// PATTERNS
// =============================================================================

var (
	safeQueryPattern = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	slugPattern      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	datePattern      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`) 
)

// =============================================================================
// SEARCH / TEXT INPUT VALIDATION
// =============================================================================

func ValidateQueryString(input string, maxLength int, fieldName string) (string, *ValidationError) {
	if input == "" {
		return "", nil
	}

	trimmed := strings.TrimSpace(input)

	if utf8.RuneCountInString(trimmed) > maxLength {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Input terlalu panjang (maksimal %d karakter)", maxLength),
		}
	}

	sanitized := safeQueryPattern.ReplaceAllString(trimmed, "")
	sanitized = strings.TrimSpace(sanitized)

	if sanitized == "" && trimmed != "" {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Input mengandung karakter yang tidak diizinkan",
		}
	}
	return sanitized, nil
}

func ValidateTextContent(input string, maxLength int, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fieldName + " tidak boleh kosong",
		}
	}
	if utf8.RuneCountInString(trimmed) > maxLength {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s melebihi batas %d karakter", fieldName, maxLength),
		}
	}
	return trimmed, nil
}

// =============================================================================
// SPECIFIC FORMAT VALIDATORS (Slug, URL, Date)
// =============================================================================

func ValidateSlug(input, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", &ValidationError{Field: fieldName, Message: fieldName + " tidak boleh kosong"}
	}
	if !slugPattern.MatchString(trimmed) {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format slug tidak valid (hanya huruf kecil, angka, dan strip)",
		}
	}
	return trimmed, nil
}

func ValidateURL(input, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", nil
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format link tidak valid (harus diawali http:// atau https://)",
		}
	}
	return trimmed, nil
}

func ValidateDate(input, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", &ValidationError{Field: fieldName, Message: fieldName + " tidak boleh kosong"}
	}
	if !datePattern.MatchString(trimmed) {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format tanggal tidak valid (gunakan format YYYY-MM-DD)",
		}
	}
	
	_, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Tanggal tidak valid atau tidak ada di kalender",
		}
	}
	
	return trimmed, nil
}

// =============================================================================
// PAGINATION (Standar Enterprise yang sinkron dengan DB Queries)
// =============================================================================

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

func ValidatePaginationParams(pageStr, limitStr string) (page, limit, offset int) {
	page = parseIntDefault(pageStr, DefaultPage)
	if page < 1 {
		page = DefaultPage
	}

	limit = parseIntDefault(limitStr, DefaultLimit)
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset = (page - 1) * limit
	return
}

func parseIntDefault(s string, fallback int) int {
	if s == "" || len(s) > 6 {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}