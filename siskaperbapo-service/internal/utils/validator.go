package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type ValidationError struct {
	Field   string
	Message string
}


func ValidateQueryString(input string, maxLength int, fieldName string) (string, *ValidationError) {
	if input == "" {
		return "", nil 
	}

	input = strings.TrimSpace(input)

	if utf8.RuneCountInString(input) > maxLength {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Input terlalu panjang (max: " + string(rune(maxLength)) + " karakter)",
		}
	}

	dangerousPattern := regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	sanitized := dangerousPattern.ReplaceAllString(input, "")

	if strings.TrimSpace(sanitized) == "" && input != "" {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Input mengandung karakter yang tidak diizinkan",
		}
	}

	return strings.TrimSpace(sanitized), nil
}

func ValidatePaginationParams(pageStr, limitStr string) (page int, limit int) {
	page = 1
	limit = 10

	if p, ok := parseInt(pageStr); ok && p > 0 {
		page = p
	}
	if l, ok := parseInt(limitStr); ok && l > 0 {
		limit = l
	}

	// Enforce limits
	if limit > 50 {
		limit = 50
	}

	return page, limit
}

func parseInt(s string) (int, bool) {
	if s == "" || len(s) > 3 {
		return 0, false
	}
	var result int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		result = result*10 + int(ch-'0')
	}
	return result, true
}


func ValidateImageFilename(original string) string {
	filename := strings.TrimSpace(original)
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")

	safePattern := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	filename = safePattern.ReplaceAllString(filename, "")

	if filename == "" {
		filename = "image"
	}

	return filename
}
