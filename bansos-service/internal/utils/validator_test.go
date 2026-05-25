package utils

import (
	"testing"
	"time"
)

func TestValidateNIK(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 16 digits", "3578010101900001", false},
		{"valid starts with 0", "0000000000000000", false},
		{"too short 15 digits", "357801010190000", true},
		{"too long 17 digits", "35780101019000011", true},
		{"with letters", "3578010101A00001", true},
		{"empty", "", true},
		{"with spaces", "3578 0101 0190 0001", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateNIK(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateNIK(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateNIK(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestValidateStatusPenyaluran(t *testing.T) {
	valid := []string{"proses", "diterima", "ditolak", "PROSES", "Diterima"}
	for _, v := range valid {
		if _, err := ValidateStatusPenyaluran(v); err != nil {
			t.Errorf("ValidateStatusPenyaluran(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "selesai", "pending", "rejected"}
	for _, v := range invalid {
		if _, err := ValidateStatusPenyaluran(v); err == nil {
			t.Errorf("ValidateStatusPenyaluran(%q) expected error, got nil", v)
		}
	}
}

func TestValidateMetodePenyaluran(t *testing.T) {
	valid := []string{"transfer_bri", "transfer_bni", "transfer_mandiri", "transfer_bca", "transfer_pos", "tunai", "TUNAI"}
	for _, v := range valid {
		if _, err := ValidateMetodePenyaluran(v); err != nil {
			t.Errorf("ValidateMetodePenyaluran(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "gopay", "ovo", "transfer"}
	for _, v := range invalid {
		if _, err := ValidateMetodePenyaluran(v); err == nil {
			t.Errorf("ValidateMetodePenyaluran(%q) expected error, got nil", v)
		}
	}
}

func TestValidateNominal(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"1000000", false},
		{"1", false},
		{"0", true},
		{"-100", true},
		{"abc", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ValidateNominal(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateNominal(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateNominal(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestValidateDateString(t *testing.T) {
	_, err := ValidateDateString("2026-01-15", "tanggal")
	if err != nil {
		t.Errorf("ValidateDateString valid date returned error: %v", err)
	}
	_, err = ValidateDateString("15/01/2026", "tanggal")
	if err == nil {
		t.Error("ValidateDateString wrong format expected error, got nil")
	}
	_, err = ValidateDateString("not-a-date", "tanggal")
	if err == nil {
		t.Error("ValidateDateString invalid string expected error, got nil")
	}
}

func TestValidatePeriode(t *testing.T) {
	mulai, _ := time.Parse("2006-01-02", "2026-01-01")
	selesai, _ := time.Parse("2006-01-02", "2026-12-31")

	if err := ValidatePeriode(mulai, selesai); err != nil {
		t.Errorf("ValidatePeriode valid range returned error: %v", err)
	}
	if err := ValidatePeriode(selesai, mulai); err == nil {
		t.Error("ValidatePeriode inverted range expected error, got nil")
	}
	if err := ValidatePeriode(mulai, mulai); err != nil {
		t.Errorf("ValidatePeriode same date (equal) should be valid: %v", err)
	}
}

func TestValidatePaginationParams_Bansos(t *testing.T) {
	page, limit, offset := ValidatePaginationParams("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: got page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ValidatePaginationParams("3", "20")
	if page != 3 || limit != 20 || offset != 40 {
		t.Errorf("page=3,limit=20: got page=%d limit=%d offset=%d", page, limit, offset)
	}
	_, limit, _ = ValidatePaginationParams("1", "999")
	if limit != 100 {
		t.Errorf("limit should be clamped to 100, got %d", limit)
	}
}

func TestValidateQueryString_Bansos(t *testing.T) {
	out, err := ValidateQueryString("penerima bansos", 100, "q")
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

func TestValidateTextContent_Bansos(t *testing.T) {
	out, err := ValidateTextContent("Nama Program", 100, "nama")
	if err != nil || out == "" {
		t.Errorf("ValidateTextContent valid: err=%v out=%q", err, out)
	}
	_, err = ValidateTextContent("", 100, "nama")
	if err == nil {
		t.Error("ValidateTextContent empty expected error")
	}
	_, err = ValidateTextContent("x", 0, "nama")
	if err == nil {
		t.Error("ValidateTextContent over max expected error")
	}
}

func BenchmarkValidateNIK(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateNIK("3578010101900001")
	}
}

func BenchmarkValidatePaginationParams_Bansos(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams("3", "20")
	}
}
