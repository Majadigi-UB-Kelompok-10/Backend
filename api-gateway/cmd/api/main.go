package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/routes"
	init_helper "github.com/Majadigi-UB-Kelompok-10/majadigi-go-shared/shared/init_helper"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize PostgreSQL database
	var pool *pgxpool.Pool
	pool = init_helper.InitializePostgreDB(os.Getenv("DATABASE_URL"))

	// Initialize Redis cache
	init_helper.InitializeRedisCache(os.Getenv("REDIS_URL"))

	// Get Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:     "Majadigi Api Gateway",
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	// Initialize Fiber app configuration
	init_helper.InitializeFiberAppConfig(app, os.Getenv("ALLOWED_ORIGINS"))

	// Setting Up
	queries := db.New(pool)
	routes.SetupDynamicEndpoint(app, queries, context.Background())
	fmt.Println("Dynamic endpoints configured")

	// Listen to port
	go func() {
		if err := app.Listen(":" + port); err != nil && err.Error() != "shutting down" {
			log.Printf("Api Gateway error: %v\n", err)
		}

		fmt.Println("Api Gateway listening on :" + port)
	}()

	// Initialize graceful shutdown listener for Fiber App, PostgreSQL, and Redis (Implicit)
	init_helper.InitializeGracefulShutdownListener(&init_helper.ShutdownType{App: app, Pool: pool})

	// Initialize graceful shutdown listener for other specific handler, worker, or services
	gracefulShutdownListener()
}

func gracefulShutdownListener() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nGraceful Shutdown for other services...")

	// TODO: add other services to shutdown

	fmt.Println("Graceful shutdown complete")
}
