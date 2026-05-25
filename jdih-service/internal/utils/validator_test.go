package utils

import (
	"testing"
)

func TestValidateQueryString_JDIH(t *testing.T) {
	out, err := ValidateQueryString("peraturan daerah", 100, "q")
	if err != nil || out == "" {
		t.Errorf("ValidateQueryString valid input: err=%v out=%q", err, out)
	}
	_, err = ValidateQueryString("x", 0, "q") // maxLength 0 -> over limit
	if err == nil {
		t.Error("ValidateQueryString over max expected error")
	}
	out, err = ValidateQueryString("", 100, "q")
	if err != nil || out != "" {
		t.Error("ValidateQueryString empty should return empty no error")
	}
}

func TestValidateTahun(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2026", false},
		{"1901", false},
		{"2200", false},
		{"1900", true},
		{"2201", true},
		{"abc", true},
		{"", true},
		{"0", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ValidateTahun(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateTahun(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateTahun(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestValidateJenisDokumen(t *testing.T) {
	valid := []string{"perda", "pergub", "peraturan", "perdes", "sk_gub", "instruksi", "se", "keputusan", "PERDA"}
	for _, v := range valid {
		if _, err := ValidateJenisDokumen(v); err != nil {
			t.Errorf("ValidateJenisDokumen(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "pp", "undang_undang", "random"}
	for _, v := range invalid {
		if _, err := ValidateJenisDokumen(v); err == nil {
			t.Errorf("ValidateJenisDokumen(%q) expected error, got nil", v)
		}
	}
}

func TestValidateStatusDokumen(t *testing.T) {
	valid := []string{"berlaku", "diubah", "dicabut", "BERLAKU"}
	for _, v := range valid {
		if _, err := ValidateStatusDokumen(v); err != nil {
			t.Errorf("ValidateStatusDokumen(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "aktif", "expired", "hapus"}
	for _, v := range invalid {
		if _, err := ValidateStatusDokumen(v); err == nil {
			t.Errorf("ValidateStatusDokumen(%q) expected error, got nil", v)
		}
	}
}

func TestValidatePDFFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"peraturan.pdf", "peraturan.pdf"},
		{"dokumen hukum", "dokumenhukum.pdf"},
		{"../../etc/passwd", "passwd.pdf"},
		{"file.PDF", "file.PDF"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ValidatePDFFilename(tt.input)
			if got != tt.want {
				t.Errorf("ValidatePDFFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidatePaginationParams_JDIH(t *testing.T) {
	page, limit, offset := ValidatePaginationParams("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: got page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ValidatePaginationParams("2", "25")
	if page != 2 || limit != 25 || offset != 25 {
		t.Errorf("page=2,limit=25: got page=%d limit=%d offset=%d", page, limit, offset)
	}
}

func TestValidateTextContent_JDIH(t *testing.T) {
	out, err := ValidateTextContent("Peraturan Daerah Provinsi", 200, "judul")
	if err != nil || out == "" {
		t.Errorf("ValidateTextContent valid: err=%v out=%q", err, out)
	}
	_, err = ValidateTextContent("", 200, "judul")
	if err == nil {
		t.Error("ValidateTextContent empty expected error")
	}
	_, err = ValidateTextContent("x", 0, "judul")
	if err == nil {
		t.Error("ValidateTextContent over max expected error")
	}
}

func TestValidateURL_JDIH(t *testing.T) {
	if _, err := ValidateURL("https://jdih.jatimprov.go.id", "url"); err != nil {
		t.Errorf("ValidateURL valid: %v", err)
	}
	out, err := ValidateURL("", "url")
	if err != nil || out != "" {
		t.Error("ValidateURL empty should be allowed")
	}
	invalid := []string{"example.com", "ftp://x.com", "javascript:alert(1)"}
	for _, v := range invalid {
		if _, err := ValidateURL(v, "url"); err == nil {
			t.Errorf("ValidateURL(%q) expected error, got nil", v)
		}
	}
}

func BenchmarkValidateTahun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateTahun("2026")
	}
}

func BenchmarkValidatePaginationParams_JDIH(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams("2", "25")
	}
}
