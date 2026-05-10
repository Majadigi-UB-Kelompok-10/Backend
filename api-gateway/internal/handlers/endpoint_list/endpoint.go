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
// Request DTOs
// ---------------------------------------------------------------------

type CreateEndpointRequest struct {
	SlugName string `json:"slug_name"`
	PageUrl  string `json:"page_url"`
}

type UpdateEndpointRequest struct {
	SlugName string `json:"slug_name"`
	PageUrl  string `json:"page_url"`
}

// ---------------------------------------------------------------------
// 01. List Endpoints
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

// ---------------------------------------------------------------------
// 02. Get Endpoint By ID
// ---------------------------------------------------------------------

func (h *EndpointHandler) GetEndpointById(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetEndpointById(ctx, id)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Endpoint tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 03. Get Endpoint By Slug
// ---------------------------------------------------------------------

func (h *EndpointHandler) GetEndpointBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetEndpointBySlug(ctx, slug)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Endpoint tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 04. Create Endpoint
// ---------------------------------------------------------------------

func (h *EndpointHandler) CreateEndpoint(c fiber.Ctx) error {
	var req CreateEndpointRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.SlugName == "" || req.PageUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "slug_name dan page_url wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.CreateEndpoint(ctx, db.CreateEndpointParams{
		SlugName: req.SlugName,
		PageUrl:  req.PageUrl,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat endpoint", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Endpoint berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 05. Update Endpoint
// ---------------------------------------------------------------------

func (h *EndpointHandler) UpdateEndpoint(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdateEndpointRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.SlugName == "" || req.PageUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "slug_name dan page_url wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.UpdateEndpoint(ctx, db.UpdateEndpointParams{
		EndpointListID: id,
		SlugName:       req.SlugName,
		PageUrl:        req.PageUrl,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui endpoint", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Endpoint berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 06. Delete Endpoint
// ---------------------------------------------------------------------

func (h *EndpointHandler) DeleteEndpoint(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteEndpoint(ctx, id); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus endpoint", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Endpoint berhasil dihapus"})
}

// ---------------------------------------------------------------------
// 07. Get All Endpoints
// ---------------------------------------------------------------------

func (h *EndpointHandler) GetAllEndpoints(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllEndpoints(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua endpoint", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
