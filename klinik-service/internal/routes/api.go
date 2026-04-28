package routes

import (
	"time"

	"github.com/farildzaky/klinik-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, hoaxHandler *handlers.HoaxHandler) {
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

	strictLimiter := limiter.New(limiter.Config{
		Max:        30,
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
		Max:        50,
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

	// ==========================================================================
	// 🌍 PUBLIC ENDPOINTS
	// ==========================================================================
	public := api.Group("/public")

	public.Get("/categories", hoaxHandler.GetCategories)
	public.Post("/reports", strictLimiter, hoaxHandler.SubmitReport)
	public.Get("/reports/track", searchLimiter, hoaxHandler.TrackReport)
	public.Get("/stats", hoaxHandler.GetDashboardStats)
	public.Get("/news", searchLimiter, hoaxHandler.GetPublicNews) 
	public.Get("/news/:slug", hoaxHandler.GetPublicNewsDetailBySlug)

	// ==========================================================================
	// 🛡️ ADMIN ENDPOINTS
	// ==========================================================================
	admin := api.Group("/admin")

	// Report Management
	admin.Get("/reports", hoaxHandler.GetAllReportsAdmin)
	admin.Post("/reports/process", strictLimiter, hoaxHandler.ProcessReportAdmin)
	admin.Post("/reports/reject", strictLimiter, hoaxHandler.RejectReportAdmin)

	// News Management (Hasil Recovery)
	admin.Get("/news", hoaxHandler.GetAllNewsAdmin)
	admin.Delete("/news/:id", strictLimiter, hoaxHandler.DeleteNewsAdmin)

	// Category Management (Hasil Recovery)
	admin.Post("/categories", strictLimiter, hoaxHandler.CreateCategoryAdmin)
}