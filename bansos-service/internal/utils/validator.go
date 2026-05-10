package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	safeQueryPattern = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	nikPattern       = regexp.MustCompile(`^[0-9]{16}$`) // Regex khusus NIK 16 Digit
)

// =============================================================================
// TEXT & SEARCH VALIDATION (Generic)
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

// =============================================================================
// BANSOS SPECIFIC VALIDATORS (Sesuai Skema Database)
// =============================================================================

// Memastikan NIK tepat 16 digit dan hanya angka
func ValidateNIK(nik string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(nik)
	if !nikPattern.MatchString(trimmed) {
		return "", &ValidationError{Field: "nik", Message: "NIK harus berupa 16 digit angka"}
	}
	return trimmed, nil
}

// Validasi ENUM status_penyaluran_type
func ValidateStatusPenyaluran(status string) (string, *ValidationError) {
	validStatus := map[string]bool{
		"proses": true, "diterima": true, "ditolak": true,
	}
	trimmed := strings.ToLower(strings.TrimSpace(status))
	if !validStatus[trimmed] {
		return "", &ValidationError{Field: "status", Message: "Status penyaluran tidak valid"}
	}
	return trimmed, nil
}

// Validasi ENUM metode_penyaluran_type
func ValidateMetodePenyaluran(metode string) (string, *ValidationError) {
	validMetode := map[string]bool{
		"transfer_bri": true, "transfer_bni": true, "transfer_mandiri": true,
		"transfer_bca": true, "transfer_pos": true, "tunai": true,
	}
	trimmed := strings.ToLower(strings.TrimSpace(metode))
	if !validMetode[trimmed] {
		return "", &ValidationError{Field: "metode", Message: "Metode penyaluran tidak valid"}
	}
	return trimmed, nil
}

// Validasi nominal harus positif (>0)
func ValidateNominal(nominalStr string) (int, *ValidationError) {
	nominal, err := strconv.Atoi(nominalStr)
	if err != nil || nominal <= 0 {
		return 0, &ValidationError{Field: "nominal", Message: "Nominal harus berupa angka lebih dari 0"}
	}
	return nominal, nil
}

// Validasi format tanggal YYYY-MM-DD
func ValidateDateString(dateStr string, fieldName string) (time.Time, *ValidationError) {
	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(dateStr))
	if err != nil {
		return time.Time{}, &ValidationError{Field: fieldName, Message: fmt.Sprintf("Format %s tidak valid (Gunakan YYYY-MM-DD)", fieldName)}
	}
	return parsedDate, nil
}

// Validasi logika rentang tanggal (Selesai >= Mulai)
func ValidatePeriode(mulai, selesai time.Time) *ValidationError {
	if selesai.Before(mulai) {
		return &ValidationError{Field: "periode_selesai", Message: "Periode selesai tidak boleh mendahului periode mulai"}
	}
	return nil
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