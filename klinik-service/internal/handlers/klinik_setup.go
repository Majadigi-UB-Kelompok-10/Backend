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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
		cache.GlobalCache.Set("hoax:stats", res, CacheTTLList)
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

		cacheKey := "hoax:news:public:q_:p1:l10"
		cache.GlobalCache.Set(cacheKey, resNews, CacheTTLList)
		slog.Info("Warmup Cache Halaman Pertama Berita Selesai")
	}
}