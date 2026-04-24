ALTER TABLE sites ADD COLUMN is_live BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE page_renders (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    locale_id BIGINT NOT NULL REFERENCES site_locales(id) ON DELETE CASCADE,
    html TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(page_id, locale_id)
);

CREATE INDEX idx_page_renders_page ON page_renders(page_id);
