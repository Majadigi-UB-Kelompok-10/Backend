package endpoint_list

import (
	"context"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
)

const ContextQueryTimeout = 5 * time.Second

type EndpointHandler struct {
	Queries *db.Queries
}

func NewEndpointHandler(q *db.Queries) *EndpointHandler {
	return &EndpointHandler{Queries: q}
}

// ---------------------------------------------------------------------
// 01. Get Endpoint List
// ---------------------------------------------------------------------
func (h *EndpointHandler) GetEndpointList(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListEndpoints(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat endpoint list", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
