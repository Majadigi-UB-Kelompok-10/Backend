package utils

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	plain := "mySecret123"

	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == plain {
		t.Fatal("hash should not equal plain text")
	}

	if !VerifyPassword(plain, hash) {
		t.Error("VerifyPassword(correct) = false, want true")
	}
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword(wrong) = true, want false")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	h1, _ := HashPassword("password")
	h2, _ := HashPassword("password")
	if h1 == h2 {
		t.Error("same password should produce different hashes due to bcrypt salt")
	}
}

func TestVerifyPassword_EmptyInputs(t *testing.T) {
	if VerifyPassword("", "$2a$12$invalididinvalidid") {
		t.Error("VerifyPassword empty password should return false")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword("benchmarkpassword")
	}
}
