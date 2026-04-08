package utils

import (
	"fmt"
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ValidationError struct {
	Field   string
	Message string
}

var (
	safeQueryPattern    = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	safeFilenamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	emailPattern        = regexp.MustCompile(`^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+\.[A-Za-z]+$`)
	phonePattern        = regexp.MustCompile(`^(?:\+62|62|0)[2-9][0-9]{7,13}$`)
)

func ValidateQueryString(input string, maxLength int, fieldName string) (string, *ValidationError) {
	if input == "" {
		return "", nil
	}

	input = strings.TrimSpace(input)

	if utf8.RuneCountInString(input) > maxLength {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Input terlalu panjang (maksimal %d karakter)", maxLength),
		}
	}

	sanitized := safeQueryPattern.ReplaceAllString(input, "")

	sanitized = html.EscapeString(sanitized)

	if strings.TrimSpace(sanitized) == "" && input != "" {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Input mengandung karakter berbahaya yang tidak diizinkan",
		}
	}

	return strings.TrimSpace(sanitized), nil
}

func ValidateTextContent(input string, maxLength int, fieldName string) (string, *ValidationError) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", &ValidationError{Field: fieldName, Message: fieldName + " tidak boleh kosong"}
	}
	if utf8.RuneCountInString(input) > maxLength {
		return "", &ValidationError{Field: fieldName, Message: fmt.Sprintf("%s melebihi batas %d karakter", fieldName, maxLength)}
	}
	
	return html.EscapeString(input), nil
}

func ValidateEmail(input string) (string, *ValidationError) {
	input = strings.TrimSpace(input)
	if !emailPattern.MatchString(input) {
		return "", &ValidationError{Field: "email", Message: "Format email tidak valid"}
	}
	return input, nil
}

func ValidatePhone(input string) (string, *ValidationError) {
	input = strings.TrimSpace(input)
	if !phonePattern.MatchString(input) {
		return "", &ValidationError{Field: "no_hp", Message: "Format nomor HP tidak valid (gunakan awalan 08 / +62 / 62)"}
	}
	return input, nil
}

func ValidateURL(input, fieldName string) (string, *ValidationError) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil 
	}

	parsedURL, err := url.ParseRequestURI(input)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format link tidak valid (harus menggunakan http:// atau https://)",
		}
	}
	return input, nil
}

func ValidateImageFilename(original string) string {
	filename := strings.TrimSpace(original)

	filename = filepath.Clean(filename)
	filename = filepath.Base(filename) 

	filename = safeFilenamePattern.ReplaceAllString(filename, "")

	if filename == "" || filename == "." {
		filename = "image_file"
	}

	return filename
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