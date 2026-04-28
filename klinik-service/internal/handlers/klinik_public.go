package handlers

import (
	"context"
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

func (h *HoaxHandler) GetCategories(c fiber.Ctx) error {
	cacheKey := "hoax:categories:all"

	if cached, found := cache.GlobalCache.Get(cacheKey); found {
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAllCategories(ctx)
	if err != nil {
		slog.Error("GetCategories DB Error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar kategori.",
		})
	}

	var responseData []map[string]interface{}
	for _, row := range data {
		responseData = append(responseData, map[string]interface{}{
			"id":   pgUUIDToStr(row.ID),
			"name": row.Name,
			"slug": row.Slug,
		})
	}

	if responseData == nil {
		responseData = []map[string]interface{}{}
	}

	res := SuccessResponse{
		Pesan: "Daftar Kategori",
		Data:  responseData,
	}

	cache.GlobalCache.Set(cacheKey, res, CacheTTLStatic)
	return c.JSON(res)
}

func (h *HoaxHandler) SubmitReport(c fiber.Ctx) error {
	var fieldErrors []FieldError

	name, errN := utils.ValidateQueryString(c.FormValue("nama"), 150, "nama")
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
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Gagal mengirim laporan karena data tidak valid.",
			Action:  "Periksa kembali isian form Anda pada kolom yang salah.",
			Errors:  fieldErrors,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancel()

	var imageUrl string
	if file, err := c.FormFile("gambar_bukti"); err == nil {
		url, errUp := h.uploadImageReal(ctx, file, "klinik_hoaks_reports")
		if errUp != nil {
			slog.Error("Upload Bukti Failed", slog.String("err", errUp.Error()))
			return c.Status(400).JSON(ErrorResponse{
				Code:    "ERR_FILE_UPLOAD",
				Message: "Gambar bukti ditolak oleh sistem: " + errUp.Error(),
				Action:  "Pastikan ukuran dan format gambar sesuai.",
			})
		}
		imageUrl = url
	}

	ticketNumber := generateTicket()

	arg := db.CreateReportParams{
		TicketNumber:  ticketNumber,
		ReporterName:  name,
		ReporterEmail: email,
		ReporterPhone: phone,
		Content:       content,
		ProofLink:     pgtype.Text{String: proofLink, Valid: proofLink != ""},
		ProofImageUrl: pgtype.Text{String: imageUrl, Valid: imageUrl != ""},
	}

	_, err := h.Queries.CreateReport(ctx, arg)
	if err != nil {
		slog.Error("Create Report DB Error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Terjadi kendala saat menyimpan laporan Anda.",
			Action:  "Sistem kami sedang sibuk. Silakan tunggu beberapa menit dan coba lagi.",
		})
	}

	sendEmailAsync(email, name, ticketNumber, "PENDING")
	slog.Info("Report Created", slog.String("ticket", ticketNumber), slog.String("email", email))

	return c.Status(201).JSON(SuccessResponse{
		Pesan: "Laporan berhasil dikirim",
		Data:  TicketCreatedResponse{TicketNumber: ticketNumber},
	})
}

func (h *HoaxHandler) TrackReport(c fiber.Ctx) error {
	ticket := strings.ToUpper(strings.TrimSpace(c.Query("no_tiket")))
	if ticket == "" {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Nomor tiket wajib diisi.",
			Action:  "Masukkan nomor tiket yang dikirimkan ke email Anda.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetReportTrackingByTicket(ctx, ticket)
	if err != nil {
		slog.Warn("Report tracking not found", slog.String("ticket", ticket))
		return c.Status(404).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data laporan tidak ditemukan.",
			Action:  "Pastikan nomor tiket sudah diketik dengan benar.",
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
		response.NewsImage = pgTextToStr(data.NewsImage)
		response.CategoryName = pgTextToStr(data.CategoryName)
	}

	return c.JSON(SuccessResponse{Pesan: "Data Ditemukan", Data: response})
}

func (h *HoaxHandler) GetDashboardStats(c fiber.Ctx) error {
	cacheKey := "hoax:stats"
	if cached, found := cache.GlobalCache.Get(cacheKey); found {
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDashboardStats(ctx)
	if err != nil {
		slog.Error("GetDashboardStats DB Error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil statistik dasbor.",
			Action:  "Silakan muat ulang halaman.",
		})
	}

	var responseData []DashboardStatItem
	for _, row := range data {
		responseData = append(responseData, DashboardStatItem{
			CategoryID:   pgUUIDToStr(row.CategoryID),
			CategoryName: row.CategoryName,
			CategorySlug: row.CategorySlug,
			TotalNews:    row.TotalNews,
		})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: responseData}
	cache.GlobalCache.Set(cacheKey, res, CacheTTLList)
	return c.JSON(res)
}

func (h *HoaxHandler) GetPublicNews(c fiber.Ctx) error {
	pageStr, limitStr := c.Query("page", "1"), c.Query("limit", "10")
	page, limit := utils.ValidatePaginationParams(pageStr, limitStr)

	keyword, _ := utils.ValidateQueryString(c.Query("search"), 100, "search")

	cacheKey := fmt.Sprintf("hoax:news:public:q_%s:p%d:l%d", keyword, page, limit)
	if cached, found := cache.GlobalCache.Get(cacheKey); found {
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	offset := (page - 1) * limit
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var responseData []PublicNewsItem
	var total int64

	if keyword != "" {
		g.Go(func() error {
			data, err := h.Queries.SearchPublicNews(ctx, db.SearchPublicNewsParams{
				Keyword:    keyword,
				LimitData:  int32(limit),
				OffsetData: int32(offset),
			})
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
			return err
		})
		g.Go(func() error {
			var err error
			total, err = h.Queries.CountSearchPublicNews(ctx, keyword)
			return err
		})
	} else {
		g.Go(func() error {
			data, err := h.Queries.GetPublicNews(ctx, db.GetPublicNewsParams{
				LimitData:  int32(limit),
				OffsetData: int32(offset),
			})
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
			return err
		})
		g.Go(func() error {
			var err error
			total, err = h.Queries.CountPublicNews(ctx)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat daftar berita.",
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

	cache.GlobalCache.Set(cacheKey, res, CacheTTLList)
	return c.JSON(res)
}

func (h *HoaxHandler) GetPublicNewsDetailBySlug(c fiber.Ctx) error {
	newsSlug := c.Params("slug")

	if newsSlug == "" {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Slug berita tidak valid.",
			Action:  "Pastikan URL yang diakses benar.",
		})
	}

	cacheKey := fmt.Sprintf("hoax:news:detail:%s", newsSlug)
	if cached, found := cache.GlobalCache.Get(cacheKey); found {
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetNewsDetailBySlug(ctx, newsSlug)
	if err != nil {
		slog.Warn("News detail not found", slog.String("slug", newsSlug))
		return c.Status(404).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Berita yang Anda cari tidak ditemukan.",
			Action:  "Berita mungkin sudah dihapus atau link kadaluarsa.",
		})
	}

	response := PublicNewsDetail{
		ID:            pgUUIDToStr(data.ID),
		Title:         data.Title,
		Description:   data.Description,
		ReferenceLink: pgTextToStr(data.ReferenceLink),
		ImageUrl:      data.ImageUrl,
		CategoryName:  data.CategoryName,
		CategorySlug:  data.CategorySlug,
		PublishedAt:   data.PublishedAt.Time.Format("02 Jan 2006 15:04 WIB"),
	}

	res := SuccessResponse{
		Pesan: "Detail Berita",
		Data:  response,
	}

	cache.GlobalCache.Set(cacheKey, res, CacheTTLDetail)
	return c.JSON(res)
}