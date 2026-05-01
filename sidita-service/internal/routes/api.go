package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"github.com/farildzaky/sidita-service/internal/handlers"
)


func SetupRoutes(
	app *fiber.App,
	destinasiHandler *handlers.DestinasiHandler,
	hotelHandler *handlers.HotelHandler,
	eventHandler *handlers.EventHandler,
) {
	api := app.Group("/api/v1")

	// =========================================================================
	// LIMITERS
	// =========================================================================

	strictLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMITED",
				Message: "Terlalu banyak request admin. Silakan tunggu 1 menit.",
				Action:  "Mohon tunggu sebelum mencoba lagi",
			})
		},
	})

	searchLimiter := limiter.New(limiter.Config{
		Max:        50,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMITED",
				Message: "Terlalu banyak request pencarian. Silakan tunggu 1 menit.",
				Action:  "Mohon tunggu sebelum mencoba lagi",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(handlers.SuccessResponse{
			Pesan: "API Sidita v1 Active",
		})
	})

	// =========================================================================
	// PUBLIC GROUP — mobile app + web (read-only)
	// =========================================================================

	public := api.Group("/public")

	public.Get("/areas", destinasiHandler.GetAllArea)

	public.Get("/destinasi", searchLimiter, destinasiHandler.ListDestinasi)
	public.Get("/destinasi/maps", destinasiHandler.GetDestinasiMaps)
	public.Get("/destinasi/recommendation", destinasiHandler.GetRecommendationDestinasi)
	public.Get("/destinasi/:slug", destinasiHandler.GetDetailDestinasi)

	public.Get("/hotel", searchLimiter, hotelHandler.ListHotel)
	public.Get("/hotel/maps", hotelHandler.GetHotelMaps)
	public.Get("/hotel/recommendation", hotelHandler.GetRecommendationHotel)

	public.Get("/event", searchLimiter, eventHandler.ListEvent)
	public.Get("/event/maps", eventHandler.GetEventMaps)
	public.Get("/event/recommendation", eventHandler.GetRecommendationEvent)
	public.Get("/event/tahun-tersedia", eventHandler.GetTahunTersedia)
	public.Get("/event/:slug", eventHandler.GetDetailEvent)

	// =========================================================================
	// ADMIN GROUP — web admin (write + read full fields)
	// =========================================================================

	admin := api.Group("/admin")

	admin.Get("/destinasi/:id", destinasiHandler.GetDestinasiByID)
	admin.Post("/destinasi", strictLimiter, destinasiHandler.CreateDestinasi)
	admin.Put("/destinasi/:id", strictLimiter, destinasiHandler.UpdateDestinasi)
	admin.Delete("/destinasi/:id", strictLimiter, destinasiHandler.DeleteDestinasi)

	admin.Get("/hotel/:id", hotelHandler.GetHotelByID)
	admin.Post("/hotel", strictLimiter, hotelHandler.CreateHotel)
	admin.Put("/hotel/:id", strictLimiter, hotelHandler.UpdateHotel)
	admin.Delete("/hotel/:id", strictLimiter, hotelHandler.DeleteHotel)

	admin.Get("/event/:id", eventHandler.GetEventByID)
	admin.Post("/event", strictLimiter, eventHandler.CreateEvent)
	admin.Put("/event/:id", strictLimiter, eventHandler.UpdateEvent)
	admin.Delete("/event/:id", strictLimiter, eventHandler.DeleteEvent)
}