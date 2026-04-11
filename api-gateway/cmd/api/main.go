package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:     "Majadigi Api Gateway",
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	// Initialize Fiber app configuration
	init_helper.InitializeFiberAppConfig(app, os.Getenv("ALLOWED_ORIGINS"))

	// Initialize graceful shutdown listener for Fiber App, PostgreSQL, and Redis (Implicit)
	init_helper.InitializeGracefulShutdownListener(&init_helper.ShutdownType{App: app, Pool: pool})

	// Initialize graceful shutdown listener for other specific handler, worker, or services
	gracefulShutdownListener()

	// Listen to port 8080
	go func() {
		if err := app.Listen(":8080"); err != nil && err.Error() != "shutting down" {
			log.Printf("Server error: %v\n", err)
		}
	}()

	fmt.Println("Server listening on :8080")
}

func gracefulShutdownListener() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nGraceful Shutdown for other services...")

	// TODO: add other services to shutdown

	fmt.Println("Graceful shutdown complete")
}
