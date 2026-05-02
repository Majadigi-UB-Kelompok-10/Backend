-- =============================================================================
-- SEED DATA: MASTER AREA
-- =============================================================================
INSERT INTO master_area (id, nama, slug) OVERRIDING SYSTEM VALUE VALUES 
    (1, 'Bangkalan', 'bangkalan'),
    (2, 'Banyuwangi', 'banyuwangi'),
    (3, 'Bojonegoro', 'bojonegoro'),
    (4, 'Bondowoso', 'bondowoso'),
    (5, 'Gresik', 'gresik'),
    (6, 'Jember', 'jember'),
    (7, 'Jombang', 'jombang'),
    (8, 'Lamongan', 'lamongan'),
    (9, 'Lumajang', 'lumajang'),
    (10, 'Magetan', 'magetan'),
    (11, 'Nganjuk', 'nganjuk'),
    (12, 'Ngawi', 'ngawi'),
    (13, 'Pacitan', 'pacitan'),
    (14, 'Pamekasan', 'pamekasan'),
    (15, 'Ponorogo', 'ponorogo'),
    (16, 'Sampang', 'sampang'),
    (17, 'Sidoarjo', 'sidoarjo'),
    (18, 'Situbondo', 'situbondo'),
    (19, 'Sumenep', 'sumenep'),
    (20, 'Surabaya', 'surabaya'),
    (21, 'Trenggalek', 'trenggalek'),
    (22, 'Tuban', 'tuban'),
    (23, 'Tulungagung', 'tulungagung'),
    (24, 'Batu', 'batu'),
    (25, 'Blitar', 'blitar'),
    (26, 'Kabupaten Blitar', 'kab-blitar'),
    (27, 'Kediri', 'kediri'),
    (28, 'Kabupaten Kediri', 'kab-kediri'),
    (29, 'Madiun', 'madiun'),
    (30, 'Kabupaten Madiun', 'kab-madiun'),
    (31, 'Malang', 'malang'),
    (32, 'Kabupaten Malang', 'kab-malang'),
    (33, 'Mojokerto', 'mojokerto'),
    (34, 'Kabupaten Mojokerto', 'kab-mojokerto'),
    (35, 'Pasuruan', 'pasuruan'),
    (36, 'Kabupaten Pasuruan', 'kab-pasuruan'),
    (37, 'Probolinggo', 'probolinggo'),
    (38, 'Kabupaten Probolinggo', 'kab-probolinggo')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- SEED DATA: BAHAN POKOK
-- =============================================================================
INSERT INTO bahan_pokok (id, nama, slug, satuan, gambar_url) OVERRIDING SYSTEM VALUE VALUES 
    (1, 'Bawang Merah / Kg', 'bawang-merah', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/bawang-merah.webp'),
    (2, 'Beras Medium / Kg', 'beras-medium', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/beras-premium.webp'),
    (3, 'Bawang Putih / Kg', 'bawang-putih', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/bawang-putih.webp'),
    (4, 'Cabai Rawit / Kg', 'cabai-rawit', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/cabai-rawit.webp'),
    (5, 'Cabai Merah / Kg', 'cabai-merah', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/cabe-merah.webp'),
    (6, 'Tepung Terigu / Kg', 'tepung-terigu', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/tepung-terigu.webp')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- SEED DATA: HARGA HARIAN
-- =============================================================================
INSERT INTO harga_harian (id, bahan_pokok_id, area_id, harga, tanggal) OVERRIDING SYSTEM VALUE VALUES 
    (1, 1, 20, 17000, '2026-03-27'),
    (2, 1, 29, 17000, '2026-03-27'),
    (3, 1, 7, 15000, '2026-03-27'),
    (4, 2, 7, 15000, '2026-03-27'),
    (6, 1, 7, 15000, '2026-03-26'),
    (7, 1, 20, 15000, '2026-03-26'),
    (8, 1, 20, 19000, '2026-03-29'),
    (9, 1, 20, 21000, '2026-03-28'),
    (11, 1, 20, 21000, '2026-03-25'),
    (12, 3, 20, 21000, '2026-03-25'),
    (13, 4, 20, 21000, '2026-03-25'),
    (15, 5, 20, 21000, '2026-03-25'),
    (16, 6, 20, 21000, '2026-03-25')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- RESET SEQUENCES
-- Sinkronisasi ID auto-increment agar tidak error saat admin insert data baru
-- =============================================================================
SELECT setval(pg_get_serial_sequence('master_area', 'id'), (SELECT COALESCE(MAX(id), 0) FROM master_area));
SELECT setval(pg_get_serial_sequence('bahan_pokok', 'id'), (SELECT COALESCE(MAX(id), 0) FROM bahan_pokok));
SELECT setval(pg_get_serial_sequence('harga_harian', 'id'), (SELECT COALESCE(MAX(id), 0) FROM harga_harian));