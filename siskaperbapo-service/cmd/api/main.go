package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
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
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"

	"github.com/cloudinary/cloudinary-go/v2"

	// "github.com/Majadigi-UB-Kelompok-10/majadigi-go-shared/shared/registry"

	"github.com/farildzaky/siskaperbapo-service/internal/cache"
	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/farildzaky/siskaperbapo-service/internal/handlers"
	"github.com/farildzaky/siskaperbapo-service/internal/routes"
	"github.com/farildzaky/siskaperbapo-service/internal/worker"
)

// =============================================================================
// BUILD INFO
// =============================================================================
var (
	serviceName = "siskaperbapo-service"
	version     = "dev"
	commit      = "unknown"
	buildTime   = "unknown"
	startedAt   = time.Now()
)

// =============================================================================
// CONFIG — typed, validated, single source of truth
// =============================================================================
type Config struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration

	DatabaseURL    string
	DBMaxConns     int32
	DBMinConns     int32
	DBMaxLifetime  time.Duration
	DBMaxIdle      time.Duration
	DBQueryTimeout time.Duration

	RedisURL      string
	CloudinaryURL string

	GatewayDBURL  string
	ServicePublic string

	AllowedOrigins []string
	BodyLimitBytes int

	RateLimitMax    int
	RateLimitWindow time.Duration

	LogLevel    slog.Level
	EnablePprof bool
	PprofPort   string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "production"),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 25*time.Second),

		DatabaseURL:    os.Getenv("DATABASE_URL"),
		DBMaxConns:     int32(getEnvInt("DB_MAX_CONNS", 30)),
		DBMinConns:     int32(getEnvInt("DB_MIN_CONNS", 5)),
		DBMaxLifetime:  getEnvDuration("DB_MAX_LIFETIME", time.Hour),
		DBMaxIdle:      getEnvDuration("DB_MAX_IDLE_TIME", 30*time.Minute),
		DBQueryTimeout: getEnvDuration("DB_QUERY_TIMEOUT", 5*time.Second),

		RedisURL:      os.Getenv("REDIS_URL"),
		CloudinaryURL: os.Getenv("CLOUDINARY_URL"), // 🚀 CLOUDINARY DIAKTIFKAN

		GatewayDBURL:  os.Getenv("GATEWAY_DATABASE_URL"),
		ServicePublic: getEnv("SERVICE_PUBLIC_URL", "http://siskaperbapo-api:8080/api/v1"),

		AllowedOrigins: parseList(getEnv("ALLOWED_ORIGINS",
			"http://localhost:3000,http://localhost:4000,http://localhost:5173")),
		BodyLimitBytes: getEnvInt("BODY_LIMIT_BYTES", 4*1024*1024),

		RateLimitMax:    getEnvInt("RATE_LIMIT_MAX", 120),
		RateLimitWindow: getEnvDuration("RATE_LIMIT_WINDOW", time.Minute),

		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		EnablePprof: getEnvBool("ENABLE_PPROF", false),
		PprofPort:   getEnv("PPROF_PORT", "6060"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.GatewayDBURL == "" {
		return nil, errors.New("GATEWAY_DATABASE_URL is required for service registration")
	}
	if cfg.DBMinConns > cfg.DBMaxConns {
		return nil, fmt.Errorf("DB_MIN_CONNS (%d) > DB_MAX_CONNS (%d)", cfg.DBMinConns, cfg.DBMaxConns)
	}
	return cfg, nil
}

// =============================================================================
// ENV HELPERS
// =============================================================================
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	switch strings.ToLower(os.Getenv(key)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return fallback
	}
}

func parseList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// =============================================================================
// SECURITY — credential masking
// =============================================================================
var (
	reConnURI  = regexp.MustCompile(`(postgres|postgresql|redis|amqp|mongodb|mysql|cloudinary)://([^:]+):([^@]+)@`)
	reKeyValue = regexp.MustCompile(`(?i)(password|passwd|pwd|token|api[_-]?key|secret|authorization)\s*[:=]\s*["']?([^"'\s,}]+)`)
	reBearer   = regexp.MustCompile(`(?i)bearer\s+([a-zA-Z0-9._\-+/=]+)`)
)

func maskSensitiveData(msg string) string {
	masked := reConnURI.ReplaceAllString(msg, `$1://$2:***@`)
	masked = reKeyValue.ReplaceAllString(masked, `$1=***`)
	masked = reBearer.ReplaceAllString(masked, `Bearer ***`)
	return masked
}

// =============================================================================
// LOGGER
// =============================================================================
func setupLogger(cfg *Config) {
	opts := &slog.HandlerOptions{
		Level:     cfg.LogLevel,
		AddSource: cfg.Environment != "production",
	}
	base := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(base).With(
		slog.String("service", serviceName),
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("env", cfg.Environment),
	)
	slog.SetDefault(logger)
}

// =============================================================================
// DATABASE POOL
// =============================================================================
func newDBPool(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	pgxCfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: tracelog.LoggerFunc(func(_ context.Context, _ tracelog.LogLevel, msg string, data map[string]interface{}) {
			if msg != "Query" {
				return
			}
			duration, _ := data["time"].(time.Duration)
			attrs := []any{
				slog.Any("sql", data["sql"]),
				slog.Duration("duration", duration),
			}
			if duration > 500*time.Millisecond {
				slog.Warn("db.slow_query", attrs...)
			} else {
				slog.Debug("db.query", attrs...)
			}
		}),
		LogLevel: tracelog.LogLevelInfo,
	}

	pgxCfg.MaxConns = cfg.DBMaxConns
	pgxCfg.MinConns = cfg.DBMinConns
	pgxCfg.MaxConnLifetime = cfg.DBMaxLifetime
	pgxCfg.MaxConnIdleTime = cfg.DBMaxIdle
	pgxCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping failed: %w", err)
	}
	return pool, nil
}

// =============================================================================
// CACHE
// =============================================================================
func initializeCache(cfg *Config) error {
	if cfg.RedisURL == "" {
		slog.Warn("REDIS_URL kosong, fallback ke SimpleCache (in-memory)")
		return nil
	}
	rc, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	cache.GlobalCache = rc
	slog.Info("redis cache terhubung")
	return nil
}

// =============================================================================
// MAIN
// =============================================================================
func main() {
	_ = godotenv.Load()

	if info, ok := debug.ReadBuildInfo(); ok && commit == "unknown" {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if len(s.Value) > 7 {
					commit = s.Value[:7]
				} else {
					commit = s.Value
				}
			case "vcs.time":
				buildTime = s.Value
			}
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	setupLogger(cfg)
	slog.Info("starting service",
		slog.String("port", cfg.Port),
		slog.String("build_time", buildTime),
	)

	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		time.Local = loc
	}

	rootCtx := context.Background()

	pool, err := newDBPool(rootCtx, cfg)
	if err != nil {
		slog.Error("init database failed", slog.String("error", maskSensitiveData(err.Error())))
		os.Exit(1)
	}
	slog.Info("postgresql terhubung", slog.Int("max_conns", int(cfg.DBMaxConns)))

	if err := initializeCache(cfg); err != nil {
		slog.Error("init cache failed", slog.String("error", maskSensitiveData(err.Error())))
		pool.Close()
		os.Exit(1)
	}

	var cld *cloudinary.Cloudinary
	if cfg.CloudinaryURL != "" {
		c, err := cloudinary.NewFromURL(cfg.CloudinaryURL)
		if err == nil {
			cld = c
			slog.Info("cloudinary terhubung")
		} else {
			slog.Error("gagal inisialisasi cloudinary", slog.String("error", maskSensitiveData(err.Error())))
		}
	} else {
		slog.Warn("CLOUDINARY_URL kosong, upload gambar akan gagal")
	}

	app := fiber.New(fiber.Config{
		AppName:      fmt.Sprintf("%s/%s", serviceName, version),
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    cfg.BodyLimitBytes,
		ErrorHandler: globalErrorHandler(cfg),
	})

	registerMiddleware(app, cfg)
	registerHealthEndpoints(app, pool)

	queries := db.New(pool)

	slog.Info("memulai background worker siskaperbapo")
	hargaWorkerPool := worker.Start(pool)

	siskaperbapoHandler := handlers.NewSiskaperbapoHandler(queries, pool, cld, hargaWorkerPool)

	routes.SetupRoutes(app, siskaperbapoHandler)
	slog.Info("routes terkonfigurasi")

	go func() {
		time.Sleep(1 * time.Second)
		siskaperbapoHandler.CacheWarmup()
	}()

	// registry.AutoRegisterFull(
	//     cfg.GatewayDBURL,
	//     "siskaperbapo",
	//     cfg.ServicePublic,
	//     "SISKAPERBAPO",
	//     "https://cdn.example.com/siskaperbapo.png",
	//     "Sistem Ketersediaan dan Harga Bahan Pokok",
	//     []string{"b1000001-0000-4000-8000-000000000008"}, // Harga Kebutuhan Pokok
	// )

	if cfg.EnablePprof {
		go startPprofServer(cfg.PprofPort)
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", slog.String("addr", ":"+cfg.Port))
		if err := app.Listen(":"+cfg.Port, fiber.ListenConfig{
			DisableStartupMessage: true,
		}); err != nil {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("server gagal start", slog.String("error", err.Error()))
	case sig := <-quit:
		slog.Info("sinyal shutdown diterima", slog.String("signal", sig.String()))
	}

	gracefulShutdown(app, pool, cfg)
}

// =============================================================================
// MIDDLEWARE STACK
// =============================================================================
func registerMiddleware(app *fiber.App, cfg *Config) {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment != "production",
		StackTraceHandler: func(c fiber.Ctx, e any) {
			slog.Error("panic recovered",
				slog.Any("panic", e),
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
				slog.String("ip", c.IP()),
			)
		},
	}))

	app.Use(requestid.New())

	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		HSTSMaxAge:            31536000,
		HSTSPreloadEnabled:    true,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	app.Use(logger.New(logger.Config{
		Format: `{"ts":"${time}","level":"info","msg":"http","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","ip":"${ip}","reqid":"${locals:requestid}","ua":"${ua}"}` + "\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))

	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: cfg.RateLimitWindow,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		Next: func(c fiber.Ctx) bool {
			p := c.Path()
			return p == "/health" || p == "/ready"
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",
				Message: "Terlalu banyak permintaan (Global Limit)",
				Action:  "Sistem mendeteksi lonjakan trafik tinggi, silakan tunggu sebentar.",
			})
		},
	}))
}

// =============================================================================
// GLOBAL ERROR HANDLER
// =============================================================================
func globalErrorHandler(cfg *Config) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var fe *fiber.Error
		if errors.As(err, &fe) {
			code = fe.Code
		}

		reqID, _ := c.Locals("requestid").(string)
		slog.Error("request error",
			slog.String("error", maskSensitiveData(err.Error())),
			slog.String("path", c.Path()),
			slog.String("method", c.Method()),
			slog.Int("status", code),
			slog.String("request_id", reqID),
		)

		message := "Terjadi kesalahan pada server Siskaperbapo"
		if cfg.Environment != "production" && code != fiber.StatusInternalServerError {
			message = err.Error()
		}

		return c.Status(code).JSON(handlers.ErrorResponse{
			Code:    fmt.Sprintf("ERR_%d", code),
			Message: message,
		})
	}
}

// =============================================================================
// HEALTH ENDPOINTS
// =============================================================================
func registerHealthEndpoints(app *fiber.App, pool *pgxpool.Pool) {
	type livenessResponse struct {
		Status  string `json:"status"`
		Service string `json:"service"`
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Uptime  string `json:"uptime"`
	}

	type readinessResponse struct {
		Status    string    `json:"status"`
		Database  string    `json:"database"`
		Cache     string    `json:"cache"`
		Timestamp time.Time `json:"timestamp"`
		Service   string    `json:"service"`
		Commit    string    `json:"commit"`
	}

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(livenessResponse{
			Status:  "alive",
			Service: serviceName,
			Version: version,
			Commit:  commit,
			Uptime:  time.Since(startedAt).String(),
		})
	})

	app.Get("/ready", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		dbStatus, cacheStatus, ready := "ok", "ok", true

		if err := pool.Ping(ctx); err != nil {
			dbStatus = "down"
			ready = false
		}
		if cache.GlobalCache == nil {
			cacheStatus = "degraded"
		}

		status, overall := fiber.StatusOK, "ready"
		if !ready {
			status, overall = fiber.StatusServiceUnavailable, "not_ready"
		}

		return c.Status(status).JSON(readinessResponse{
			Status:    overall,
			Database:  dbStatus,
			Cache:     cacheStatus,
			Timestamp: time.Now(),
			Service:   serviceName,
			Commit:    commit,
		})
	})
}

// =============================================================================
// PPROF
// =============================================================================
func startPprofServer(port string) {
	addr := "127.0.0.1:" + port
	slog.Warn("pprof enabled — bind localhost-only", slog.String("addr", addr))
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("pprof server error", slog.String("error", err.Error()))
	}
}

// =============================================================================
// GRACEFUL SHUTDOWN
// =============================================================================
func gracefulShutdown(app *fiber.App, pool *pgxpool.Pool, cfg *Config) {
	slog.Info("memulai graceful shutdown", slog.Duration("timeout", cfg.ShutdownTimeout))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("shutdown server gagal", slog.String("error", err.Error()))
	} else {
		slog.Info("server http berhenti menerima request")
	}

	if rc, ok := cache.GlobalCache.(*cache.RedisCache); ok {
		slog.Info("menutup koneksi redis")
		rc.Close()
	}

	slog.Info("menutup koneksi database")
	pool.Close()

	slog.Info("shutdown selesai dengan aman",
		slog.Duration("total_uptime", time.Since(startedAt)))
}
