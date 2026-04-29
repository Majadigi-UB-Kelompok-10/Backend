# Service Migration Guide

> Step-by-step playbook to migrate a service to the canonical blueprint
> pattern (`klinik-blueprint-v1.0`).

This guide is a **checklist**. Follow it in order. Each phase has a verification
step — if it fails, fix before moving on. **Don't skip phases**.

---

## Overview: 5 Phases

```
Phase 1: Foundation    — Cache, validators, slug (no handler changes yet)
Phase 2: DTOs          — Request/Response types
Phase 3: Schema        — DB migration + sqlc regenerate
Phase 4: Handlers      — Refactor to use new helpers
Phase 5: Verification  — Build, runtime, smoke test, push, CI
```

Each phase **builds successfully on its own**. If you finish Phase 1 and stop
there, the codebase compiles. This minimizes risk.

---

## Pre-Flight Checklist

Before starting migration:

```bash
# 1. Make sure klinik-blueprint is up to date locally
cd /Users/dzaky/programming/CAPSTONE-BE
git fetch --tags
git checkout klinik-blueprint-v1.0
ls klinik-service/   # familiarize with the structure

# 2. Create migration branch
git checkout dev/full_capstone_be      # or your base branch
git pull origin dev/full_capstone_be
git checkout -b dev/<service>_migration

# 3. Verify klinik builds locally (sanity check)
cd klinik-service
go build ./cmd/api
echo "Klinik build: $?"  # MUST be 0
```

If klinik build fails, **stop**. The blueprint reference is broken; investigate
before attempting migration.

---

## Phase 1: Foundation

### 1.1 Cache Layer

**Action**: Replace cache files with klinik versions.

```bash
SERVICE=<service-name>   # e.g., sidita-service

cd /Users/dzaky/programming/CAPSTONE-BE/$SERVICE

# Backup originals
cp internal/cache/cache.go internal/cache/cache.go.bak
cp internal/cache/cache_redis.go internal/cache/cache_redis.go.bak

# Copy from klinik
cp ../klinik-service/internal/cache/cache.go internal/cache/cache.go
cp ../klinik-service/internal/cache/cache_redis.go internal/cache/cache_redis.go
```

**Customize**: Adjust `RedisCache` pool tuning per service traffic profile:
- High-traffic public-facing service → `PoolSize=100, MinIdleConns=10`
- Internal/admin service → `PoolSize=30, MinIdleConns=3`

### 1.2 Validators

```bash
cp ../klinik-service/internal/utils/validator.go internal/utils/validator.go
```

**Note**: This adds `ValidateEmail`, `ValidatePhone`, `ValidateURL`, etc. Some
might be unused in your service — leave them. Unused validators don't hurt.

### 1.3 Slug Generator (only if needed)

If your service has URL-friendly identifiers (`/destinasi/<slug>`,
`/event/<slug>`, etc.):

```bash
cp ../klinik-service/internal/utils/slug.go internal/utils/slug.go
```

If not, **skip**. Don't add code "just in case" (YAGNI).

### 1.4 Verification

```bash
go build ./internal/cache/
echo "Cache: $?"          # MUST be 0

go build ./internal/utils/
echo "Utils: $?"          # MUST be 0
```

If both exit 0, **Phase 1 complete**. Phase 2 next.

If errors, paste them — likely:
- Import paths point to klinik (find/replace `klinik-service` → `<your-service>`)
- Module path different (check `go.mod`)

---

## Phase 2: DTOs

### 2.1 Base DTO

```bash
# Copy dto_base.go from klinik (universal, shouldn't differ)
cp ../klinik-service/internal/handlers/dto_base.go internal/handlers/dto_base.go
```

This contains `ErrorResponse`, `SuccessResponse`, `FieldError`, `PaginationMeta`.

### 2.2 Domain DTO (already exists, just adjust)

Your service likely has `dto_<domain>.go` already. Update to:

- ✅ Add `omitempty` to optional fields
- ✅ Create separate Request DTOs for POST endpoints (`Create<X>Request`)
- ✅ Match column names from your schema
- ✅ Add comments per struct explaining purpose

**Don't** copy `dto_klinik.go` blindly — your domain is different.

### 2.3 Verification

```bash
go build ./internal/handlers/  # may have errors if handlers reference old fields
                               # That's OK; we'll fix in Phase 4.

# Just check DTO file itself compiles:
go vet ./internal/handlers/dto_base.go ./internal/handlers/dto_<domain>.go
```

---

## Phase 3: Schema + sqlc

### 3.1 Schema Migration

Update `db/migrations/001.schema.sql`:

- Add `CHECK` constraints for emails, URLs, slugs (where applicable)
- Add `trigger_set_timestamp` trigger if you have `updated_at` columns
- Use `UUID` for public-facing PKs (replace `BIGSERIAL` if currently used)
- Add ENUMs for status fields
- Drop redundant `CREATE INDEX` for UNIQUE columns (already auto-indexed)

**Warning**: If your service is in production with existing data, schema changes
need real migration (ALTER TABLE), not just rewrite. For dev/test, wipe volume
and rebuild from scratch.

### 3.2 Update Queries

Edit `db/queries/queries.sql`:

- Replace `SELECT *` with explicit columns
- Add per-section comments
- Use `sqlc.narg(...)` for nullable args, `sqlc.arg(...)` for required
- Add `RETURNING` clauses where you need the inserted row's auto-generated values

### 3.3 Regenerate sqlc

```bash
sqlc generate
```

If errors:
- "column not found in schema" → schema and query out of sync
- "unsupported type" → check `sqlc.yaml` config

### 3.4 Verification

```bash
go build ./internal/db/    # MUST be 0; if not, sqlc generated bad code

# Check generated code looks reasonable
ls -la internal/db/        # should have models.go, queries.sql.go, db.go
```

---

## Phase 4: Handlers

This is the most involved phase. Order: utility files first, then handlers.

### 4.1 handler_utils.go

```bash
cp ../klinik-service/internal/handlers/handler_utils.go internal/handlers/handler_utils.go
```

**Customize**:
- Change `CACHE_TTL_<DOMAIN>` env var name to match your `.env`
- Adjust `CacheTTLList`, `CacheTTLDetail`, `CacheTTLMaps` defaults if needed
- Adjust import paths: `farildzaky/klinik-service` → `farildzaky/<your-service>`

### 4.2 Setup File

Create `internal/handlers/<service>_setup.go` based on
`klinik-service/internal/handlers/klinik_setup.go`. Include:

- `Handler` struct (Queries, DB, optional Cld for Cloudinary)
- Constructor `NewXxxHandler(...)`
- pgtype helpers (`pgTextToStr`, `pgUUIDToStr`, `parseUUID`)
- `uploadImageReal` if your service uploads images
- `sendEmailAsync` if your service sends emails
- `CacheWarmup` for popular pre-loaded data

### 4.3 Public + Admin Handlers

Refactor existing `<service>_public.go` and `<service>_admin.go`:

- Use `respondCached(c, key)` instead of `cache.GlobalCache.Get(...)` + `c.Send(...)`
- Use `cacheJSON(c, key, ttl, data)` instead of `cache.GlobalCache.Set` + `c.JSON`
- Use `parsePagination(c)` instead of manual page/limit parsing
- Use `validationErrorResponse(c, err)` for `*utils.ValidationError`
- Use structured logging with dot notation (`public.x.error`, not "X Error")
- Use `errgroup` for parallel queries (e.g., list + count)
- Use `tx.Begin / Commit / Rollback` for atomic multi-write operations
- Centralize cache invalidation in `invalidatePublicCache()`

**Common adjustments**:
- Different domain → different cache key prefix
- No image upload → remove Cloudinary code
- No email → remove SendGrid code

### 4.4 Verification

```bash
go build ./cmd/api
echo "Build cmd/api: $?"   # MUST be 0

go build ./...
echo "Build full: $?"      # MUST be 0
```

If errors, paste them. Common ones:
- `unknown field X in db.YParams` → schema/query/sqlc out of sync
- `cannot use Y as Z` → DTO field type mismatch
- `undefined: parsePagination` → forgot to copy handler_utils.go

---

## Phase 5: Verification

### 5.1 Routes Update

Edit `internal/routes/api.go`:

- Use `app.Group("/api/v1")` then sub-groups `/public` and `/admin`
- Tiered rate limit: `strictLimiter` (30/min) for writes, `searchLimiter`
  (50/min) for search/track
- Adjust handler method names to match yours

### 5.2 Env Vars

Update `.env.encrypted` with the 13 tuning vars from BLUEPRINT.md:

```bash
# Decrypt
dotenvx decrypt

# Edit .env, append the 13 vars
# (See BLUEPRINT.md "Tuning" section)

# Re-encrypt
dotenvx encrypt

# Cleanup plaintext
rm .env
```

### 5.3 Local Build + Runtime

```bash
go build ./cmd/api
echo "Build: $?"  # 0

go build ./...
echo "Full: $?"   # 0

# Wipe DB volume + restart
cd /Users/dzaky/programming/CAPSTONE-BE
docker compose down

# Hapus volume service ini DOANG (ganti nama sesuai compose project kamu)
docker volume rm capstone-project_<service>-postgres-data
docker volume rm capstone-project_<service>-redis-data 2>/dev/null

docker compose up -d
sleep 15

# Cek log service-db (no errors)
docker compose logs <service>-db --tail=50

# Cek log service-api (listening, registered)
docker compose logs <service>-api --tail=30
```

### 5.4 Smoke Test Endpoints

```bash
PORT=<service-port>   # check docker-compose.yml ports

curl http://localhost:$PORT/health
curl http://localhost:$PORT/ready
curl http://localhost:$PORT/api/v1/
curl http://localhost:$PORT/api/v1/public/<endpoint>
# etc
```

All should return HTTP 200 with valid JSON. If 404, check route paths.
If 500, paste service-api log.

### 5.5 Commit + Push

Atomic commits per topic:

```bash
git add <service>/.env.encrypted
git commit -m "feat(<service>): add enterprise tuning env vars"

git add <service>/db/
git commit -m "perf(<service>): improve schema with constraints, triggers, ENUM"

git add <service>/internal/db/
git commit -m "chore(<service>): regenerate sqlc"

git add <service>/internal/cache/
git commit -m "refactor(<service>): enterprise cache with LRU, stats"

git add <service>/internal/utils/
git commit -m "refactor(<service>): add comprehensive validators"

git add <service>/internal/handlers/
git commit -m "refactor(<service>): apply blueprint handler patterns"

git add <service>/internal/routes/
git commit -m "refactor(<service>): tiered rate limit + public/admin grouping"

git add <service>/cmd/api/main.go
git commit -m "refactor(<service>): finalize main.go enterprise pattern"

git push origin dev/<service>_migration
```

### 5.6 CI Verification

Open GitHub Actions tab. Wait for all workflow runs to be green:
- ✅ `<service>` build job
- ✅ Other services not broken (pattern doesn't affect them)

If red, paste log of failed step.

### 5.7 Optional: Tag

After CI green, tag if it's a meaningful milestone:

```bash
git tag -a <service>-migration-v1.0 -m "Migrated <service> to blueprint v1.0"
git push origin <service>-migration-v1.0
```

---

## Common Issues & Fixes

### Issue: `cannot find module providing package`

```
go: github.com/farildzaky/<service>/internal/utils: no Go files in ...
```

**Cause**: Module path mismatch.

**Fix**:
```bash
# Check go.mod
head -1 go.mod
# Make sure imports match this module path
```

### Issue: `undefined: parsePagination`

**Cause**: `handler_utils.go` not present or wrong content.

**Fix**: Copy from klinik again, adjust import paths.

### Issue: Schema reload not happening

```
docker compose up
# but old data still there
```

**Cause**: Volume not wiped.

**Fix**:
```bash
docker compose down
docker volume rm capstone-project_<service>-postgres-data
docker compose up -d
```

### Issue: `violates check constraint`

```
ERROR: new row for relation "X" violates check constraint
```

**Cause**: Seed data doesn't match new constraints (e.g., empty string for
nullable URL field).

**Fix**: Update seed:
- `''` → `NULL`
- Slugs lowercase kebab-case
- Emails plain (not markdown)
- URLs with `http://` or `https://`

### Issue: Build fails after sqlc generate

```
internal/db/queries.sql.go:XXX: undefined: ...
```

**Cause**: schema and queries out of sync.

**Fix**: Make sure all column references in queries.sql exist in schema.sql.

---

## Service-Specific Notes

### sidita-service

- Has multiple domains (destinasi, hotel, event)
- Multiple `dto_*.go` files (already structured well)
- Uses Cloudinary (similar to klinik)
- Should follow klinik blueprint closely

### siskaperbapo-service

- Read-heavy (price/commodity data)
- Likely no Cloudinary, no SendGrid
- Simpler than klinik — fewer DTOs
- Cache TTL probably longer (data updates daily, not realtime)

### user-service

- Adds **authentication** dimension (JWT, password hashing)
- Schema includes sessions, tokens
- Different rate limit profile (login attempts)
- Migrate AFTER you're confident with the pattern

### api-gateway

- Different beast — proxies requests to other services
- No DB models per se
- Service registry consumer (not producer)
- Migrate LAST when other services are stable

---

## When Things Go Sideways

If migration becomes a mess and you want to start over:

```bash
git stash                              # save current work
git checkout dev/<service>_migration   # back to start
git reset --hard origin/dev/full_capstone_be  # nuke local changes
# Or: git reset --hard <last-good-commit>

# Start fresh, follow phases in order, don't skip
```

**Don't be afraid to start over**. Better to migrate cleanly than push messy code.

---

## After Migration

After migrating a service:

1. ✅ Update BLUEPRINT.md if you discovered a better pattern
2. ✅ Add service to "migrated services" list (if you keep one)
3. ✅ Document any service-specific deviations from the blueprint
4. ✅ If the blueprint pattern needs an update, do it in `klinik-service` first,
   then migrate the change to other services

---

**See also**: 
- `BLUEPRINT.md` — pattern reference
- `klinik-service/` — concrete implementation
- Tag `klinik-blueprint-v1.0` — frozen reference point
