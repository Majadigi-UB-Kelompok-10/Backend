package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"github.com/farildzaky/sidita-service/internal/handlers"
)

func SetupRoutes(app *fiber.App, destinasiHandler *handlers.DestinasiHandler, hotelHandler *handlers.HotelHandler, eventHandler *handlers.EventHandler) {
	api := app.Group("/api/v1")

	actionLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Terlalu banyak request. Silakan tunggu 1 menit.",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("API Sidita v1 Active")
	})

	api.Get("/areas", destinasiHandler.GetAllArea)

	api.Get("/destinasi/maps", destinasiHandler.GetDestinasiMaps)
	api.Get("/destinasi", destinasiHandler.ListDestinasi)
	api.Get("/destinasi/:slug", destinasiHandler.GetDetailDestinasi)
	api.Post("/destinasi", actionLimiter, destinasiHandler.CreateDestinasi)
	api.Put("/destinasi/:id", actionLimiter, destinasiHandler.UpdateDestinasi)
	api.Delete("/destinasi/:id", destinasiHandler.DeleteDestinasi)

	api.Get("/hotel/maps", hotelHandler.GetHotelMaps)
	api.Get("/hotel", hotelHandler.ListHotel)
	api.Get("/hotel/:slug", hotelHandler.GetDetailHotel)
	api.Post("/hotel", actionLimiter, hotelHandler.CreateHotel)
	api.Put("/hotel/:id", actionLimiter, hotelHandler.UpdateHotel)
	api.Delete("/hotel/:id", hotelHandler.DeleteHotel)

	api.Get("/event/maps", eventHandler.GetEventMaps)
	api.Get("/event", eventHandler.ListEvent)
	api.Get("/event/:slug", eventHandler.GetDetailEvent)
	api.Post("/event", actionLimiter, eventHandler.CreateEvent)
	api.Put("/event/:id", actionLimiter, eventHandler.UpdateEvent)
	api.Delete("/event/:id", eventHandler.DeleteEvent)
}