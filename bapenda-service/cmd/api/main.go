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
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/db"
	"github.com/farildzaky/bapenda-service/internal/handlers"
	"github.com/farildzaky/bapenda-service/internal/routes"
)

func maskSensitiveData(msg string) string {
	if strings.Contains(msg, "postgres://") {
		return "database connection error (credentials masked)"
	}
	return msg
}

func initializeCache() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		redisCache, err := cache.NewRedisCache(redisURL)
		if err != nil {
			log.Fatalf("Gagal inisiasi Redis: %v\n", err)
		}
		cache.GlobalCache = redisCache
		fmt.Println("Redis Bapenda Terhubung")
	} else {
		log.Println("Peringatan: REDIS_URL tidak ditemukan, menggunakan SimpleCache (in-memory)")
	}
}

func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000,http://localhost:4000,http://localhost:5173"
	}
	var result []string
	for _, origin := range strings.Split(origins, ",") {
		result = append(result, strings.TrimSpace(origin))
	}
	return result
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan variabel lingkungan sistem")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	config, errConf := pgxpool.ParseConfig(dbURL)
	if errConf != nil {
		log.Fatalf("Gagal parse konfigurasi DB: %s\n", maskSensitiveData(errConf.Error()))
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Koneksi database gagal: %s\n", maskSensitiveData(err.Error()))
	}
	fmt.Println("PostgreSQL Bapenda Terhubung")

	initializeCache()

	app := fiber.New(fiber.Config{
		AppName:     "Bapenda Service",
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))

	app.Use(limiter.New(limiter.Config{
		Max:        120,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Terlalu banyak permintaan, silakan coba lagi nanti"})
		},
	}))

	queries := db.New(pool)
	bapendaHandler := handlers.NewBapendaHandler(queries)
	routes.SetupRoutes(app, bapendaHandler)
	fmt.Println("Routes Bapenda Terkonfigurasi")

	// =========================================================
	// JALANKAN CACHE WARMUP SEBELUM SERVER MENERIMA REQUEST
	// =========================================================
	bapendaHandler.RunCacheWarmup()

	go func() {
		fmt.Printf("Bapenda Service berjalan di port :%s\n", port)
		if err := app.Listen(":" + port); err != nil && err.Error() != "shutting down" {
			log.Printf("Server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nMenerima sinyal shutdown, mematikan server secara bertahap...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Kesalahan saat mematikan server HTTP: %v\n", err)
	}

	fmt.Println("Menutup koneksi database...")
	pool.Close()

	if redisCache, ok := cache.GlobalCache.(*cache.RedisCache); ok {
		fmt.Println("Menutup koneksi Redis...")
		redisCache.Close()
	}

	fmt.Println("Proses shutdown selesai")
}