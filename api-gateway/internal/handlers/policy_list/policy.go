package policy_list

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const ContextQueryTimeout = 5 * time.Second

type PolicyHandler struct {
	Queries *db.Queries
}

func NewPolicyHandler(q *db.Queries) *PolicyHandler {
	return &PolicyHandler{Queries: q}
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

type CreatePolicyRequest struct {
	ServiceListID string          `json:"service_list_id"`
	Benefit       json.RawMessage `json:"benefit"`
	Instruction   json.RawMessage `json:"instruction"`
}

type UpdatePolicyRequest struct {
	Benefit     json.RawMessage `json:"benefit"`
	Instruction json.RawMessage `json:"instruction"`
}

// ---------------------------------------------------------------------
// 01. Get Policy By Service ID
// ---------------------------------------------------------------------

func (h *PolicyHandler) GetPolicyByServiceId(c fiber.Ctx) error {
	serviceID, err := parseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetPolicyByServiceId(ctx, serviceID)
	if err != nil {
		return c.Status(404).JSON(handlers.ErrorResponse{Error: "Data kebijakan tidak ditemukan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 02. Create Policy
// ---------------------------------------------------------------------

func (h *PolicyHandler) CreatePolicy(c fiber.Ctx) error {
	var req CreatePolicyRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}
	if req.ServiceListID == "" {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Validasi gagal", Detail: "service_list_id wajib diisi"})
	}

	serviceID, err := parseUUID(req.ServiceListID)
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "service_list_id tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	benefit := []byte("[]")
	if len(req.Benefit) > 0 {
		benefit = req.Benefit
	}

	instruction := []byte("[]")
	if len(req.Instruction) > 0 {
		instruction = req.Instruction
	}

	data, err := h.Queries.CreatePolicy(ctx, db.CreatePolicyParams{
		ServiceListID: serviceID,
		Benefit:       benefit,
		Instruction:   instruction,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal membuat data kebijakan", Detail: err.Error()})
	}
	return c.Status(201).JSON(handlers.SuccessResponse{Pesan: "Data kebijakan berhasil dibuat", Data: data})
}

// ---------------------------------------------------------------------
// 03. Update Policy
// ---------------------------------------------------------------------

func (h *PolicyHandler) UpdatePolicy(c fiber.Ctx) error {
	serviceID, err := parseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	var req UpdatePolicyRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "Request tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	benefit := []byte("[]")
	if len(req.Benefit) > 0 {
		benefit = req.Benefit
	}

	instruction := []byte("[]")
	if len(req.Instruction) > 0 {
		instruction = req.Instruction
	}

	data, err := h.Queries.UpdatePolicy(ctx, db.UpdatePolicyParams{
		ServiceListID: serviceID,
		Benefit:       benefit,
		Instruction:   instruction,
	})
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memperbarui data kebijakan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Data kebijakan berhasil diperbarui", Data: data})
}

// ---------------------------------------------------------------------
// 04. Delete Policy
// ---------------------------------------------------------------------

func (h *PolicyHandler) DeletePolicy(c fiber.Ctx) error {
	serviceID, err := parseUUID(c.Params("service_id"))
	if err != nil {
		return c.Status(400).JSON(handlers.ErrorResponse{Error: "UUID tidak valid", Detail: err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeletePolicy(ctx, serviceID); err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal menghapus data kebijakan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Data kebijakan berhasil dihapus"})
}

// ---------------------------------------------------------------------
// 05. Get All Policies
// ---------------------------------------------------------------------

func (h *PolicyHandler) GetAllPolicies(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllPolicies(ctx)
	if err != nil {
		return c.Status(500).JSON(handlers.ErrorResponse{Error: "Gagal memuat semua data kebijakan", Detail: err.Error()})
	}
	return c.JSON(handlers.SuccessResponse{Pesan: "Sukses", Data: data})
}
