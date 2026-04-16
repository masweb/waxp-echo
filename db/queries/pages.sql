-- name: CreatePage :one
INSERT INTO pages (site_id, blog_id, parent_id, type, layout, published_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, site_id, blog_id, parent_id, type, layout, published_at, created_at, updated_at;

-- name: GetPageByID :one
SELECT id, site_id, blog_id, parent_id, type, layout, published_at, created_at, updated_at
FROM pages
WHERE id = $1 AND site_id = $2;

-- name: UpdatePage :one
UPDATE pages
SET parent_id = $1, layout = $2, published_at = $3, updated_at = NOW()
WHERE id = $4 AND site_id = $5
RETURNING id, site_id, blog_id, parent_id, type, layout, published_at, created_at, updated_at;

-- name: DeletePage :exec
DELETE FROM pages WHERE id = $1 AND site_id = $2;

-- name: CreatePageSlug :one
INSERT INTO page_slugs (page_id, locale_id, slug)
VALUES ($1, $2, $3)
RETURNING id, page_id, locale_id, slug;

-- name: GetPageSlugsByPageID :many
SELECT id, page_id, locale_id, slug
FROM page_slugs
WHERE page_id = $1;

-- name: DeletePageSlugsByPageID :exec
DELETE FROM page_slugs WHERE page_id = $1;

-- name: CreatePageSeo :one
INSERT INTO page_seo (page_id, locale_id, title, description)
VALUES ($1, $2, $3, $4)
RETURNING id, page_id, locale_id, title, description;

-- name: GetPageSeoByPageID :many
SELECT id, page_id, locale_id, title, description
FROM page_seo
WHERE page_id = $1;

-- name: DeletePageSeoByPageID :exec
DELETE FROM page_seo WHERE page_id = $1;

-- name: GetPageSlugsByPageIDs :many
SELECT id, page_id, locale_id, slug
FROM page_slugs
WHERE page_id = ANY($1::BIGINT[]);

-- name: GetPageSeoByPageIDs :many
SELECT id, page_id, locale_id, title, description
FROM page_seo
WHERE page_id = ANY($1::BIGINT[]);

-- name: GetBlogByID :one
SELECT id, site_id, created_at, updated_at
FROM blogs
WHERE id = $1 AND site_id = $2;

-- name: GetRootPageBySite :one
SELECT p.id, p.site_id, p.blog_id, p.parent_id, p.type, p.layout, p.published_at, p.created_at, p.updated_at
FROM pages p
JOIN page_slugs ps ON ps.page_id = p.id
JOIN site_locales sl ON sl.id = ps.locale_id
WHERE p.site_id = $1 AND p.type = 'page' AND p.parent_id IS NULL
  AND sl.is_default = true AND ps.slug = ''
LIMIT 1;

-- name: GetPageRoutes :many
WITH RECURSIVE page_tree AS (
    SELECT p.id, ps.locale_id, sl.code, sl.is_default, ps.slug::TEXT AS path
    FROM pages p
    JOIN page_slugs ps ON ps.page_id = p.id
    JOIN site_locales sl ON sl.id = ps.locale_id
    WHERE p.site_id = $1 AND p.type = 'page' AND p.parent_id IS NULL
      AND p.published_at IS NOT NULL
  UNION ALL
    SELECT c.id, cs.locale_id, sl.code, sl.is_default,
           CASE WHEN pt.path = '' THEN cs.slug ELSE pt.path || '/' || cs.slug END
    FROM pages c
    JOIN page_slugs cs ON cs.page_id = c.id
    JOIN site_locales sl ON sl.id = cs.locale_id
    JOIN page_tree pt ON pt.id = c.parent_id AND pt.locale_id = cs.locale_id
    WHERE c.site_id = $1 AND c.type = 'page' AND c.published_at IS NOT NULL
)
SELECT id AS page_id, locale_id, code AS locale_code, is_default, path
FROM page_tree ORDER BY path;

-- name: GetBlogRoutes :many
SELECT b.id AS blog_id, sl.code AS locale_code, sl.is_default, bs.slug AS path
FROM blogs b
JOIN blog_slugs bs ON bs.blog_id = b.id
JOIN site_locales sl ON sl.id = bs.locale_id
WHERE b.site_id = $1
ORDER BY bs.slug;

-- name: GetPostRoutes :many
WITH RECURSIVE post_tree AS (
    SELECT p.id, p.blog_id, ps.locale_id, ps.slug::TEXT AS path
    FROM pages p
    JOIN page_slugs ps ON ps.page_id = p.id
    WHERE p.site_id = $1 AND p.type = 'post' AND p.parent_id IS NULL
      AND p.published_at IS NOT NULL
  UNION ALL
    SELECT c.id, c.blog_id, cs.locale_id,
           CASE WHEN pt.path = '' THEN cs.slug ELSE pt.path || '/' || cs.slug END
    FROM pages c
    JOIN page_slugs cs ON cs.page_id = c.id
    JOIN post_tree pt ON pt.id = c.parent_id AND pt.locale_id = cs.locale_id
    WHERE c.site_id = $1 AND c.type = 'post' AND c.published_at IS NOT NULL
)
SELECT pt.id AS page_id, pt.blog_id, pt.locale_id, sl.code AS locale_code, sl.is_default,
       CASE WHEN pt.path = '' THEN bs.slug ELSE (bs.slug || '/' || pt.path)::TEXT END AS path
FROM post_tree pt
JOIN blog_slugs bs ON bs.blog_id = pt.blog_id AND bs.locale_id = pt.locale_id
JOIN site_locales sl ON sl.id = pt.locale_id
ORDER BY path;
