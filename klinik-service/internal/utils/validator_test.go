package utils

import (
	"testing"
)

func TestValidateQueryString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantErr   bool
		expected  string
	}{
		{"Normal Input", "Pencarian Valid", 50, false, "Pencarian Valid"},
		{"Kosong", "", 50, false, ""},
		{"Melebihi Max Length", "Ini adalah teks yang sangat panjang melebihi batas dua puluh karakter", 20, true, ""},
		{"Hanya Karakter Berbahaya", "<script>!@#</script>", 50, true, ""}, 
		{"Campur Karakter Valid dan Berbahaya", "Valid!@#Data", 50, false, "ValidData"}, 
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateQueryString(tt.input, tt.maxLength, "test_field")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueryString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ValidateQueryString() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidateTextContent(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantErr   bool
		expected  string
	}{
		{"Normal", "Isi konten deskripsi", 50, false, "Isi konten deskripsi"},
		{"Kosong", "   ", 50, true, ""},
		{"Terlalu Panjang", "A B C D E F", 5, true, ""},
		{"HTML Escape", "A < B", 50, false, "A &lt; B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateTextContent(tt.input, tt.maxLength, "test_field")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTextContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ValidateTextContent() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected string
	}{
		{"Valid Email", "dzaky@gmail.com", false, "dzaky@gmail.com"},
		{"Valid Email dengan Simbol", "dzaky.filkom+test@ub.ac.id", false, "dzaky.filkom+test@ub.ac.id"},
		{"Invalid Tanpa Domain", "dzaky@", true, ""},
		{"Invalid Tanpa TLD", "dzaky@gmail", true, ""},
		{"Kosong", "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ValidateEmail() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected string
	}{
		{"Awalan 08", "081234567890", false, "081234567890"},
		{"Awalan 62", "6281234567890", false, "6281234567890"},
		{"Awalan +62", "+6281234567890", false, "+6281234567890"},
		{"Invalid Huruf", "08123ABCD", true, ""},
		{"Invalid Terlalu Pendek", "0812", true, ""},
		{"Invalid Awalan Salah", "1234567890", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePhone(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePhone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ValidatePhone() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected string
	}{
		{"Valid HTTP", "http://google.com", false, "http://google.com"},
		{"Valid HTTPS", "https://ub.ac.id/path?q=1", false, "https://ub.ac.id/path?q=1"},
		{"Boleh Kosong", "", false, ""},
		{"Invalid Scheme FTP", "ftp://file.com", true, ""},
		{"Invalid Teks Biasa", "bukan-link", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateURL(tt.input, "test_link")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ValidateURL() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidateImageFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal", "gambar.jpg", "gambar.jpg"},
		{"Path Injection (Hack Attempt)", "../../../etc/passwd", "passwd"},
		{"Hanya Simbol", "!@#$%.png", ".png"},
		{"Kosong Total", "", "image_file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateImageFilename(tt.input)
			if got != tt.expected {
				t.Errorf("ValidateImageFilename() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name          string
		pageStr       string
		limitStr      string
		expectedPage  int
		expectedLimit int
	}{
		{"Normal Input", "2", "20", 2, 20},
		{"Input Kosong (Fallback Default 1, 10)", "", "", 1, 10},
		{"Input Huruf (Fallback Default 1, 10)", "abc", "xyz", 1, 10},
		{"Limit Melebihi Maksimal (Dipaksa 50)", "1", "1000", 1, 50}, 
		{"Limit Mendekati Maksimal tapi Valid", "1", "99", 1, 50},    
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotLimit := ValidatePaginationParams(tt.pageStr, tt.limitStr)
			if gotPage != tt.expectedPage {
				t.Errorf("ValidatePaginationParams() gotPage = %v, expected %v", gotPage, tt.expectedPage)
			}
			if gotLimit != tt.expectedLimit {
				t.Errorf("ValidatePaginationParams() gotLimit = %v, expected %v", gotLimit, tt.expectedLimit)
			}
		})
	}
}