package routes

import (
	"context"
	"fmt"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers"
	"github.com/gofiber/fiber/v3"
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
		id := route.EndpointListID
		url := route.PageUrl

		// Register the GET endpoint using the slug_name
		api.Get("/"+slug, func(c fiber.Ctx) error {

			// For now, just return a success message acknowledging the slug
			return c.JSON(handlers.SuccessResponse{
				Pesan: "Sukses",
				Data: map[string]any{
					"slug": slug,
					"id":   id,
					"url":  url,
				},
			})
		})
	}

	fmt.Println("Successfully registered", len(routes), "dynamic endpoints")
	return nil
}
