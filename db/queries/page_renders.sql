-- name: UpsertPageRender :one
INSERT INTO page_renders (page_id, locale_id, html, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (page_id, locale_id)
DO UPDATE SET html = EXCLUDED.html, updated_at = NOW()
RETURNING id, page_id, locale_id, html, updated_at;

-- name: GetPageRenderByPageAndLocale :one
SELECT id, page_id, locale_id, html, updated_at
FROM page_renders
WHERE page_id = $1 AND locale_id = $2;

-- name: DeletePageRendersByPageID :exec
DELETE FROM page_renders WHERE page_id = $1;

-- name: GetPublishedPageSlug :one
SELECT ps.page_id, sl.id AS locale_id, sl.code AS locale_code, sl.is_default
FROM page_slugs ps
JOIN site_locales sl ON sl.id = ps.locale_id
JOIN pages p ON p.id = ps.page_id
WHERE p.site_id = $1 AND ps.slug = $2 AND sl.code = $3 AND p.published_at IS NOT NULL;

-- name: GetAllPublishedPageIDs :many
SELECT DISTINCT p.id FROM pages p WHERE p.site_id = $1 AND p.published_at IS NOT NULL;
