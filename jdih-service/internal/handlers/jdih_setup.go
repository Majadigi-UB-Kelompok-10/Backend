package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/farildzaky/jdih-service/internal/cache"
	"github.com/farildzaky/jdih-service/internal/db"
)

const (
	ContextQueryTimeout  = 5 * time.Second
	ContextUploadTimeout = 30 * time.Second
	maxPdfSizeBytes      = 10 * 1024 * 1024 
)

// =====================================================================
// STRUCT & CONSTRUCTOR
// =====================================================================

type JdihHandler struct {
	Queries *db.Queries
	DB      *pgxpool.Pool
	Cld     *cloudinary.Cloudinary
	Cache   cache.Cache
}

func NewJdihHandler(q *db.Queries, dbPool *pgxpool.Pool, cld *cloudinary.Cloudinary, c cache.Cache) *JdihHandler {
	return &JdihHandler{
		Queries: q,
		DB:      dbPool,
		Cld:     cld,
		Cache:   c,
	}
}

// =====================================================================
// SETUP ROUTES
// =====================================================================

func (h *JdihHandler) SetupRoutes(router fiber.Router) {
	jdih := router.Group("/jdih")

	public := jdih.Group("/public")
	public.Get("/pengumuman", h.PublicGetPengumuman)
	public.Get("/search", h.PublicSearchDokumen)
	public.Get("/dokumen/:jenis", h.PublicListDokumenByJenis)
	public.Get("/dokumen/:jenis/tahun", h.PublicGetFilterTahun)
	public.Get("/dokumen/detail/:id", h.PublicGetDetailDokumen)


}

// =====================================================================
// PDF UPLOAD — Cloudinary (Khusus JDIH)
// =====================================================================

func (h *JdihHandler) uploadPDFReal(ctx context.Context, fileHeader *multipart.FileHeader, folderName string) (string, int32, error) {
	if h.Cld == nil {
		return "", 0, fmt.Errorf("layanan penyimpanan dokumen belum terkonfigurasi")
	}
	if fileHeader.Size > maxPdfSizeBytes {
		return "", 0, fmt.Errorf("ukuran dokumen maksimal 10MB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", 0, fmt.Errorf("gagal membuka file dokumen")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", 0, fmt.Errorf("gagal membaca file dokumen")
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", 0, fmt.Errorf("gagal reset pointer file")
	}

	mimeType := http.DetectContentType(buffer)
	if mimeType != "application/pdf" {
		return "", 0, fmt.Errorf("format file harus berupa PDF")
	}

	sizeKb := int32(fileHeader.Size / 1024) 

	resCld, err := h.Cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:       folderName,
		ResourceType: "raw", 
	})
	if err != nil {
		return "", 0, fmt.Errorf("gagal upload ke server: %v", err)
	}

	return resCld.SecureURL, sizeKb, nil
}

// =====================================================================
// CACHE WARMUP (Dipanggil saat server Start)
// =====================================================================
func (h *JdihHandler) CacheWarmup() {
	slog.Info("jdih.cache.warmup.start")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pengumuman, err := h.Queries.ListPengumumanTerbaru(ctx, 3) 
	if err == nil {
		resData := make([]PengumumanResponse, 0, len(pengumuman))
		for _, p := range pengumuman {
			var isiStr string
			if p.Isi.Valid {
				isiStr = p.Isi.String
			}
			resData = append(resData, PengumumanResponse{
				ID:      p.ID,
				Judul:   p.Judul,
				Isi:     isiStr,
				Tanggal: p.CreatedAt.Time.Format("02 Jan 2006"),
			})
		}

		resPengumuman := SuccessResponse{
			Pesan: "Berhasil mengambil pengumuman terbaru",
			Data:  resData,
		}

		h.cacheJSONToWarmup("jdih:public:pengumuman", 24*time.Hour, resPengumuman)
	}

	slog.Info("jdih.cache.warmup.success",
		slog.String("duration", time.Since(startTime).String()),
	)
}

func (h *JdihHandler) cacheJSONToWarmup(key string, ttl time.Duration, data interface{}) {
	if b, err := sonic.Marshal(data); err == nil {
		h.Cache.Set(key, b, ttl)
	}
}