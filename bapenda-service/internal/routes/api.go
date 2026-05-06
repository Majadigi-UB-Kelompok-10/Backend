package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"github.com/farildzaky/bapenda-service/internal/handlers"
)

func SetupRoutes(app *fiber.App, h *handlers.BapendaHandler) {
	api := app.Group("/api/v1")



	// Search/read endpoints: more permissive (50/min)
	searchLimiter := limiter.New(limiter.Config{
		Max:        50,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: rateLimitResponse,
	})

	// Mutation endpoints (POST/PUT/DELETE): stricter (30/min)
	actionLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: rateLimitResponse,
	})



	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(handlers.SuccessResponse{
			Pesan: "Bapenda API v1 Active",
		})
	})

	// =========================================================================
	// PUBLIC: CEK PAJAK
	// =========================================================================

	api.Post("/pajak/info", actionLimiter, h.GetInfoPajak)

	// =========================================================================
	// PUBLIC: NJKB / KALKULATOR PAJAK
	// =========================================================================

	njkb := api.Group("/njkb")
	njkb.Get("/jenis", searchLimiter, h.GetDropdownJenis)
	njkb.Get("/merk", searchLimiter, h.GetDropdownMerk)
	njkb.Get("/model", searchLimiter, h.GetDropdownModel)
	njkb.Get("/tipe", searchLimiter, h.GetDropdownTipe)
	njkb.Get("/tahun", searchLimiter, h.GetDropdownTahun)
	njkb.Post("/kalkulasi", actionLimiter, h.HitungKalkulasiNJKB)



	admin := api.Group("/admin")

	admin.Get("/pajak", searchLimiter, h.GetAllInfoPajakAdmin)
	admin.Get("/pajak/:id", h.GetDetailPajakAdmin)
	admin.Post("/pajak", actionLimiter, h.CreateInfoPajak)
	admin.Put("/pajak/:id", actionLimiter, h.UpdateInfoPajak)
	admin.Delete("/pajak/:id", actionLimiter, h.DeleteInfoPajak)

	admin.Get("/njkb", searchLimiter, h.GetAllMasterNjkbAdmin)
	admin.Get("/njkb/:id", searchLimiter, h.GetDetailMasterNjkbAdmin)
	admin.Post("/njkb", actionLimiter, h.CreateMasterNjkb)
	admin.Put("/njkb/:id", actionLimiter, h.UpdateMasterNjkb)
	admin.Delete("/njkb/:id", actionLimiter, h.DeleteMasterNjkb)
}

func rateLimitResponse(c fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
		Code:    "ERR_RATE_LIMIT",
		Message: "Terlalu banyak request",
		Action:  "Silakan tunggu 1 menit sebelum mencoba lagi",
	})
}