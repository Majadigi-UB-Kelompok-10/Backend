package routes

import (
	"context"
	"fmt"
	"sync"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	zstdUtil "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/util"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
)

type RouteRegistry struct {
	sync.RWMutex
	Routes map[string]string
}

var GatewayRegistry = &RouteRegistry{
	Routes: make(map[string]string),
}

func RefreshEndpoints(ctx context.Context, queries *db.Queries) error {
	dbRoutes, err := queries.ListEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	newRoutes := make(map[string]string)
	for _, r := range dbRoutes {
		newRoutes[r.SlugName] = r.PageUrl
	}

	GatewayRegistry.Lock()
	GatewayRegistry.Routes = newRoutes
	GatewayRegistry.Unlock()

	fmt.Printf("RAM Synced: %d rute dinamis siap digunakan\n", len(newRoutes))
	return nil
}

func SetupDynamicEndpoint(app *fiber.App) {
	api := app.Group("/api/v1")

	// Apply zstd compression middleware to all routes in this group
	api.Use(zstdUtil.ZstdCompressionMiddleware)

	// Route to list all available endpoints
	api.Get("/endpoints", func(c fiber.Ctx) error {
		GatewayRegistry.RLock()
		defer GatewayRegistry.RUnlock()

		return c.JSON(fiber.Map{
			"pesan": "Sukses",
			"data":  GatewayRegistry.Routes,
		})
	})

	api.All("/:slug/*", func(c fiber.Ctx) error {
		slug := c.Params("slug")

		GatewayRegistry.RLock()
		targetUrl, exists := GatewayRegistry.Routes[slug]
		GatewayRegistry.RUnlock()

		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Endpoint layanan tidak ditemukan.",
			})
		}

		extraPath := c.Params("*")
		if extraPath != "" {
			targetUrl = targetUrl + "/" + extraPath
		}

		if err := proxy.Do(c, targetUrl); err != nil {
			fmt.Printf("Proxy error untuk slug %s ke %s: %v\n", slug, targetUrl, err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "Bad Gateway",
				"message": "Layanan sedang mengalami gangguan.",
			})
		}

		return nil
	})
}
