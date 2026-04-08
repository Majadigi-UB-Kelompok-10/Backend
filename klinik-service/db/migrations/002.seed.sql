-- 1. BERSIHKAN DATABASE (Mulai dari Kertas Kosong)
TRUNCATE TABLE hoax_news, hoax_reports, hoax_categories CASCADE;

-- 2. SEED KATEGORI (UUID 1-4)
INSERT INTO hoax_categories (id, name, slug) VALUES
('11111111-1111-1111-1111-111111111111', 'Berita Hoaks', 'berita-hoaks'),
('22222222-2222-2222-2222-222222222222', 'Disinformasi', 'disinformasi'),
('33333333-3333-3333-3333-333333333333', 'Fakta', 'fakta'),
('44444444-4444-4444-4444-444444444444', 'Hate Speech', 'hate-speech');

-- 3. SEED LAPORAN WARGA (Mix Status: PENDING, PROCESSED, REJECTED)
INSERT INTO hoax_reports (id, ticket_number, reporter_name, reporter_email, reporter_phone, content, proof_link, status, created_at) VALUES
-- Laporan 1: Selesai Diproses (PROCESSED)
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'KLX-ABC123', 'Ahmad Santoso', 'ahmad@gmail.com', '081234567890', 'Tolong cek info ini, katanya minum air rebusan bawang putih mentah bisa sembuhkan virus mematikan dalam 5 menit. Grup WA keluarga heboh.', 'https://facebook.com/hoaksbawang', 'PROCESSED', NOW() - INTERVAL '3 days'),

-- Laporan 2: Selesai Diproses (PROCESSED)
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'KLX-DEF456', 'Siti Aminah', 'siti@yahoo.com', '081987654321', 'Ada video gempa megathrust menghancurkan balai kota hari ini, padahal saya di lokasi aman-aman saja.', 'https://tiktok.com/@penyebarhoaks', 'PROCESSED', NOW() - INTERVAL '2 days'),

-- Laporan 3: Masih Menunggu (PENDING)
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'KLX-GHI789', 'Budi Hartono', 'budi.h@gmail.com', '085612345678', 'Ada pesan SMS dan link telegram yang mengatasnamakan kementerian, katanya saya dapat bansos Rp 5 Juta.', 'https://t.me/bansospalsu', 'PENDING', NOW() - INTERVAL '5 hours'),

-- Laporan 4: Ditolak (REJECTED)
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'KLX-JKL012', 'Iseng Doang', 'iseng@gmail.com', '081111111111', 'Tolong pak, tetangga saya sepertinya pakai pesugihan babi ngepet karena tiap malam keliling kampung bawa lilin.', '', 'REJECTED', NOW() - INTERVAL '1 days');

-- 4. SEED BERITA KLARIFIKASI (Tersambung ke Laporan 1 & 2)
INSERT INTO hoax_news (id, report_id, category_id, title, slug, description, reference_link, image_url, published_at) VALUES
-- Berita 1 (Dari Laporan 1) -> Masuk Kategori Berita Hoaks
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 
'[HOAKS] Air Rebusan Bawang Putih Sembuhkan Virus dalam 5 Menit', 
'hoaks-air-rebusan-bawang-putih', 
'Beredar informasi di WhatsApp yang mengklaim bahwa air rebusan bawang putih dapat menyembuhkan berbagai virus berbahaya dalam hitungan menit. Faktanya, menurut para ahli medis dan WHO, tidak ada bukti ilmiah yang mendukung klaim tersebut. Bawang putih memang memiliki sifat antimikroba alami, namun bukan obat mutlak untuk infeksi virus.', 
'https://www.kominfo.go.id/content/detail/hoaks-bawang-putih', 
'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg', -- Placeholder Image
NOW() - INTERVAL '2 days'),

-- Berita 2 (Dari Laporan 2) -> Masuk Kategori Disinformasi
('ffffffff-ffff-ffff-ffff-ffffffffffff', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222', 
'[DISINFORMASI] Video Gempa Megathrust Hancurkan Balai Kota Hari Ini', 
'disinformasi-video-gempa-megathrust-balai-kota', 
'Sebuah video di TikTok memperlihatkan guncangan dahsyat yang diklaim sebagai gempa megathrust terkini di Balai Kota. Setelah dilakukan penelusuran fakta, video tersebut adalah rekaman kejadian gempa di negara lain pada tahun 2018 yang diedit ulang dan disebarkan dengan narasi yang menyesatkan.', 
'https://www.bmkg.go.id/klarifikasi-gempa', 
'https://res.cloudinary.com/demo/image/upload/v1618471325/sample.jpg', -- Placeholder Image
NOW() - INTERVAL '1 days');