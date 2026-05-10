package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/farildzaky/sinaker-service/internal/cache"
	"github.com/farildzaky/sinaker-service/internal/db"
)

const (
	ContextQueryTimeout  = 5 * time.Second
	ContextUploadTimeout = 15 * time.Second
	maxImageSizeBytes    = 2 * 1024 * 1024 
)



type SinakerHandler struct {
	Queries *db.Queries
	DB      *pgxpool.Pool
	Cld     *cloudinary.Cloudinary
	Cache   cache.Cache
}

func NewSinakerHandler(q *db.Queries, dbPool *pgxpool.Pool, cld *cloudinary.Cloudinary, c cache.Cache) *SinakerHandler {
	return &SinakerHandler{
		Queries: q,
		DB:      dbPool,
		Cld:     cld,
		Cache:   c,
	}
}

// =====================================================================
// SETUP ROUTES
// =====================================================================

func (h *SinakerHandler) SetupRoutes(router fiber.Router) {
	sinaker := router.Group("/sinaker")
	public := sinaker.Group("/public")

	public.Get("/kota", h.GetKotaBLK)
	public.Get("/blk", h.GetListBLK)
	public.Get("/blk/:id/kejuruan", h.GetKejuruanByBLK)

	public.Post("/pendaftaran", h.SubmitPendaftaran)
	public.Post("/cek-status", h.CekStatus)
}

// =====================================================================
// IMAGE UPLOAD — Cloudinary
// =====================================================================

func (h *SinakerHandler) uploadImageReal(ctx context.Context, fileHeader *multipart.FileHeader, folderName string) (string, error) {
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
		return "", fmt.Errorf("format foto harus JPG/PNG/WEBP")
	}

	resCld, err := h.Cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folderName,
	})
	if err != nil {
		return "", fmt.Errorf("gagal upload ke server: %v", err)
	}
	return resCld.SecureURL, nil
}

// =====================================================================
// EMAIL — SendGrid Async
// =====================================================================

// sendEmailStatusAsync dipakai saat ada perubahan status pendaftaran (Pending -> Diterima/Ditolak)
func sendEmailStatusAsync(toEmail, namaLengkap, blkNama, kejuruanNama, status string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("email.panic_recovered", slog.Any("error", r), slog.String("email", toEmail))
			}
		}()

		apiKey := os.Getenv("SENDGRID_API_KEY")
		senderEmail := os.Getenv("SENDGRID_SENDER_EMAIL")

		if apiKey == "" || senderEmail == "" {
			slog.Warn("email.config_missing", slog.String("note", "SendGrid env belum diset, kirim email di-skip"))
			return
		}

		from := mail.NewEmail("Tim BLK Jawa Timur", senderEmail)
		subject := "Update Status Pendaftaran BLK Anda"
		to := mail.NewEmail(namaLengkap, toEmail)

		var color string
		if status == "diterima" {
			color = "green"
		} else if status == "ditolak" {
			color = "red"
		} else {
			color = "blue"
		}

		plainText := fmt.Sprintf("Halo %s, status pendaftaran Anda di %s untuk kejuruan %s saat ini adalah: %s.",
			namaLengkap, blkNama, kejuruanNama, strings.ToUpper(status))

		htmlContent := fmt.Sprintf(`
            <h3>Halo %s,</h3>
            <p>Berikut adalah update mengenai pendaftaran pelatihan vokasi Anda:</p>
            <ul>
                <li><b>Lokasi BLK:</b> %s</li>
                <li><b>Kejuruan:</b> %s</li>
            </ul>
            <p>Status Pendaftaran Anda saat ini: <b style="color:%s;">%s</b>.</p>
            <hr><small>Tim Disnakertrans Jawa Timur (Sinaker)</small>
        `, namaLengkap, blkNama, kejuruanNama, color, strings.ToUpper(status))

		message := mail.NewSingleEmail(from, subject, to, plainText, htmlContent)
		client := sendgrid.NewSendClient(apiKey)

		response, err := client.Send(message)
		switch {
		case err != nil:
			slog.Error("email.send_failed", slog.String("error", err.Error()), slog.String("to", toEmail))
		case response.StatusCode >= 400:
			slog.Error("email.rejected", slog.Int("status_code", response.StatusCode), slog.String("body", response.Body))
		default:
			slog.Info("email.sent", slog.String("to", toEmail), slog.String("status", status))
		}
	}()
}

// =====================================================================
// CACHE WARMUP (Untuk dipanggil saat server Start)
// =====================================================================
func (h *SinakerHandler) CacheWarmup() {
	slog.Info("sinaker.cache.warmup.start")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kotaList, err := h.Queries.GetDistinctKabKotaBLK(ctx)
	if err == nil {
		resKota := SuccessResponse{
			Pesan: "Berhasil mengambil data kota",
			Data:  kotaList,
		}
		cache.GlobalCache.Set("sinaker:public:blk:kota", resKota, 24*time.Hour)
	}

	// 2. Warmup List Semua BLK (Tanpa Filter)
	blks, err := h.Queries.ListAktifBLK(ctx, pgtype.Text{})
	if err == nil {
		resData := make([]BLKResponse, 0, len(blks))
		for _, b := range blks {
			var kec *string
			if b.Kecamatan.Valid {
				kec = &b.Kecamatan.String
			}
			var lat, lng *float64
			if b.Lat.Valid {
				latVal := b.Lat.Float64
				lat = &latVal
			}
			if b.Lng.Valid {
				lngVal := b.Lng.Float64
				lng = &lngVal
			}

			resData = append(resData, BLKResponse{
				ID:        b.ID,
				Nama:      b.Nama,
				Alamat:    b.Alamat,
				KabKota:   b.KabKota,
				Kecamatan: kec,
				Slug:      b.Slug,
				Lat:       lat,
				Lng:       lng,
			})
		}
		resBlk := SuccessResponse{
			Pesan: "Berhasil mengambil data BLK",
			Data:  resData,
		}
		cache.GlobalCache.Set("sinaker:public:blk:list", resBlk, 2*time.Hour)
	}

	slog.Info("sinaker.cache.warmup.success",
		slog.String("duration", time.Since(startTime).String()),
	)
}