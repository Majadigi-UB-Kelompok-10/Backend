package utils

import (
	"testing"
)

func TestValidateNIK_Sinaker(t *testing.T) {
	valid := []string{"3578010101900001", "1234567890123456"}
	for _, v := range valid {
		if _, err := ValidateNIK(v, "nik"); err != nil {
			t.Errorf("ValidateNIK(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "123456789012345", "12345678901234567", "123456789012345a"}
	for _, v := range invalid {
		if _, err := ValidateNIK(v, "nik"); err == nil {
			t.Errorf("ValidateNIK(%q) expected error, got nil", v)
		}
	}
}

func TestValidateEmail_Sinaker(t *testing.T) {
	valid := []string{"user@example.com", "test.user@mail.co.id"}
	for _, v := range valid {
		if _, err := ValidateEmail(v, "email"); err != nil {
			t.Errorf("ValidateEmail(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "notanemail", "missing@tld", "@nodomain.com"}
	for _, v := range invalid {
		if _, err := ValidateEmail(v, "email"); err == nil {
			t.Errorf("ValidateEmail(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePhone_Sinaker(t *testing.T) {
	valid := []string{"081234567890", "+6281234567890", "6281234567890"}
	for _, v := range valid {
		if _, err := ValidatePhone(v, "no_hp"); err != nil {
			t.Errorf("ValidatePhone(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "12345", "1234567890123456"}
	for _, v := range invalid {
		if _, err := ValidatePhone(v, "no_hp"); err == nil {
			t.Errorf("ValidatePhone(%q) expected error, got nil", v)
		}
	}
}

func TestValidateDate_Sinaker(t *testing.T) {
	if _, err := ValidateDate("2026-01-15", "tanggal"); err != nil {
		t.Errorf("ValidateDate valid: %v", err)
	}
	invalid := []string{"", "15/01/2026", "2026-13-01", "not-a-date"}
	for _, v := range invalid {
		if _, err := ValidateDate(v, "tanggal"); err == nil {
			t.Errorf("ValidateDate(%q) expected error, got nil", v)
		}
	}
}

func TestValidateJenisKelamin(t *testing.T) {
	valid := []string{"laki_laki", "perempuan", "LAKI_LAKI", "Perempuan"}
	for _, v := range valid {
		if _, err := ValidateJenisKelamin(v); err != nil {
			t.Errorf("ValidateJenisKelamin(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "laki", "female", "male", "l"}
	for _, v := range invalid {
		if _, err := ValidateJenisKelamin(v); err == nil {
			t.Errorf("ValidateJenisKelamin(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePendidikan(t *testing.T) {
	valid := []string{"sd", "smp", "sma_smk", "d3", "s1", "s2", "s3", "tidak_sekolah", "SD"}
	for _, v := range valid {
		if _, err := ValidatePendidikan(v, "pendidikan"); err != nil {
			t.Errorf("ValidatePendidikan(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "sma", "sarjana", "phd"}
	for _, v := range invalid {
		if _, err := ValidatePendidikan(v, "pendidikan"); err == nil {
			t.Errorf("ValidatePendidikan(%q) expected error, got nil", v)
		}
	}
}

func TestValidateStatusPendaftaran(t *testing.T) {
	valid := []string{"pending", "diterima", "ditolak", "PENDING"}
	for _, v := range valid {
		if _, err := ValidateStatusPendaftaran(v); err != nil {
			t.Errorf("ValidateStatusPendaftaran(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "proses", "selesai", "rejected"}
	for _, v := range invalid {
		if _, err := ValidateStatusPendaftaran(v); err == nil {
			t.Errorf("ValidateStatusPendaftaran(%q) expected error, got nil", v)
		}
	}
}

func TestValidateSlug_Sinaker(t *testing.T) {
	if _, err := ValidateSlug("pelatihan-kerja-2026", "slug"); err != nil {
		t.Errorf("ValidateSlug valid: %v", err)
	}
	invalid := []string{"", "Has Capital", "has space", "special!"}
	for _, v := range invalid {
		if _, err := ValidateSlug(v, "slug"); err == nil {
			t.Errorf("ValidateSlug(%q) expected error, got nil", v)
		}
	}
}

func TestValidateTextContent_Sinaker(t *testing.T) {
	out, err := ValidateTextContent("Pelatihan Komputer Dasar", 200, "nama")
	if err != nil || out == "" {
		t.Errorf("ValidateTextContent valid: err=%v out=%q", err, out)
	}
	_, err = ValidateTextContent("", 200, "nama")
	if err == nil {
		t.Error("ValidateTextContent empty expected error")
	}
}

func TestValidatePaginationParams_Sinaker(t *testing.T) {
	page, limit, offset := ValidatePaginationParams("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ValidatePaginationParams("2", "15")
	if page != 2 || limit != 15 || offset != 15 {
		t.Errorf("page=2,limit=15: page=%d limit=%d offset=%d", page, limit, offset)
	}
	_, limit, _ = ValidatePaginationParams("1", "999")
	if limit != 100 {
		t.Errorf("limit clamped to 100, got %d", limit)
	}
}

func TestValidateURL_Sinaker(t *testing.T) {
	if _, err := ValidateURL("https://sinaker.jatim.go.id", "url"); err != nil {
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

func BenchmarkValidateNIK_Sinaker(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateNIK("3578010101900001", "nik")
	}
}

func BenchmarkValidatePaginationParams_Sinaker(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams("2", "15")
	}
}
