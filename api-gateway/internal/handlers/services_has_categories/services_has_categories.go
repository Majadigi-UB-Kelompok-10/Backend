package services_has_categories

import (
	"context"
	"strings"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type ServicesCategoriesHandler struct {
	Queries *db.Queries
}

func NewServicesCategoriesHandler(q *db.Queries) *ServicesCategoriesHandler {
	return &ServicesCategoriesHandler{Queries: q}
}


// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

type AddCategoryRequest struct {
	CategoryListID string `json:"category_list_id"`
}

// ---------------------------------------------------------------------
// 01. List Categories By Service ID
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) ListCategoriesByServiceId(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListCategoriesByServiceId(ctx, serviceID)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat kategori layanan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. List Services By Category ID
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) ListServicesByCategoryId(c fiber.Ctx) error {
	categoryID, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListServicesByCategoryId(ctx, categoryID)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat layanan berdasarkan kategori", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 03. Add Category To Service
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) AddCategoryToService(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID service tidak valid", Detail: err.Error()})
	}

	var req AddCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.CategoryListID == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "category_list_id wajib diisi"})
	}

	categoryID, err := handlers.ParseUUID(req.CategoryListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "category_list_id tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.AddCategoryToService(ctx, db.AddCategoryToServiceParams{
		ServiceListID:  serviceID,
		CategoryListID: categoryID,
	}); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menambahkan kategori ke layanan", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Kategori berhasil ditambahkan ke layanan"})
}

// ---------------------------------------------------------------------
// 04. Remove Category From Service
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) RemoveCategoryFromService(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID service tidak valid", Detail: err.Error()})
	}

	categoryID, err := handlers.ParseUUID(c.Params("category_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID kategori tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.RemoveCategoryFromService(ctx, db.RemoveCategoryFromServiceParams{
		ServiceListID:  serviceID,
		CategoryListID: categoryID,
	}); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus kategori dari layanan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Kategori berhasil dihapus dari layanan"})
}

// ---------------------------------------------------------------------
// 05. Get All Services Has Categories
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) GetAllServicesHasCategories(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllServicesHasCategories(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua relasi layanan-kategori", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 06. Get Personalized Services
// GET /api/v1/services/personalized?category_ids=uuid1,uuid2,uuid3
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) GetPersonalizedServices(c fiber.Ctx) error {
	raw := c.Query("category_ids")
	if raw == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "category_ids wajib diisi"})
	}

	parts := strings.Split(raw, ",")
	categoryUUIDs := make([]pgtype.UUID, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		uid, err := handlers.ParseUUID(p)
		if err != nil {
			return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: p + ": " + err.Error()})
		}
		categoryUUIDs = append(categoryUUIDs, uid)
	}

	if len(categoryUUIDs) == 0 {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "category_ids tidak boleh kosong"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListServicesByCategories(ctx, categoryUUIDs)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat layanan personal", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 07. Get All Normalized Service Category
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) GetAllNormalizedServiceCategory(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllNormalizedServiceCategory(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat normalized service category", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 08. Get Normalized Service Category By Id
// ---------------------------------------------------------------------

func (h *ServicesCategoriesHandler) GetNormalizedServiceCategoryById(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetNormalizedServiceCategoryById(ctx, serviceID)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat normalized service category berdasarkan ID", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
