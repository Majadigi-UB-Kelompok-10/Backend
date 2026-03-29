package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/farildzaky/siskaperbapo-service/internal/handlers"
	"github.com/farildzaky/siskaperbapo-service/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/farildzaky/siskaperbapo-service/internal/worker"
)


func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		// Development defaults
		origins = "http://localhost:3000,http://localhost:3001,http://localhost:5173"
	}
	var result []string
	for _, origin := range strings.Split(origins, ",") {
		result = append(result, strings.TrimSpace(origin))
	}
	return result
}

// maskSensitiveData masks sensitive information from error messages
func maskSensitiveData(msg string) string {
	// Mask database URL credentials
	if strings.Contains(msg, "postgres://") {
		msg = "database connection error (credentials masked)"
	}
	// Mask cloudinary credentials
	if strings.Contains(msg, "cloudinary://") {
		msg = "cloudinary error (credentials masked)"
	}
	return msg
}

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Println("⚠️ File .env tidak ditemukan, menggunakan system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	config, errConf := pgxpool.ParseConfig(dbURL)
	if errConf != nil {
		log.Fatalf("❌ Gagal parse config DB: %s\n", maskSensitiveData(errConf.Error()))
	}
	
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("❌ Gagal konek ke database: %s\n", maskSensitiveData(err.Error()))
	}
	fmt.Println("✓ Database PostgreSQL connected")

	cld, errCld := cloudinary.New()
	if errCld != nil {
		log.Fatalf("❌ Gagal inisialisasi Cloudinary: %s\n", maskSensitiveData(errCld.Error()))
	}
	fmt.Println("✓ Cloudinary initialized")

	app := fiber.New(fiber.Config{
		AppName:     "Siskaperbapo Service",
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})	
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: getAllowedOrigins(),
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 3600,
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Too Many Requests"})
		},
	}))
	app.Get("/uploads/*", static.New("./uploads"))

	queries := db.New(pool)
	hargaWorker := worker.NewHargaWorkerPool(queries, 10, 1000)
	bpHandler := handlers.NewBahanPokokHandler(queries, cld, hargaWorker) 
	routes.SetupRoutes(app, bpHandler)

	fmt.Println("✓ Routes configured")

	// Run cache warmup asynchronously
	go func() {
		time.Sleep(1 * time.Second) // Wait for app to fully start
		bpHandler.CacheWarmup()
	}()

	// Start server in background
	go func() {
		if err := app.Listen(":8080"); err != nil && err.Error() != "shutting down" {
			log.Printf("❌ Server error: %v\n", err)
		}
	}()
	fmt.Println("✓ Server listening on :8080")

	// =====================================================================
	// GRACEFUL SHUTDOWN - Wait for signals
	// =====================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\n⏱️ Shutdown signal received, starting graceful shutdown...")
	
	// Graceful shutdown with 30 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 1. Stop accepting new requests and wait for in-flight requests
	fmt.Println("📍 Shutting down HTTP server...")
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("⚠️ HTTP server shutdown error: %v\n", err)
	}
	fmt.Println("✓ HTTP server shutdown complete")

	// 2. Stop worker pool
	fmt.Println("📍 Stopping worker pool...")
	hargaWorker.Shutdown()
	fmt.Println("✓ Worker pool stopped")

	// 3. Close database connection pool
	fmt.Println("📍 Closing database connections...")
	pool.Close()
	fmt.Println("✓ Database connections closed")

	fmt.Println("✓ Graceful shutdown complete - goodbye! 👋")
}
