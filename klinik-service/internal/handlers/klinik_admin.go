package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/farildzaky/klinik-service/internal/cache"
	"github.com/farildzaky/klinik-service/internal/db"
	"github.com/farildzaky/klinik-service/internal/notify"
	"github.com/farildzaky/klinik-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// =============================================================================
// ADMIN: NEWS — list & delete
// =============================================================================

func (h *HoaxHandler) GetAllNewsAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.GetAllNewsAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllNewsAdmin(gCtx, db.GetAllNewsAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllNewsAdmin(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.news.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat data berita admin",
		})
	}

	responseData := make([]AdminNewsItem, 0, len(data))
	for _, r := range data {
		responseData = append(responseData, AdminNewsItem{
			ID:           pgUUIDToStr(r.ID),
			Title:        r.Title,
			Slug:         r.Slug,
			CategoryName: r.CategoryName,
			PublishedAt:  r.PublishedAt.Time.Format("2006-01-02 15:04 WIB"),
			CreatedAt:    r.CreatedAt.Time.Format("2006-01-02 15:04 WIB"),
			TicketNumber: pgTextToStr(r.TicketNumber),
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Berita tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteNews(ctx, pgNewsID); err != nil {
		slog.Error("admin.news.delete_error",
			slog.String("id", newsID),
			slog.String("err", err.Error()),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menghapus berita",
		})
	}

	invalidatePublicCache()

	return c.JSON(SuccessResponse{Pesan: "Berita berhasil dihapus secara permanen"})
}

// =============================================================================
// ADMIN: CATEGORY
// =============================================================================

func (h *HoaxHandler) CreateCategoryAdmin(c fiber.Ctx) error {
	name, errN := utils.ValidateQueryString(c.FormValue("name"), 50, "name")
	if name == "" || errN != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Nama kategori tidak valid atau kosong",
		})
	}

	slug := utils.GenerateSlug(name)
	iconUrl, _ := utils.ValidateURL(c.FormValue("icon_url"), "icon_url")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Queries.CreateCategory(ctx, db.CreateCategoryParams{
		Name:    name,
		Slug:    slug,
		IconUrl: pgtype.Text{String: iconUrl, Valid: iconUrl != ""},
	})
	if err != nil {
		slog.Error("admin.category.create_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Code:    "ERR_CONFLICT",
			Message: "Gagal membuat kategori (mungkin nama sudah ada)",
		})
	}

	cache.GlobalCache.Delete("hoax:categories:all")
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Kategori berhasil ditambahkan",
	})
}

// =============================================================================
// ADMIN: REPORTS — list & moderate
// =============================================================================

func (h *HoaxHandler) GetAllReportsAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	statusFilter := c.Query("status")

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var statusArg db.NullReportStatus
	if statusFilter != "" {
		statusArg = db.NullReportStatus{
			ReportStatus: db.ReportStatus(statusFilter),
			Valid:        true,
		}
	}

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.GetAllReportsAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllReportsAdmin(gCtx, db.GetAllReportsAdminParams{
			StatusFilter: statusArg,
			LimitData:    int32(limit),
			OffsetData:   int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllReportsAdmin(gCtx, statusArg)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.reports.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat data laporan",
		})
	}

	responseData := make([]AdminReportItem, 0, len(data))
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

	desc, errD := utils.ValidateTextContent(c.FormValue("description"), 5000, "description")
	if desc == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "description", Message: "Deskripsi tidak boleh kosong"})
	} else if errD != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: "description", Message: errD.Message})
	}

	refLink, errL := utils.ValidateURL(c.FormValue("reference_link"), "reference_link")
	if errL != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: errL.Field, Message: errL.Message})
	}

	if len(fieldErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Data publikasi berita tidak lengkap",
			Action:  "Periksa kembali isian formulir admin",
			Errors:  fieldErrors,
		})
	}

	newsSlug := utils.GenerateSlug(title)

	pgReportID, err1 := parseUUID(reportID)
	pgCatID, err2 := parseUUID(catID)
	if err1 != nil || err2 != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format ID Laporan atau Kategori tidak valid",
		})
	}

	file, errFile := c.FormFile("admin_image")
	if errFile != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Gambar bukti admin belum diunggah",
			Action:  "Wajib mengunggah gambar stempel HOAX/FAKTA",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		slog.Error("admin.tx.begin_failed", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memulai transaksi sistem",
		})
	}
	defer func() { _ = tx.Rollback(ctx) }()

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	imageUrl, errUp := h.uploadImageReal(uploadCtx, file, "klinik_hoaks_news")
	if errUp != nil {
		slog.Error("admin.upload.failed", slog.String("err", errUp.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_FILE_UPLOAD",
			Message: "Gagal upload gambar admin: " + errUp.Error(),
		})
	}

	qtx := h.Queries.WithTx(tx)

	if _, err := qtx.CreateNewsClarification(ctx, db.CreateNewsClarificationParams{
		ReportID:      pgReportID,
		CategoryID:    pgCatID,
		Title:         title,
		Slug:          newsSlug,
		Description:   desc,
		ReferenceLink: pgtype.Text{String: refLink, Valid: refLink != ""},
		ImageUrl:      imageUrl,
	}); err != nil {
		slog.Warn("admin.tx.insert_news_failed — gambar Cloudinary mungkin orphan", slog.String("imageUrl", imageUrl), slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menyimpan berita",
		})
	}

	if err := qtx.UpdateReportStatus(ctx, db.UpdateReportStatusParams{
		ReportID: pgReportID,
		Status:   "PROCESSED",
	}); err != nil {
		slog.Warn("admin.tx.update_status_failed — gambar Cloudinary mungkin orphan", slog.String("imageUrl", imageUrl), slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memperbarui status laporan",
		})
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Warn("admin.tx.commit_failed — gambar Cloudinary mungkin orphan", slog.String("imageUrl", imageUrl), slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memfinalisasi transaksi",
		})
	}

	invalidatePublicCache()
	notify.ByService("/klinik-hoaks", "Berita Baru: "+title, "Berita edukasi baru telah dipublikasikan. Ketuk untuk membaca selengkapnya.", nil)

	if reporterEmail := c.FormValue("reporter_email"); reporterEmail != "" {
		reporterName := c.FormValue("reporter_name")
		if reporterName == "" {
			reporterName = "Pelapor"
		}
		ticketNumber := c.FormValue("ticket_number")
		if ticketNumber == "" {
			ticketNumber = "Laporan Anda"
		}
		sendEmailAsync(reporterEmail, reporterName, ticketNumber, "PROCESSED")
	}

	slog.Info("admin.report.processed",
		slog.String("report_id", reportID),
		slog.String("news_slug", newsSlug),
	)

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Berita dipublikasi",
	})
}

func (h *HoaxHandler) RejectReportAdmin(c fiber.Ctx) error {
	reportID := c.FormValue("report_id")
	reason := c.FormValue("reason")

	if reportID == "" || reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Laporan dan Alasan penolakan wajib diisi",
		})
	}

	pgReportID, err := parseUUID(reportID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format ID Laporan tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.UpdateReportStatus(ctx, db.UpdateReportStatusParams{
		ReportID: pgReportID,
		Status:   "REJECTED",
	}); err != nil {
		slog.Error("admin.report.reject_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menolak laporan",
		})
	}

	if reporterEmail := c.FormValue("reporter_email"); reporterEmail != "" {
		reporterName := c.FormValue("reporter_name")
		if reporterName == "" {
			reporterName = "Pelapor"
		}
		ticketNumber := c.FormValue("ticket_number")
		if ticketNumber != "" {
			statusEmail := fmt.Sprintf("DITOLAK. Alasan Admin: %s", reason)
			sendEmailAsync(reporterEmail, reporterName, ticketNumber, statusEmail)
		}
	}

	slog.Info("admin.report.rejected",
		slog.String("report_id", reportID),
		slog.String("reason", reason),
	)

	return c.JSON(SuccessResponse{
		Pesan: "Laporan berhasil ditolak dan pelapor telah dinotifikasi",
	})
}

// =============================================================================
// CACHE INVALIDATION — centralize biar konsisten
// =============================================================================


func invalidatePublicCache() {
	cache.GlobalCache.Delete("hoax:stats")
	cache.GlobalCache.DeleteByPrefix("hoax:news:public")
	cache.GlobalCache.DeleteByPrefix("hoax:news:detail")
}