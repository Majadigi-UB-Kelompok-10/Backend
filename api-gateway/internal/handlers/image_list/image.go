package image_list

import (
	"context"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type ImageHandler struct {
	Queries *db.Queries
}

func NewImageHandler(q *db.Queries) *ImageHandler {
	return &ImageHandler{Queries: q}
}


// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

type CreateImageRequest struct {
	ServiceListID string `json:"service_list_id"`
	ImageUrl      string `json:"image_url"`
	SemanticLabel string `json:"semantic_label"`
}

type UpdateImageRequest struct {
	ImageUrl      string `json:"image_url"`
	SemanticLabel string `json:"semantic_label"`
}

// ---------------------------------------------------------------------
// 01. List Images By Service ID
// ---------------------------------------------------------------------

func (h *ImageHandler) ListImagesByServiceId(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListImagesByServiceId(ctx, serviceID)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat daftar gambar", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. Get Image By ID
// ---------------------------------------------------------------------

func (h *ImageHandler) GetImageById(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetImageById(ctx, id)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Gambar tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 03. Create Image
// ---------------------------------------------------------------------

func (h *ImageHandler) CreateImage(c fiber.Ctx) error {
	var req CreateImageRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.ServiceListID == "" || req.ImageUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "service_list_id dan image_url wajib diisi"})
	}

	serviceID, err := handlers.ParseUUID(req.ServiceListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "service_list_id tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	label := pgtype.Text{}
	if req.SemanticLabel != "" {
		label = pgtype.Text{String: req.SemanticLabel, Valid: true}
	}

	data, err := h.Queries.CreateImage(ctx, db.CreateImageParams{
		ServiceListID: serviceID,
		ImageUrl:      req.ImageUrl,
		SemanticLabel: label,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat gambar", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Gambar berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 04. Update Image
// ---------------------------------------------------------------------

func (h *ImageHandler) UpdateImage(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdateImageRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.ImageUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "image_url wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	label := pgtype.Text{}
	if req.SemanticLabel != "" {
		label = pgtype.Text{String: req.SemanticLabel, Valid: true}
	}

	data, err := h.Queries.UpdateImage(ctx, db.UpdateImageParams{
		ImageListID:   id,
		ImageUrl:      req.ImageUrl,
		SemanticLabel: label,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui gambar", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Gambar berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 05. Delete Image
// ---------------------------------------------------------------------

func (h *ImageHandler) DeleteImage(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteImage(ctx, id); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus gambar", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Gambar berhasil dihapus"})
}

// ---------------------------------------------------------------------
// 06. Get All Images
// ---------------------------------------------------------------------

func (h *ImageHandler) GetAllImages(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllImages(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua gambar", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
