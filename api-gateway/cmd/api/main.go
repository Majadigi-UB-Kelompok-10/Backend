package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	gatewayAdmin "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/gateway_admin"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/routes"
	init_helper "github.com/Majadigi-UB-Kelompok-10/majadigi-go-shared/shared/init_helper"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var pool *pgxpool.Pool
	pool = init_helper.InitializePostgreDB(os.Getenv("DATABASE_URL"))

	init_helper.InitializeRedisCache(os.Getenv("REDIS_URL"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	app := fiber.New(fiber.Config{
		AppName:     "Majadigi Api Gateway",
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Use(compress.New(compress.Config{Level: compress.LevelBestCompression}))
	app.Use(etag.New(etag.ConfigDefault))

	init_helper.InitializeFiberAppConfig(app, os.Getenv("ALLOWED_ORIGINS"))

	queries := db.New(pool)

	if err := routes.RefreshEndpoints(context.Background(), queries); err != nil {
		log.Printf("Peringatan: Gagal load rute awal: %v\n", err)
	}

	adminHandler := gatewayAdmin.NewGatewayAdminHandler(queries)
	app.Post("/api/v1/admin/gateway/sync", adminHandler.SyncNow)

	routes.SetupPublicRoutes(app, queries)
	routes.SetupAdminRoutes(app, queries)
	routes.SetupDynamicEndpoint(app)
	fmt.Println("Dynamic endpoints configured")

	go startRealtimeObserver(pool, queries)

	go func() {
		if err := app.Listen(":" + port); err != nil && err.Error() != "shutting down" {
			log.Printf("Api Gateway error: %v\n", err)
		}

		fmt.Println("Api Gateway listening on :" + port)
	}()

	init_helper.InitializeGracefulShutdownListener(&init_helper.ShutdownType{App: app, Pool: pool})

	gracefulShutdownListener()
}

func gracefulShutdownListener() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nGraceful Shutdown for other services...")

	fmt.Println("Graceful shutdown complete")
}

func startRealtimeObserver(pool *pgxpool.Pool, queries *db.Queries) {
	fmt.Println("Observer Real-time dijalankan...")

	for {
		ctx := context.Background()
		conn, err := pool.Acquire(ctx)
		if err != nil {
			fmt.Printf("❌ Gagal mengambil koneksi listener: %v. Coba lagi 5 detik...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		_, err = conn.Exec(ctx, "LISTEN gateway_refresh_channel")
		if err != nil {
			fmt.Printf("❌ Gagal melakukan LISTEN: %v\n", err)
			conn.Release()
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Listener PostgreSQL...")

		for {
			ctxWait, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			_, err := conn.Conn().WaitForNotification(ctxWait)
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					if pingErr := conn.Ping(ctx); pingErr != nil {
						fmt.Printf("Koneksi mati diam-diam oleh Docker (Ping gagal). Reconnecting...\n")
						break
					}
					continue
				}

				fmt.Printf("Koneksi terputus tiba-tiba: %v. Reconnecting...\n", err)
				break
			}

			fmt.Println("Sinyal UPDATE diterima! Memperbarui rute di RAM...")
			if err := routes.RefreshEndpoints(ctx, queries); err != nil {
				fmt.Printf("Gagal sinkronisasi RAM: %v\n", err)
			}
		}

		conn.Release()
		time.Sleep(2 * time.Second)
	}
}
