package main

import (
	"context"
	"errors"
	"log/slog"
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
		BodyLimit:   1 * 1024 * 1024,
	})

	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(etag.New(etag.ConfigDefault))

	init_helper.InitializeFiberAppConfig(app, os.Getenv("ALLOWED_ORIGINS"))

	queries := db.New(pool)

	if err := routes.RefreshEndpoints(context.Background(), queries); err != nil {
		slog.Warn("gateway.routes.initial_load_failed", "error", err)
	}

	adminHandler := gatewayAdmin.NewGatewayAdminHandler(queries)
	app.Post("/api/v1/admin/gateway/sync", adminHandler.SyncNow)

	routes.SetupPublicRoutes(app, queries)
	routes.SetupAdminRoutes(app, queries)
	routes.SetupDynamicEndpoint(app)
	slog.Info("gateway.routes.ready")

	observerCtx, observerCancel := context.WithCancel(context.Background())
	defer observerCancel()
	go startRealtimeObserver(observerCtx, pool, queries)

	go func() {
		if err := app.Listen(":" + port); err != nil && err.Error() != "shutting down" {
			slog.Error("gateway.listen.error", "error", err)
		}
	}()

	init_helper.InitializeGracefulShutdownListener(&init_helper.ShutdownType{App: app, Pool: pool})

	gracefulShutdownListener()
}

func gracefulShutdownListener() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("gateway.shutdown.start")
	slog.Info("gateway.shutdown.complete")
}

func startRealtimeObserver(ctx context.Context, pool *pgxpool.Pool, queries *db.Queries) {
	slog.Info("observer.realtime.start")

	for {
		select {
		case <-ctx.Done():
			slog.Info("observer.realtime.shutdown")
			return
		default:
		}

		connCtx, connCancel := context.WithTimeout(ctx, 10*time.Second)
		conn, err := pool.Acquire(connCtx)
		connCancel()
		if err != nil {
			slog.Error("observer.acquire.error", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		_, err = conn.Exec(ctx, "LISTEN gateway_refresh_channel")
		if err != nil {
			slog.Error("observer.listen.error", "error", err)
			conn.Release()
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		slog.Info("observer.listen.ready")

		for {
			select {
			case <-ctx.Done():
				conn.Release()
				return
			default:
			}

			ctxWait, cancel := context.WithTimeout(ctx, 15*time.Second)
			_, err := conn.Conn().WaitForNotification(ctxWait)
			cancel()

			if err != nil {
				if errors.Is(err, context.Canceled) {
					conn.Release()
					return
				}
				if errors.Is(err, context.DeadlineExceeded) {
					if pingErr := conn.Ping(ctx); pingErr != nil {
						slog.Warn("observer.connection.dead", "error", pingErr)
						break
					}
					continue
				}
				slog.Warn("observer.connection.lost", "error", err)
				break
			}

			slog.Info("observer.refresh.received")
			if err := routes.RefreshEndpoints(ctx, queries); err != nil {
				slog.Error("observer.refresh.error", "error", err)
			}
		}

		conn.Release()
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
