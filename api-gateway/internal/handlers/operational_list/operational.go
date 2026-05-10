package operational_list

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type OperationalHandler struct {
	Queries *db.Queries
}

func NewOperationalHandler(q *db.Queries) *OperationalHandler {
	return &OperationalHandler{Queries: q}
}


// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

type CreateOperationalRequest struct {
	ServiceListID   string          `json:"service_list_id"`
	ServiceUrl      string          `json:"service_url"`
	Address         string          `json:"address"`
	OperationalHour json.RawMessage `json:"operational_hour"`
	SocialMedia     json.RawMessage `json:"social_media"`
}

type UpdateOperationalRequest struct {
	ServiceUrl      string          `json:"service_url"`
	Address         string          `json:"address"`
	OperationalHour json.RawMessage `json:"operational_hour"`
	SocialMedia     json.RawMessage `json:"social_media"`
}

// ---------------------------------------------------------------------
// 01. Get Operational By Service ID
// ---------------------------------------------------------------------

func (h *OperationalHandler) GetOperationalByServiceId(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetOperationalByServiceId(ctx, serviceID)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Data operasional tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. Create Operational
// ---------------------------------------------------------------------

func (h *OperationalHandler) CreateOperational(c fiber.Ctx) error {
	var req CreateOperationalRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.ServiceListID == "" || req.ServiceUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "service_list_id dan service_url wajib diisi"})
	}

	serviceID, err := handlers.ParseUUID(req.ServiceListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "service_list_id tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	address := pgtype.Text{}
	if req.Address != "" {
		address = pgtype.Text{String: req.Address, Valid: true}
	}

	opHour := []byte("{}")
	if len(req.OperationalHour) > 0 {
		opHour = req.OperationalHour
	}

	socialMedia := []byte("{}")
	if len(req.SocialMedia) > 0 {
		socialMedia = req.SocialMedia
	}

	data, err := h.Queries.CreateOperational(ctx, db.CreateOperationalParams{
		ServiceListID:   serviceID,
		ServiceUrl:      req.ServiceUrl,
		Address:         address,
		OperationalHour: opHour,
		SocialMedia:     socialMedia,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat data operasional", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Data operasional berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 03. Update Operational
// ---------------------------------------------------------------------

func (h *OperationalHandler) UpdateOperational(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdateOperationalRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.ServiceUrl == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "service_url wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	address := pgtype.Text{}
	if req.Address != "" {
		address = pgtype.Text{String: req.Address, Valid: true}
	}

	opHour := []byte("{}")
	if len(req.OperationalHour) > 0 {
		opHour = req.OperationalHour
	}

	socialMedia := []byte("{}")
	if len(req.SocialMedia) > 0 {
		socialMedia = req.SocialMedia
	}

	data, err := h.Queries.UpdateOperational(ctx, db.UpdateOperationalParams{
		ServiceListID:   serviceID,
		ServiceUrl:      req.ServiceUrl,
		Address:         address,
		OperationalHour: opHour,
		SocialMedia:     socialMedia,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui data operasional", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Data operasional berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 04. Delete Operational
// ---------------------------------------------------------------------

func (h *OperationalHandler) DeleteOperational(c fiber.Ctx) error {
	serviceID, err := handlers.ParseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteOperational(ctx, serviceID); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus data operasional", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Data operasional berhasil dihapus"})
}

// ---------------------------------------------------------------------
// 05. Get All Operationals
// ---------------------------------------------------------------------

func (h *OperationalHandler) GetAllOperationals(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllOperationals(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua data operasional", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
