package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/farildzaky/klinik-service/internal/cache"
	"github.com/farildzaky/klinik-service/internal/db"
	"github.com/farildzaky/klinik-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	ContextQueryTimeout  = 5 * time.Second
	ContextUploadTimeout = 15 * time.Second
)

type HoaxHandler struct {
	Queries *db.Queries
	DB      *pgxpool.Pool
	Cld     *cloudinary.Cloudinary
}

func NewHoaxHandler(q *db.Queries, db *pgxpool.Pool, cld *cloudinary.Cloudinary) *HoaxHandler {
	return &HoaxHandler{Queries: q, DB: db, Cld: cld}
}

func pgTextToStr(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func pgUUIDToStr(u pgtype.UUID) string {
	if u.Valid {
		return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
	}
	return ""
}

func parseUUID(id string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(id)
	return u, err
}

func generateTicket() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return fmt.Sprintf("KLX-%s", strings.ToUpper(hex.EncodeToString(bytes)))
}

func (h *HoaxHandler) uploadImageReal(ctx context.Context, fileHeader *multipart.FileHeader, folderName string) (string, error) {
	if h.Cld == nil {
		return "", fmt.Errorf("layanan penyimpanan gambar belum terkonfigurasi")
	}
	if fileHeader.Size > 2*1024*1024 {
		return "", fmt.Errorf("ukuran gambar maksimal 2MB")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file gambar")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	file.Read(buffer)
	file.Seek(0, 0)
	mimeType := http.DetectContentType(buffer)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return "", fmt.Errorf("format gambar harus JPG/PNG/WEBP")
	}

	resCld, errUpload := h.Cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folderName,
	})
	if errUpload != nil {
		return "", fmt.Errorf("gagal upload ke server: %v", errUpload)
	}

	return resCld.SecureURL, nil
}

func sendEmailAsync(toEmail, name, ticket, status string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Email panic recovered", slog.Any("error", r), slog.String("ticket", ticket))
			}
		}()

		apiKey := os.Getenv("SENDGRID_API_KEY")
		senderEmail := os.Getenv("SENDGRID_SENDER_EMAIL")

		if apiKey == "" || senderEmail == "" {
			slog.Warn("SendGrid belum disetting di .env, bypass kirim email", slog.String("ticket", ticket))
			return
		}

		from := mail.NewEmail("Tim Klinik Hoaks", senderEmail)
		subject := "Update Laporan Klinik Hoaks Anda - " + ticket
		to := mail.NewEmail(name, toEmail)
		
		plainTextContent := fmt.Sprintf("Halo %s, laporan Anda (%s) saat ini berstatus: %s.", name, ticket, status)
		htmlContent := fmt.Sprintf(`
			<h3>Halo %s,</h3>
			<p>Laporan Anda dengan nomor tiket <b>%s</b> saat ini berstatus: <b style="color:blue;">%s</b>.</p>
			<p>Terima kasih telah berkontribusi!</p>
			<hr><small>Tim Klinik Hoaks</small>
		`, name, ticket, status)

		message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		client := sendgrid.NewSendClient(apiKey)
		
		response, err := client.Send(message)
		if err != nil {
			slog.Error("Gagal mengirim email via SendGrid", slog.String("error", err.Error()), slog.String("to", toEmail))
		} else if response.StatusCode >= 400 {
			slog.Error("SendGrid menolak email", slog.Int("status_code", response.StatusCode), slog.String("body", response.Body))
		} else {
			slog.Info("Email SendGrid terkirim", slog.String("to", toEmail), slog.String("ticket", ticket))
		}
	}()
}

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

	// 3. Format Response
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

	// Simpan ke Cache selama 1 Jam
	cache.GlobalCache.Set(cacheKey, res, 1*time.Hour)
	return c.JSON(res)
}


// =====================================================================
// 1. PUBLIC API
// =====================================================================

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
	cache.GlobalCache.Set(cacheKey, res, 5*time.Minute)
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

	cache.GlobalCache.Set(cacheKey, res, 2*time.Minute)
	return c.JSON(res)
}

func (h *HoaxHandler) CacheWarmup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("Memulai proses Cache Warmup...")

	statsData, err := h.Queries.GetDashboardStats(ctx)
	if err == nil {
		var responseData []DashboardStatItem
		for _, row := range statsData {
			responseData = append(responseData, DashboardStatItem{
				CategoryID:   pgUUIDToStr(row.CategoryID),
				CategoryName: row.CategoryName,
				CategorySlug: row.CategorySlug,
				TotalNews:    row.TotalNews,
			})
		}
		res := SuccessResponse{Pesan: "Sukses", Data: responseData}
		cache.GlobalCache.Set("hoax:stats", res, 5*time.Minute)
		slog.Info("Warmup Cache Statistik Dasbor Selesai")
	}

	newsData, errNews := h.Queries.GetPublicNews(ctx, db.GetPublicNewsParams{
		LimitData:  10,
		OffsetData: 0,
	})
	totalNews, errCount := h.Queries.CountPublicNews(ctx)
	
	if errNews == nil && errCount == nil {
		var responseNews []PublicNewsItem
		for _, row := range newsData {
			responseNews = append(responseNews, PublicNewsItem{
				ID:           pgUUIDToStr(row.ID),
				Title:        row.Title,
				Slug:         row.Slug,
				ImageUrl:     row.ImageUrl,
				CategoryName: row.CategoryName,
				CategorySlug: row.CategorySlug,
				PublishedAt:  row.PublishedAt.Time.Format("02 Jan 2006"),
			})
		}
		if responseNews == nil {
			responseNews = []PublicNewsItem{}
		}

		resNews := SuccessResponse{
			Pesan:      "Sukses",
			Data:       responseNews,
			Pagination: &PaginationMeta{Page: 1, Limit: 10, Total: totalNews},
		}
		
		cacheKey := "hoax:news:public:p1:l10"
		cache.GlobalCache.Set(cacheKey, resNews, 2*time.Minute)
		slog.Info("Warmup Cache Halaman Pertama Berita Selesai")
	}
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

	cache.GlobalCache.Set(cacheKey, res, 10*time.Minute)
	return c.JSON(res)
}

// =====================================================================
// 2. ADMIN API
// =====================================================================

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
	reason := c.FormValue("reason") // Admin wajib mengisi alasan penolakan

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