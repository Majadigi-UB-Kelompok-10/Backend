package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	userServiceURL string
	internalKey    string
	httpClient     = &http.Client{Timeout: 5 * time.Second}
)

func Init() {
	userServiceURL = os.Getenv("USER_SERVICE_URL")
	internalKey = os.Getenv("INTERNAL_SERVICE_KEY")
}

// ByService sends FCM only to users who favorited the given service slug.
func ByService(slug, title, body string, data map[string]string) {
	if userServiceURL == "" {
		return
	}
	go func() {
		payload := map[string]any{
			"service_slug": slug,
			"title":        title,
			"body":         body,
			"data":         data,
		}
		b, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
			userServiceURL+"/api/v1/internal/notify-by-service", bytes.NewReader(b))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Key", internalKey)
		resp, err := httpClient.Do(req)
		if err != nil {
			slog.Warn("notify.by_service.error", slog.String("slug", slug), slog.String("err", err.Error()))
			return
		}
		defer resp.Body.Close()
	}()
}
