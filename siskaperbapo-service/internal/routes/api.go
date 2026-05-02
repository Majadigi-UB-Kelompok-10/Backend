package routes

import (
	"time"

	"github.com/farildzaky/siskaperbapo-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, handler *handlers.SiskaperbapoHandler) {
	api := app.Group("/api/v1")



	searchLimiter := limiter.New(limiter.Config{
		Max:        50,
		Expiration: 1 * time.Minute,
		LimitReached: rateLimitResponse("Terlalu banyak request pencarian. Silakan tunggu 1 menit."),
	})

	actionLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: rateLimitResponse("Terlalu banyak request simpan data. Silakan tunggu 1 menit."),
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(handlers.SuccessResponse{
			Pesan: "API Siskaperbapo v1 Active",
		})
	})

	public := api.Group("/public")

	public.Get("/areas", searchLimiter, handler.GetAllAreas)
	public.Get("/bahan-pokok", searchLimiter, handler.GetAllBahanPokok)
	public.Get("/bahan-pokok/:slug", searchLimiter, handler.GetDetailBahanPokok)
	
	public.Get("/sdui-page", searchLimiter, handler.GetSduiMainData)


	admin := api.Group("/admin")

	admin.Post("/bahan-pokok", actionLimiter, handler.CreateBahanPokok)
	admin.Put("/bahan-pokok/:id", actionLimiter, handler.UpdateBahanPokok)
	admin.Delete("/bahan-pokok/:id", actionLimiter, handler.DeleteBahanPokok)

	admin.Post("/harga-harian", actionLimiter, handler.CreateHargaHarian)
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