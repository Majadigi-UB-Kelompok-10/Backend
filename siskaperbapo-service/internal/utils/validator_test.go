package utils

import (
	"testing"
)

func TestValidateQueryString_Siskaperbapo(t *testing.T) {
	out, err := ValidateQueryString("beras medium", 100, "q")
	if err != nil || out == "" {
		t.Errorf("ValidateQueryString valid: err=%v out=%q", err, out)
	}
	_, err = ValidateQueryString("toolong", 3, "q")
	if err == nil {
		t.Error("ValidateQueryString over max expected error")
	}
	out, err = ValidateQueryString("", 100, "q")
	if err != nil || out != "" {
		t.Error("ValidateQueryString empty should return empty no error")
	}
}

func TestValidateSlug_Siskaperbapo(t *testing.T) {
	valid := []string{"beras-medium", "minyak-goreng-2026", "a1"}
	for _, v := range valid {
		if _, err := ValidateSlug(v, "slug"); err != nil {
			t.Errorf("ValidateSlug(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "Has Space", "Capital", "special!@"}
	for _, v := range invalid {
		if _, err := ValidateSlug(v, "slug"); err == nil {
			t.Errorf("ValidateSlug(%q) expected error, got nil", v)
		}
	}
}

func TestValidateURL_Siskaperbapo(t *testing.T) {
	if _, err := ValidateURL("https://example.com", "url"); err != nil {
		t.Errorf("ValidateURL valid https: %v", err)
	}
	out, err := ValidateURL("", "url")
	if err != nil || out != "" {
		t.Error("ValidateURL empty should return empty no error")
	}
	invalid := []string{"example.com", "ftp://example.com", "javascript:alert(1)"}
	for _, v := range invalid {
		if _, err := ValidateURL(v, "url"); err == nil {
			t.Errorf("ValidateURL(%q) expected error, got nil", v)
		}
	}
}

func TestValidateDate_Siskaperbapo(t *testing.T) {
	if _, err := ValidateDate("2026-05-23", "tanggal"); err != nil {
		t.Errorf("ValidateDate valid: %v", err)
	}
	invalid := []string{"", "23/05/2026", "2026-13-01", "notadate"}
	for _, v := range invalid {
		if _, err := ValidateDate(v, "tanggal"); err == nil {
			t.Errorf("ValidateDate(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePaginationParams_Siskaperbapo(t *testing.T) {
	page, limit, offset := ValidatePaginationParams("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ValidatePaginationParams("4", "50")
	if page != 4 || limit != 50 || offset != 150 {
		t.Errorf("page=4,limit=50: page=%d limit=%d offset=%d", page, limit, offset)
	}
	_, limit, _ = ValidatePaginationParams("1", "999")
	if limit != 100 {
		t.Errorf("limit clamped to 100, got %d", limit)
	}
}

func TestValidateTextContent_Siskaperbapo(t *testing.T) {
	out, err := ValidateTextContent("Beras Medium Premium", 200, "nama")
	if err != nil || out == "" {
		t.Errorf("ValidateTextContent valid: err=%v out=%q", err, out)
	}
	_, err = ValidateTextContent("", 200, "nama")
	if err == nil {
		t.Error("ValidateTextContent empty expected error")
	}
	_, err = ValidateTextContent("x", 0, "nama")
	if err == nil {
		t.Error("ValidateTextContent over max expected error")
	}
}

func BenchmarkValidateSlug_Siskaperbapo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateSlug("beras-medium-2026", "slug")
	}
}

func BenchmarkValidatePaginationParams_Siskaperbapo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams("4", "50")
	}
}
