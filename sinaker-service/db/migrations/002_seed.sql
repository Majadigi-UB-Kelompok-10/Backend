-- ==========================================
-- SEED DATA BLK (Balai Latihan Kerja)
-- ==========================================
-- Kita biarkan PostgreSQL meng-generate ID secara otomatis
INSERT INTO blk (nama, alamat, kab_kota, kecamatan, slug, lat, lng) VALUES
('UPT BLK Surabaya', 'Jl. Dukuh Menanggal III/29', 'Surabaya', 'Gayungan', 'upt-blk-surabaya', -7.339682, 112.726588),
('UPT BLK Singosari', 'Jl. Raya Singosari No.1', 'Kabupaten Malang', 'Singosari', 'upt-blk-singosari', -7.881729, 112.663110),
('UPT BLK Kediri', 'Jl. Kapt. Tendean No.2', 'Kota Kediri', 'Pesantren', 'upt-blk-kediri', -7.838541, 112.030632),
('UPT BLK Jember', 'Jl. Basuki Rahmat No.202', 'Kabupaten Jember', 'Kaliwates', 'upt-blk-jember', -8.182412, 113.678121);

-- ==========================================
-- SEED DATA KEJURUAN
-- ==========================================
-- Kita cari blk_id secara dinamis berdasarkan slug-nya. Jauh lebih aman!

-- Kejuruan BLK Surabaya
INSERT INTO kejuruan (blk_id, nama) VALUES
((SELECT id FROM blk WHERE slug = 'upt-blk-surabaya'), 'Teknologi Informasi dan Komunikasi (TIK)'),
((SELECT id FROM blk WHERE slug = 'upt-blk-surabaya'), 'Teknik Otomotif'),
((SELECT id FROM blk WHERE slug = 'upt-blk-surabaya'), 'Desain Grafis');

-- Kejuruan BLK Singosari
INSERT INTO kejuruan (blk_id, nama) VALUES
((SELECT id FROM blk WHERE slug = 'upt-blk-singosari'), 'Teknik Manufaktur'),
((SELECT id FROM blk WHERE slug = 'upt-blk-singosari'), 'Teknik Las (Welding)'),
((SELECT id FROM blk WHERE slug = 'upt-blk-singosari'), 'Otomasi Industri');

-- Kejuruan BLK Kediri
INSERT INTO kejuruan (blk_id, nama) VALUES
((SELECT id FROM blk WHERE slug = 'upt-blk-kediri'), 'Tata Boga'),
((SELECT id FROM blk WHERE slug = 'upt-blk-kediri'), 'Tata Busana / Menjahit'),
((SELECT id FROM blk WHERE slug = 'upt-blk-kediri'), 'Processing Hasil Pertanian');

-- Kejuruan BLK Jember
INSERT INTO kejuruan (blk_id, nama) VALUES
((SELECT id FROM blk WHERE slug = 'upt-blk-jember'), 'Teknik Sepeda Motor'),
((SELECT id FROM blk WHERE slug = 'upt-blk-jember'), 'Teknologi Informasi (Web Dev)'),
((SELECT id FROM blk WHERE slug = 'upt-blk-jember'), 'Bahasa Asing (Jepang)');