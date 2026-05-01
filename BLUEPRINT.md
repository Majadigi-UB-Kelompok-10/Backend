# Capstone-BE Service Blueprint

> **Status**: Active blueprint based on `klinik-blueprint-v1.0`  
> **Reference tag**: `git checkout klinik-blueprint-v1.0`  
> **Reference service**: `klinik-service/`  
> **Last updated**: April 2026

This document defines the canonical pattern for all microservices in
Capstone-BE. The pattern was first established in `klinik-service` and
has been verified end-to-end (build, runtime, CI green, tagged).

---

## Why This Blueprint Exists

Without a shared blueprint:
- Each developer reinvents the wheel
- Code style drifts service-to-service
- Bug fixes don't propagate (because patterns differ)
- Onboarding new team members takes weeks

With this blueprint:
- New service = follow checklist, ship in days
- Bug fix in pattern → applied uniformly across services
- Code reviews focus on business logic, not boilerplate
- Architecture conversations have a shared vocabulary

---

## Folder Structure (Canonical)

```
service-name/
├── cmd/api/main.go              # Entry point (~470 LOC enterprise pattern)
├── db/
│   ├── migrations/
│   │   ├── 001.schema.sql       # Tables, indexes, triggers, functions
│   │   └── 002.seed.sql         # Test data (NULL for nullable, not '')
│   └── queries/
│       └── queries.sql          # SQLC source — NEVER edit internal/db/*
├── internal/
│   ├── cache/
│   │   ├── cache.go             # SimpleCache: LRU + TTL cleanup + atomic stats
│   │   └── cache_redis.go       # RedisCache: pool tuning + error counter
│   ├── db/                      # SQLC GENERATED — don't edit manually
│   ├── handlers/
│   │   ├── handler_utils.go     # TTL constants, cache helpers, parsePagination
│   │   ├── dto_base.go          # ErrorResponse, SuccessResponse, FieldError
│   │   ├── dto_<domain>.go      # Domain DTOs (Request + Response)
│   │   ├── <service>_setup.go   # Constructor, helpers, cache warmup
│   │   ├── <service>_public.go  # Citizen-facing endpoints
│   │   └── <service>_admin.go   # Admin endpoints
│   ├── routes/api.go            # Route setup with tiered rate limiting
│   └── utils/
│       ├── validator.go         # All validators (email, phone, URL, UUID, pagination)
│       └── slug.go              # ONLY if service needs URL slugs
├── docker-compose.yml
├── Dockerfile.dev
├── .env.encrypted               # dotenvx
├── .env.keys                    # private key — NEVER commit (.gitignore)
├── go.mod / go.sum
├── sqlc.yaml
├── Makefile                     # Standard tasks: build, test, lint, etc
└── README.md                    # Service-specific docs
```

---

## Environment Variables Schema

### 🔴 Required (service crashes if missing)

```env
DATABASE_URL=postgres://user:pass@host:port/db
GATEWAY_DATABASE_URL=postgres://...    # for service registry auto-register
```

### 🟡 Strongly Recommended

```env
PORT=8080
ENVIRONMENT=development                 # development | staging | production
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

REDIS_URL=redis://...

POSTGRES_USER=...                       # Used by docker-compose
POSTGRES_PASSWORD=...
POSTGRES_DB=...
```

### 🟢 Tuning (all have defaults in code; override per environment)

```env
# Service registry
SERVICE_PUBLIC_URL=http://<service>-api:8080/api/v1

# Lifecycle
SHUTDOWN_TIMEOUT=25s

# DB connection pool
DB_MAX_CONNS=30
DB_MIN_CONNS=5
DB_MAX_LIFETIME=1h
DB_MAX_IDLE_TIME=30m
DB_QUERY_TIMEOUT=5s

# HTTP limits
BODY_LIMIT_BYTES=4194304                # 4MB; bump if upload-heavy (e.g. 10MB)
RATE_LIMIT_MAX=120                      # global rate limit per minute per IP
RATE_LIMIT_WINDOW=1m

# Observability
LOG_LEVEL=debug                         # debug | info | warn | error
ENABLE_PPROF=true
PPROF_PORT=6060

# Cache TTL (naming can differ per domain)
CACHE_TTL_LIST=10m
CACHE_TTL_DETAIL=30m
CACHE_TTL_MAPS=1h
CACHE_TTL_<DOMAIN>=24h                  # e.g. STATIC, AREA, MASTER
```

### Domain-Specific (only if service needs)

```env
# Image upload
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name

# Email notifications
SENDGRID_API_KEY=SG.xxx
[email protected]

# Add more as needed (e.g., Twilio, S3, etc.)
```

---

## main.go — 12 Required Components

The canonical `main.go` has these components in this order:

1. **Build info variables** — populated by `runtime/debug.ReadBuildInfo()`
   ```go
   var (
       serviceName = "..."
       version     = "dev"      // overridden by ldflags in production
       commit      = "unknown"  // populated from VCS
       buildTime   = "unknown"
       startedAt   = time.Now()
   )
   ```

2. **Typed Config struct** with `loadConfig()` validation
   ```go
   type Config struct {
       Port string; Environment string; ShutdownTimeout time.Duration
       DatabaseURL string; DBMaxConns int32; ...
       // returns error if required vars missing
   }
   ```

3. **Env helpers** — `getEnv`, `getEnvInt`, `getEnvDuration`, `getEnvBool`,
   `parseList`, `parseLogLevel`

4. **Credential masking regex** for safe logging
   ```go
   reConnURI  = matches postgres://, redis://, cloudinary://, etc
   reKeyValue = matches password=xxx, token=xxx, api_key=xxx
   reBearer   = matches Bearer <token>
   ```

5. **Structured logger setup** with service metadata baked in
   ```go
   slog.New(...).With(
       slog.String("service", serviceName),
       slog.String("version", version),
       slog.String("commit", commit),
       slog.String("env", cfg.Environment),
   )
   ```

6. **DB pool with smart tracer**
   - Pool tuning from config (max/min conns, lifetime, idle time)
   - Tracer logs slow queries (>500ms) at WARN level
   - Normal queries at DEBUG level (visible only in dev)

7. **Cache initialization with fallback**
   - Try Redis first; fallback to in-memory SimpleCache if Redis URL empty
   - Log warning if falling back (not recommended for multi-replica)

8. **External services init** (best-effort)
   - Cloudinary, SendGrid, etc. — log error if init fails but don't crash

9. **Health endpoints** in `registerHealthEndpoints()`
   - `/health` → liveness check (always returns 200 if process alive)
   - `/ready` → readiness check (DB ping + cache check; 503 if not ready)
   - Uses **local types** inside function (NOT in dto_*.go — see below)

10. **Pprof gated** — bind 127.0.0.1:6060 only, disabled by default
    ```go
    if cfg.EnablePprof {
        go startPprofServer(cfg.PprofPort)
    }
    ```

11. **Background goroutines**
    - `go handler.CacheWarmup()` — pre-load popular data
    - `go registry.AutoRegister(...)` — register to gateway

12. **Graceful shutdown** in correct order
    - HTTP server (stop accepting new requests, drain in-flight)
    - Cache (close Redis if used)
    - DB pool (close connections)

---

## Database Schema Patterns

### Use UUID for Public Resources

```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```

**Why**: Prevents enumeration attacks (`/news/1, /news/2, /news/3, ...`).

**Exception**: Use `BIGSERIAL` only if the resource is internal (admin-only).

### Use ENUM for Status Fields

```sql
CREATE TYPE report_status AS ENUM ('PENDING', 'PROCESSED', 'REJECTED');

CREATE TABLE hoax_reports (
    ...
    status report_status NOT NULL DEFAULT 'PENDING'
);
```

**Why**: Type-safe at DB level. Can't insert typo `'pendng'`.

### Use CHECK Constraints for Format Validation

```sql
-- Email
reporter_email VARCHAR(150) NOT NULL 
    CHECK (reporter_email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$')

-- URL (nullable)
proof_link VARCHAR(255) 
    CHECK (proof_link IS NULL OR proof_link ~* '^https?://[^\s]+$')

-- Slug
slug VARCHAR(255) UNIQUE NOT NULL 
    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
```

**Why**: Defense in depth. Even if handler validation forgotten, DB rejects bad data.

### Auto-Update `updated_at` via Trigger

```sql
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_<table>
    BEFORE UPDATE ON <table>
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
```

**Why**: Don't rely on application code to set `updated_at`. Easy to forget.

### Generated tsvector for Full-Text Search (if applicable)

```sql
search_vector tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(description, '')), 'B')
) STORED

CREATE INDEX idx_<table>_search ON <table> USING GIN (search_vector);
```

**Why**: Title weight A > description weight B in ranking. Auto-updated.

### DB-Side Generation for Race-Prone IDs

```sql
CREATE OR REPLACE FUNCTION generate_ticket_number()
RETURNS TEXT AS $$
BEGIN
    RETURN 'XX-' || UPPER(SUBSTRING(MD5(random()::text || clock_timestamp()::text), 1, 8));
END;
$$ LANGUAGE plpgsql;

ALTER TABLE my_reports 
    ALTER COLUMN ticket_number SET DEFAULT generate_ticket_number();
```

**Why**: Race-condition-proof. Generating in Go can collide under concurrent submits.

### Anti-Patterns to Avoid

- ❌ Empty string `''` for nullable column (use `NULL`)
- ❌ Manual `INSERT` of UNIQUE field that has DB default
- ❌ `CREATE INDEX` for column that already has UNIQUE constraint (auto-creates index)
- ❌ Storing user input HTML-escaped (escape on output, not input)
- ❌ `SELECT *` — list columns explicitly

---

## Routes Pattern

```go
api := app.Group("/api/v1")

// Tiered rate limiting
strictLimiter := limiter.New(limiter.Config{
    Max:        30,           // for write/destructive actions
    Expiration: 1 * time.Minute,
})
searchLimiter := limiter.New(limiter.Config{
    Max:        50,           // for search/track endpoints
    Expiration: 1 * time.Minute,
})

// API root (gateway healthcheck)
api.Get("/", func(c fiber.Ctx) error {
    return c.SendString("MyService API v1 Active")
})

// Public group — citizen-accessible
public := api.Group("/public")
public.Get("/categories", h.GetCategories)
public.Post("/reports", strictLimiter, h.SubmitReport)
public.Get("/reports/track", searchLimiter, h.TrackReport)
public.Get("/news", searchLimiter, h.GetPublicNews)
public.Get("/news/:slug", h.GetPublicNewsDetailBySlug)

// Admin group — moderator-only (add auth middleware later)
admin := api.Group("/admin")
admin.Get("/reports", h.GetAllReportsAdmin)
admin.Post("/reports/process", strictLimiter, h.ProcessReportAdmin)
admin.Delete("/news/:id", strictLimiter, h.DeleteNewsAdmin)
```

**Why this structure?**
- ✅ Self-documenting URL (you know `/admin/...` is privileged)
- ✅ Auth middleware easy to add (`admin.Use(requireAuth)`)
- ✅ Separate rate limits per intent (search vs action vs default)

---

## Handler Patterns

### Public GET Handler (with cache)

```go
func (h *MyHandler) GetCategories(c fiber.Ctx) error {
    cacheKey := "myservice:categories:all"
    if respondCached(c, cacheKey) {
        return nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    data, err := h.Queries.GetAllCategories(ctx)
    if err != nil {
        slog.Error("public.categories.error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code:    "ERR_INTERNAL_DB",
            Message: "Failed to fetch categories",
        })
    }

    responseData := make([]CategoryItem, 0, len(data))
    for _, row := range data {
        responseData = append(responseData, CategoryItem{...})
    }

    res := SuccessResponse{Pesan: "Daftar Kategori", Data: responseData}
    return cacheJSON(c, cacheKey, CacheTTLStatic, res)
}
```

### Public POST Handler (with validation)

```go
func (h *MyHandler) SubmitReport(c fiber.Ctx) error {
    var fieldErrors []FieldError

    name, errN := utils.ValidateQueryString(c.FormValue("nama"), 150, "nama")
    if errN != nil { fieldErrors = append(fieldErrors, FieldError{...}) }

    email, errE := utils.ValidateEmail(c.FormValue("email"))
    if errE != nil { fieldErrors = append(fieldErrors, FieldError{...}) }

    // ... more validations ...

    if len(fieldErrors) > 0 {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code:    "ERR_VALIDATION",
            Message: "Input invalid",
            Errors:  fieldErrors,
        })
    }

    // ... DB write, cache invalidation, async tasks ...
}
```

### Paginated List with Parallel Queries

```go
func (h *MyHandler) GetPublicNews(c fiber.Ctx) error {
    page, limit, offset := parsePagination(c)
    keyword, _ := utils.ValidateQueryString(c.Query("search"), 100, "search")

    cacheKey := fmt.Sprintf("myservice:news:q_%s:p%d:l%d", keyword, page, limit)
    if respondCached(c, cacheKey) {
        return nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    g, gCtx := errgroup.WithContext(ctx)
    var data []db.GetNewsRow
    var total int64

    g.Go(func() error {
        var err error
        data, err = h.Queries.GetNews(gCtx, db.GetNewsParams{
            LimitData:  int32(limit),
            OffsetData: int32(offset),
        })
        return err
    })
    g.Go(func() error {
        var err error
        total, err = h.Queries.CountNews(gCtx)
        return err
    })

    if err := g.Wait(); err != nil { /* handle */ }

    res := SuccessResponse{
        Data:       data,
        Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
    }
    return cacheJSON(c, cacheKey, CacheTTLList, res)
}
```

### Admin with Transaction

```go
func (h *MyHandler) ProcessReport(c fiber.Ctx) error {
    // ... validation ...

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    tx, err := h.DB.Begin(ctx)
    if err != nil { /* handle */ }
    defer func() { _ = tx.Rollback(ctx) }()  // safe to call after Commit

    qtx := h.Queries.WithTx(tx)

    if _, err := qtx.CreateNews(ctx, ...); err != nil { /* handle */ }
    if err := qtx.UpdateReportStatus(ctx, ...); err != nil { /* handle */ }

    if err := tx.Commit(ctx); err != nil { /* handle */ }

    invalidatePublicCache()  // centralized cache invalidation

    // ... async email, logging, etc ...
}
```

---

## Cache Patterns

### TTL Tier Selection

| Use Case | Constant | Default | When to Use |
|---|---|---|---|
| List endpoints | `CacheTTLList` | 10m | Data changes regularly |
| Detail per item | `CacheTTLDetail` | 30m | Detail rarely changes once published |
| Relationships | `CacheTTLMaps` | 1h | Cross-table joins, complex aggregations |
| Master data | `CacheTTLStatic`/`Area` | 24h | Categories, areas, fixed lookup data |

### Cache Key Convention

```
{service}:{resource}:{specifier}
```

Examples:
- `hoax:categories:all`
- `hoax:stats`
- `hoax:news:public:q_<keyword>:p<page>:l<limit>`
- `hoax:news:detail:<slug>`

### Centralized Invalidation

Don't scatter `cache.GlobalCache.Delete(...)` throughout handlers. Centralize:

```go
// In <service>_admin.go
func invalidatePublicCache() {
    cache.GlobalCache.Delete("myservice:stats")
    cache.GlobalCache.DeleteByPrefix("myservice:news:public")
    cache.GlobalCache.DeleteByPrefix("myservice:news:detail")
}
```

Call this whenever data that affects public-facing cache changes.

---

## Logging Conventions

### Format

JSON structured logs via `slog` (Go 1.21+ stdlib).

### Naming: `context.event`

```go
slog.Info("public.report.created", ...)
slog.Error("admin.tx.commit_failed", ...)
slog.Warn("cache.warmup.stats_failed", ...)
slog.Info("auth.login.success", ...)
```

### Required Fields

- Error message (already credential-masked)
- Domain identifier (ticket, slug, ID)
- Request ID (if available)

### Anti-Pattern

```go
// ❌ Bad
slog.Info("Report Created", slog.String("ticket", t))

// ✅ Good
slog.Info("public.report.created",
    slog.String("ticket", t),
    slog.String("email", maskEmail(e)),
)
```

---

## DTO Conventions

### Where to put what

| Type | File | Reason |
|---|---|---|
| Used by every handler (`ErrorResponse`, etc.) | `dto_base.go` | Universal |
| Domain-specific (Request + Response) | `dto_<domain>.go` | Per-feature |
| Used in 1 function only | Local type inside function | Encapsulation |
| Health/ready response | Local type in `main.go` | Operational, not API contract |

**Anti-pattern**: putting `LivenessResponse` / `ReadinessResponse` in `dto_*.go`.
These are operational details, not user-facing API contracts.

### Request vs Response DTOs

Always have **both** for write endpoints:

```go
// Request — bound to incoming body/form
type CreateReportRequest struct {
    ReporterName  string `json:"reporter_name"`
    ReporterEmail string `json:"reporter_email"`
    // ...
}

// Response — what API returns
type TicketCreatedResponse struct {
    TicketNumber string `json:"ticket_number"`
    CreatedAt    string `json:"created_at"`
}
```

### Use `omitempty` for Optional Fields

```go
type ReportTrackingResponse struct {
    ReportID  string `json:"report_id"`
    NewsID    string `json:"news_id,omitempty"`     // optional
    NewsSlug  string `json:"news_slug,omitempty"`   // optional
}
```

---

## What's NOT Part of the Blueprint

These are **service-specific**, NOT pattern violations:

- Table names (`hoax_news`, `kendaraan_pajak`, `destinasi_wisata`)
- Cache key prefixes (`hoax:`, `pajak:`, `wisata:`)
- Ticket/ID format (`KH-`, `BPN-`, `WST-`)
- TTL values (klinik 24h vs bapenda 12h — different data freshness needs)
- Domain DTO fields
- External integrations (Cloudinary in klinik, none in bapenda)
- Env var domain naming (`CACHE_TTL_STATIC` vs `CACHE_TTL_AREA`)

**The pattern is consistent. The implementation is domain-specific.**

---

## Migration Order Recommendation

The order in which to migrate remaining services matters:

1. **sidita-service** ← Next, similar complexity to klinik (Cloudinary, multi-domain)
2. **siskaperbapo-service** ← Simpler, mostly read-heavy
3. **user-service** ← Auth concerns, do after pattern is stable
4. **api-gateway** ← Last, since it integrates all services

**Rationale**: Start with services most similar to klinik (so the blueprint is
directly applicable). Save the unique services (auth, gateway) for last when
you've internalized the pattern.

---

## Future Phase 2: Shared Module

After **3+ services** stable on this blueprint, extract common code to
`majadigi-go-shared`:

**Candidates for extraction**:
- `validator` package (email, URL, UUID, pagination)
- `cache` interface + `SimpleCache` + `RedisCache`
- `httputil` for credential masking, structured error responses
- `health` package for `/health` `/ready` handlers
- `buildinfo` for version/commit/buildTime helpers

**DO NOT extract until**:
- ✅ At least 3 services use the same pattern (Rule of Three)
- ✅ All 3 have been running stable for 2+ weeks
- ✅ Pattern hasn't changed in last 2 services migrated

**See**: `MIGRATION_GUIDE.md` for step-by-step migration instructions.

---

## Quick Reference Map

| Topic | Look in `klinik-service/` |
|---|---|
| Enterprise main.go | `cmd/api/main.go` |
| LRU cache + stats | `internal/cache/cache.go` |
| Redis cache + pool | `internal/cache/cache_redis.go` |
| Cache helpers | `internal/handlers/handler_utils.go` |
| Validators | `internal/utils/validator.go` |
| Slug generator | `internal/utils/slug.go` |
| Public handler | `internal/handlers/klinik_public.go` |
| Admin + transaction | `internal/handlers/klinik_admin.go` |
| Setup + cache warmup | `internal/handlers/klinik_setup.go` |
| Schema enterprise | `db/migrations/001.schema.sql` |
| Queries enterprise | `db/queries/queries.sql` |
| Routes pattern | `internal/routes/api.go` |
| Env tuning | `.env.encrypted` |

---

**Maintainer**: farildzaky  
**Reference**: `klinik-service` @ tag `klinik-blueprint-v1.0`  
**See also**: `MIGRATION_GUIDE.md` for service migration steps
