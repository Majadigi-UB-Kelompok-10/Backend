package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/farildzaky/klinik-service/internal/cache"
	"github.com/farildzaky/klinik-service/internal/db"
	"github.com/farildzaky/klinik-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// =========================================================
// 🚀 RECOVERY: FUNGSI ADMIN YANG HILANG DIMAKAN COPILOT
// =========================================================

func (h *HoaxHandler) GetAllNewsAdmin(c fiber.Ctx) error {
	pageStr, limitStr := c.Query("page", "1"), c.Query("limit", "10")
	page, limit := utils.ValidatePaginationParams(pageStr, limitStr)
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	var data []db.GetAllNewsAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllNewsAdmin(ctx, db.GetAllNewsAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})

	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllNewsAdmin(ctx)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("GetAllNewsAdmin DB Error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat data berita admin.",
		})
	}

	var responseData []AdminNewsItem
	for _, r := range data {
		responseData = append(responseData, AdminNewsItem{
			ID:           pgUUIDToStr(r.ID),
			Title:        r.Title,
			CategoryName: r.CategoryName,
			PublishedAt:  r.PublishedAt.Time.Format("2006-01-02 15:04 WIB"),
			CreatedAt:    r.CreatedAt.Time.Format("2006-01-02 15:04 WIB"),
			TicketNumber: pgTextToStr(r.TicketNumber),
		})
	}
	if responseData == nil {
		responseData = []AdminNewsItem{}
	}

	return c.JSON(SuccessResponse{
		Pesan:      "Data Berita Klarifikasi",
		Data:       responseData,
		Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
	})
}

func (h *HoaxHandler) DeleteNewsAdmin(c fiber.Ctx) error {
	newsID := c.Params("id")
	pgNewsID, err := parseUUID(newsID)
	if err != nil || newsID == "" {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Berita tidak valid.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	errDelete := h.Queries.DeleteNews(ctx, pgNewsID)
	if errDelete != nil {
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menghapus berita dari database.",
		})
	}

	// Hapus Cache yang berkaitan
	cache.GlobalCache.Delete("hoax:stats")
	cache.GlobalCache.DeleteByPrefix("hoax:news:public")

	return c.JSON(SuccessResponse{Pesan: "Berita berhasil dihapus secara permanen"})
}

func (h *HoaxHandler) CreateCategoryAdmin(c fiber.Ctx) error {
	name, errN := utils.ValidateQueryString(c.FormValue("name"), 50, "name")
	if name == "" || errN != nil {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Nama kategori tidak valid atau kosong.",
		})
	}

	slug := utils.GenerateSlug(name)
	iconUrl := c.FormValue("icon_url")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Queries.CreateCategory(ctx, db.CreateCategoryParams{
		Name:    name,
		Slug:    slug,
		IconUrl: pgtype.Text{String: iconUrl, Valid: iconUrl != ""},
	})

	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal membuat kategori (mungkin nama sudah ada).",
		})
	}

	cache.GlobalCache.Delete("hoax:categories:all")
	return c.Status(201).JSON(SuccessResponse{Pesan: "Kategori berhasil ditambahkan"})
}

func (h *HoaxHandler) GetAllReportsAdmin(c fiber.Ctx) error {
	pageStr, limitStr := c.Query("page", "1"), c.Query("limit", "10")
	page, limit := utils.ValidatePaginationParams(pageStr, limitStr)
	offset := (page - 1) * limit
	statusFilter := c.Query("status")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var data []db.GetAllReportsAdminRow
	var total int64

	var statusArg db.NullReportStatus
	if statusFilter != "" {
		statusArg = db.NullReportStatus{ReportStatus: db.ReportStatus(statusFilter), Valid: true}
	}

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllReportsAdmin(ctx, db.GetAllReportsAdminParams{
			StatusFilter: statusArg,
			LimitData:    int32(limit),
			OffsetData:   int32(offset),
		})
		return err
	})

	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllReportsAdmin(ctx, statusArg)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("GetAllReportsAdmin DB Error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat data laporan admin.",
			Action:  "Periksa koneksi database atau coba muat ulang halaman.",
		})
	}

	var responseData []AdminReportItem
	for _, r := range data {
		responseData = append(responseData, AdminReportItem{
			ID:            pgUUIDToStr(r.ID),
			TicketNumber:  r.TicketNumber,
			ReporterName:  r.ReporterName,
			ReporterEmail: r.ReporterEmail,
			Content:       r.Content,
			ProofLink:     pgTextToStr(r.ProofLink),
			ProofImageUrl: pgTextToStr(r.ProofImageUrl),
			Status:        string(r.Status),
			CreatedAt:     r.CreatedAt.Time.Format("2006-01-02 15:04 WIB"),
		})
	}
	if responseData == nil {
		responseData = []AdminReportItem{}
	}

	return c.JSON(SuccessResponse{
		Pesan:      "Data Laporan",
		Data:       responseData,
		Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
	})
}

func (h *HoaxHandler) ProcessReportAdmin(c fiber.Ctx) error {
	var fieldErrors []FieldError

	reportID := c.FormValue("report_id")
	if reportID == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "report_id", Message: "ID Laporan wajib disertakan"})
	}

	catID := c.FormValue("category_id")
	if catID == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "category_id", Message: "Kategori wajib dipilih"})
	}

	title, errT := utils.ValidateQueryString(c.FormValue("title"), 255, "title")
	if title == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "title", Message: "Judul tidak boleh kosong"})
	} else if errT != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: "title", Message: errT.Message})
	}

	newsSlug := utils.GenerateSlug(title)

	desc, errD := utils.ValidateTextContent(c.FormValue("description"), 5000, "description")
	if desc == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "description", Message: "Deskripsi tidak boleh kosong"})
	} else if errD != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: "description", Message: errD.Message})
	}

	refLink, errL := utils.ValidateURL(c.FormValue("reference_link"), "reference_link")
	if errL != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: "reference_link", Message: errL.Message})
	}

	if len(fieldErrors) > 0 {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Data publikasi berita tidak lengkap atau salah.",
			Action:  "Periksa kembali isian formulir admin.",
			Errors:  fieldErrors,
		})
	}

	file, errFile := c.FormFile("admin_image")
	if errFile != nil {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Gambar bukti admin belum diunggah.",
			Action:  "Wajib mengunggah gambar stempel HOAX/FAKTA.",
		})
	}

	ctxCld, cancelCld := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelCld()

	imageUrl, errUp := h.uploadImageReal(ctxCld, file, "klinik_hoaks_news")
	if errUp != nil {
		slog.Error("Admin Image Upload Failed", slog.String("err", errUp.Error()))
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_FILE_UPLOAD",
			Message: "Gagal memproses gambar admin: " + errUp.Error(),
			Action:  "Pastikan koneksi internet stabil dan ukuran gambar < 2MB.",
		})
	}

	pgReportID, err1 := parseUUID(reportID)
	pgCatID, err2 := parseUUID(catID)
	if err1 != nil || err2 != nil {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format ID Laporan atau Kategori tidak valid.",
			Action:  "Pastikan memproses dari data yang benar di dashboard.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	tx, errTx := h.DB.Begin(ctx)
	if errTx != nil {
		slog.Error("Failed to start transaction", slog.String("err", errTx.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memulai transaksi sistem.",
			Action:  "Silakan coba lagi dalam beberapa saat.",
		})
	}
	defer tx.Rollback(ctx)

	qtx := h.Queries.WithTx(tx)

	newsArg := db.CreateNewsClarificationParams{
		ReportID:      pgReportID,
		CategoryID:    pgCatID,
		Title:         title,
		Slug:          newsSlug,
		Description:   desc,
		ReferenceLink: pgtype.Text{String: refLink, Valid: refLink != ""},
		ImageUrl:      imageUrl,
	}

	_, errInsert := qtx.CreateNewsClarification(ctx, newsArg)
	if errInsert != nil {
		slog.Error("Failed to insert news", slog.String("err", errInsert.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menyimpan berita ke database.",
			Action:  "Silakan coba lagi dalam beberapa saat.",
		})
	}

	errUpdate := qtx.UpdateReportStatus(ctx, db.UpdateReportStatusParams{
		ReportID: pgReportID,
		Status:   "PROCESSED",
	})
	if errUpdate != nil {
		slog.Error("Failed to update report status", slog.String("err", errUpdate.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memperbarui status laporan.",
			Action:  "Silakan coba lagi dalam beberapa saat.",
		})
	}

	if errCommit := tx.Commit(ctx); errCommit != nil {
		slog.Error("Failed to commit transaction", slog.String("err", errCommit.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memfinalisasi transaksi data.",
			Action:  "Silakan coba lagi dalam beberapa saat.",
		})
	}

	cache.GlobalCache.Delete("hoax:stats")
	cache.GlobalCache.DeleteByPrefix("hoax:news:public")

	reporterEmail := c.FormValue("reporter_email")
	reporterName := c.FormValue("reporter_name")
	if reporterName == "" {
		reporterName = "Pelapor"
	}
	ticketNumber := c.FormValue("ticket_number")
	if ticketNumber == "" {
		ticketNumber = "Laporan Anda"
	}

	if reporterEmail != "" {
		sendEmailAsync(reporterEmail, reporterName, ticketNumber, "PROCESSED")
	} else {
		slog.Warn("Email pelapor kosong, notifikasi SendGrid dilewati", slog.String("report_id", reportID))
	}

	slog.Info("Report Processed and News Created", slog.String("report_id", reportID), slog.String("news_slug", newsSlug))

	return c.Status(201).JSON(SuccessResponse{Pesan: "Berita dipublikasi"})
}

func (h *HoaxHandler) RejectReportAdmin(c fiber.Ctx) error {
	reportID := c.FormValue("report_id")
	reason := c.FormValue("reason")

	if reportID == "" || reason == "" {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Laporan dan Alasan penolakan wajib diisi.",
			Action:  "Pastikan Anda menuliskan alasan kenapa laporan ini ditolak.",
		})
	}

	pgReportID, err := parseUUID(reportID)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format ID Laporan tidak valid.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	errUpdate := h.Queries.UpdateReportStatus(ctx, db.UpdateReportStatusParams{
		ReportID: pgReportID,
		Status:   "REJECTED",
	})

	if errUpdate != nil {
		slog.Error("Failed to reject report", slog.String("err", errUpdate.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menolak laporan.",
		})
	}

	reporterEmail := c.FormValue("reporter_email")
	reporterName := c.FormValue("reporter_name")
	ticketNumber := c.FormValue("ticket_number")

	if reporterEmail != "" && ticketNumber != "" {
		if reporterName == "" {
			reporterName = "Pelapor"
		}
		statusEmail := fmt.Sprintf("DITOLAK. Alasan Admin: %s", reason)
		sendEmailAsync(reporterEmail, reporterName, ticketNumber, statusEmail)
	}

	slog.Info("Report Rejected", slog.String("report_id", reportID), slog.String("reason", reason))

	return c.Status(200).JSON(SuccessResponse{
		Pesan: "Laporan berhasil ditolak dan pelapor telah dinotifikasi.",
	})
}