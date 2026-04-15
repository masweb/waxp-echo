CREATE TABLE page_seo (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    locale_id BIGINT NOT NULL REFERENCES site_locales(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(page_id, locale_id)
);

CREATE INDEX idx_page_seo_page ON page_seo(page_id);
CREATE INDEX idx_page_seo_locale ON page_seo(locale_id);