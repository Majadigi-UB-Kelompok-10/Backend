package routes

import (
	"time"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/farildzaky/sidita-service/internal/handlers"
)

func SetupRoutes(app *fiber.App, destinasiHandler *handlers.DestinasiHandler) {
	api := app.Group("/api/v1")

	actionLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Terlalu banyak request, sistem butuh nafas. Tunggu 1 menit ya!",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error { 
		return c.SendString("🏖️ API Sidita v1 Menyala Bossku!")
	})

	api.Get("/destinasi", destinasiHandler.ListDestinasi)
	api.Get("/destinasi/:slug", destinasiHandler.GetDetailDestinasi)
	api.Get("/areas", destinasiHandler.GetAllArea)
	api.Post("/destinasi", actionLimiter, destinasiHandler.CreateDestinasi)

}