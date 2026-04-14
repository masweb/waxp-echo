-- name: CreateSite :one
INSERT INTO sites (name, domain, options)
VALUES ($1, $2, $3)
RETURNING id, name, domain, options, created_at, updated_at;

-- name: GetSiteByID :one
SELECT id, name, domain, options, created_at, updated_at FROM sites WHERE id = $1;

-- name: GetSiteByDomain :one
SELECT id, name, domain, options, created_at, updated_at FROM sites WHERE domain = $1;

-- name: ListSites :many
SELECT id, name, domain, options, created_at, updated_at
FROM sites
WHERE ($1::BIGINT IS NULL OR id > $1)
ORDER BY id ASC
LIMIT $2;

-- name: CountSites :one
SELECT COUNT(*) FROM sites;

-- name: UpdateSite :one
UPDATE sites
SET name = $1, domain = $2, options = $3, updated_at = NOW()
WHERE id = $4
RETURNING id, name, domain, options, created_at, updated_at;

-- name: DeleteSite :exec
DELETE FROM sites WHERE id = $1;
