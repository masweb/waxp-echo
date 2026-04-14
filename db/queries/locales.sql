-- name: CreateSiteLocale :one
INSERT INTO site_locales (site_id, code, is_default)
VALUES ($1, $2, $3)
RETURNING id, site_id, code, is_default, created_at;

-- name: GetSiteLocaleByID :one
SELECT id, site_id, code, is_default, created_at FROM site_locales WHERE id = $1;

-- name: GetSiteLocaleByCodeAndSite :one
SELECT id, site_id, code, is_default, created_at FROM site_locales WHERE code = $1 AND site_id = $2;

-- name: ListSiteLocales :many
SELECT id, site_id, code, is_default, created_at
FROM site_locales
WHERE site_id = $1
ORDER BY is_default DESC, code ASC;

-- name: DeleteSiteLocale :exec
DELETE FROM site_locales WHERE id = $1 AND site_id = $2;

-- name: DeleteSiteLocaleByCode :exec
DELETE FROM site_locales WHERE code = $1 AND site_id = $2;
