package routes

import (
	"time"

	"github.com/farildzaky/siskaperbapo-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, bpHandler *handlers.BahanPokokHandler) {
	api := app.Group("/api/v1")

	postLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Terlalu banyak request simpan data, tunggu sebentar.",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("API Siskaperbapo Menyala Bossku!")
	})

	api.Get("/bahan-pokok", bpHandler.GetAllBahanPokok)
	api.Get("/bahan-pokok/:slug", bpHandler.GetDetailBahanPokok)
	api.Get("/areas", bpHandler.GetAllAreas)
	api.Post("/bahan-pokok", postLimiter, bpHandler.CreateBahanPokok)
	api.Post("/harga-harian", postLimiter, bpHandler.CreateHargaHarian)
	api.Put("/harga-harian/:id", postLimiter, bpHandler.UpdateHargaHarian)
	api.Delete("/harga-harian/:id", postLimiter, bpHandler.DeleteHargaHarian)
	api.Put("/bahan-pokok/:id", postLimiter, bpHandler.UpdateBahanPokok)
	api.Delete("/bahan-pokok/:id", postLimiter, bpHandler.DeleteBahanPokok)

	api.Get("/siskaperbapo-page", bpHandler.GetSduiMainData)
}
