package integration_list

import (
	"context"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type IntegrationHandler struct {
	Queries *db.Queries
}

func NewIntegrationHandler(q *db.Queries) *IntegrationHandler {
	return &IntegrationHandler{Queries: q}
}


// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

type CreateIntegrationRequest struct {
	ServiceListID  string `json:"service_list_id"`
	EndpointListID string `json:"endpoint_list_id"`
	Title          string `json:"title"`
	IconUrl        string `json:"icon_url"`
}

type UpdateIntegrationRequest struct {
	EndpointListID string `json:"endpoint_list_id"`
	Title          string `json:"title"`
	IconUrl        string `json:"icon_url"`
}

// ---------------------------------------------------------------------
// 01. List Integrations By Service ID
// ---------------------------------------------------------------------

func (h *IntegrationHandler) ListIntegrationsByServiceId(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListIntegrationsByServiceId(ctx, serviceID)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat daftar integrasi", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. Get Integration By ID
// ---------------------------------------------------------------------

func (h *IntegrationHandler) GetIntegrationById(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetIntegrationById(ctx, id)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Integrasi tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 03. Create Integration
// ---------------------------------------------------------------------

func (h *IntegrationHandler) CreateIntegration(c fiber.Ctx) error {
	var req CreateIntegrationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.ServiceListID == "" || req.EndpointListID == "" || req.Title == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "service_list_id, endpoint_list_id, dan title wajib diisi"})
	}

	serviceID, err := handlers.ParseUUID(req.ServiceListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "service_list_id tidak valid", Detail: err.Error()})
	}
	endpointID, err := handlers.ParseUUID(req.EndpointListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "endpoint_list_id tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	iconUrl := pgtype.Text{}
	if req.IconUrl != "" {
		iconUrl = pgtype.Text{String: req.IconUrl, Valid: true}
	}

	data, err := h.Queries.CreateIntegration(ctx, db.CreateIntegrationParams{
		ServiceListID:  serviceID,
		EndpointListID: endpointID,
		Title:          req.Title,
		IconUrl:        iconUrl,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat integrasi", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Integrasi berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 04. Update Integration
// ---------------------------------------------------------------------

func (h *IntegrationHandler) UpdateIntegration(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdateIntegrationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.EndpointListID == "" || req.Title == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "endpoint_list_id dan title wajib diisi"})
	}

	endpointID, err := handlers.ParseUUID(req.EndpointListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "endpoint_list_id tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	iconUrl := pgtype.Text{}
	if req.IconUrl != "" {
		iconUrl = pgtype.Text{String: req.IconUrl, Valid: true}
	}

	data, err := h.Queries.UpdateIntegration(ctx, db.UpdateIntegrationParams{
		IntegrationListID: id,
		EndpointListID:    endpointID,
		Title:             req.Title,
		IconUrl:           iconUrl,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui integrasi", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Integrasi berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 05. Delete Integration
// ---------------------------------------------------------------------

func (h *IntegrationHandler) DeleteIntegration(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteIntegration(ctx, id); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus integrasi", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Integrasi berhasil dihapus"})
}

// ---------------------------------------------------------------------
// 06. Get All Integrations
// ---------------------------------------------------------------------

func (h *IntegrationHandler) GetAllIntegrations(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllIntegrations(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua integrasi", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
