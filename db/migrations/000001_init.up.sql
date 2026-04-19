CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE sites (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL UNIQUE,
    options JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE site_locales (
    id BIGSERIAL PRIMARY KEY,
    site_id BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(site_id, code)
);

CREATE TABLE blogs (
    id BIGSERIAL PRIMARY KEY,
    site_id BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blog_slugs (
    id BIGSERIAL PRIMARY KEY,
    blog_id BIGINT NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    locale_id BIGINT NOT NULL REFERENCES site_locales(id) ON DELETE CASCADE,
    slug VARCHAR(500) NOT NULL,
    UNIQUE(blog_id, locale_id)
);

CREATE TABLE pages (
    id BIGSERIAL PRIMARY KEY,
    site_id BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    blog_id BIGINT REFERENCES blogs(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES pages(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL DEFAULT 'page',
    layout JSONB NOT NULL DEFAULT '{}',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pages_type_check CHECK (type IN ('page', 'post')),
    CONSTRAINT pages_blog_only_posts CHECK (
        (type = 'page' AND blog_id IS NULL) OR
        (type = 'post' AND blog_id IS NOT NULL)
    )
);

CREATE TABLE page_slugs (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    locale_id BIGINT NOT NULL REFERENCES site_locales(id) ON DELETE CASCADE,
    slug VARCHAR(500) NOT NULL,
    UNIQUE(page_id, locale_id)
);

CREATE TABLE page_seo (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    locale_id BIGINT NOT NULL REFERENCES site_locales(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(page_id, locale_id)
);

CREATE TABLE section_counters (
    site_id BIGINT NOT NULL PRIMARY KEY REFERENCES sites(id) ON DELETE CASCADE,
    current_value BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE block_counters (
    site_id BIGINT NOT NULL PRIMARY KEY REFERENCES sites(id) ON DELETE CASCADE,
    current_value BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE media (
    id BIGSERIAL PRIMARY KEY,
    filename VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    url VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pages_site ON pages(site_id);
CREATE INDEX idx_pages_site_id_id ON pages(site_id, id);
CREATE INDEX idx_pages_blog ON pages(blog_id);
CREATE INDEX idx_pages_parent ON pages(parent_id);
CREATE INDEX idx_pages_type ON pages(type);
CREATE INDEX idx_page_slugs_slug ON page_slugs(slug);
CREATE INDEX idx_page_seo_page ON page_seo(page_id);
CREATE INDEX idx_page_seo_locale ON page_seo(locale_id);
CREATE INDEX idx_pages_layout ON pages USING gin(layout);
CREATE INDEX idx_section_counters_site ON section_counters(site_id);
CREATE INDEX idx_media_created_at ON media(created_at DESC);

INSERT INTO users (email, password_hash)
VALUES ('admin@waxp.com', '$2a$10$ysLGEY/eUOniH2eVzRGpQ.SmVS7PfQZOLaQ4QGgfcpF0E8uO98Tz6')
ON CONFLICT (email) DO NOTHING;
