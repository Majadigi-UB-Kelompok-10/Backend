package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	// Import Registry Gateway
	"github.com/Majadigi-UB-Kelompok-10/majadigi-go-shared/shared/registry"

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
			slog.Error("Gagal inisiasi Redis Bapenda", slog.String("error", err.Error()))
			os.Exit(1)
		}
		cache.GlobalCache = redisCache
		slog.Info("Redis Cache Bapenda Terhubung")
	} else {
		slog.Warn("REDIS_URL tidak ditemukan, menggunakan SimpleCache (in-memory)")
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
	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(loggerHandler))

	loc, _ := time.LoadLocation("Asia/Jakarta")
	time.Local = loc

	if err := godotenv.Load(); err != nil {
		slog.Info("Info: File .env tidak ditemukan, menggunakan variabel lingkungan sistem")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	config, errConf := pgxpool.ParseConfig(dbURL)
	if errConf != nil {
		slog.Error("Gagal parse konfigurasi DB", slog.String("error", maskSensitiveData(errConf.Error())))
		os.Exit(1)
	}

	config.MaxConns = 30 
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("Koneksi database gagal", slog.String("error", maskSensitiveData(err.Error())))
		os.Exit(1)
	}
	slog.Info("PostgreSQL Bapenda Terhubung")

	initializeCache()

	app := fiber.New(fiber.Config{
		AppName:      "Bapenda Service",
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(helmet.New(helmet.Config{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000,
		HSTSPreloadEnabled: true,
	}))

	app.Use(logger.New(logger.Config{
		Format: `{"time":"${time}", "level":"INFO", "method":"${method}", "path":"${path}", "status":${status}, "latency":"${latency}"}` + "\n",
	}))

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

	app.Get("/health", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dbStatus, redisStatus := "OK", "OK"
		if errPing := pool.Ping(ctx); errPing != nil {
			dbStatus = "DOWN"
		}
		if cache.GlobalCache == nil {
			redisStatus = "DOWN"
		}

		status, httpCode := "healthy", 200
		if dbStatus == "DOWN" || redisStatus == "DOWN" {
			status, httpCode = "degraded", 503
		}

		return c.Status(httpCode).JSON(fiber.Map{
			"status":    status,
			"database":  dbStatus,
			"redis":     redisStatus,
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "bapenda-api",
		})
	})

	queries := db.New(pool)
	bapendaHandler := handlers.NewBapendaHandler(queries)
	routes.SetupRoutes(app, bapendaHandler)
	slog.Info("Routes Bapenda Terkonfigurasi")


	bapendaHandler.RunCacheWarmup()

	gatewayDBUrl := os.Getenv("GATEWAY_DATABASE_URL")
	registry.AutoRegister(gatewayDBUrl, "bapenda", "http://bapenda-api:8080/api/v1")
	slog.Info("Proses AutoRegister Bapenda ke Gateway telah dipanggil")

	go func() {
		fmt.Printf("\n🚀 Bapenda Service berjalan di port :%s\n", port)
		if err := app.Listen(":" + port); err != nil && err.Error() != "shutting down" {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Menerima sinyal shutdown, mematikan server secara bertahap...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("Kesalahan saat mematikan server HTTP", slog.String("error", err.Error()))
	}

	slog.Info("Menutup koneksi database Bapenda...")
	pool.Close()

	if redisCache, ok := cache.GlobalCache.(*cache.RedisCache); ok {
		slog.Info("Menutup koneksi Redis Bapenda...")
		redisCache.Close()
	}

	slog.Info("Proses shutdown Bapenda selesai dengan aman.")
}