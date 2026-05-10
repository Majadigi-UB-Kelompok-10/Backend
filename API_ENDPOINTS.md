# Majadigi API Endpoints

> Base URL gateway (development): `http://localhost:<GATEWAY_PORT>`
> User-service diakses via gateway slug proxy: `/api/v1/<slug>/...`

---

## Autentikasi & User (user-service)

### Public (tanpa token)

| Method | Path | Body | Keterangan |
|--------|------|------|------------|
| `POST` | `/auth/register` | JSON (lihat bawah) | Daftar akun baru, kirim email verifikasi |
| `GET` | `/auth/verify-email` | `?token=xxx` | Verifikasi email dari link yang dikirim |
| `POST` | `/auth/login` | `{ "email", "password" }` | Login (email harus sudah diverifikasi) |
| `POST` | `/auth/refresh` | `{ "refresh_token" }` | Perbarui access token |
| `POST` | `/auth/logout` | `{ "refresh_token" }` | Revoke refresh token |

**Register body:**
```json
{
  "first_name": "Budi",
  "last_name": "Santoso",
  "phone": "+6281234567890",
  "email": "budi@email.com",
  "nik": "3512345678901234",
  "address": "Jl. Mawar No. 1, Surabaya",
  "birth_date": "01/17/1998",
  "gender": "LAKI_LAKI",
  "password": "min8karakter",
  "confirm_password": "min8karakter"
}
```
> Gender: `LAKI_LAKI` | `PEREMPUAN` · `birth_date` format: `mm/dd/yyyy`

---

### Protected — User (perlu `Authorization: Bearer <access_token>`)

| Method | Path | Body / Query | Keterangan |
|--------|------|-------------|------------|
| `GET` | `/auth/me` | — | Profil user sendiri |
| `PUT` | `/auth/me` | `{ "first_name", "last_name", "phone", "address", "birth_date", "gender" }` | Update profil |
| `GET` | `/auth/preferences` | — | Ambil daftar kategori pilihan user |
| `POST` | `/auth/preferences` | `{ "category_ids": ["uuid", ...] }` | Simpan semua pilihan kategori (replace all) |
| `GET` | `/auth/favorites` | — | Ambil daftar service favorit user |
| `POST` | `/auth/favorites/:service_id` | — | Tambah service ke favorit |
| `DELETE` | `/auth/favorites/:service_id` | — | Hapus service dari favorit |

---

### Protected — Admin (perlu role `admin` atau `superadmin`)

| Method | Path | Query | Keterangan |
|--------|------|-------|------------|
| `GET` | `/admin/users` | `?page=1&limit=20` | List semua user |
| `GET` | `/admin/users/:id` | — | Detail user |
| `DELETE` | `/admin/users/:id` | — | Nonaktifkan user |
| `PUT` | `/admin/users/:id/role` | `{ "role": "admin" }` | Ubah role (superadmin only) |

> Role valid: `user` \| `admin` \| `superadmin`

---

## Flow Utama Aplikasi

```
1. Daftar
   POST /auth/register  →  201 "cek email Anda"

2. Verifikasi email
   GET  /auth/verify-email?token=xxx  →  200 "email terverifikasi"

3. Login
   POST /auth/login  →  { access_token, refresh_token }

4. Pertama kali masuk — pilih kategori
   GET  /api/v1/categories              ← tampilkan semua (gunakan is_popular untuk grouping)
   POST /auth/preferences               ← simpan pilihan user

5. Halaman Beranda
   GET  /auth/favorites                 ← layanan favorit user (icon grid)

6. Halaman Layanan
   GET  /api/v1/services                ← semua layanan
   GET  /api/v1/services?search=trans   ← cari layanan
   POST /auth/favorites/:service_id     ← tambah ke favorit  (+)
   DELETE /auth/favorites/:service_id   ← hapus dari favorit (-)

7. Klik layanan  →  langsung diarahkan ke halaman microservice via slug proxy
```

---

## Kategori (api-gateway)

| Method | Path | Query | Keterangan |
|--------|------|-------|------------|
| `GET` | `/api/v1/categories` | — | Semua kategori (termasuk field `is_popular`) |
| `GET` | `/api/v1/categories/:id` | — | Detail satu kategori |
| `GET` | `/api/v1/categories/:id/services` | — | Semua service dalam kategori ini |

### Daftar Kategori

| UUID | Nama | Populer |
|------|------|---------|
| `b1000001-0000-4000-8000-000000000001` | Layanan Darurat | ✅ |
| `b1000001-0000-4000-8000-000000000002` | Transportasi Publik | ✅ |
| `b1000001-0000-4000-8000-000000000003` | Layanan Pajak | ✅ |
| `b1000001-0000-4000-8000-000000000004` | Informasi Ketenagakerjaan | ✅ |
| `b1000001-0000-4000-8000-000000000005` | Produk Hukum | — |
| `b1000001-0000-4000-8000-000000000006` | Cek Hoax | — |
| `b1000001-0000-4000-8000-000000000007` | Bantuan Sosial | — |
| `b1000001-0000-4000-8000-000000000008` | Harga Kebutuhan Pokok | — |
| `b1000001-0000-4000-8000-000000000009` | Eksplorasi Pariwisata | — |
| `b1000001-0000-4000-8000-000000000010` | Layanan Kesehatan | — |

---

## Services (api-gateway)

| Method | Path | Query | Keterangan |
|--------|------|-------|------------|
| `GET` | `/api/v1/services` | `?search=keyword` | Semua service; opsional filter nama |
| `GET` | `/api/v1/services/personalized` | `?category_ids=uuid1,uuid2` | Services sesuai kategori pilihan |
| `GET` | `/api/v1/services/:id` | — | Detail satu service |
| `GET` | `/api/v1/services/:id/categories` | — | Kategori milik service ini |

---

## Setup Superadmin (manual via DB)

Setelah register normal, jalankan SQL berikut di `user-db`:

```sql
UPDATE users
SET role = 'superadmin', email_verified = true
WHERE email = 'email_kamu@domain.com';
```

Jalankan via:
```bash
docker exec -it user-db psql -U dzaky -d user_db
```

---

## Admin CRUD — api-gateway (perlu token admin)

> Header: `Authorization: Bearer <access_token>` (role `admin` atau `superadmin`)

### Services

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/services` | Tambah service baru |
| `PUT` | `/api/v1/admin/services/:id` | Edit service |
| `DELETE` | `/api/v1/admin/services/:id` | Hapus service |
| `POST` | `/api/v1/admin/services/:id/categories` | Masukkan service ke kategori |
| `DELETE` | `/api/v1/admin/services/:id/categories/:category_id` | Keluarkan service dari kategori |

**POST `/api/v1/admin/services`** — `title` dan `icon_url` wajib:
```json
{
  "title": "Nama Layanan",
  "description": "Deskripsi layanan (opsional)",
  "icon_url": "https://cdn.example.com/icon.png"
}
```

**PUT `/api/v1/admin/services/:id`** — sama dengan POST, semua field wajib:
```json
{
  "title": "Nama Layanan Baru",
  "description": "Deskripsi baru",
  "icon_url": "https://cdn.example.com/icon-baru.png"
}
```

**POST `/api/v1/admin/services/:id/categories`** — assign service ke kategori:
```json
{ "category_list_id": "b1000001-0000-4000-8000-000000000002" }
```

### Kategori

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/categories` | Tambah kategori baru |
| `PUT` | `/api/v1/admin/categories/:id` | Edit kategori |
| `DELETE` | `/api/v1/admin/categories/:id` | Hapus kategori |

**POST & PUT `/api/v1/admin/categories`** — `name` wajib:
```json
{
  "name": "Nama Kategori",
  "description": "Deskripsi (opsional)"
}
```

### Endpoints (slug proxy)

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/endpoints` | Daftarkan microservice baru |
| `PUT` | `/api/v1/admin/endpoints/:id` | Edit endpoint |
| `DELETE` | `/api/v1/admin/endpoints/:id` | Hapus endpoint |

**POST & PUT `/api/v1/admin/endpoints`**:
```json
{
  "slug_name": "transjatim",
  "page_url": "http://transjatim-service:8080"
}
```
> `slug_name` dipakai sebagai prefix di proxy: `/api/v1/transjatim/...`

### Images

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/images` | Upload gambar untuk service |
| `PUT` | `/api/v1/admin/images/:id` | Edit gambar |
| `DELETE` | `/api/v1/admin/images/:id` | Hapus gambar |

**POST `/api/v1/admin/images`**:
```json
{
  "service_list_id": "<uuid-service>",
  "image_url": "https://cdn.example.com/image.jpg",
  "semantic_label": "Foto gedung kantor"
}
```

**PUT `/api/v1/admin/images/:id`**:
```json
{
  "image_url": "https://cdn.example.com/image-baru.jpg",
  "semantic_label": "Label baru"
}
```

### Integrations

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/integrations` | Tambah integrasi |
| `PUT` | `/api/v1/admin/integrations/:id` | Edit integrasi |
| `DELETE` | `/api/v1/admin/integrations/:id` | Hapus integrasi |

**POST `/api/v1/admin/integrations`**:
```json
{
  "service_list_id": "<uuid-service>",
  "endpoint_list_id": "<uuid-endpoint>",
  "title": "Nama Integrasi",
  "icon_url": "https://cdn.example.com/icon.png"
}
```

**PUT `/api/v1/admin/integrations/:id`**:
```json
{
  "endpoint_list_id": "<uuid-endpoint-baru>"
}
```

### Operational

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/operational` | Set jam operasional service |
| `PUT` | `/api/v1/admin/operational/service/:service_id` | Edit jam operasional |
| `DELETE` | `/api/v1/admin/operational/service/:service_id` | Hapus jam operasional |

**POST & PUT `/api/v1/admin/operational`**:
```json
{
  "service_list_id": "<uuid-service>",
  "service_url": "https://layanan.jatimprov.go.id",
  "address": "Jl. Pahlawan No. 1, Surabaya",
  "operational_hour": {
    "senin_jumat": "08:00 - 16:00",
    "sabtu": "08:00 - 12:00",
    "minggu": "Tutup"
  },
  "social_media": {
    "instagram": "@majadigi",
    "twitter": "@majadigi"
  }
}
```
> `operational_hour` dan `social_media` bebas struktur JSON-nya.

### Policies

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/v1/admin/policies` | Tambah kebijakan layanan |
| `PUT` | `/api/v1/admin/policies/service/:service_id` | Edit kebijakan |
| `DELETE` | `/api/v1/admin/policies/service/:service_id` | Hapus kebijakan |

**POST `/api/v1/admin/policies`**:
```json
{
  "service_list_id": "<uuid-service>",
  "benefit": ["Akses 24 jam", "Gratis biaya admin"],
  "instruction": ["Buka aplikasi", "Pilih menu layanan", "Isi formulir"]
}
```

**PUT `/api/v1/admin/policies/service/:service_id`**:
```json
{
  "benefit": ["Benefit baru"],
  "instruction": ["Langkah 1", "Langkah 2"]
}
```
> `benefit` dan `instruction` bebas — bisa array string atau object.

---

## Proxy Dinamis (microservices)

> Format: `<METHOD> /api/v1/<slug>/<path>`

| Slug | Service |
|------|---------|
| *(terdaftar di tabel endpoint_list)* | Sesuai `page_url` di DB |

Contoh: `GET /api/v1/transjatim/buses` → diteruskan ke transjatim-service

---

## Catatan

- **Access token** berlaku **15 menit** — perbarui via `POST /auth/refresh`
- **Refresh token** berlaku **30 hari** — simpan di secure storage
- Registrasi tidak langsung dapat token; user harus verifikasi email dulu sebelum bisa login
- Gunakan `is_popular: true` pada response `/api/v1/categories` untuk kelompok "Populer" di layar onboarding
- Klik service di halaman Layanan → tidak ada detail page, langsung masuk via slug proxy ke microservice
