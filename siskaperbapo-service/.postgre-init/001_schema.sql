CREATE TABLE master_area (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(100) UNIQUE NOT NULL, 
    slug VARCHAR(100) UNIQUE NOT NULL  
);

CREATE TABLE bahan_pokok (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    satuan VARCHAR(50) DEFAULT 'kg' NOT NULL, 
    gambar_url TEXT
);

CREATE TABLE harga_harian (
    id SERIAL PRIMARY KEY,
    bahan_pokok_id INT NOT NULL REFERENCES bahan_pokok(id),
    area_id INT NOT NULL REFERENCES master_area(id),
    harga INTEGER NOT NULL,
    tanggal DATE NOT NULL,
    CONSTRAINT uq_harga_per_area_tanggal 
        UNIQUE (bahan_pokok_id, area_id, tanggal)     
);


CREATE INDEX idx_hh_bahan_tanggal ON harga_harian(bahan_pokok_id, tanggal DESC);
CREATE INDEX idx_hh_tanggal_only ON harga_harian(tanggal DESC);
CREATE INDEX idx_bahan_pokok_nama_lower ON bahan_pokok(LOWER(nama));

CREATE INDEX idx_bahan_pokok_slug ON bahan_pokok(slug);
CREATE INDEX idx_master_area_slug ON master_area(slug);