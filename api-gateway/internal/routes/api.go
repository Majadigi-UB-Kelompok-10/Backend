package routes

import (
	"context"
	"fmt"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
)

func SetupDynamicEndpoint(app *fiber.App, queries *db.Queries, ctx context.Context) error {
	api := app.Group("/api/v1")

	routes, err := queries.ListEndpoints(ctx)
	if err != nil {
		return err
	}

	// Generate based on Data
	for _, route := range routes {
		slug := route.SlugName
		url := route.PageUrl

		// Register the GET endpoint using the slug_name
		api.Get("/"+slug, func(c fiber.Ctx) error {

			// For now, just return a success message acknowledging the slug
			// return c.JSON(handlers.SuccessResponse{
			// 	Pesan: "Sukses",
			// 	Data: map[string]any{
			// 		"slug": slug,
			// 		"id":   id,
			// 		"url":  url,
			// 	},
			// })

			if err := proxy.Do(c, url); err != nil {
				fmt.Printf("Proxy error for slug %s to %s: %v\n", slug, url, err)
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
					"error": "Service temporarily unavailable",
				})
			}

			return nil
		})
	}

	fmt.Println("Successfully registered", len(routes), "dynamic endpoints")
	return nil
}
