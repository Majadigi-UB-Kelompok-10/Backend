package middleware_test

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/middleware"
)

const testSecret = "test-secret-for-gateway-middleware"

func init() {
	os.Setenv("JWT_SECRET", testSecret)
}

func makeToken(role, tokenType string, ttl time.Duration) string {
	claims := jwt.MapClaims{
		"uid":  "user-123",
		"role": role,
		"type": tokenType,
		"exp":  time.Now().Add(ttl).Unix(),
		"iat":  time.Now().Unix(),
		"sub":  "user-123",
	}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	return t
}

func newApp() *fiber.App {
	app := fiber.New()
	app.Get("/protected", middleware.RequireAuth(), func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})
	app.Get("/admin-only", middleware.RequireAuth(), middleware.RequireAdmin, func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})
	return app
}

func TestRequireAuth_NoHeader(t *testing.T) {
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("no header: want 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_MalformedHeader(t *testing.T) {
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("malformed header: want 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt")
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("invalid token: want 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	claims := jwt.MapClaims{
		"uid": "u1", "role": "user", "type": "access",
		"exp": time.Now().Add(time.Minute).Unix(),
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("wrong secret: want 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	tok := makeToken("user", "access", -time.Hour)
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("expired: want 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_RefreshTokenRejected(t *testing.T) {
	tok := makeToken("user", "refresh", time.Minute)
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("refresh token as access: want 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	tok := makeToken("user", "access", time.Minute)
	app := newApp()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("valid user token: want 200, got %d", resp.StatusCode)
	}
}

func TestRequireAdmin_UserRoleBlocked(t *testing.T) {
	tok := makeToken("user", "access", time.Minute)
	app := newApp()
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 403 {
		t.Errorf("user role on admin route: want 403, got %d", resp.StatusCode)
	}
}

func TestRequireAdmin_AdminRoleAllowed(t *testing.T) {
	tok := makeToken("admin", "access", time.Minute)
	app := newApp()
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("admin role: want 200, got %d", resp.StatusCode)
	}
}

func TestRequireAdmin_SuperadminAllowed(t *testing.T) {
	tok := makeToken("superadmin", "access", time.Minute)
	app := newApp()
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("superadmin role: want 200, got %d", resp.StatusCode)
	}
}

func BenchmarkRequireAuth(b *testing.B) {
	tok := makeToken("admin", "access", time.Minute)
	app := newApp()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		app.Test(req)
	}
}
