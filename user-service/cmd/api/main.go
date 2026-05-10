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

	"github.com/Majadigi-UB-Kelompok-10/majadigi-go-shared/shared/registry"
	"github.com/bytedance/sonic"
	"github.com/farildzaky/user-service/internal/cache"
	"github.com/farildzaky/user-service/internal/db"
	"github.com/farildzaky/user-service/internal/handlers"
	"github.com/farildzaky/user-service/internal/routes"
	"github.com/farildzaky/user-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
)

var (
	serviceName = "user-service"
	version     = "dev"
	commit      = "unknown"
	buildTime   = "unknown"
	startedAt   = time.Now()
)

// =============================================================================
// CONFIG
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
	GatewayDBURL  string
	ServicePublic string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

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
		GatewayDBURL:  os.Getenv("GATEWAY_DATABASE_URL"),
		ServicePublic: getEnv("SERVICE_PUBLIC_URL", "http://user-api:8080/api/v1"),

		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  getEnvDuration("JWT_ACCESS_TTL", utils.DefaultAccessTTL),
		JWTRefreshTTL: getEnvDuration("JWT_REFRESH_TTL", utils.DefaultRefreshTTL),

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
		return nil, errors.New("DATABASE_URL wajib diisi")
	}
	if cfg.GatewayDBURL == "" {
		return nil, errors.New("GATEWAY_DATABASE_URL wajib diisi")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET wajib diisi")
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
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
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
// SECURITY
// =============================================================================

var (
	reConnURI  = regexp.MustCompile(`(postgres|postgresql|redis|mongodb|mysql)://([^:]+):([^@]+)@`)
	reKeyValue = regexp.MustCompile(`(?i)(password|passwd|token|api[_-]?key|secret|authorization)\s*[:=]\s*["']?([^"'\s,}]+)`)
	reBearer   = regexp.MustCompile(`(?i)bearer\s+([a-zA-Z0-9._\-+/=]+)`)
)

func maskSensitiveData(msg string) string {
	masked := reConnURI.ReplaceAllString(msg, `$1://$2:***@`)
	masked = reKeyValue.ReplaceAllString(masked, `$1=***`)
	return reBearer.ReplaceAllString(masked, `Bearer ***`)
}

// =============================================================================
// LOGGER
// =============================================================================

func setupLogger(cfg *Config) {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel, AddSource: cfg.Environment != "production"}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)).With(
		slog.String("service", serviceName),
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("env", cfg.Environment),
	))
}

// =============================================================================
// DATABASE
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
			dur, _ := data["time"].(time.Duration)
			attrs := []any{slog.Any("sql", data["sql"]), slog.Duration("duration", dur)}
			if dur > 500*time.Millisecond {
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
		slog.Warn("REDIS_URL kosong, fallback ke in-memory SimpleCache")
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
	slog.Info("starting service", slog.String("port", cfg.Port), slog.String("build_time", buildTime))

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
	authHandler := handlers.NewAuthHandler(queries, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	routes.SetupRoutes(app, authHandler, cfg.JWTSecret)
	slog.Info("routes terkonfigurasi")

	go authHandler.CacheWarmup()

	cleanupCtx, cleanupCancel := context.WithCancel(rootCtx)
	defer cleanupCancel()
	go authHandler.StartTokenCleanup(cleanupCtx)

	go func() {
		registry.AutoRegister(cfg.GatewayDBURL, "user", cfg.ServicePublic)
		slog.Info("auto-register ke gateway selesai")
	}()

	if cfg.EnablePprof {
		go startPprofServer(cfg.PprofPort)
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", slog.String("addr", ":"+cfg.Port))
		if err := app.Listen(":"+cfg.Port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
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
// MIDDLEWARE
// =============================================================================

func registerMiddleware(app *fiber.App, cfg *Config) {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment != "production",
		StackTraceHandler: func(c fiber.Ctx, e any) {
			slog.Error("panic recovered", slog.Any("panic", e), slog.String("path", c.Path()))
		},
	}))
	app.Use(requestid.New())
	app.Use(helmet.New(helmet.Config{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		HSTSMaxAge:         31536000,
		HSTSPreloadEnabled: true,
	}))
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: `{"ts":"${time}","level":"info","msg":"http","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","ip":"${ip}","reqid":"${locals:requestid}"}` + "\n",
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
		Max:          cfg.RateLimitMax,
		Expiration:   cfg.RateLimitWindow,
		KeyGenerator: func(c fiber.Ctx) string { return c.IP() },
		Next: func(c fiber.Ctx) bool {
			p := c.Path()
			return p == "/health" || p == "/ready"
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",
				Message: "Terlalu banyak permintaan",
			})
		},
	}))
}

// =============================================================================
// HEALTH ENDPOINTS
// =============================================================================

func registerHealthEndpoints(app *fiber.App, pool *pgxpool.Pool) {
	type livenessResp struct {
		Status  string    `json:"status"`
		Uptime  string    `json:"uptime"`
		Service string    `json:"service"`
		Version string    `json:"version"`
		At      time.Time `json:"at"`
	}
	type readinessResp struct {
		Status string `json:"status"`
		DB     string `json:"db"`
		Cache  string `json:"cache"`
	}

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(livenessResp{
			Status:  "healthy",
			Uptime:  time.Since(startedAt).Round(time.Second).String(),
			Service: serviceName,
			Version: version,
			At:      time.Now(),
		})
	})

	app.Get("/ready", func(c fiber.Ctx) error {
		status, dbStatus, cacheStatus, code := "ready", "ok", "ok", 200
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "error: " + err.Error()
			status = "not ready"
			code = 503
		}
		cache.GlobalCache.Set("__health__", "ok", 5*time.Second)
		return c.Status(code).JSON(readinessResp{Status: status, DB: dbStatus, Cache: cacheStatus})
	})
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
		slog.Error("request error",
			slog.String("error", maskSensitiveData(err.Error())),
			slog.String("path", c.Path()),
			slog.Int("status", code),
		)
		msg := "Terjadi kesalahan pada server"
		if cfg.Environment != "production" {
			msg = err.Error()
		}
		return c.Status(code).JSON(handlers.ErrorResponse{Code: "ERR_INTERNAL", Message: msg})
	}
}

// =============================================================================
// GRACEFUL SHUTDOWN
// =============================================================================

func gracefulShutdown(app *fiber.App, pool *pgxpool.Pool, cfg *Config) {
	slog.Info("memulai graceful shutdown...")
	shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := app.ShutdownWithContext(shutCtx); err != nil {
		slog.Error("server shutdown error", slog.String("error", err.Error()))
	}
	if sc, ok := cache.GlobalCache.(interface{ Close() }); ok {
		sc.Close()
	}
	pool.Close()
	slog.Info("shutdown selesai", slog.Duration("uptime", time.Since(startedAt)))
}

// =============================================================================
// PPROF
// =============================================================================

func startPprofServer(port string) {
	addr := "127.0.0.1:" + port
	slog.Info("pprof listening", slog.String("addr", addr))
	srv := &http.Server{Addr: addr}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Warn("pprof server error", slog.String("err", err.Error()))
	}
}
