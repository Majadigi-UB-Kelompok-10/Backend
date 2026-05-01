package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/farildzaky/klinik-service/internal/cache"
	"github.com/farildzaky/klinik-service/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/sync/errgroup" 
)

const (
	ContextQueryTimeout  = 5 * time.Second
	ContextUploadTimeout = 15 * time.Second

	maxImageSizeBytes = 2 * 1024 * 1024 
)

type HoaxHandler struct {
	Queries *db.Queries
	DB      *pgxpool.Pool
	Cld     *cloudinary.Cloudinary
}

func NewHoaxHandler(q *db.Queries, dbPool *pgxpool.Pool, cld *cloudinary.Cloudinary) *HoaxHandler {
	return &HoaxHandler{Queries: q, DB: dbPool, Cld: cld}
}

// =============================================================================
// PGTYPE HELPERS — convert pgx types ↔ Go strings
// =============================================================================

func pgTextToStr(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func pgUUIDToStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func parseUUID(id string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(id)
	return u, err
}

// =============================================================================
// IMAGE UPLOAD — Cloudinary
// =============================================================================

func (h *HoaxHandler) uploadImageReal(ctx context.Context, fileHeader *multipart.FileHeader, folderName string) (string, error) {
	if h.Cld == nil {
		return "", fmt.Errorf("layanan penyimpanan gambar belum terkonfigurasi")
	}
	if fileHeader.Size > maxImageSizeBytes {
		return "", fmt.Errorf("ukuran gambar maksimal 2MB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file gambar")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", fmt.Errorf("gagal membaca file gambar")
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("gagal reset pointer file")
	}

	mimeType := http.DetectContentType(buffer)
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return "", fmt.Errorf("format gambar harus JPG/PNG/WEBP")
	}

	resCld, err := h.Cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folderName,
	})
	if err != nil {
		return "", fmt.Errorf("gagal upload ke server: %v", err)
	}
	return resCld.SecureURL, nil
}

// =============================================================================
// EMAIL — SendGrid async with panic recovery
// =============================================================================

func sendEmailAsync(toEmail, name, ticket, status string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("email.panic_recovered",
					slog.Any("error", r),
					slog.String("ticket", ticket),
				)
			}
		}()

		apiKey := os.Getenv("SENDGRID_API_KEY")
		senderEmail := os.Getenv("SENDGRID_SENDER_EMAIL")

		if apiKey == "" || senderEmail == "" {
			slog.Warn("email.config_missing",
				slog.String("ticket", ticket),
				slog.String("note", "SendGrid env belum diset, kirim email di-skip"),
			)
			return
		}

		from := mail.NewEmail("Tim Klinik Hoaks", senderEmail)
		subject := "Update Laporan Klinik Hoaks Anda - " + ticket
		to := mail.NewEmail(name, toEmail)

		plainText := fmt.Sprintf("Halo %s, laporan Anda (%s) saat ini berstatus: %s.",
			name, ticket, status)

		htmlContent := fmt.Sprintf(`
            <h3>Halo %s,</h3>
            <p>Laporan Anda dengan nomor tiket <b>%s</b> saat ini berstatus: <b style="color:blue;">%s</b>.</p>
            <p>Terima kasih telah berkontribusi!</p>
            <hr><small>Tim Klinik Hoaks</small>
        `, name, ticket, status)

		message := mail.NewSingleEmail(from, subject, to, plainText, htmlContent)
		client := sendgrid.NewSendClient(apiKey)

		response, err := client.Send(message)
		switch {
		case err != nil:
			slog.Error("email.send_failed",
				slog.String("error", err.Error()),
				slog.String("to", toEmail),
			)
		case response.StatusCode >= 400:
			slog.Error("email.rejected",
				slog.Int("status_code", response.StatusCode),
				slog.String("body", response.Body),
			)
		default:
			slog.Info("email.sent",
				slog.String("to", toEmail),
				slog.String("ticket", ticket),
			)
		}
	}()
}

// =============================================================================
// CACHE WARMUP 
// =============================================================================

func (h *HoaxHandler) CacheWarmup() {
	slog.Info("cache.warmup.start")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		statsData, err := h.Queries.GetDashboardStats(gCtx)
		if err != nil {
			return fmt.Errorf("warmup stats: %w", err)
		}

		responseData := make([]DashboardStatItem, 0, len(statsData))
		for _, row := range statsData {
			responseData = append(responseData, DashboardStatItem{
				CategoryID:   pgUUIDToStr(row.CategoryID),
				CategoryName: row.CategoryName,
				CategorySlug: row.CategorySlug,
				IconUrl:      pgTextToStr(row.IconUrl),
				TotalNews:    row.TotalNews,
			})
		}
		cache.GlobalCache.Set(
			"hoax:stats",
			SuccessResponse{Pesan: "Sukses", Data: responseData},
			CacheTTLList,
		)
		return nil
	})

	g.Go(func() error {
		newsData, errNews := h.Queries.GetPublicNews(gCtx, db.GetPublicNewsParams{
			LimitData:  10,
			OffsetData: 0,
		})
		if errNews != nil {
			return fmt.Errorf("warmup news data: %w", errNews)
		}

		totalNews, errCount := h.Queries.CountPublicNews(gCtx)
		if errCount != nil {
			return fmt.Errorf("warmup news count: %w", errCount)
		}

		responseNews := make([]PublicNewsItem, 0, len(newsData))
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

		res := SuccessResponse{
			Pesan:      "Sukses",
			Data:       responseNews,
			Pagination: &PaginationMeta{Page: 1, Limit: 10, Total: totalNews},
		}
		cache.GlobalCache.Set("hoax:news:public:q_:p1:l10", res, CacheTTLList)
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Warn("cache.warmup.partial_failure",
			slog.String("err", err.Error()),
			slog.String("note", "service tetap jalan, cache akan diisi on-demand"),
		)
		return
	}

	slog.Info("cache.warmup.success",
		slog.String("duration", time.Since(startTime).String()),
	)
}