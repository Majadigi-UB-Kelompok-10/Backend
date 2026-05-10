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

// =============================================================================
// VALIDATION ERROR
// =============================================================================
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
	safeQueryPattern    = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	safeFilenamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)

// =============================================================================
// TEXT & SEARCH VALIDATION
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
	return sanitized, nil
}

func ValidateTextContent(input string, maxLength int, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", &ValidationError{Field: fieldName, Message: fieldName + " tidak boleh kosong"}
	}
	if utf8.RuneCountInString(trimmed) > maxLength {
		return "", &ValidationError{Field: fieldName, Message: fmt.Sprintf("%s melebihi batas %d karakter", fieldName, maxLength)}
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
		return "", &ValidationError{Field: fieldName, Message: "Format link tidak valid (harus http:// atau https://)"}
	}
	return trimmed, nil
}

// =============================================================================
// JDIH SPECIFIC VALIDATORS (Sesuai ENUM dan CHECK di PostgreSQL)
// =============================================================================

func ValidateTahun(tahunStr string) (int16, *ValidationError) {
	tahun, err := strconv.Atoi(tahunStr)
	// Constraint chk_dokumen_tahun (tahun > 1900 AND tahun <= 2200)
	if err != nil || tahun <= 1900 || tahun > 2200 {
		return 0, &ValidationError{Field: "tahun", Message: "Tahun harus berupa angka antara 1901 - 2200"}
	}
	return int16(tahun), nil
}

func ValidateJenisDokumen(jenis string) (string, *ValidationError) {
	validJenis := map[string]bool{
		"perda": true, "pergub": true, "peraturan": true, "perdes": true,
		"sk_gub": true, "instruksi": true, "se": true, "keputusan": true,
	}
	trimmed := strings.ToLower(strings.TrimSpace(jenis))
	if !validJenis[trimmed] {
		return "", &ValidationError{Field: "jenis", Message: "Jenis dokumen tidak valid"}
	}
	return trimmed, nil
}

func ValidateStatusDokumen(status string) (string, *ValidationError) {
	validStatus := map[string]bool{
		"berlaku": true, "diubah": true, "dicabut": true,
	}
	trimmed := strings.ToLower(strings.TrimSpace(status))
	if !validStatus[trimmed] {
		return "", &ValidationError{Field: "status", Message: "Status dokumen tidak valid"}
	}
	return trimmed, nil
}

// =============================================================================
// PDF FILENAME VALIDATOR (Khusus Dokumen JDIH)
// =============================================================================
func ValidatePDFFilename(original string) string {
	filename := strings.TrimSpace(original)
	filename = filepath.Base(filename)
	filename = safeFilenamePattern.ReplaceAllString(filename, "")
	
	// Pastikan file berakhiran .pdf
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}
	
	if filename == ".pdf" || filename == "" {
		filename = "dokumen_hukum.pdf"
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