package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

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



const (
	DefaultPage     = 1
	DefaultLimit    = 10
	MaxLimit        = 100
	maxPageDigits   = 6 
	maxLimitDigits  = 3 
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


var platNomorPattern = regexp.MustCompile(`^[A-Z0-9]{3,15}$`)


func ValidatePlatNomor(plat string) (string, *ValidationError) {
	cleaned := strings.ToUpper(strings.ReplaceAll(plat, " ", ""))
	if cleaned == "" {
		return "", &ValidationError{Field: "plat_nomor", Message: "Plat nomor wajib diisi"}
	}
	if !platNomorPattern.MatchString(cleaned) {
		return "", &ValidationError{
			Field:   "plat_nomor",
			Message: "Plat nomor harus berupa huruf dan angka (3-15 karakter)",
		}
	}
	return cleaned, nil
}

var nomorRangkaPattern = regexp.MustCompile(`^[A-Z0-9]{5}$`)

func ValidateNomorRangka(rangka string) (string, *ValidationError) {
	cleaned := strings.ToUpper(strings.TrimSpace(rangka))
	if !nomorRangkaPattern.MatchString(cleaned) {
		return "", &ValidationError{
			Field:   "nomor_rangka",
			Message: "Nomor rangka harus 5 karakter terakhir (huruf/angka)",
		}
	}
	return cleaned, nil
}



var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func SanitizeFilename(original string) string {
	filename := strings.TrimSpace(original)
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = unsafeFilenameChars.ReplaceAllString(filename, "")
	if filename == "" {
		filename = "image"
	}
	return filename
}