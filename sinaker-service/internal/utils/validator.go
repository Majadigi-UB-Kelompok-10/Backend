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
// PATTERNS (Tambahan untuk Sinaker)
// =============================================================================

var (
	safeQueryPattern = regexp.MustCompile(`[^a-zA-Z0-9 \-_./]`)
	slugPattern      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	datePattern      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	nikPattern       = regexp.MustCompile(`^\d{16}$`)                                      
	emailPattern     = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`) 
	phonePattern     = regexp.MustCompile(`^[0-9+]{9,15}$`)                                
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
// SPECIFIC FORMAT VALIDATORS (Sinaker Edition)
// =============================================================================

func ValidateNIK(input, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if !nikPattern.MatchString(trimmed) {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format NIK tidak valid (harus persis 16 digit angka)",
		}
	}
	return trimmed, nil
}

func ValidateEmail(input, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if !emailPattern.MatchString(trimmed) {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format Email tidak valid",
		}
	}
	return trimmed, nil
}

func ValidatePhone(input, fieldName string) (string, *ValidationError) {
	trimmed := strings.TrimSpace(input)
	if !phonePattern.MatchString(trimmed) {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Format Nomor HP tidak valid (9-15 digit angka, boleh diawali +)",
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
		return "", nil // Seringkali URL (seperti foto) bisa jadi optional sebelum di-upload
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

// =============================================================================
// ENUM VALIDATORS
// =============================================================================

func ValidateJenisKelamin(input string) (string, *ValidationError) {
	val := strings.ToLower(strings.TrimSpace(input))
	if val != "laki_laki" && val != "perempuan" {
		return "", &ValidationError{
			Field:   "jenis_kelamin",
			Message: "Jenis kelamin harus 'laki_laki' atau 'perempuan'",
		}
	}
	return val, nil
}

func ValidatePendidikan(input, fieldName string) (string, *ValidationError) {
	validPendidikan := map[string]bool{
		"tidak_sekolah": true, "sd": true, "smp": true, "sma_smk": true,
		"d3": true, "s1": true, "s2": true, "s3": true,
	}

	val := strings.ToLower(strings.TrimSpace(input))
	if !validPendidikan[val] {
		return "", &ValidationError{
			Field:   fieldName,
			Message: "Tingkat pendidikan tidak valid",
		}
	}
	return val, nil
}

func ValidateStatusPendaftaran(input string) (string, *ValidationError) {
	val := strings.ToLower(strings.TrimSpace(input))
	if val != "pending" && val != "diterima" && val != "ditolak" {
		return "", &ValidationError{
			Field:   "status",
			Message: "Status harus 'pending', 'diterima', atau 'ditolak'",
		}
	}
	return val, nil
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