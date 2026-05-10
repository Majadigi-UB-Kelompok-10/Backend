package category_list

import (
	"context"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type CategoryHandler struct {
	Queries *db.Queries
}

func NewCategoryHandler(q *db.Queries) *CategoryHandler {
	return &CategoryHandler{Queries: q}
}


// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ---------------------------------------------------------------------
// 01. List Categories
// ---------------------------------------------------------------------

func (h *CategoryHandler) ListCategories(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListCategories(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat daftar kategori", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. Get Category By ID
// ---------------------------------------------------------------------

func (h *CategoryHandler) GetCategoryById(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetCategoryById(ctx, id)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Kategori tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 03. Create Category
// ---------------------------------------------------------------------

func (h *CategoryHandler) CreateCategory(c fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.Name == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "name wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	desc := pgtype.Text{}
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	data, err := h.Queries.CreateCategory(ctx, db.CreateCategoryParams{
		Name:        req.Name,
		Description: desc,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat kategori", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Kategori berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 05. Update Category
// ---------------------------------------------------------------------

func (h *CategoryHandler) UpdateCategory(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.Name == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "name wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	desc := pgtype.Text{}
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	data, err := h.Queries.UpdateCategory(ctx, db.UpdateCategoryParams{
		CategoryListID: id,
		Name:           req.Name,
		Description:    desc,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui kategori", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Kategori berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 06. Delete Category
// ---------------------------------------------------------------------

func (h *CategoryHandler) DeleteCategory(c fiber.Ctx) error {
	id, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteCategory(ctx, id); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus kategori", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Kategori berhasil dihapus"})
}

