package utils

import (
	"testing"
)

func TestValidateQueryString_Transjatim(t *testing.T) {
	out, err := ValidateQueryString("surabaya malang", 100, "q")
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

func TestValidateSlug_Transjatim(t *testing.T) {
	valid := []string{"bus-reguler-surabaya", "rute-123", "a"}
	for _, v := range valid {
		if _, err := ValidateSlug(v, "slug"); err != nil {
			t.Errorf("ValidateSlug(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "Has Space", "Capital", "has!char"}
	for _, v := range invalid {
		if _, err := ValidateSlug(v, "slug"); err == nil {
			t.Errorf("ValidateSlug(%q) expected error, got nil", v)
		}
	}
}

func TestValidateURL_Transjatim(t *testing.T) {
	if _, err := ValidateURL("https://transjatim.id", "url"); err != nil {
		t.Errorf("ValidateURL valid: %v", err)
	}
	out, err := ValidateURL("", "url")
	if err != nil || out != "" {
		t.Error("ValidateURL empty should be allowed")
	}
	invalid := []string{"example.com", "ftp://x.com", "not-a-url"}
	for _, v := range invalid {
		if _, err := ValidateURL(v, "url"); err == nil {
			t.Errorf("ValidateURL(%q) expected error, got nil", v)
		}
	}
}

func TestValidateDate_Transjatim(t *testing.T) {
	if _, err := ValidateDate("2026-12-31", "tanggal"); err != nil {
		t.Errorf("ValidateDate valid: %v", err)
	}
	invalid := []string{"", "31-12-2026", "2026/12/31", "notadate", "2026-13-01"}
	for _, v := range invalid {
		if _, err := ValidateDate(v, "tanggal"); err == nil {
			t.Errorf("ValidateDate(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePaginationParams_Transjatim(t *testing.T) {
	page, limit, offset := ValidatePaginationParams("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ValidatePaginationParams("5", "10")
	if page != 5 || limit != 10 || offset != 40 {
		t.Errorf("page=5,limit=10: page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, _, _ = ValidatePaginationParams("-1", "10")
	if page != 1 {
		t.Errorf("negative page should default to 1, got %d", page)
	}
}

func TestValidateTextContent_Transjatim(t *testing.T) {
	out, err := ValidateTextContent("Bus Trans Jatim Reguler", 200, "nama")
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

func BenchmarkValidateSlug_Transjatim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateSlug("bus-reguler-2026", "slug")
	}
}

func BenchmarkValidatePaginationParams_Transjatim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams("5", "10")
	}
}
