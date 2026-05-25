package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"github.com/farildzaky/bansos-service/internal/handlers"
)

func SetupRoutes(app *fiber.App, bansosHandler *handlers.BansosHandler) {
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

	// strictLimiter: Untuk Create/Update/Delete (Admin)
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

	// searchLimiter: Untuk Cek NIK & Detail Bansos (Public)
	searchLimiter := limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",
				Message: "Terlalu banyak pencarian",
				Action:  "Silakan tunggu sebentar sebelum melakukan pengecekan lagi.",
			})
		},
	})

	// ==========================================================
	// BANSOS ROUTES
	// ==========================================================

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("BANSOS API v1 Active")
	})

	// ----------------------------------------------------------
	// 1. PUBLIC ROUTES (Web / Mobile Warga)
	// ----------------------------------------------------------
	public := api.Group("/public")

	// Cek Bansos berdasarkan NIK (Info Bansos)
	// Contoh: GET /api/v1/public/cek-bansos?nik=1234567890123456
	public.Get("/cek-bansos", searchLimiter, bansosHandler.PublicCekBansos)

	// Detail Penyaluran Bansos
	public.Get("/bansos/detail/:id", searchLimiter, bansosHandler.PublicGetDetailBansos)

	// ----------------------------------------------------------
	// 2. ADMIN ROUTES (CMS / Dashboard Admin)
	// ----------------------------------------------------------
	admin := api.Group("/admin")

	// A. Manajemen Program Bansos (PKH, BPNT, dll)
	admin.Get("/program", bansosHandler.AdminListProgram)
	admin.Post("/program", strictLimiter, bansosHandler.AdminCreateProgram)

	// B. Manajemen Penerima (Warga yang terdaftar NIK-nya)
	admin.Get("/penerima", bansosHandler.AdminListPenerima)
	admin.Post("/penerima", strictLimiter, bansosHandler.AdminCreatePenerima)

	// C. Manajemen Penyaluran Bansos (History pembagian bansos)
	admin.Get("/penyaluran", bansosHandler.AdminListPenyaluran)
	admin.Post("/penyaluran", strictLimiter, bansosHandler.AdminCreatePenyaluran)
	admin.Patch("/penyaluran/:id/status", strictLimiter, bansosHandler.AdminUpdateStatusPenyaluran)
	admin.Delete("/penyaluran/:id", strictLimiter, bansosHandler.AdminDeletePenyaluran)
}