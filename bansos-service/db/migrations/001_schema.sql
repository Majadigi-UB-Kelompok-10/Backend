CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE status_penyaluran_type AS ENUM ('proses', 'diterima', 'ditolak');

-- transfer_pos = Via Kantor Pos, sesuai skema penyaluran bansos nasional
CREATE TYPE metode_penyaluran_type AS ENUM (
    'transfer_bri', 'transfer_bni', 'transfer_mandiri',
    'transfer_bca', 'transfer_pos', 'tunai'
);

-- =====================================================================
-- PROGRAM BANSOS (master data: PKH, BPNT, dst)
-- =====================================================================
CREATE TABLE program_bansos (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama       VARCHAR(100) NOT NULL,
    kode       VARCHAR(30)  NOT NULL,
    deskripsi  TEXT,
    aktif      BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_program_kode UNIQUE (kode),
    CONSTRAINT uq_program_nama UNIQUE (nama)
);

-- =====================================================================
-- PENERIMA (data penerima bansos, lookup by NIK)
-- =====================================================================
CREATE TABLE penerima (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nik        VARCHAR(16)  NOT NULL,
    nama       VARCHAR(200) NOT NULL,
    alamat     VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_penerima_nik  UNIQUE (nik),
    CONSTRAINT chk_penerima_nik CHECK (nik ~ '^[0-9]{16}$')
);

-- =====================================================================
-- PENYALURAN (record distribusi bansos per penerima per program)
-- =====================================================================
CREATE TABLE penyaluran (
    id              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    penerima_id     INT NOT NULL REFERENCES penerima(id) ON DELETE CASCADE,
    program_id      INT NOT NULL REFERENCES program_bansos(id),
    nominal         INT NOT NULL,
    periode_mulai   DATE NOT NULL,
    periode_selesai DATE NOT NULL,
    metode          metode_penyaluran_type NOT NULL DEFAULT 'transfer_bri',
    status          status_penyaluran_type NOT NULL DEFAULT 'proses',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_penyaluran_penerima_program_periode UNIQUE (penerima_id, program_id, periode_mulai),
    CONSTRAINT chk_nominal_positif                    CHECK (nominal > 0),
    CONSTRAINT chk_periode_valid                      CHECK (periode_selesai >= periode_mulai)
);

CREATE INDEX idx_penerima_nik        ON penerima(nik);
CREATE INDEX idx_penyaluran_penerima ON penyaluran(penerima_id);
CREATE INDEX idx_penyaluran_program  ON penyaluran(program_id);
CREATE INDEX idx_penyaluran_status   ON penyaluran(status);

CREATE TRIGGER set_ts_program    BEFORE UPDATE ON program_bansos FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_penerima   BEFORE UPDATE ON penerima       FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_penyaluran BEFORE UPDATE ON penyaluran     FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
