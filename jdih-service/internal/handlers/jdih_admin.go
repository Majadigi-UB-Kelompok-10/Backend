package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/farildzaky/jdih-service/internal/db"
	"github.com/farildzaky/jdih-service/internal/utils"
)

func (h *JdihHandler) invalidateJdihCache() {
	h.Cache.DeleteByPrefix("jdih:public:")
}

// =============================================================================
// ADMIN: LIST ALL DOKUMEN
// =============================================================================

func (h *JdihHandler) AdminListDokumen(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	var jenisArg db.NullJenisDokumenType
	if jns := c.Query("jenis"); jns != "" {
		if j, err := utils.ValidateJenisDokumen(jns); err == nil {
			jenisArg = db.NullJenisDokumenType{JenisDokumenType: db.JenisDokumenType(j), Valid: true}
		}
	}

	var tahunArg pgtype.Int2
	if thn := c.Query("tahun"); thn != "" {
		if t, err := utils.ValidateTahun(thn); err == nil {
			tahunArg = pgtype.Int2{Int16: t, Valid: true}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListAllDokumenAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListAllDokumenAdmin(gCtx, db.ListAllDokumenAdminParams{
			Jenis:      jenisArg,
			Tahun:      tahunArg,
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllDokumen(gCtx, db.CountAllDokumenParams{
			Jenis: jenisArg,
			Tahun: tahunArg,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memuat dokumen"})
	}

	// Re-use DTO dari public agar standar
	resData := make([]DokumenItemResponse, 0, len(data))
	for _, r := range data {
		resData = append(resData, DokumenItemResponse{
			ID:         r.ID,
			Jenis:      string(r.Jenis),
			Judul:      r.Judul,
			Tanggal:    r.TanggalPenetapan.Time.Format("02 Jan 2006"),
			Status:     string(r.Status),
			JumlahView: r.JumlahView,
		})
	}

	return c.JSON(PaginatedResponse{
		Pesan:      "Data Semua Dokumen",
		Data:       resData,
		Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
	})
}

// =============================================================================
// ADMIN: CREATE DOKUMEN + SUBJEK + UPLOAD PDF
// =============================================================================

func (h *JdihHandler) AdminCreateDokumen(c fiber.Ctx) error {
	jenis, errJ := utils.ValidateJenisDokumen(c.FormValue("jenis"))
	tahun, errT := utils.ValidateTahun(c.FormValue("tahun"))
	status, errS := utils.ValidateStatusDokumen(c.FormValue("status"))
	
	if errJ != nil || errT != nil || errS != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format Jenis, Tahun, atau Status tidak valid"})
	}

	nomor := c.FormValue("nomor")
	judul := c.FormValue("judul")
	ringkasan := c.FormValue("ringkasan")
	urusan := c.FormValue("urusan_pemerintahan")
	tglPenetapanStr := c.FormValue("tanggal_penetapan") 
	
	parsedDate, errDate := time.Parse("2006-01-02", tglPenetapanStr)
	if errDate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format tanggal_penetapan harus YYYY-MM-DD"})
	}

	file, errFile := c.FormFile("pdf_file")
	if errFile != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "File PDF wajib diunggah"})
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	pdfUrl, pdfSizeKb, errUp := h.uploadPDFReal(uploadCtx, file, fmt.Sprintf("jdih_dokumen/%d", tahun))
	if errUp != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_UPLOAD", Message: errUp.Error()})
	}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancelDB()

	tx, errTx := h.DB.Begin(ctxDB)
	if errTx != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memulai transaksi"})
	}
	defer func() { _ = tx.Rollback(ctxDB) }()

	qtx := h.Queries.WithTx(tx)

	dokID, errDoc := qtx.CreateDokumen(ctxDB, db.CreateDokumenParams{
		Jenis:              db.JenisDokumenType(jenis),
		Nomor:              nomor,
		Tahun:              tahun,
		Judul:              judul,
		Ringkasan:          pgtype.Text{String: ringkasan, Valid: ringkasan != ""},
		TanggalPenetapan:   pgtype.Date{Time: parsedDate, Valid: true},
		Status:             db.StatusDokumenType(status),
		PdfUrl:             pdfUrl,
		PdfSizeKb:          pgtype.Int4{Int32: pdfSizeKb, Valid: true},
		UrusanPemerintahan: pgtype.Text{String: urusan, Valid: urusan != ""},
	})

	if errDoc != nil {
		slog.Error("admin.dokumen.create_failed", slog.String("err", errDoc.Error()))
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Code: "ERR_CONFLICT", Message: "Gagal menyimpan dokumen. Pastikan Nomor dan Tahun tidak duplikat."})
	}

	rawSubjek := c.FormValue("subjek")
	if rawSubjek != "" {
		tags := strings.Split(rawSubjek, ",")
		for _, t := range tags {
			tagName := strings.TrimSpace(t)
			if tagName == "" {
				continue
			}
			subjekRow, errSub := qtx.GetOrCreateSubjek(ctxDB, tagName)
			if errSub == nil {
				_ = qtx.AddSubjekToDokumen(ctxDB, db.AddSubjekToDokumenParams{
					DokumenID: dokID,
					SubjekID:  subjekRow.ID,
				})
			}
		}
	}

	if errCommit := tx.Commit(ctxDB); errCommit != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memfinalisasi data dokumen"})
	}

	h.invalidateJdihCache()

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Dokumen berhasil dipublikasikan",
		Data: map[string]interface{}{"id": dokID, "pdf_url": pdfUrl},
	})
}

// =============================================================================
// ADMIN: DELETE DOKUMEN
// =============================================================================

func (h *JdihHandler) AdminDeleteDokumen(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteDokumen(ctx, int32(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal menghapus dokumen"})
	}

	h.invalidateJdihCache()
	return c.JSON(SuccessResponse{Pesan: "Dokumen berhasil dihapus"})
}

// =============================================================================
// ADMIN: PENGUMUMAN
// =============================================================================

func (h *JdihHandler) AdminCreatePengumuman(c fiber.Ctx) error {
	judul := c.FormValue("judul")
	isi := c.FormValue("isi")

	if strings.TrimSpace(judul) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Judul pengumuman wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Queries.CreatePengumuman(ctx, db.CreatePengumumanParams{
		Judul: judul,
		Isi:   pgtype.Text{String: isi, Valid: isi != ""},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal membuat pengumuman"})
	}

	h.invalidateJdihCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Pesan: "Pengumuman berhasil dipublikasikan"})
}