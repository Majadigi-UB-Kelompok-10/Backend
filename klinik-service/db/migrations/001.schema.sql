CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE report_status AS ENUM ('PENDING', 'PROCESSED', 'REJECTED');

CREATE TABLE hoax_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL,
    slug        VARCHAR(50) NOT NULL,
    icon_url    VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_hoax_category_slug UNIQUE (slug),
    CONSTRAINT uq_hoax_category_name UNIQUE (name)
);

CREATE TABLE hoax_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number   VARCHAR(20) NOT NULL,
    reporter_name   VARCHAR(150) NOT NULL,
    reporter_email  VARCHAR(150) NOT NULL CHECK (reporter_email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    reporter_phone  VARCHAR(20) NOT NULL,
    content         TEXT NOT NULL,
    proof_link      VARCHAR(255),
    proof_image_url VARCHAR(255),
    status          report_status NOT NULL DEFAULT 'PENDING',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_hoax_ticket UNIQUE (ticket_number)
);

-- 3. TABEL BERITA PUBLIKASI
CREATE TABLE hoax_news (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id       UUID REFERENCES hoax_reports(id) ON DELETE SET NULL,
    category_id     UUID NOT NULL REFERENCES hoax_categories(id) ON DELETE RESTRICT,
    title           VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description     TEXT NOT NULL,
    reference_link  VARCHAR(255),
    image_url       VARCHAR(255) NOT NULL,
    published_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    search_vector   tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description, '')), 'B')
    ) STORED
);

-- ==========================================
-- INDEXING STRATEGY (Sangat Krusial untuk Performa)
-- ==========================================

CREATE INDEX idx_hoax_reports_ticket ON hoax_reports(ticket_number);
CREATE INDEX idx_hoax_news_category ON hoax_news(category_id);
CREATE INDEX idx_hoax_news_published ON hoax_news(published_at DESC);

CREATE INDEX idx_hoax_news_title_trgm ON hoax_news USING GIN (title gin_trgm_ops);

CREATE INDEX idx_hoax_news_search ON hoax_news USING GIN (search_vector);

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_reports
BEFORE UPDATE ON hoax_reports
FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_news
BEFORE UPDATE ON hoax_news
FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();