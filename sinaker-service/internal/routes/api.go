package routes

import (
	"time"

	"github.com/farildzaky/sinaker-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, h *handlers.SinakerHandler) {
	api := app.Group("/api/v1")

	// =========================================================================
	// RATE LIMITERS (Proteksi Anti-Spam / DDoS)
	// =========================================================================

	searchLimiter := limiter.New(limiter.Config{
		Max:          50,
		Expiration:   1 * time.Minute,
		LimitReached: rateLimitResponse("Terlalu banyak request pencarian. Silakan tunggu 1 menit."),
	})

	actionLimiter := limiter.New(limiter.Config{
		Max:          30,
		Expiration:   1 * time.Minute,
		LimitReached: rateLimitResponse("Terlalu banyak request simpan/update data. Silakan tunggu 1 menit."),
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(handlers.SuccessResponse{
			Pesan: "API Sinaker v1 Active",
		})
	})



	// =====================================================================
	// PUBLIC ROUTES (Aplikasi Warga / Mobile)
	// =====================================================================
	public := api.Group("/public")

	// BLK & Kejuruan (Read-only -> searchLimiter)
	public.Get("/kota", searchLimiter, h.GetKotaBLK)
	public.Get("/blk", searchLimiter, h.GetListBLK)
	public.Get("/blk/:id/kejuruan", searchLimiter, h.GetKejuruanByBLK)

	// Pendaftaran & Cek Status
	public.Post("/pendaftaran", actionLimiter, h.SubmitPendaftaran) 
	public.Post("/cek-status", searchLimiter, h.CekStatus)          

	// =====================================================================
	// ADMIN ROUTES (Dashboard CMS)
	// =====================================================================
	admin := api.Group("/admin")

	// Kelola BLK
	admin.Get("/blk", searchLimiter, h.AdminGetListBLK)
	admin.Post("/blk", actionLimiter, h.AdminCreateBLK)

	// Kelola Pendaftaran & Status
	admin.Get("/pendaftaran", searchLimiter, h.AdminGetListPendaftaran)
	admin.Patch("/pendaftaran/:id/status", actionLimiter, h.AdminUpdateStatusPendaftaran)
}

// rateLimitResponse adalah helper untuk menyeragamkan format JSON saat kena limit
func rateLimitResponse(message string) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
			Code:    "ERR_RATE_LIMIT",
			Message: message,
			Action:  "Sistem mendeteksi lonjakan trafik. Silakan tunggu sebentar sebelum mencoba lagi.",
		})
	}
}