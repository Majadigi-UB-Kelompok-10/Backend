package services_list

import (
	"context"
	"strconv"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type ServiceHandler struct {
	Queries *db.Queries
}

func NewServiceHandler(q *db.Queries) *ServiceHandler {
	return &ServiceHandler{Queries: q}
}

func parseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
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
// 01. List Services (paginated)
// ---------------------------------------------------------------------

func (h *ServiceHandler) ListServices(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListServices(ctx, db.ListServicesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat daftar layanan", Detail: err.Error()})
	}

	return c.JSON(handlers.SuccessResponse{
		Pesan: "Sukses",
		Data:  data,
		Pagination: handlers.PaginationMeta{
			Page:  page,
			Limit: limit,
		},
	})
}

// ---------------------------------------------------------------------
// 02. Get Service By ID
// ---------------------------------------------------------------------

func (h *ServiceHandler) GetServiceById(c fiber.Ctx) error {
	id, err := parseUUID(c.Params("id"))
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
// 03. Get Service By Title
// ---------------------------------------------------------------------

func (h *ServiceHandler) GetServiceByTitle(c fiber.Ctx) error {
	title := c.Params("title")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetServiceByTitle(ctx, title)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Layanan tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 04. Get All Normalized Service-Category
// ---------------------------------------------------------------------

func (h *ServiceHandler) GetAllNormalized(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllNormalizedServiceCategory(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat data layanan ternormalisasi", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 05. Get Normalized Service-Category By ID
// ---------------------------------------------------------------------

func (h *ServiceHandler) GetNormalizedById(c fiber.Ctx) error {
	id, err := parseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetNormalizedServiceCategoryById(ctx, id)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Layanan tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 06. Create Service
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
	id, err := parseUUID(c.Params("id"))
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
	id, err := parseUUID(c.Params("id"))
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
