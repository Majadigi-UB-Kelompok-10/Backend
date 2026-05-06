-- =============================================================================
-- SEED DATA: MASTER KELAS
-- =============================================================================
-- Kita biarkan Postgres membuatkan ID-nya secara otomatis

INSERT INTO master_kelas (nama, slug) VALUES
('VVIP', 'vvip'),
('VIP', 'vip'),
('Kelas I', 'kelas-1'),
('Kelas II', 'kelas-2'),
('Kelas III', 'kelas-3'),
('ICU / NICU / PICU', 'icu-nicu-picu')
ON CONFLICT (slug) DO NOTHING;

-- =============================================================================
-- SEED DATA: RUANGAN
-- =============================================================================
-- Kita gunakan SELECT untuk mencari ID kelas berdasarkan slug-nya, 
-- sehingga relasinya dijamin 100% akurat tanpa harus menebak angka ID.

INSERT INTO ruangan (kelas_id, nama, slug, kapasitas, terisi) VALUES
-- Data Ruangan VVIP
((SELECT id FROM master_kelas WHERE slug = 'vvip'), 'Paviliun VVIP Anggrek', 'paviliun-vvip-anggrek', 1, 0),
((SELECT id FROM master_kelas WHERE slug = 'vvip'), 'Paviliun VVIP Mawar', 'paviliun-vvip-mawar', 1, 1),

-- Data Ruangan VIP
((SELECT id FROM master_kelas WHERE slug = 'vip'), 'Paviliun VIP Melati 1', 'paviliun-vip-melati-1', 2, 1),
((SELECT id FROM master_kelas WHERE slug = 'vip'), 'Paviliun VIP Melati 2', 'paviliun-vip-melati-2', 2, 0),

-- Data Ruangan Kelas I
((SELECT id FROM master_kelas WHERE slug = 'kelas-1'), 'Ruang Rawat Inap Teratai (Kelas I)', 'ruang-rawat-inap-teratai-kelas-1', 4, 3),
((SELECT id FROM master_kelas WHERE slug = 'kelas-1'), 'Ruang Rawat Inap Flamboyan (Kelas I)', 'ruang-rawat-inap-flamboyan-kelas-1', 4, 4),

-- Data Ruangan Kelas II
((SELECT id FROM master_kelas WHERE slug = 'kelas-2'), 'Ruang Rawat Inap Kenanga (Kelas II)', 'ruang-rawat-inap-kenanga-kelas-2', 6, 4),
((SELECT id FROM master_kelas WHERE slug = 'kelas-2'), 'Ruang Rawat Inap Bougenville (Kelas II)', 'ruang-rawat-inap-bougenville-kelas-2', 6, 2),

-- Data Ruangan Kelas III
((SELECT id FROM master_kelas WHERE slug = 'kelas-3'), 'Bangsal Dahlia (Kelas III)', 'bangsal-dahlia-kelas-3', 12, 8),
((SELECT id FROM master_kelas WHERE slug = 'kelas-3'), 'Bangsal Tulip (Kelas III)', 'bangsal-tulip-kelas-3', 10, 10),
((SELECT id FROM master_kelas WHERE slug = 'kelas-3'), 'Bangsal Asoka (Kelas III)', 'bangsal-asoka-kelas-3', 15, 5),

-- Data Ruangan ICU/Kritis
((SELECT id FROM master_kelas WHERE slug = 'icu-nicu-picu'), 'ICU Sentral', 'icu-sentral', 8, 7),
((SELECT id FROM master_kelas WHERE slug = 'icu-nicu-picu'), 'NICU (Neonatal ICU)', 'nicu-neonatal-icu', 5, 5),
((SELECT id FROM master_kelas WHERE slug = 'icu-nicu-picu'), 'HCU Jantung', 'hcu-jantung', 4, 1)
ON CONFLICT (slug) DO NOTHING;