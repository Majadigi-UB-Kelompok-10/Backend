package routes

import (
	"time"

	"github.com/farildzaky/transjatim-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, h *handlers.TransJatimHandler) {
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
		LimitReached: rateLimitResponse("Terlalu banyak request simpan data. Silakan tunggu 1 menit."),
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(handlers.SuccessResponse{
			Pesan: "API Trans Jatim v1 Active",
		})
	})

	// =========================================================================
	// PUBLIC GROUP (Aplikasi Warga)
	// =========================================================================

	public := api.Group("/public")

	// Pencarian & Info Warga
	public.Get("/terminals", searchLimiter, h.GetActiveTerminals)
	public.Get("/jadwal/search", searchLimiter, h.SearchJadwal)
	public.Get("/jadwal/:id", searchLimiter, h.GetJadwalDetail)
	public.Get("/harga", searchLimiter, h.GetInformasiHarga)

	// =========================================================================
	// ADMIN GROUP (Dashboard Petugas Trans Jatim)
	// =========================================================================

	admin := api.Group("/admin")

	// 1. TERMINAL
	admin.Get("/terminals", searchLimiter, h.GetAllTerminalsAdmin)
	admin.Post("/terminals", actionLimiter, h.CreateTerminal)
	admin.Put("/terminals/:slug", actionLimiter, h.UpdateTerminal)
	admin.Patch("/terminals/:slug/deactivate", actionLimiter, h.DeactivateTerminal)

	// 2. BUS
	admin.Get("/buses", searchLimiter, h.GetAllBusesAdmin)
	admin.Post("/buses", actionLimiter, h.CreateBus)
	admin.Put("/buses/:id", actionLimiter, h.UpdateBus)
	admin.Delete("/buses/:id", actionLimiter, h.DeleteBus)

	// 3. RUTE
	admin.Get("/rute", searchLimiter, h.GetAllRutesAdmin)
	admin.Post("/rute", actionLimiter, h.CreateRute)
	admin.Put("/rute/:slug", actionLimiter, h.UpdateRute)
	admin.Patch("/rute/:slug/deactivate", actionLimiter, h.DeactivateRute)
	admin.Post("/rute/:slug/generate-geometry", actionLimiter, h.AdminGenerateRuteGeometry)

	// 4. JADWAL
	admin.Get("/jadwal", searchLimiter, h.GetAllJadwalsAdmin)
	admin.Post("/jadwal", actionLimiter, h.CreateJadwal)
	admin.Put("/jadwal/:id", actionLimiter, h.UpdateJadwal)
	admin.Delete("/jadwal/:id", actionLimiter, h.DeleteJadwal)

	// 5. HARGA
	admin.Get("/harga", searchLimiter, h.GetAllHargaAdmin)
	admin.Post("/harga", actionLimiter, h.CreateHarga)
	admin.Put("/harga/:id", actionLimiter, h.UpdateHarga)
	admin.Delete("/harga/:id", actionLimiter, h.DeleteHarga)
}

func rateLimitResponse(message string) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
			Code:    "ERR_RATE_LIMIT",
			Message: message,
			Action:  "Sistem mendeteksi lonjakan trafik. Silakan tunggu sebentar sebelum mencoba lagi.",
		})
	}
}