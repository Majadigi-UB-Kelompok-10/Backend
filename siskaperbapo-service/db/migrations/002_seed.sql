-- =============================================================================
-- SEED DATA: MASTER AREA
-- =============================================================================
INSERT INTO master_area (nama, slug) VALUES 
    ('Bangkalan', 'bangkalan'),
    ('Banyuwangi', 'banyuwangi'),
    ('Bojonegoro', 'bojonegoro'),
    ('Bondowoso', 'bondowoso'),
    ('Gresik', 'gresik'),
    ('Jember', 'jember'),
    ('Jombang', 'jombang'),
    ('Lamongan', 'lamongan'),
    ('Lumajang', 'lumajang'),
    ('Magetan', 'magetan'),
    ('Nganjuk', 'nganjuk'),
    ('Ngawi', 'ngawi'),
    ('Pacitan', 'pacitan'),
    ('Pamekasan', 'pamekasan'),
    ('Ponorogo', 'ponorogo'),
    ('Sampang', 'sampang'),
    ('Sidoarjo', 'sidoarjo'),
    ('Situbondo', 'situbondo'),
    ('Sumenep', 'sumenep'),
    ('Surabaya', 'surabaya'),
    ('Trenggalek', 'trenggalek'),
    ('Tuban', 'tuban'),
    ('Tulungagung', 'tulungagung'),
    ('Batu', 'batu'),
    ('Blitar', 'blitar'),
    ('Kabupaten Blitar', 'kab-blitar'),
    ('Kediri', 'kediri'),
    ('Kabupaten Kediri', 'kab-kediri'),
    ('Madiun', 'madiun'),
    ('Kabupaten Madiun', 'kab-madiun'),
    ('Malang', 'malang'),
    ('Kabupaten Malang', 'kab-malang'),
    ('Mojokerto', 'mojokerto'),
    ('Kabupaten Mojokerto', 'kab-mojokerto'),
    ('Pasuruan', 'pasuruan'),
    ('Kabupaten Pasuruan', 'kab-pasuruan'),
    ('Probolinggo', 'probolinggo'),
    ('Kabupaten Probolinggo', 'kab-probolinggo')
ON CONFLICT (slug) DO NOTHING;

-- =============================================================================
-- SEED DATA: BAHAN POKOK
-- =============================================================================
INSERT INTO bahan_pokok (nama, slug, satuan, gambar_url) VALUES 
    ('Bawang Merah / Kg', 'bawang-merah', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/bawang-merah.webp'),
    ('Beras Medium / Kg', 'beras-medium', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/beras-premium.webp'),
    ('Bawang Putih / Kg', 'bawang-putih', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/bawang-putih.webp'),
    ('Cabai Rawit / Kg', 'cabai-rawit', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/cabai-rawit.webp'),
    ('Cabai Merah / Kg', 'cabai-merah', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/cabe-merah.webp'),
    ('Tepung Terigu / Kg', 'tepung-terigu', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/tepung-terigu.webp')
ON CONFLICT (slug) DO NOTHING;

-- =============================================================================
-- SEED DATA: HARGA HARIAN
-- =============================================================================
INSERT INTO harga_harian (bahan_pokok_id, area_id, harga, tanggal) VALUES 
    -- Bawang Merah
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 17000, '2026-03-27'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'madiun'), 17000, '2026-03-27'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'jombang'), 15000, '2026-03-27'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'jombang'), 15000, '2026-03-26'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 15000, '2026-03-26'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 19000, '2026-03-29'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 21000, '2026-03-28'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-merah'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 21000, '2026-03-25'),
    
    -- Beras Medium
    ((SELECT id FROM bahan_pokok WHERE slug = 'beras-medium'), (SELECT id FROM master_area WHERE slug = 'jombang'), 15000, '2026-03-27'),
    
    -- Lain-lain (Surabaya - 25 Maret)
    ((SELECT id FROM bahan_pokok WHERE slug = 'bawang-putih'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 21000, '2026-03-25'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'cabai-rawit'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 21000, '2026-03-25'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'cabai-merah'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 21000, '2026-03-25'),
    ((SELECT id FROM bahan_pokok WHERE slug = 'tepung-terigu'), (SELECT id FROM master_area WHERE slug = 'surabaya'), 21000, '2026-03-25')
ON CONFLICT (bahan_pokok_id, area_id, tanggal) DO NOTHING;