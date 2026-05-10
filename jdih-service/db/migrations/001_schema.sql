CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 8 jenis dari Akses Cepat di UI
CREATE TYPE jenis_dokumen_type AS ENUM (
    'perda', 'pergub', 'peraturan', 'perdes',
    'sk_gub', 'instruksi', 'se', 'keputusan'
);

CREATE TYPE status_dokumen_type AS ENUM ('berlaku', 'diubah', 'dicabut');

-- =====================================================================
-- DOKUMEN HUKUM
-- =====================================================================
CREATE TABLE dokumen_hukum (
    id                  INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    jenis               jenis_dokumen_type NOT NULL,
    nomor               VARCHAR(100) NOT NULL,
    tahun               SMALLINT NOT NULL,
    judul               TEXT NOT NULL,
    ringkasan           TEXT,
    tanggal_penetapan   DATE NOT NULL,
    status              status_dokumen_type NOT NULL DEFAULT 'berlaku',
    pdf_url             TEXT NOT NULL,
    pdf_size_kb         INT,
    urusan_pemerintahan VARCHAR(200),
    jumlah_view         INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_dokumen_jenis_nomor_tahun UNIQUE (jenis, nomor, tahun),
    CONSTRAINT chk_dokumen_tahun    CHECK (tahun > 1900 AND tahun <= 2200),
    CONSTRAINT chk_pdf_size_kb      CHECK (pdf_size_kb IS NULL OR pdf_size_kb > 0),
    CONSTRAINT chk_jumlah_view      CHECK (jumlah_view >= 0)
);

-- =====================================================================
-- SUBJEK (tag metadata, tampil di Detail Dokumen)
-- =====================================================================
CREATE TABLE subjek (
    id   INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    CONSTRAINT uq_subjek_nama UNIQUE (nama)
);

CREATE TABLE dokumen_subjek (
    dokumen_id INT NOT NULL REFERENCES dokumen_hukum(id) ON DELETE CASCADE,
    subjek_id  INT NOT NULL REFERENCES subjek(id)        ON DELETE CASCADE,
    PRIMARY KEY (dokumen_id, subjek_id)
);

-- =====================================================================
-- PENGUMUMAN (carousel TERBARU di homepage)
-- =====================================================================
CREATE TABLE pengumuman (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    judul      VARCHAR(300) NOT NULL,
    isi        TEXT,
    aktif      BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dokumen_jenis         ON dokumen_hukum(jenis);
CREATE INDEX idx_dokumen_tahun         ON dokumen_hukum(tahun);
CREATE INDEX idx_dokumen_jenis_tahun   ON dokumen_hukum(jenis, tahun);
CREATE INDEX idx_dokumen_status        ON dokumen_hukum(status);
CREATE INDEX idx_dokumen_view          ON dokumen_hukum(jumlah_view DESC);
CREATE INDEX idx_dokumen_tanggal       ON dokumen_hukum(tanggal_penetapan DESC);
CREATE INDEX idx_dokumen_judul_trgm    ON dokumen_hukum USING GIN (judul gin_trgm_ops);
CREATE INDEX idx_dokumen_nomor_trgm    ON dokumen_hukum USING GIN (nomor gin_trgm_ops);
CREATE INDEX idx_dokumen_subjek_dok    ON dokumen_subjek(dokumen_id);

CREATE TRIGGER set_ts_dokumen    BEFORE UPDATE ON dokumen_hukum FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_pengumuman BEFORE UPDATE ON pengumuman    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
