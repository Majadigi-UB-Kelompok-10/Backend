-- =============================================================================
-- KLINIK HOAX SERVICE — SEED DATA
-- Compatible dengan schema baru (auto-ticket, slug check, URL validation)
-- =============================================================================

-- 1. CLEAN SLATE
TRUNCATE TABLE hoax_news, hoax_reports, hoax_categories CASCADE;

-- =============================================================================
-- 2. CATEGORIES
-- =============================================================================
INSERT INTO hoax_categories (id, name, slug, icon_url) VALUES
('11111111-1111-1111-1111-111111111111', 'Berita Hoaks', 'berita-hoaks', 'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg'),
('22222222-2222-2222-2222-222222222222', 'Disinformasi', 'disinformasi', 'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg'),
('33333333-3333-3333-3333-333333333333', 'Fakta',        'fakta',        'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg'),
('44444444-4444-4444-4444-444444444444', 'Hate Speech',  'hate-speech',  'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg');

-- =============================================================================
-- 3. REPORTS
-- ticket_number SENGAJA TIDAK DI-SET — biar pakai auto-generate format KH-XXXXXXXX
-- proof_link kosong = NULL (bukan empty string)
-- =============================================================================
INSERT INTO hoax_reports (id, reporter_name, reporter_email, reporter_phone, content, proof_link, status, created_at) VALUES
-- Laporan 1: Selesai Diproses (PROCESSED) — punya proof_link
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
 'Ahmad Santoso', 'ahmad@gmail.com', '081234567890',
 'Tolong cek info ini, katanya minum air rebusan bawang putih mentah bisa sembuhkan virus mematikan dalam 5 menit. Grup WA keluarga heboh.',
 'https://facebook.com/hoaksbawang',
 'PROCESSED',
 NOW() - INTERVAL '3 days'),

-- Laporan 2: Selesai Diproses (PROCESSED) — punya proof_link
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
 'Siti Aminah', 'siti@yahoo.com', '081987654321',
 'Ada video gempa megathrust menghancurkan balai kota hari ini, padahal saya di lokasi aman-aman saja.',
 'https://tiktok.com/@penyebarhoaks',
 'PROCESSED',
 NOW() - INTERVAL '2 days'),

-- Laporan 3: Masih Menunggu (PENDING) — punya proof_link
('cccccccc-cccc-cccc-cccc-cccccccccccc',
 'Budi Hartono', 'budi.h@gmail.com', '085612345678',
 'Ada pesan SMS dan link telegram yang mengatasnamakan kementerian, katanya saya dapat bansos Rp 5 Juta.',
 'https://t.me/bansospalsu',
 'PENDING',
 NOW() - INTERVAL '5 hours'),

-- Laporan 4: Ditolak (REJECTED) — TANPA proof_link → pake NULL
('dddddddd-dddd-dddd-dddd-dddddddddddd',
 'Iseng Doang', 'iseng@gmail.com', '081111111111',
 'Tolong pak, tetangga saya sepertinya pakai pesugihan babi ngepet karena tiap malam keliling kampung bawa lilin.',
 NULL,
 'REJECTED',
 NOW() - INTERVAL '1 days');

-- =============================================================================
-- 4. NEWS (KLARIFIKASI)
-- Tersambung ke laporan 1 & 2 yang sudah PROCESSED
-- =============================================================================
INSERT INTO hoax_news (id, report_id, category_id, title, slug, description, reference_link, image_url, published_at) VALUES
-- Berita 1 (klarifikasi Laporan 1) → Kategori Berita Hoaks
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
 '11111111-1111-1111-1111-111111111111',
 '[HOAKS] Air Rebusan Bawang Putih Sembuhkan Virus dalam 5 Menit',
 'hoaks-air-rebusan-bawang-putih',
 'Beredar informasi di WhatsApp yang mengklaim bahwa air rebusan bawang putih dapat menyembuhkan berbagai virus berbahaya dalam hitungan menit. Faktanya, menurut para ahli medis dan WHO, tidak ada bukti ilmiah yang mendukung klaim tersebut. Bawang putih memang memiliki sifat antimikroba alami, namun bukan obat mutlak untuk infeksi virus.',
 'https://www.kominfo.go.id/content/detail/hoaks-bawang-putih',
 'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg',
 NOW() - INTERVAL '2 days'),

-- Berita 2 (klarifikasi Laporan 2) → Kategori Disinformasi
('ffffffff-ffff-ffff-ffff-ffffffffffff',
 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
 '22222222-2222-2222-2222-222222222222',
 '[DISINFORMASI] Video Gempa Megathrust Hancurkan Balai Kota Hari Ini',
 'disinformasi-video-gempa-megathrust-balai-kota',
 'Sebuah video di TikTok memperlihatkan guncangan dahsyat yang diklaim sebagai gempa megathrust terkini di Balai Kota. Setelah dilakukan penelusuran fakta, video tersebut adalah rekaman kejadian gempa di negara lain pada tahun 2018 yang diedit ulang dan disebarkan dengan narasi yang menyesatkan.',
 'https://www.bmkg.go.id/klarifikasi-gempa',
 'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg',
 NOW() - INTERVAL '1 days');