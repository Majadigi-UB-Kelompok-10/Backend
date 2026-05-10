CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE jenis_kelamin_type   AS ENUM ('laki_laki', 'perempuan');
CREATE TYPE pendidikan_type       AS ENUM ('tidak_sekolah', 'sd', 'smp', 'sma_smk', 'd3', 's1', 's2', 's3');
CREATE TYPE status_daftar_type    AS ENUM ('pending', 'diterima', 'ditolak');

-- =====================================================================
-- BLK (Balai Latihan Kerja)
-- =====================================================================
CREATE TABLE blk (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama       VARCHAR(200) NOT NULL,
    alamat     TEXT NOT NULL,
    kab_kota   VARCHAR(100) NOT NULL,
    kecamatan  VARCHAR(100),
    slug       VARCHAR(250) NOT NULL,
    lat        DOUBLE PRECISION,
    lng        DOUBLE PRECISION,
    aktif      BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_blk_slug UNIQUE (slug),
    CONSTRAINT uq_blk_nama UNIQUE (nama)
);

-- =====================================================================
-- KEJURUAN (per BLK, untuk dropdown form pendaftaran)
-- =====================================================================
CREATE TABLE kejuruan (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    blk_id     INT NOT NULL REFERENCES blk(id) ON DELETE CASCADE,
    nama       VARCHAR(150) NOT NULL,
    aktif      BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_kejuruan_blk_nama UNIQUE (blk_id, nama)
);

-- =====================================================================
-- PENDAFTARAN
-- =====================================================================
CREATE TABLE pendaftaran (
    id                     INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    blk_id                 INT NOT NULL REFERENCES blk(id),
    kejuruan_id            INT NOT NULL REFERENCES kejuruan(id),
    nik                    VARCHAR(16)  NOT NULL,
    nama_lengkap           VARCHAR(200) NOT NULL,
    tempat_lahir           VARCHAR(100) NOT NULL,
    tanggal_lahir          DATE NOT NULL,
    email                  VARCHAR(150) NOT NULL,
    jenis_kelamin          jenis_kelamin_type NOT NULL,
    provinsi               VARCHAR(100) NOT NULL,
    kab_kota               VARCHAR(100) NOT NULL,
    kecamatan              VARCHAR(100) NOT NULL,
    kelurahan              VARCHAR(100) NOT NULL,
    rt                     VARCHAR(5)  NOT NULL,
    rw                     VARCHAR(5)  NOT NULL,
    alamat_lengkap         TEXT NOT NULL,
    penyandang_disabilitas BOOLEAN NOT NULL DEFAULT false,
    no_wa                  VARCHAR(20) NOT NULL,
    no_wa_darurat          VARCHAR(20) NOT NULL,
    pendidikan_terakhir    pendidikan_type NOT NULL,
    pendidikan_sekarang    pendidikan_type NOT NULL,
    asal_sekolah           VARCHAR(200),
    jurusan                VARCHAR(200),
    foto_url               TEXT NOT NULL,
    status                 status_daftar_type NOT NULL DEFAULT 'pending',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_pendaftaran_blk_nik UNIQUE (blk_id, nik),
    CONSTRAINT chk_pendaftaran_nik    CHECK (nik ~ '^[0-9]{16}$')
);

CREATE INDEX idx_blk_kab_kota      ON blk(kab_kota);
CREATE INDEX idx_blk_nama_trgm     ON blk USING GIN (nama gin_trgm_ops);
CREATE INDEX idx_kejuruan_blk      ON kejuruan(blk_id) WHERE aktif = true;
CREATE INDEX idx_pendaftaran_nik   ON pendaftaran(nik);
CREATE INDEX idx_pendaftaran_blk   ON pendaftaran(blk_id);
CREATE INDEX idx_pendaftaran_status ON pendaftaran(status);

CREATE TRIGGER set_ts_blk        BEFORE UPDATE ON blk        FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_kejuruan   BEFORE UPDATE ON kejuruan   FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_pendaftaran BEFORE UPDATE ON pendaftaran FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
