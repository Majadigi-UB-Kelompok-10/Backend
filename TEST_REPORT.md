# Test Report — CAPSTONE-BE

## Unit Test

### api-gateway

| File | Test Case | Hasil |
|---|---|---|
| `middleware/auth_test.go` | `TestRequireAuth_NoHeader` — tanpa Authorization header → 401 | PASS |
| | `TestRequireAuth_MalformedHeader` — header bukan "Bearer ..." → 401 | PASS |
| | `TestRequireAuth_InvalidToken` — token bukan JWT valid → 401 | PASS |
| | `TestRequireAuth_WrongSecret` — token ditandatangani secret salah → 401 | PASS |
| | `TestRequireAuth_ExpiredToken` — token sudah kedaluwarsa → 401 | PASS |
| | `TestRequireAuth_RefreshTokenRejected` — refresh token dipakai sebagai access → 401 | PASS |
| | `TestRequireAuth_ValidToken` — access token valid → 200 | PASS |
| | `TestRequireAdmin_UserRoleBlocked` — role "user" pada rute admin → 403 | PASS |
| | `TestRequireAdmin_AdminRoleAllowed` — role "admin" pada rute admin → 200 | PASS |
| | `TestRequireAdmin_SuperadminAllowed` — role "superadmin" pada rute admin → 200 | PASS |
| | `BenchmarkRequireAuth` — throughput validasi JWT | PASS |
| `routes/routes_test.go` | `TestRoutes_UnknownSlug_404` — slug tidak terdaftar di registry → 404 | PASS |
| | `TestRoutes_Admin_NoToken_401` — `/:slug/admin/*` tanpa token → 401 | PASS |
| | `TestRoutes_Admin_UserRole_403` — `/:slug/admin/*` dengan role "user" → 403 | PASS |
| | `TestRoutes_Admin_AdminToken_200` — `/:slug/admin/*` dengan token admin → diproksi (200) | PASS |
| | `TestRoutes_Admin_SuperadminToken_200` — `/:slug/admin/*` dengan token superadmin → diproksi (200) | PASS |
| | `TestRoutes_Public_NoToken_200` — `/:slug/public/*` tanpa token → diproksi (200) | PASS |
| | `TestRoutes_QueryString_Forwarded` — query string diteruskan ke downstream | PASS |
| | `TestRoutes_Registry_ConcurrentAccess` — 50 goroutine baca+tulis GatewayRegistry bersamaan (race detector aktif) | PASS |

### bansos-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | ValidateNIK, StatusPenyaluran, MetodePenyaluran, Nominal, DateString, Periode, Pagination, QueryString, TextContent | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### bapenda-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | SanitizeQueryString, ParsePagination, ValidatePlatNomor, ValidateNomorRangka, SanitizeFilename, ValidationError | PASS |

### jdih-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | QueryString, Tahun (1901–2200), JenisDokumen, StatusDokumen, PDFFilename, Pagination, TextContent, URL | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### klinik-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | Email, Phone, URL, UUID, PaginationParams, ImageFilename | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### rssa-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | SanitizeQueryString, ValidateSlug, ParsePagination | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### sidita-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | Email, Phone, URL, Slug, QueryString, TextContent, PaginationParams, ImageFilename | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### sinaker-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | NIK, Email, Phone, Date, JenisKelamin, Pendidikan, StatusPendaftaran, Slug, TextContent, Pagination, URL | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### siskaperbapo-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | QueryString, Slug, URL, Date, Pagination, TextContent | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### transjatim-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | QueryString, Slug, URL, Date, Pagination, TextContent | PASS |
| `utils/slug_test.go` | GenerateSlug — format, uniqueness, karakter aman, panjang maks | PASS |

### user-service

| File | Test Case | Hasil |
|---|---|---|
| `utils/validator_test.go` | Email, Phone (+62), NIK, Password, BirthDate, Pagination | PASS |
| `utils/password_test.go` | HashPassword, VerifyPassword, bcrypt cost 12, salt uniqueness | PASS |
| `utils/jwt_test.go` | GenerateAccessToken, GenerateRefreshToken, ValidateToken, HashToken, expired token rejection | PASS |

---

## `make test` — Ringkasan

```
api-gateway          PASS
bansos-service       PASS
bapenda-service      PASS
jdih-service         PASS
klinik-service       PASS
rssa-service         PASS
sidita-service       PASS
sinaker-service      PASS
siskaperbapo-service PASS
transjatim-service   PASS
user-service         PASS
```

**11/11 service PASS** — 22 file test, ~160 test case, 11 benchmark.

---

## Coverage

Coverage diukur per package dengan `go test -coverprofile`. Nilai *utils* adalah coverage package utils saja; nilai *module* mencakup seluruh package termasuk handler, repository, db yang belum ada tesnya.

| Service | Coverage utils | Coverage module |
|---|---|---|
| api-gateway | middleware 95.5%, routes 13.3% | 3.4% |
| bansos-service | 94.2% | 7.7% |
| bapenda-service | 97.8% | 4.7% |
| jdih-service | 91.8% | 6.8% |
| klinik-service | 76.0% | 5.5% |
| rssa-service | 93.5% | 6.3% |
| sidita-service | 84.3% | 4.2% |
| sinaker-service | 83.9% | 7.7% |
| siskaperbapo-service | 78.8% | 5.5% |
| transjatim-service | 92.6% | 4.5% |
| user-service | 83.1% | 5.3% |

---

## Load Test — 10.000 User Simultan

**Tool**: `scripts/loadtest/main.go` — Go goroutine, tiap request IP unik (`X-Forwarded-For`)  
**Target**: `http://localhost:8888/health`

| Metrik | Nilai |
|---|---|
| Total request | 10.000 |
| Sukses (2xx) | 10.000 (100%) |
| Throughput | ~26.800 req/s |
| p50 | ~3 ms |
| p99 | ~37 ms |
| Max | ~85 ms |

---

## E2E Test

**Script**: `scripts/e2e/main.go` | Jalankan: `make e2e`

| Step | Endpoint | Ekspektasi |
|---|---|---|
| 1 | `GET /health` | 200 |
| 2 | `GET /api/v1/categories` | 200 (public, tanpa token) |
| 3 | `POST /api/v1/admin/categories` tanpa token | 401 |
| 4 | `POST /api/v1/auth/login` password salah | 401 |
| 5 | `POST /api/v1/auth/login` kredensial valid | 200 + access_token |
| 6 | `GET /api/v1/auth/me` dengan token | 200 |
| 7 | `GET /api/v1/auth/me` tanpa token | 401 |
| 8 | `POST /api/v1/sinaker-api/admin/blk` tanpa token | 401 (dynamic routing auth) |

---

## Keamanan yang Diterapkan

| Kontrol | Detail |
|---|---|
| **Rate Limiting — umum** | 100 req/IP/menit via Fiber v3 + Redis |
| **Rate Limiting — login** | 5 req/IP/menit (brute-force protection) |
| **Rate Limiting — register** | 10 req/IP/menit |
| **Auth gateway — rute statis** | `/api/v1/admin/*` wajib token admin |
| **Auth gateway — rute dinamis** | `/:slug/admin/*` wajib token admin (di-enforce sebelum proxy ke downstream) |
| **JWT** | HMAC-SHA256, access 15 menit, refresh 7 hari, refresh token di-hash SHA-256 di DB |
| **Password** | bcrypt cost 12, salt otomatis |
| **Input Validation** | Validator domain-spesifik per service (NIK, plat nomor, tanggal, enum, dll) |
| **XSS / Path Traversal** | `SanitizeQueryString`, `SanitizeFilename` membuang karakter berbahaya dan `../` |
| **SQL Injection** | Seluruh query parameterized (sqlc-generated) |
| **URL Validation** | Hanya `http://` dan `https://` diterima |
| **Proxy Trust** | `TrustProxyConfig{Loopback: true, Private: true}` untuk XFF dari Docker bridge |
