package utils

import (
	"fmt"
	"regexp"
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

var (
	emailPattern = regexp.MustCompile(`^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
	phonePattern = regexp.MustCompile(`^\+62[0-9]{8,13}$`)
	nikPattern   = regexp.MustCompile(`^[0-9]{16}$`)
)

func ValidateEmail(email string) (string, *ValidationError) {
	v := strings.TrimSpace(email)
	if v == "" {
		return "", &ValidationError{Field: "email", Message: "email wajib diisi"}
	}
	if !emailPattern.MatchString(v) {
		return "", &ValidationError{Field: "email", Message: "format email tidak valid"}
	}
	return strings.ToLower(v), nil
}

func ValidatePhone(phone string) (string, *ValidationError) {
	v := strings.TrimSpace(phone)
	if v == "" {
		return "", &ValidationError{Field: "phone", Message: "nomor HP wajib diisi"}
	}
	if !phonePattern.MatchString(v) {
		return "", &ValidationError{Field: "phone", Message: "format nomor HP tidak valid (gunakan format +628xxxxxxxx)"}
	}
	return v, nil
}

func ValidateNIK(nik string) (string, *ValidationError) {
	v := strings.TrimSpace(nik)
	if v == "" {
		return "", &ValidationError{Field: "nik", Message: "NIK wajib diisi"}
	}
	if !nikPattern.MatchString(v) {
		return "", &ValidationError{Field: "nik", Message: "NIK harus 16 digit angka"}
	}
	return v, nil
}

func ValidatePassword(password string) *ValidationError {
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password minimal 8 karakter"}
	}
	return nil
}

func ValidateQueryString(input string, maxLength int, fieldName string) (string, *ValidationError) {
	if input == "" {
		return "", nil
	}
	trimmed := strings.TrimSpace(input)
	if utf8.RuneCountInString(trimmed) > maxLength {
		return "", &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("input terlalu panjang (maksimal %d karakter)", maxLength),
		}
	}
	return trimmed, nil
}

func ValidateBirthDate(raw string) (time.Time, *ValidationError) {
	if raw == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("01/02/2006", raw)
	if err != nil {
		return time.Time{}, &ValidationError{
			Field:   "birth_date",
			Message: "format tanggal lahir tidak valid (gunakan mm/dd/yyyy)",
		}
	}
	if t.After(time.Now()) {
		return time.Time{}, &ValidationError{
			Field:   "birth_date",
			Message: "tanggal lahir tidak boleh di masa depan",
		}
	}
	return t, nil
}

func ValidatePaginationParams(pageStr, limitStr string) (page, limit, offset int) {
	page = 1
	limit = 10
	if p := strings.TrimSpace(pageStr); p != "" {
		if n, err := fmt.Sscanf(p, "%d", &page); n != 1 || err != nil || page < 1 {
			page = 1
		}
	}
	if l := strings.TrimSpace(limitStr); l != "" {
		if n, err := fmt.Sscanf(l, "%d", &limit); n != 1 || err != nil || limit < 1 || limit > 100 {
			limit = 10
		}
	}
	offset = (page - 1) * limit
	return
}
