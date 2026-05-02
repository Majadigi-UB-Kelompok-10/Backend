CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE master_area (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uq_master_area_nama UNIQUE (nama),
    CONSTRAINT uq_master_area_slug UNIQUE (slug),
    CONSTRAINT chk_area_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);


CREATE TABLE bahan_pokok (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,
    satuan      VARCHAR(50) NOT NULL DEFAULT 'kg',
    gambar_url  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_bahan_pokok_nama UNIQUE (nama),
    CONSTRAINT uq_bahan_pokok_slug UNIQUE (slug),
    CONSTRAINT chk_bahan_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT chk_bahan_gambar CHECK (gambar_url IS NULL OR gambar_url ~* '^https?://[^\s]+$')
);

CREATE TABLE harga_harian (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bahan_pokok_id  INT NOT NULL REFERENCES bahan_pokok(id) ON DELETE CASCADE,
    area_id         INT NOT NULL REFERENCES master_area(id) ON DELETE CASCADE,
    harga           INTEGER NOT NULL,
    tanggal         DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_harga_per_area_tanggal UNIQUE (bahan_pokok_id, area_id, tanggal),
    CONSTRAINT chk_harga_harian_positif CHECK (harga >= 0)
);



CREATE INDEX idx_bahan_pokok_nama_trgm ON bahan_pokok USING GIN (nama gin_trgm_ops);

CREATE INDEX idx_hh_bahan_tanggal ON harga_harian(bahan_pokok_id, tanggal DESC);

CREATE INDEX idx_hh_bahan_area_tanggal ON harga_harian(bahan_pokok_id, area_id, tanggal DESC);

CREATE INDEX idx_hh_tanggal_desc ON harga_harian(tanggal DESC);



CREATE TRIGGER set_timestamp_master_area
    BEFORE UPDATE ON master_area
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_bahan_pokok
    BEFORE UPDATE ON bahan_pokok
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_harga_harian
    BEFORE UPDATE ON harga_harian
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();