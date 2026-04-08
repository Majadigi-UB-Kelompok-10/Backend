package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/farildzaky/klinik-service/internal/handlers"
)

func SetupRoutes(app *fiber.App, hoaxHandler *handlers.HoaxHandler) {
	api := app.Group("/api/v1")


	strictLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",	
				Message: "Terlalu banyak permintaan",
				Action:  "Sistem mendeteksi aktivitas spam. Silakan tunggu 1 menit.",
			})
		},
	})

	searchLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",	
				Message: "Terlalu banyak pencarian",
				Action:  "Silakan tunggu 1 menit sebelum melakukan pelacakan/pencarian lagi.",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Klinik Hoaks API v1 Active")
	})

	public := api.Group("/public")

	public.Get("/categories", hoaxHandler.GetCategories)

	public.Post("/reports", strictLimiter, hoaxHandler.SubmitReport)
	
	public.Get("/reports/track", searchLimiter, hoaxHandler.TrackReport)

	public.Get("/stats", hoaxHandler.GetDashboardStats)

	public.Get("/news", hoaxHandler.GetPublicNews)
	public.Get("/news/:slug", hoaxHandler.GetPublicNewsDetailBySlug)

	// ==========================================================================
	// ADMIN ENDS

	admin := api.Group("/admin")

	admin.Get("/reports", hoaxHandler.GetAllReportsAdmin)
	
	admin.Post("/reports/process", strictLimiter, hoaxHandler.ProcessReportAdmin)

	admin.Post("/reports/reject", strictLimiter, hoaxHandler.RejectReportAdmin)
}