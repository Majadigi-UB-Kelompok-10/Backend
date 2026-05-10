-- Seed: 10 kategori layanan (fixed UUIDs, dipakai oleh AutoRegisterFull tiap service)
ALTER TABLE category_list ADD COLUMN IF NOT EXISTS is_popular BOOLEAN NOT NULL DEFAULT false;

INSERT INTO category_list (category_list_id, name, description, is_popular) VALUES
    ('b1000001-0000-4000-8000-000000000001', 'Layanan Darurat',          'Layanan terkait kedaruratan dan respons cepat seperti IGD rumah sakit.',    true),
    ('b1000001-0000-4000-8000-000000000002', 'Transportasi Publik',       'Layanan informasi transportasi umum di Jawa Timur.',                         true),
    ('b1000001-0000-4000-8000-000000000003', 'Layanan Pajak',             'Layanan administrasi dan informasi pajak kendaraan bermotor.',                true),
    ('b1000001-0000-4000-8000-000000000004', 'Informasi Ketenagakerjaan', 'Layanan pelatihan, bursa kerja, dan informasi ketenagakerjaan.',              true),
    ('b1000001-0000-4000-8000-000000000005', 'Produk Hukum',              'Peraturan daerah, regulasi, dan produk hukum Jawa Timur.',                   false),
    ('b1000001-0000-4000-8000-000000000006', 'Cek Hoax',                  'Layanan verifikasi informasi dan pelaporan hoaks.',                           false),
    ('b1000001-0000-4000-8000-000000000007', 'Bantuan Sosial',            'Informasi dan akses program bantuan sosial pemerintah.',                     false),
    ('b1000001-0000-4000-8000-000000000008', 'Harga Kebutuhan Pokok',     'Pemantauan harga dan ketersediaan bahan pokok harian.',                      false),
    ('b1000001-0000-4000-8000-000000000009', 'Eksplorasi Pariwisata',     'Informasi destinasi wisata dan atraksi di Jawa Timur.',                      false),
    ('b1000001-0000-4000-8000-000000000010', 'Layanan Kesehatan',         'Informasi rumah sakit, klinik, dan layanan kesehatan.',                      false)
ON CONFLICT (category_list_id) DO UPDATE
    SET is_popular  = EXCLUDED.is_popular,
        name        = EXCLUDED.name,
        description = EXCLUDED.description;
