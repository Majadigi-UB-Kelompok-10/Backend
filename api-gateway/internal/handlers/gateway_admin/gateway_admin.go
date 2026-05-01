package gateway_admin

import (
	"context"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/routes"
	"github.com/gofiber/fiber/v3"
)

type GatewayAdminHandler struct {
	Queries *db.Queries
}

func NewGatewayAdminHandler(q *db.Queries) *GatewayAdminHandler {
	return &GatewayAdminHandler{Queries: q}
}

func (h *GatewayAdminHandler) SyncNow(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := routes.RefreshEndpoints(ctx, h.Queries); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal sinkron",
			"pesan": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Gateway berhasil disinkronkan secara manual!",
	})
}
