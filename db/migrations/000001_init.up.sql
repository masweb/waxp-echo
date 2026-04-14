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

CREATE TABLE blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    block_group UUID NOT NULL,
    locale_id BIGINT NOT NULL REFERENCES site_locales(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(block_group, locale_id)
);

CREATE INDEX idx_pages_site ON pages(site_id);
CREATE INDEX idx_pages_blog ON pages(blog_id);
CREATE INDEX idx_pages_parent ON pages(parent_id);
CREATE INDEX idx_pages_type ON pages(type);
CREATE INDEX idx_page_slugs_slug ON page_slugs(slug);
CREATE INDEX idx_blocks_site ON blocks(site_id);
CREATE INDEX idx_blocks_group ON blocks(block_group);
CREATE INDEX idx_pages_layout ON pages USING gin(layout);
