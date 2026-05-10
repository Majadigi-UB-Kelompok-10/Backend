package services_list

import (
	"context"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type ServiceHandler struct {
	Queries *db.Queries
}

func NewServiceHandler(q *db.Queries) *ServiceHandler {
	return &ServiceHandler{Queries: q}
}


// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

type CreateServiceRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IconUrl     string `json:"icon_url"`
}

type UpdateServiceRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IconUrl     string `json:"icon_url"`
}

// ---------------------------------------------------------------------
// 01. List Services (all, with optional ?search= filter)
// ---------------------------------------------------------------------

func (h *ServiceHandler) ListServices(c fiber.Ctx) error {
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if search != "" {
		data, err := h.Queries.SearchServices(ctx, "%"+search+"%")
		if err != nil {
			return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal mencari layanan", Detail: err.Error()})
		}
		return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
	}

	data, err := h.Queries.GetAllServices(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat daftar layanan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. Get Service By ID
// ---------------------------------------------------------------------

func (h *ServiceHandler) GetServiceById(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetServiceById(ctx, id)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Layanan tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 03. Create Service
// ---------------------------------------------------------------------

func (h *ServiceHandler) CreateService(c fiber.Ctx) error {
	var req CreateServiceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.Title == "" || req.IconUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "title dan icon_url wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	desc := pgtype.Text{}
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	data, err := h.Queries.CreateService(ctx, db.CreateServiceParams{
		Title:       req.Title,
		Description: desc,
		IconUrl:     req.IconUrl,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat layanan", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Layanan berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 07. Update Service
// ---------------------------------------------------------------------

func (h *ServiceHandler) UpdateService(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdateServiceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.Title == "" || req.IconUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "title dan icon_url wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	desc := pgtype.Text{}
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	data, err := h.Queries.UpdateService(ctx, db.UpdateServiceParams{
		ServiceListID: id,
		Title:         req.Title,
		Description:   desc,
		IconUrl:       req.IconUrl,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui layanan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Layanan berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 08. Delete Service
// ---------------------------------------------------------------------

func (h *ServiceHandler) DeleteService(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteService(ctx, id); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus layanan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Layanan berhasil dihapus"})
}

// ---------------------------------------------------------------------
// 09. Get All Services (unpaginated)
// ---------------------------------------------------------------------

func (h *ServiceHandler) GetAllServices(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllServices(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua layanan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
