package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-unit-tests"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := GenerateAccessToken("user-123", "user", testSecret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Role != "user" {
		t.Errorf("Role = %q, want %q", claims.Role, "user")
	}
	if claims.Type != TokenTypeAccess {
		t.Errorf("Type = %q, want %q", claims.Type, TokenTypeAccess)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken("user-456", "admin", testSecret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-456")
	}
	if claims.Type != TokenTypeRefresh {
		t.Errorf("Type = %q, want %q", claims.Type, TokenTypeRefresh)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, _ := GenerateAccessToken("u1", "user", testSecret, time.Minute)
	_, err := ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("ValidateToken with wrong secret expected error, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	claims := Claims{
		UserID: "u1",
		Role:   "user",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "u1",
		},
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to create expired test token: %v", err)
	}
	_, err = ValidateToken(tokenStr, testSecret)
	if err == nil {
		t.Error("ValidateToken expired token expected error, got nil")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	_, err := ValidateToken("not.a.valid.jwt", testSecret)
	if err == nil {
		t.Error("ValidateToken malformed expected error, got nil")
	}
	_, err = ValidateToken("", testSecret)
	if err == nil {
		t.Error("ValidateToken empty token expected error, got nil")
	}
}

func TestHashToken(t *testing.T) {
	h1 := HashToken("mytoken")
	h2 := HashToken("mytoken")
	if h1 != h2 {
		t.Error("HashToken is not deterministic")
	}
	if HashToken("mytoken") == HashToken("othertoken") {
		t.Error("different tokens should produce different hashes")
	}
	if len(h1) != 64 {
		t.Errorf("HashToken should produce 64-char hex string, got %d", len(h1))
	}
}

func TestGenerateAccessToken_DefaultTTL(t *testing.T) {
	token, err := GenerateAccessToken("u1", "user", testSecret, 0)
	if err != nil {
		t.Fatalf("GenerateAccessToken with 0 TTL failed: %v", err)
	}
	claims, _ := ValidateToken(token, testSecret)
	if claims == nil {
		t.Fatal("claims is nil")
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 14*time.Minute {
		t.Errorf("default TTL should be ~15min, remaining=%v", remaining)
	}
}
