package routes

import (
	"time"

	"github.com/farildzaky/rssa-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, h *handlers.RSSAHandler) {
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
			Pesan: "API RSSA v1 Active",
		})
	})

	// =========================================================================
	// PUBLIC GROUP 
	// =========================================================================

	public := api.Group("/public")

	public.Get("/summary", searchLimiter, h.GetHeaderSummary)
	public.Get("/kelas", searchLimiter, h.GetMasterKelas)
	public.Get("/ruangan", searchLimiter, h.GetListRuangan)

	// =========================================================================
	// ADMIN GROUP 
	// =========================================================================

	admin := api.Group("/admin")

	admin.Post("/ruangan", actionLimiter, h.CreateRuanganAdmin)
	admin.Put("/ruangan/:id", actionLimiter, h.UpdateRuanganAdmin)
	admin.Delete("/ruangan/:id", actionLimiter, h.DeleteRuanganAdmin)
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