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
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/farildzaky/sidita-service/internal/db"
	"github.com/farildzaky/sidita-service/internal/cache"
	"github.com/farildzaky/sidita-service/internal/handlers"
	"github.com/farildzaky/sidita-service/internal/routes"
)

func initializeCache() {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL != "" {
		redisCache, err := cache.NewRedisCache(redisURL)
		if err != nil {
			log.Fatalf("Failed to initialize Redis cache: %v\n", err)
		}
		cache.GlobalCache = redisCache
		fmt.Println("Redis cache initialized")
	} else {
		fmt.Println("Using SimpleCache (in-memory)")
	}
}

func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000,http://localhost:3001,http://localhost:5173"
	}
	var result []string
	for _, origin := range strings.Split(origins, ",") {
		result = append(result, strings.TrimSpace(origin))
	}
	return result
}

func maskSensitiveData(msg string) string {
	if strings.Contains(msg, "postgres://") {
		msg = "database connection error (credentials masked)"
	}
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
		log.Fatalf("Failed to connect to database: %s\n", maskSensitiveData(err.Error()))
	}
	fmt.Println("Database PostgreSQL connected")

	cld, errCld := cloudinary.New()
	if errCld != nil {
		log.Fatalf("Failed to initialize Cloudinary: %s\n", maskSensitiveData(errCld.Error()))
	}
	fmt.Println("Cloudinary initialized")

	initializeCache()

	app := fiber.New(fiber.Config{
		AppName:     "Sidita Service",
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Too Many Requests. Santai dulu bossku!"})
		},
	}))

	queries := db.New(pool)
	destinasiHandler := handlers.NewDestinasiHandler(queries, cld)

	routes.SetupRoutes(app, destinasiHandler)
	fmt.Println("Routes configured")

	// 9. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Port default Sidita
	}

	go func() {
		if err := app.Listen(":" + port); err != nil && err.Error() != "shutting down" {
			log.Printf("Server error: %v\n", err)
		}
	}()
	fmt.Printf("Server Sidita listening on port %s\n", port)

	// =====================================================================
	// 10. GRACEFUL SHUTDOWN
	// =====================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutdown signal received, starting graceful shutdown...") 

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	fmt.Println("🔄 Shutting down HTTP server...")
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}
	fmt.Println("HTTP server shutdown complete")

	fmt.Println("Closing database connections...")
	pool.Close()
	fmt.Println("Database connections closed")

	// Optional: Close Redis Connection jika menggunakan RedisCache
	if redisCache, ok := cache.GlobalCache.(*cache.RedisCache); ok {
		fmt.Println("Closing Redis connection...")
		redisCache.Close()
		fmt.Println("Redis connection closed")
	}

	fmt.Println("Graceful shutdown complete. Bye bossku!")
}