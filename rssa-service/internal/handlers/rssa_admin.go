package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/farildzaky/rssa-service/internal/db"
	"github.com/farildzaky/rssa-service/internal/notify"
	"github.com/farildzaky/rssa-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// ADMIN: CREATE RUANGAN
// =============================================================================

func (h *RSSAHandler) CreateRuanganAdmin(c fiber.Ctx) error {
	var req CreateRuanganRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_INVALID_JSON",
			Message: "Format data tidak valid",
		})
	}

	// Validasi field wajib
	if err := requireFields(c, map[string]string{
		"nama":     req.Nama,
		"kelas_id": fmt.Sprintf("%d", req.KelasID),
	}); err != nil {
		return err
	}

	if strings.TrimSpace(req.Nama) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Nama ruangan tidak boleh kosong atau hanya berisi spasi!",
		})
	}
	
	if req.KelasID == 0 { 
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Kelas ID wajib diisi!",
		})
	}

	ruanganSlug := utils.GenerateSlug(req.Nama)

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_,err := h.Queries.CreateRuangan(ctx, db.CreateRuanganParams{
		KelasID:   req.KelasID,
		Nama:      req.Nama,
		Slug:      ruanganSlug, // Insert slug ke DB
		Kapasitas: req.Kapasitas,
		Terisi:    req.Terisi,
	})
	if err != nil {
		slog.Error("admin.ruangan.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_INSERT",
			Message: "Gagal menambah data ruangan",
		})
	}

	// Invalidasi cache publik agar jumlah kamar langsung terupdate!
	invalidateRuanganCache()
	notify.ByService("/rssa", "Fasilitas Baru RSSA", "Ruangan baru tersedia: "+req.Nama, nil)

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Data ruangan berhasil ditambahkan",
	})
}

// =============================================================================
// ADMIN: UPDATE RUANGAN
// =============================================================================

func (h *RSSAHandler) UpdateRuanganAdmin(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID ruangan tidak valid",
		})
	}

	var req UpdateRuanganRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_INVALID_JSON",
			Message: "Format data tidak valid",
		})
	}

	// Validasi field wajib
	if err := requireFields(c, map[string]string{
		"nama":     req.Nama,
		"kelas_id": fmt.Sprintf("%d", req.KelasID),
	}); err != nil {
		return err
	}

	if strings.TrimSpace(req.Nama) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Nama ruangan tidak boleh kosong atau hanya berisi spasi!",
		})
	}
	
	if req.KelasID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Kelas ID wajib diisi!",
		})
	}

	ruanganSlug := utils.GenerateSlug(req.Nama)

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_,err = h.Queries.UpdateRuangan(ctx, db.UpdateRuanganParams{
		ID:        int32(id),
		KelasID:   req.KelasID,
		Nama:      req.Nama,
		Slug:      ruanganSlug, // Update slug di DB
		Kapasitas: req.Kapasitas,
		Terisi:    req.Terisi,
	})
	if err != nil {
		slog.Error("admin.ruangan.update.error", slog.String("err", err.Error()), slog.Int("id", id))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_UPDATE",
			Message: "Gagal memperbarui data ruangan",
		})
	}

	invalidateRuanganCache()

	return c.JSON(SuccessResponse{
		Pesan: "Data ruangan berhasil diperbarui",
	})
}

// =============================================================================
// ADMIN: DELETE RUANGAN
// =============================================================================

func (h *RSSAHandler) DeleteRuanganAdmin(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID ruangan tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteRuangan(ctx, int32(id)); err != nil {
		slog.Error("admin.ruangan.delete.error", slog.String("err", err.Error()), slog.Int("id", id))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_DELETE",
			Message: "Gagal menghapus data ruangan",
		})
	}

	invalidateRuanganCache()

	return c.JSON(SuccessResponse{
		Pesan: "Data ruangan berhasil dihapus",
	})
}