package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type result struct {
	name   string
	pass   bool
	detail string
}

var (
	baseURL  string
	email    string
	password string
	client   = &http.Client{Timeout: 10 * time.Second}
	results  []result
)

func main() {
	flag.StringVar(&baseURL, "url", "http://localhost:8888", "base URL of api-gateway")
	flag.StringVar(&email, "email", "admin@example.com", "admin email")
	flag.StringVar(&password, "password", "admin123456", "admin password")
	flag.Parse()

	fmt.Printf("E2E Test — %s\n\n", baseURL)

	accessToken := ""

	step("GET /health returns 200", func() (bool, string) {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200, fmt.Sprintf("status=%d", resp.StatusCode)
	})

	step("GET /api/v1/categories returns 200 (public)", func() (bool, string) {
		resp, err := client.Get(baseURL + "/api/v1/categories")
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200, fmt.Sprintf("status=%d", resp.StatusCode)
	})

	step("POST /api/v1/admin/categories without token returns 401", func() (bool, string) {
		body := bytes.NewBufferString(`{"name":"test","slug":"test","icon":"icon"}`)
		resp, err := client.Post(baseURL+"/api/v1/admin/categories", "application/json", body)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 401, fmt.Sprintf("status=%d (want 401)", resp.StatusCode)
	})

	step("POST /api/v1/auth/login with wrong password returns 401", func() (bool, string) {
		body := jsonBody(map[string]string{"email": email, "password": "wrongpassword"})
		resp, err := client.Post(baseURL+"/api/v1/auth/login", "application/json", body)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 401, fmt.Sprintf("status=%d (want 401)", resp.StatusCode)
	})

	step("POST /api/v1/auth/login with valid credentials returns 200", func() (bool, string) {
		body := jsonBody(map[string]string{"email": email, "password": password})
		resp, err := client.Post(baseURL+"/api/v1/auth/login", "application/json", body)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			return false, fmt.Sprintf("status=%d body=%s", resp.StatusCode, string(b))
		}
		var res map[string]any
		json.NewDecoder(resp.Body).Decode(&res)
		data, _ := res["data"].(map[string]any)
		tokens, _ := data["tokens"].(map[string]any)
		accessToken, _ = tokens["access_token"].(string)
		if accessToken == "" {
			return false, "access_token not found in response"
		}
		return true, "access_token received"
	})

	step("GET /api/v1/auth/me with token returns 200", func() (bool, string) {
		if accessToken == "" {
			return false, "skipped: no access token"
		}
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200, fmt.Sprintf("status=%d", resp.StatusCode)
	})

	step("GET /api/v1/auth/me without token returns 401", func() (bool, string) {
		resp, err := client.Get(baseURL + "/api/v1/auth/me")
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 401, fmt.Sprintf("status=%d (want 401)", resp.StatusCode)
	})

	step("Dynamic admin route without token returns 401", func() (bool, string) {
		body := bytes.NewBufferString(`{}`)
		resp, err := client.Post(baseURL+"/api/v1/sinaker-api/admin/blk", "application/json", body)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode == 401, fmt.Sprintf("status=%d (want 401)", resp.StatusCode)
	})

	printSummary()
}

func step(name string, fn func() (bool, string)) {
	pass, detail := fn()
	results = append(results, result{name: name, pass: pass, detail: detail})
	status := "PASS"
	if !pass {
		status = "FAIL"
	}
	fmt.Printf("  [%s] %s — %s\n", status, name, detail)
}

func jsonBody(v any) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func printSummary() {
	pass, fail := 0, 0
	for _, r := range results {
		if r.pass {
			pass++
		} else {
			fail++
		}
	}
	fmt.Printf("\n========================================\n")
	fmt.Printf("E2E HASIL: %d/%d lulus\n", pass, pass+fail)
	fmt.Printf("========================================\n")
	if fail > 0 {
		fmt.Println("\nGagal:")
		for _, r := range results {
			if !r.pass {
				fmt.Printf("  - %s (%s)\n", r.name, r.detail)
			}
		}
		os.Exit(1)
	}
}
