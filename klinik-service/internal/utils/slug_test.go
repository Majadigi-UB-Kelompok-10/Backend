package utils

import (
	"strings"
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantPrefix string 
	}{
		{"Judul Biasa", "Awas Hoaks Air Garam", "awas-hoaks-air-garam"},
		{"Judul dengan Spasi Lebar dan Simbol", "  CEK FAKTA !!! Bahaya...   ", "cek-fakta-bahaya"},
		{"Judul Kepanjangan", "Satu Dua Tiga Empat Lima Enam Tujuh Delapan Sembilan Sepuluh Sebelas Dua Belas", "satu-dua-tiga-empat-lima-enam-tujuh-delapan-sembilan-sepuluh"}, // Terpotong max 60 karakter
		{"Judul Simbol Saja", "!!! ??? @@@", "berita"}, // Fallback karena semua simbol dihapus
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.input)

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("GenerateSlug() = %v, want prefix %v", got, tt.wantPrefix)
			}

			if len(got) != len(tt.wantPrefix)+5 {
				t.Errorf("GenerateSlug() = %v, panjang tidak sesuai (harus ada -xxxx di akhir)", got)
			}
		})
	}
}