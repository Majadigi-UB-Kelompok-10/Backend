package routes

import (
	"time"

	"github.com/farildzaky/jdih-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, jdihHandler *handlers.JdihHandler) {
	// ==========================================================
	// HEALTH CHECK
	// ==========================================================
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(200).JSON(struct {
			Status    string    `json:"status"`
			Timestamp time.Time `json:"timestamp"`
		}{
			Status:    "healthy",
			Timestamp: time.Now(),
		})
	})

	api := app.Group("/api/v1")

	// ==========================================================
	// RATE LIMITERS
	// ==========================================================
	
	// strictLimiter: Untuk Create/Delete (Admin/Upload PDF)
	strictLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",
				Message: "Terlalu banyak permintaan",
				Action:  "Sistem mendeteksi aktivitas berlebih. Silakan tunggu 1 menit.",
			})
		},
	})

	// searchLimiter: Untuk Pencarian & List Dokumen (Public)
	// Sengaja di-set 100 karena JDIH sangat read-heavy (warga sering buka-buka dokumen)
	searchLimiter := limiter.New(limiter.Config{
		Max:        100, 
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",
				Message: "Terlalu banyak pencarian dokumen",
				Action:  "Silakan tunggu 1 menit sebelum melakukan pencarian lagi.",
			})
		},
	})

	// ==========================================================
	// JDIH ROUTES
	// ==========================================================

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("JDIH API v1 Active")
	})

	// ----------------------------------------------------------
	// 1. PUBLIC ROUTES (Web / Mobile Warga)
	// ----------------------------------------------------------
	public := api.Group("/public")

	// Homepage
	public.Get("/pengumuman", jdihHandler.PublicGetPengumuman)
	
	// Pencarian Dokumen
	public.Get("/search", searchLimiter, jdihHandler.PublicSearchDokumen)
	
	// Detail Dokumen (View PDF + nambah View Counter)
	public.Get("/dokumen/detail/:id", searchLimiter, jdihHandler.PublicGetDetailDokumen)
	
	// Akses Cepat per Jenis
	public.Get("/dokumen/:jenis", searchLimiter, jdihHandler.PublicListDokumenByJenis)
	public.Get("/dokumen/:jenis/tahun", jdihHandler.PublicGetFilterTahun)

	// ----------------------------------------------------------
	// 2. ADMIN ROUTES (CMS JDIH)
	// ----------------------------------------------------------
	admin := api.Group("/admin")

	// Manajemen Dokumen Hukum
	admin.Get("/dokumen", jdihHandler.AdminListDokumen)
	admin.Post("/dokumen", strictLimiter, jdihHandler.AdminCreateDokumen)
	admin.Delete("/dokumen/:id", strictLimiter, jdihHandler.AdminDeleteDokumen)

	// Manajemen Pengumuman
	admin.Post("/pengumuman", strictLimiter, jdihHandler.AdminCreatePengumuman)
}