package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
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


var (
	safeQueryPattern    = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	safeFilenamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	emailPattern        = regexp.MustCompile(`^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
	phonePattern        = regexp.MustCompile(`^(?:\+62|62|0)8[1-9][0-9]{7,11}$`)
	slugPattern         = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// =============================================================================
// SEARCH/TEXT INPUT VALIDATION
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


func ValidateEmail(input string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(strings.ToLower(input))
	if trimmed == "" {
		return "", &ValidationError{Field: "email", Message: "Email tidak boleh kosong"}
	}
	if !emailPattern.MatchString(trimmed) {
		return "", &ValidationError{Field: "email", Message: "Format email tidak valid"}
	}
	return trimmed, nil
}

func ValidatePhone(input string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(input, " ", ""))
	if trimmed == "" {
		return "", &ValidationError{Field: "no_hp", Message: "Nomor HP tidak boleh kosong"}
	}
	if !phonePattern.MatchString(trimmed) {
		return "", &ValidationError{
			Field:   "no_hp",
			Message: "Format nomor HP tidak valid (gunakan awalan 08, +62, atau 62)",
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
			Message: "Format link tidak valid (harus http:// atau https://)",
		}
	}
	return trimmed, nil
}


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

// =============================================================================
// FILENAME — buat upload (image)
// =============================================================================


func ValidateImageFilename(original string) string {
	filename := strings.TrimSpace(original)
	filename = filepath.Base(filename)
	filename = safeFilenamePattern.ReplaceAllString(filename, "")
	if filename == "" || filename == "." {
		filename = "image_file"
	}
	return filename
}

// =============================================================================
// PAGINATION
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