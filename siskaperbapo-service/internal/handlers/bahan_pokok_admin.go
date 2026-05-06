package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/farildzaky/siskaperbapo-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// =====================================================================
// ADMIN: CREATE BAHAN POKOK
// =====================================================================
func (h *SiskaperbapoHandler) CreateBahanPokok(c fiber.Ctx) error {
	nama := strings.TrimSpace(c.FormValue("nama"))
	satuan := strings.TrimSpace(c.FormValue("satuan"))

	if err := requireFields(c, map[string]string{"nama": nama}); err != nil {
		return err
	}
	if satuan == "" {
		satuan = "kg"
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	exists, errCheck := h.Queries.CheckBahanPokokNamaExists(ctxDb, nama)
	if errCheck != nil {
		slog.Error("admin.bahan_pokok.nama_check_error", slog.String("err", errCheck.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memvalidasi nama komoditas",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Code:    "ERR_CONFLICT",
			Message: "Bahan pokok dengan nama ini sudah ada",
		})
	}

	gambarHeader, errT := c.FormFile("gambar")
	if errT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "File gambar wajib diupload",
		})
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	gambarURL, gambarPubID, errUp := uploadImage(uploadCtx, h.Cld, gambarHeader, "siskaperbapo/komoditas")
	if errUp != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_FILE_UPLOAD",
			Message: "Gagal upload gambar: " + errUp.Error(),
		})
	}

	slugBaru := generateSlug(nama)

	idBaru, errDb := h.Queries.CreateBahanPokok(ctxDb, db.CreateBahanPokokParams{
		Nama:      nama,
		Slug:      slugBaru,
		Satuan:    satuan,
		GambarUrl: pgtype.Text{String: gambarURL, Valid: true},
	})

	if errDb != nil {
		destroyImageAsync(h.Cld, gambarPubID) 
		var pgErr *pgconn.PgError
		if errors.As(errDb, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code:    "ERR_CONFLICT",
				Message: "Bahan pokok ini sudah ada di database!",
			})
		}
		slog.Error("admin.bahan_pokok.create_error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menyimpan ke database",
		})
	}

	invalidateAllBahanCache()

	slog.Info("admin.bahan_pokok.created", slog.Int("id", int(idBaru.ID)), slog.String("nama", idBaru.Nama))

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Komoditas berhasil disimpan.",
		Data: BahanPokokItemResponse{
			ID:        idBaru.ID,
			Nama:      idBaru.Nama,
			Slug:      idBaru.Slug,
			Satuan:    satuan,
			GambarUrl: gambarURL,
		},
	})
}

// =====================================================================
// ADMIN: UPDATE BAHAN POKOK
// =====================================================================
func (h *SiskaperbapoHandler) UpdateBahanPokok(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID harus berupa angka yang valid",
		})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	oldData, errDbRead := h.Queries.GetBahanPokokByID(ctxDb, int32(id))
	if errDbRead != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data bahan pokok tidak ditemukan",
		})
	}

	finalImageUrl := oldData.GambarUrl
	nama := strings.TrimSpace(c.FormValue("nama"))
	satuan := strings.TrimSpace(c.FormValue("satuan"))

	finalNama := oldData.Nama
	finalSlug := oldData.Slug

	if satuan == "" {
		satuan = oldData.Satuan
	}

	if nama != "" && !strings.EqualFold(nama, oldData.Nama) {
		exists, errCheck := h.Queries.CheckBahanPokokNamaExistsExcluding(ctxDb, db.CheckBahanPokokNamaExistsExcludingParams{
			Nama: nama,
			ID:   int32(id),
		})
		if errCheck != nil {
			slog.Error("admin.bahan_pokok.nama_check_error", slog.String("err", errCheck.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Code:    "ERR_INTERNAL_DB",
				Message: "Gagal memvalidasi nama komoditas",
			})
		}
		if exists {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code:    "ERR_CONFLICT",
				Message: "Bahan pokok dengan nama ini sudah ada",
			})
		}
		finalNama = nama
		finalSlug = generateSlug(nama)
	}

	var newPubID, oldPubIDToDelete string

	if gambarBaru, errFile := c.FormFile("gambar"); errFile == nil {
		uploadCtx, cancelUp := context.WithTimeout(context.Background(), ContextUploadTimeout)
		defer cancelUp()

		url, pubID, errUp := uploadImage(uploadCtx, h.Cld, gambarBaru, "siskaperbapo/komoditas")
		if errUp != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code:    "ERR_FILE_UPLOAD",
				Message: "Gagal upload gambar baru: " + errUp.Error(),
			})
		}
		newPubID = pubID
		finalImageUrl = pgtype.Text{String: url, Valid: true}
		oldPubIDToDelete = utils.ExtractPublicID(oldData.GambarUrl.String)
	}

	ctxUpdate, cancelUpdate := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelUpdate()

	dataUpdate, errDb := h.Queries.UpdateBahanPokok(ctxUpdate, db.UpdateBahanPokokParams{
		ID:        int32(id),
		Nama:      finalNama,
		Slug:      finalSlug,
		Satuan:    satuan,
		GambarUrl: finalImageUrl,
	})

	if errDb != nil {
		destroyImageAsync(h.Cld, newPubID)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengupdate bahan pokok",
		})
	}

	destroyImageAsync(h.Cld, oldPubIDToDelete)
	invalidateAllBahanCache()

	slog.Info("admin.bahan_pokok.updated", slog.Int("id", int(dataUpdate.ID)), slog.String("nama", dataUpdate.Nama))

	return c.JSON(SuccessResponse{
		Pesan: "Komoditas berhasil diperbarui.",
		Data: BahanPokokItemResponse{
			ID:        dataUpdate.ID,
			Nama:      dataUpdate.Nama,
			Slug:      dataUpdate.Slug,
			Satuan:    satuan,
			GambarUrl: finalImageUrl.String,
		},
	})
}

// =====================================================================
// ADMIN: DELETE BAHAN POKOK
// =====================================================================
func (h *SiskaperbapoHandler) DeleteBahanPokok(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID tidak valid",
		})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	oldData, errDbRead := h.Queries.GetBahanPokokByID(ctxDb, int32(id))
	if errDbRead != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data tidak ditemukan",
		})
	}

	if err := h.Queries.DeleteBahanPokok(ctxDb, int32(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menghapus bahan pokok",
		})
	}

	destroyImageAsync(h.Cld, utils.ExtractPublicID(oldData.GambarUrl.String))
	invalidateAllBahanCache()

	slog.Info("admin.bahan_pokok.deleted", slog.Int("id", id), slog.String("nama", oldData.Nama))

	return c.JSON(SuccessResponse{Pesan: "Bahan pokok berhasil dihapus."})
}

// =====================================================================
// ADMIN: UPSERT HARGA HARIAN (Menggantikan Create/Update lama)
// =====================================================================
func (h *SiskaperbapoHandler) CreateHargaHarian(c fiber.Ctx) error {
	var req UpsertHargaRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format JSON tidak valid",
		})
	}

	if req.Harga <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Harga harus lebih besar dari 0",
		})
	}

	parsedTime, errT := time.Parse("2006-01-02", req.Tanggal)
	if errT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format tanggal salah (gunakan YYYY-MM-DD)",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, errDb := h.Queries.UpsertHargaHarian(ctx, db.UpsertHargaHarianParams{
		BahanPokokID: req.BahanPokokID,
		AreaID:       req.AreaID,
		Harga:        req.Harga,
		Tanggal:      pgtype.Date{Time: parsedTime, Valid: true},
	})

	if errDb != nil {
		slog.Error("admin.harga.upsert_error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menyimpan harga",
		})
	}

	invalidateAllBahanCache()

	slog.Info("admin.harga.upserted", slog.Int("bahan_pokok_id", int(req.BahanPokokID)), slog.String("tanggal", req.Tanggal))

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Harga berhasil dicatat.",
		Data:  req,
	})
}