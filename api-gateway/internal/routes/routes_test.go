package routes_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/routes"
)

const routeSecret = "test-secret-for-dynamic-routes"

func init() {
	os.Setenv("JWT_SECRET", routeSecret)
}

func routeToken(role string) string {
	claims := jwt.MapClaims{
		"uid":  "u1",
		"role": role,
		"type": "access",
		"exp":  time.Now().Add(time.Minute).Unix(),
		"sub":  "u1",
	}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(routeSecret))
	return t
}

func echo200() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
}

func appWith(slug, base string) *fiber.App {
	routes.GatewayRegistry.Lock()
	routes.GatewayRegistry.Routes[slug] = base
	routes.GatewayRegistry.Unlock()
	app := fiber.New()
	routes.SetupDynamicEndpoint(app)
	return app
}

func TestRoutes_UnknownSlug_404(t *testing.T) {
	routes.GatewayRegistry.Lock()
	delete(routes.GatewayRegistry.Routes, "no-such")
	routes.GatewayRegistry.Unlock()

	app := fiber.New()
	routes.SetupDynamicEndpoint(app)

	req := httptest.NewRequest("GET", "/api/v1/no-such/public/x", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 404 {
		t.Errorf("unknown slug: want 404, got %d", resp.StatusCode)
	}
}

func TestRoutes_Admin_NoToken_401(t *testing.T) {
	ds := echo200()
	defer ds.Close()
	app := appWith("r-a", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-a/admin/items", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("admin no token: want 401, got %d", resp.StatusCode)
	}
}

func TestRoutes_Admin_UserRole_403(t *testing.T) {
	ds := echo200()
	defer ds.Close()
	app := appWith("r-b", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-b/admin/items", nil)
	req.Header.Set("Authorization", "Bearer "+routeToken("user"))
	resp, _ := app.Test(req)
	if resp.StatusCode != 403 {
		t.Errorf("user role on admin: want 403, got %d", resp.StatusCode)
	}
}

func TestRoutes_Admin_AdminToken_200(t *testing.T) {
	ds := echo200()
	defer ds.Close()
	app := appWith("r-c", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-c/admin/items", nil)
	req.Header.Set("Authorization", "Bearer "+routeToken("admin"))
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("admin token: want 200, got %d", resp.StatusCode)
	}
}

func TestRoutes_Admin_SuperadminToken_200(t *testing.T) {
	ds := echo200()
	defer ds.Close()
	app := appWith("r-d", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-d/admin/items", nil)
	req.Header.Set("Authorization", "Bearer "+routeToken("superadmin"))
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("superadmin token: want 200, got %d", resp.StatusCode)
	}
}

func TestRoutes_Public_NoToken_200(t *testing.T) {
	ds := echo200()
	defer ds.Close()
	app := appWith("r-e", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-e/public/items", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("public no token: want 200, got %d", resp.StatusCode)
	}
}

func TestRoutes_QueryString_Forwarded(t *testing.T) {
	received := ""
	ds := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.URL.RawQuery
		w.WriteHeader(200)
	}))
	defer ds.Close()
	app := appWith("r-f", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-f/public/items?page=2&limit=10", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("query forward: want 200, got %d", resp.StatusCode)
	}
	if received != "page=2&limit=10" {
		t.Errorf("query string not forwarded: got %q", received)
	}
}

func TestRoutes_Admin_PathPreserved(t *testing.T) {
	receivedPath := ""
	ds := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(200)
	}))
	defer ds.Close()
	app := appWith("r-g", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-g/admin/jadwal", nil)
	req.Header.Set("Authorization", "Bearer "+routeToken("admin"))
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("admin path: want 200, got %d", resp.StatusCode)
	}
	if receivedPath != "/admin/jadwal" {
		t.Errorf("admin path not preserved: got %q, want %q", receivedPath, "/admin/jadwal")
	}
}

func TestRoutes_Admin_ExactPath_401(t *testing.T) {
	ds := echo200()
	defer ds.Close()
	app := appWith("r-h", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-h/admin", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("/admin exact path no token: want 401, got %d", resp.StatusCode)
	}
}

func TestRoutes_Admin_ExactPath_PathPreserved(t *testing.T) {
	receivedPath := ""
	ds := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(200)
	}))
	defer ds.Close()
	app := appWith("r-i", ds.URL)

	req := httptest.NewRequest("GET", "/api/v1/r-i/admin", nil)
	req.Header.Set("Authorization", "Bearer "+routeToken("admin"))
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("/admin exact: want 200, got %d", resp.StatusCode)
	}
	if receivedPath != "/admin" {
		t.Errorf("exact admin path not preserved: got %q, want %q", receivedPath, "/admin")
	}
}

func TestRoutes_Registry_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			slug := fmt.Sprintf("conc-%d", n)
			routes.GatewayRegistry.Lock()
			routes.GatewayRegistry.Routes[slug] = "http://localhost:9999"
			routes.GatewayRegistry.Unlock()
		}(i)
		go func(n int) {
			defer wg.Done()
			slug := fmt.Sprintf("conc-%d", n)
			routes.GatewayRegistry.RLock()
			_ = routes.GatewayRegistry.Routes[slug]
			routes.GatewayRegistry.RUnlock()
		}(i)
	}
	wg.Wait()
}
