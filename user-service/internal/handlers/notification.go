package handlers

import (
	"context"
	"log/slog"

	"github.com/farildzaky/user-service/internal/fcm"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------------
// POST /api/v1/internal/notify-by-service
// Header: X-Internal-Key
// Body: {"service_slug":"siskaperbapo","title":"...","body":"...","data":{}}
// Sends FCM only to users who favorited the given service slug.
// ---------------------------------------------------------------------------

type NotificationHandler struct {
	Pool          *pgxpool.Pool
	InternalKey   string
	GatewayDBURL  string
}

func NewNotificationHandler(pool *pgxpool.Pool, internalKey, gatewayDBURL string) *NotificationHandler {
	return &NotificationHandler{Pool: pool, InternalKey: internalKey, GatewayDBURL: gatewayDBURL}
}

// ---------------------------------------------------------------------------
// POST /api/v1/auth/device-token
// Body: {"token": "fcm-device-token", "platform": "android"|"ios"}
// ---------------------------------------------------------------------------

type RegisterTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func (h *NotificationHandler) RegisterDeviceToken(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req RegisterTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Request tidak valid"})
	}
	if req.Token == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "token wajib diisi"})
	}
	if req.Platform != "android" && req.Platform != "ios" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "platform harus 'android' atau 'ios'"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Pool.Exec(ctx, `
		INSERT INTO device_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, platform)
		DO UPDATE SET token = EXCLUDED.token, updated_at = NOW()
	`, userID, req.Token, req.Platform)
	if err != nil {
		slog.Error("device_token.register.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menyimpan device token"})
	}

	return c.JSON(SuccessResponse{Pesan: "Device token berhasil disimpan"})
}

// ---------------------------------------------------------------------------
// DELETE /api/v1/auth/device-token
// Body: {"platform": "android"|"ios"}
// ---------------------------------------------------------------------------

func (h *NotificationHandler) DeleteDeviceToken(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req struct {
		Platform string `json:"platform"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Request tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	query := `DELETE FROM device_tokens WHERE user_id = $1`
	args := []any{userID}
	if req.Platform == "android" || req.Platform == "ios" {
		query += ` AND platform = $2`
		args = append(args, req.Platform)
	}

	if _, err := h.Pool.Exec(ctx, query, args...); err != nil {
		slog.Error("device_token.delete.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menghapus device token"})
	}

	return c.JSON(SuccessResponse{Pesan: "Device token berhasil dihapus"})
}

// ---------------------------------------------------------------------------
// POST /api/v1/internal/notify
// Header: X-Internal-Key: <INTERNAL_SERVICE_KEY>
// Body: {"user_id":"uuid","title":"...","body":"...","data":{"key":"val"}}
// ---------------------------------------------------------------------------

type NotifyRequest struct {
	UserID string            `json:"user_id"`
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Data   map[string]string `json:"data"`
}

func (h *NotificationHandler) SendNotification(c fiber.Ctx) error {
	if h.InternalKey != "" && c.Get("X-Internal-Key") != h.InternalKey {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Internal key tidak valid"})
	}

	var req NotifyRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Request tidak valid"})
	}
	if req.UserID == "" || req.Title == "" || req.Body == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "user_id, title, dan body wajib diisi"})
	}

	if !fcm.Enabled() {
		return c.Status(503).JSON(ErrorResponse{Code: "ERR_FCM_DISABLED", Message: "FCM tidak aktif"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	rows, err := h.Pool.Query(ctx,
		`SELECT token FROM device_tokens WHERE user_id = $1`, req.UserID,
	)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal mengambil device token"})
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err == nil {
			tokens = append(tokens, token)
		}
	}

	if len(tokens) == 0 {
		return c.JSON(SuccessResponse{Pesan: "Tidak ada device token terdaftar untuk user ini"})
	}

	if err := fcm.SendToTokens(ctx, tokens, req.Title, req.Body, req.Data); err != nil {
		slog.Error("fcm.send.error", slog.String("user_id", req.UserID), slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_FCM", Message: "Gagal mengirim notifikasi"})
	}

	slog.Info("fcm.sent", slog.String("user_id", req.UserID), slog.Int("tokens", len(tokens)))
	return c.JSON(SuccessResponse{Pesan: "Notifikasi berhasil dikirim"})
}

// ---------------------------------------------------------------------------
// POST /api/v1/internal/notify-all
// Header: X-Internal-Key: <INTERNAL_SERVICE_KEY>
// Body: {"title":"...","body":"...","data":{"key":"val"}}
// ---------------------------------------------------------------------------

type NotifyAllRequest struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

func (h *NotificationHandler) SendNotificationAll(c fiber.Ctx) error {
	if h.InternalKey != "" && c.Get("X-Internal-Key") != h.InternalKey {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Internal key tidak valid"})
	}

	var req NotifyAllRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Request tidak valid"})
	}
	if req.Title == "" || req.Body == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "title dan body wajib diisi"})
	}

	if !fcm.Enabled() {
		return c.Status(503).JSON(ErrorResponse{Code: "ERR_FCM_DISABLED", Message: "FCM tidak aktif"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	rows, err := h.Pool.Query(ctx, `SELECT DISTINCT token FROM device_tokens`)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal mengambil device tokens"})
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err == nil {
			tokens = append(tokens, token)
		}
	}

	if len(tokens) == 0 {
		return c.JSON(SuccessResponse{Pesan: "Tidak ada device token terdaftar"})
	}

	if err := fcm.SendToTokens(ctx, tokens, req.Title, req.Body, req.Data); err != nil {
		slog.Error("fcm.broadcast.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_FCM", Message: "Gagal mengirim notifikasi broadcast"})
	}

	slog.Info("fcm.broadcast.sent", slog.Int("tokens", len(tokens)))
	return c.JSON(SuccessResponse{Pesan: "Notifikasi broadcast berhasil dikirim"})
}

type NotifyByServiceRequest struct {
	ServiceSlug string            `json:"service_slug"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data"`
}

func (h *NotificationHandler) NotifyByService(c fiber.Ctx) error {
	if h.InternalKey != "" && c.Get("X-Internal-Key") != h.InternalKey {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Internal key tidak valid"})
	}

	var req NotifyByServiceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Request tidak valid"})
	}
	if req.ServiceSlug == "" || req.Title == "" || req.Body == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "service_slug, title, dan body wajib diisi"})
	}

	if !fcm.Enabled() {
		return c.Status(503).JSON(ErrorResponse{Code: "ERR_FCM_DISABLED", Message: "FCM tidak aktif"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	// Lookup service_list_id by slug from gateway DB
	gatewayPool, err := pgxpool.New(ctx, h.GatewayDBURL)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal terhubung ke gateway DB"})
	}
	defer gatewayPool.Close()

	var serviceID string
	err = gatewayPool.QueryRow(ctx, `
		SELECT il.service_list_id::text
		FROM integration_list il
		JOIN endpoint_list el ON il.endpoint_list_id = el.endpoint_list_id
		WHERE el.slug_name = $1
		LIMIT 1
	`, req.ServiceSlug).Scan(&serviceID)
	if err != nil {
		slog.Warn("notify.by_service.slug_not_found", slog.String("slug", req.ServiceSlug))
		return c.JSON(SuccessResponse{Pesan: "Service tidak ditemukan, notifikasi dilewati"})
	}

	// Get device tokens for users who favorited this service
	rows, err := h.Pool.Query(ctx, `
		SELECT DISTINCT dt.token
		FROM device_tokens dt
		JOIN user_service_favorites usf ON dt.user_id = usf.user_id
		WHERE usf.service_id = $1
	`, serviceID)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal mengambil device tokens"})
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err == nil {
			tokens = append(tokens, token)
		}
	}

	if len(tokens) == 0 {
		return c.JSON(SuccessResponse{Pesan: "Tidak ada pengguna yang memfavoritkan service ini"})
	}

	if err := fcm.SendToTokens(ctx, tokens, req.Title, req.Body, req.Data); err != nil {
		slog.Error("fcm.by_service.error", slog.String("slug", req.ServiceSlug), slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_FCM", Message: "Gagal mengirim notifikasi"})
	}

	slog.Info("fcm.by_service.sent", slog.String("slug", req.ServiceSlug), slog.Int("tokens", len(tokens)))
	return c.JSON(SuccessResponse{Pesan: "Notifikasi berhasil dikirim ke pengguna service ini"})
}
