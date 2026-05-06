package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/farildzaky/klinik-service/internal/cache"
	"github.com/farildzaky/klinik-service/internal/db"
	"github.com/farildzaky/klinik-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// =============================================================================
// PUBLIC: CATEGORIES
// =============================================================================

func (h *HoaxHandler) GetCategories(c fiber.Ctx) error {
	cacheKey := "hoax:categories:all"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllCategories(ctx)
	if err != nil {
		slog.Error("public.categories.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar kategori",
		})
	}

	responseData := make([]CategoryItem, 0, len(data))
	for _, row := range data {
		responseData = append(responseData, CategoryItem{
			ID:      pgUUIDToStr(row.ID),
			Name:    row.Name,
			Slug:    row.Slug,
			IconUrl: pgTextToStr(row.IconUrl),
		})
	}

	res := SuccessResponse{Pesan: "Daftar Kategori", Data: responseData}
	return cacheJSON(c, cacheKey, CacheTTLStatic, res)
}

// =============================================================================
// PUBLIC: SUBMIT REPORT
// =============================================================================

func (h *HoaxHandler) SubmitReport(c fiber.Ctx) error {
	var fieldErrors []FieldError

	rawName := c.FormValue("nama")
	name, errN := utils.ValidateQueryString(rawName, 150, "nama")
	if name == "" {
		msg := "Nama tidak boleh kosong"
		if errN != nil {
			msg = errN.Message
		}
		fieldErrors = append(fieldErrors, FieldError{Field: "nama", Message: msg})
	}

	email, errE := utils.ValidateEmail(c.FormValue("email"))
	if errE != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: errE.Field, Message: errE.Message})
	}

	phone, errP := utils.ValidatePhone(c.FormValue("no_hp"))
	if errP != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: errP.Field, Message: errP.Message})
	}

	content, errC := utils.ValidateTextContent(c.FormValue("isi_laporan"), 1000, "isi_laporan")
	if errC != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: errC.Field, Message: errC.Message})
	}

	proofLink, errL := utils.ValidateURL(c.FormValue("link_bukti"), "link_bukti")
	if errL != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: errL.Field, Message: errL.Message})
	}

	if len(fieldErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Gagal mengirim laporan karena data tidak valid",
			Action:  "Periksa kembali isian form Anda pada kolom yang salah",
			Errors:  fieldErrors,
		})
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	var imageUrl string
	if file, err := c.FormFile("gambar_bukti"); err == nil {
		url, errUp := h.uploadImageReal(uploadCtx, file, "klinik_hoaks_reports")
		if errUp != nil {
			slog.Error("public.upload.failed", slog.String("err", errUp.Error()))
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code:    "ERR_FILE_UPLOAD",
				Message: "Gambar bukti ditolak: " + errUp.Error(),
				Action:  "Pastikan ukuran dan format gambar sesuai",
			})
		}
		imageUrl = url
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	res, err := h.Queries.CreateReport(ctx, db.CreateReportParams{
		ReporterName:  name,
		ReporterEmail: email,
		ReporterPhone: phone,
		Content:       content,
		ProofLink:     pgtype.Text{String: proofLink, Valid: proofLink != ""},
		ProofImageUrl: pgtype.Text{String: imageUrl, Valid: imageUrl != ""},
	})
	if err != nil {
		if imageUrl != "" {
			slog.Warn("public.report.create_failed — gambar Cloudinary mungkin orphan", slog.String("imageUrl", imageUrl), slog.String("err", err.Error()))
		} else {
			slog.Error("public.report.create_failed", slog.String("err", err.Error()))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Terjadi kendala saat menyimpan laporan",
			Action:  "Sistem sedang sibuk, silakan coba lagi sebentar",
		})
	}

	cache.GlobalCache.Delete("hoax:stats")

	sendEmailAsync(email, name, res.TicketNumber, "PENDING")
	slog.Info("public.report.created",
		slog.String("ticket", res.TicketNumber),
		slog.String("email", email),
	)

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Laporan berhasil dikirim",
		Data: TicketCreatedResponse{
			TicketNumber: res.TicketNumber,
			CreatedAt:    res.CreatedAt.Time.Format(time.RFC3339),
		},
	})
}

// =============================================================================
// PUBLIC: TRACK REPORT BY TICKET
// =============================================================================

func (h *HoaxHandler) TrackReport(c fiber.Ctx) error {
	rawTicket := c.Query("no_tiket")
	ticket := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(rawTicket), " ", ""))

	if ticket == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Nomor tiket wajib diisi",
			Action:  "Masukkan nomor tiket yang dikirim ke email Anda",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetReportTrackingByTicket(ctx, ticket)
	if err != nil {
		slog.Warn("public.track.not_found", slog.String("ticket", ticket))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data laporan tidak ditemukan",
			Action:  "Pastikan nomor tiket sudah benar",
		})
	}

	response := ReportTrackingResponse{
		ReportID:     pgUUIDToStr(data.ReportID),
		TicketNumber: data.TicketNumber,
		ReporterName: data.ReporterName,
		ReportStatus: string(data.ReportStatus),
		ReportedAt:   data.ReportedAt.Time.Format(time.RFC3339),
	}

	if data.NewsID.Valid {
		response.NewsID = pgUUIDToStr(data.NewsID)
		response.NewsTitle = pgTextToStr(data.NewsTitle)
		response.NewsSlug = pgTextToStr(data.NewsSlug)
		response.NewsImage = pgTextToStr(data.NewsImage)
		response.CategoryName = pgTextToStr(data.CategoryName)
	}

	return c.JSON(SuccessResponse{Pesan: "Data Ditemukan", Data: response})
}

// =============================================================================
// PUBLIC: DASHBOARD STATS
// =============================================================================

func (h *HoaxHandler) GetDashboardStats(c fiber.Ctx) error {
	cacheKey := "hoax:stats"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDashboardStats(ctx)
	if err != nil {
		slog.Error("public.stats.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil statistik dasbor",
		})
	}

	responseData := make([]DashboardStatItem, 0, len(data))
	for _, row := range data {
		responseData = append(responseData, DashboardStatItem{
			CategoryID:   pgUUIDToStr(row.CategoryID),
			CategoryName: row.CategoryName,
			CategorySlug: row.CategorySlug,
			IconUrl:      pgTextToStr(row.IconUrl),
			TotalNews:    row.TotalNews,
		})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: responseData}
	return cacheJSON(c, cacheKey, CacheTTLList, res)
}

// =============================================================================
// PUBLIC: NEWS LIST + SEARCH (paginated)
// =============================================================================

func (h *HoaxHandler) GetPublicNews(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	keyword, _ := utils.ValidateQueryString(c.Query("search"), 100, "search")

	cacheKey := fmt.Sprintf("hoax:news:public:q_%s:p%d:l%d", keyword, page, limit)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var responseData []PublicNewsItem
	var total int64

	if keyword != "" {
		g.Go(func() error {
			data, err := h.Queries.SearchPublicNews(gCtx, db.SearchPublicNewsParams{
				Keyword:    keyword,
				LimitData:  int32(limit),
				OffsetData: int32(offset),
			})
			if err != nil {
				return err
			}
			for _, row := range data {
				responseData = append(responseData, PublicNewsItem{
					ID:           pgUUIDToStr(row.ID),
					Title:        row.Title,
					Slug:         row.Slug,
					ImageUrl:     row.ImageUrl,
					CategoryName: row.CategoryName,
					CategorySlug: row.CategorySlug,
					PublishedAt:  row.PublishedAt.Time.Format("02 Jan 2006"),
				})
			}
			return nil
		})
		g.Go(func() error {
			var err error
			total, err = h.Queries.CountSearchPublicNews(gCtx, keyword)
			return err
		})
	} else {
		g.Go(func() error {
			data, err := h.Queries.GetPublicNews(gCtx, db.GetPublicNewsParams{
				LimitData:  int32(limit),
				OffsetData: int32(offset),
			})
			if err != nil {
				return err
			}
			for _, row := range data {
				responseData = append(responseData, PublicNewsItem{
					ID:           pgUUIDToStr(row.ID),
					Title:        row.Title,
					Slug:         row.Slug,
					ImageUrl:     row.ImageUrl,
					CategoryName: row.CategoryName,
					CategorySlug: row.CategorySlug,
					PublishedAt:  row.PublishedAt.Time.Format("02 Jan 2006"),
				})
			}
			return nil
		})
		g.Go(func() error {
			var err error
			total, err = h.Queries.CountPublicNews(gCtx)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		slog.Error("public.news.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat daftar berita",
		})
	}

	if responseData == nil {
		responseData = []PublicNewsItem{}
	}

	pesan := "Daftar Berita"
	if keyword != "" {
		pesan = "Hasil Pencarian: " + keyword
	}

	res := SuccessResponse{
		Pesan:      pesan,
		Data:       responseData,
		Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
	}
	return cacheJSON(c, cacheKey, CacheTTLList, res)
}

// =============================================================================
// PUBLIC: NEWS DETAIL BY SLUG
// =============================================================================

func (h *HoaxHandler) GetPublicNewsDetailBySlug(c fiber.Ctx) error {
	newsSlug := strings.TrimSpace(c.Params("slug"))
	if newsSlug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Slug berita tidak valid",
			Action:  "Pastikan URL benar",
		})
	}

	cacheKey := "hoax:news:detail:" + newsSlug
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetNewsDetailBySlug(ctx, newsSlug)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Warn("public.news_detail.timeout", slog.String("slug", newsSlug))
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Code:    "ERR_TIMEOUT",
				Message: "Server sedang sibuk",
			})
		}
		slog.Warn("public.news_detail.not_found", slog.String("slug", newsSlug))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Berita yang dicari tidak ditemukan",
			Action:  "Berita mungkin sudah dihapus atau link kadaluarsa",
		})
	}

	response := PublicNewsDetail{
		ID:            pgUUIDToStr(data.ID),
		Title:         data.Title,
		Slug:          data.Slug,
		Description:   data.Description,
		ReferenceLink: pgTextToStr(data.ReferenceLink),
		ImageUrl:      data.ImageUrl,
		CategoryName:  data.CategoryName,
		CategorySlug:  data.CategorySlug,
		PublishedAt:   data.PublishedAt.Time.Format("02 Jan 2006 15:04 WIB"),
	}

	res := SuccessResponse{Pesan: "Detail Berita", Data: response}
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}