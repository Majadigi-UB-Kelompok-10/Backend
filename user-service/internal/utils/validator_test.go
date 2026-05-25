package utils

import (
	"testing"
	"time"
)

func TestValidateEmail_User(t *testing.T) {
	valid := []string{"user@example.com", "first.last@mail.co.id", "user+tag@sub.domain.com"}
	for _, v := range valid {
		if _, err := ValidateEmail(v); err != nil {
			t.Errorf("ValidateEmail(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "notanemail", "missing@tld", "@nodomain.com", "no-at-sign"}
	for _, v := range invalid {
		if _, err := ValidateEmail(v); err == nil {
			t.Errorf("ValidateEmail(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePhone_User(t *testing.T) {
	valid := []string{"+6281234567890", "+628123456789"}
	for _, v := range valid {
		if _, err := ValidatePhone(v); err != nil {
			t.Errorf("ValidatePhone(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "081234567890", "6281234567890", "12345", "abcdefghijk"}
	for _, v := range invalid {
		if _, err := ValidatePhone(v); err == nil {
			t.Errorf("ValidatePhone(%q) expected error, got nil", v)
		}
	}
}

func TestValidateNIK_User(t *testing.T) {
	if _, err := ValidateNIK("3578010101900001"); err != nil {
		t.Errorf("ValidateNIK valid: %v", err)
	}
	invalid := []string{"", "123456789012345", "12345678901234567", "abc1234567890123"}
	for _, v := range invalid {
		if _, err := ValidateNIK(v); err == nil {
			t.Errorf("ValidateNIK(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("password123"); err != nil {
		t.Errorf("ValidatePassword valid: %v", err)
	}
	if err := ValidatePassword("12345678"); err != nil {
		t.Errorf("ValidatePassword exactly 8 chars: %v", err)
	}
	if err := ValidatePassword("short"); err == nil {
		t.Error("ValidatePassword too short expected error")
	}
	if err := ValidatePassword(""); err == nil {
		t.Error("ValidatePassword empty expected error")
	}
}

func TestValidateBirthDate(t *testing.T) {
	_, err := ValidateBirthDate("01/15/1995")
	if err != nil {
		t.Errorf("ValidateBirthDate valid MM/DD/YYYY: %v", err)
	}

	_, err = ValidateBirthDate("")
	if err != nil {
		t.Errorf("ValidateBirthDate empty should be allowed (optional): %v", err)
	}

	_, err = ValidateBirthDate("1995-01-15")
	if err == nil {
		t.Error("ValidateBirthDate wrong format YYYY-MM-DD expected error")
	}

	future := time.Now().AddDate(0, 0, 1).Format("01/02/2006")
	_, err = ValidateBirthDate(future)
	if err == nil {
		t.Error("ValidateBirthDate future date expected error")
	}
}

func TestValidatePaginationParams_User(t *testing.T) {
	page, limit, offset := ValidatePaginationParams("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ValidatePaginationParams("2", "20")
	if page != 2 || limit != 20 || offset != 20 {
		t.Errorf("page=2,limit=20: page=%d limit=%d offset=%d", page, limit, offset)
	}
	_, limit, _ = ValidatePaginationParams("1", "200")
	if limit != 10 {
		t.Errorf("limit>100 should fall back to 10, got %d", limit)
	}
}
