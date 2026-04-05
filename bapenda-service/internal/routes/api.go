package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/farildzaky/bapenda-service/internal/handlers"
)

func SetupRoutes(app *fiber.App, bapendaHandler *handlers.BapendaHandler) {
	api := app.Group("/api/v1")

	actionLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Terlalu banyak request. Silakan tunggu 1 menit.",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Bapenda API v1 Active")
	})

	// ==========================================
	// 1. INFO PAJAK
	// ==========================================
	api.Post("/pajak/info", actionLimiter, bapendaHandler.GetInfoPajak)

	// ==========================================
	// 2. NILAI JUAL 
	// ==========================================
	njkb := api.Group("/njkb")
	njkb.Get("/jenis", bapendaHandler.GetDropdownJenis)
	njkb.Get("/merk", bapendaHandler.GetDropdownMerk)
	njkb.Get("/model", bapendaHandler.GetDropdownModel)
	njkb.Get("/tipe", bapendaHandler.GetDropdownTipe)
	njkb.Get("/tahun", bapendaHandler.GetDropdownTahun)
	
	njkb.Post("/kalkulasi", actionLimiter, bapendaHandler.HitungKalkulasiNJKB)

	// ==========================================
	// 3. ADMIN ENDPOINTS 
	// ==========================================
	admin := api.Group("/admin")
    
    admin.Get("/pajak", bapendaHandler.GetAllInfoPajakAdmin)      
    admin.Get("/pajak/:plat", bapendaHandler.GetDetailPajakAdmin) 
    admin.Post("/pajak", bapendaHandler.CreateInfoPajak)          
    admin.Put("/pajak/:plat", bapendaHandler.UpdateInfoPajak)     
    admin.Delete("/pajak/:plat", bapendaHandler.DeleteInfoPajak)  
    
	admin.Get("/njkb", bapendaHandler.GetAllMasterNjkbAdmin)
    admin.Get("/njkb/:id", bapendaHandler.GetDetailMasterNjkbAdmin)
    admin.Post("/njkb", bapendaHandler.CreateMasterNjkb)
    admin.Put("/njkb/:id", bapendaHandler.UpdateMasterNjkb)
    admin.Delete("/njkb/:id", bapendaHandler.DeleteMasterNjkb)

}