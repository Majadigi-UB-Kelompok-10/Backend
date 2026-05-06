CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- MASTER KELAS
-- ==========================================
CREATE TABLE master_kelas (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama        VARCHAR(50) NOT NULL,
    slug        VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_master_kelas_nama UNIQUE (nama),
    CONSTRAINT uq_master_kelas_slug UNIQUE (slug),
    CONSTRAINT chk_kelas_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

-- ==========================================
-- TABEL RUANGAN
-- ==========================================
CREATE TABLE ruangan (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kelas_id    INT NOT NULL REFERENCES master_kelas(id) ON DELETE RESTRICT,
    nama        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,
    kapasitas   INT NOT NULL DEFAULT 0,
    terisi      INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Kalkulasi otomatis di level Database (Tidak perlu hitung di Golang lagi!)
    tersedia    INT GENERATED ALWAYS AS (kapasitas - terisi) STORED,

    CONSTRAINT uq_ruangan_slug UNIQUE (slug),
    CONSTRAINT chk_ruangan_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT chk_ruangan_kapasitas CHECK (kapasitas >= 0),
    CONSTRAINT chk_ruangan_terisi CHECK (terisi >= 0 AND terisi <= kapasitas)
);

-- ==========================================
-- INDEXING SAKTI
-- ==========================================
CREATE INDEX idx_ruangan_kelas_id ON ruangan(kelas_id);
CREATE INDEX idx_ruangan_nama_trgm ON ruangan USING GIN (nama gin_trgm_ops);
CREATE INDEX idx_ruangan_created_at ON ruangan(created_at DESC);

-- ==========================================
-- TRIGGERS
-- ==========================================
CREATE TRIGGER set_timestamp_master_kelas
    BEFORE UPDATE ON master_kelas
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_ruangan
    BEFORE UPDATE ON ruangan
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();