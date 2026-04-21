-- name: CreateMedia :one
INSERT INTO media (filename, mime_type, size, url, thumbnail_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMediaByID :one
SELECT * FROM media
WHERE id = $1;

-- name: ListMedia :many
SELECT * FROM media
WHERE ($1::BIGINT IS NULL OR id > $1)
ORDER BY id ASC
LIMIT $2;

-- name: DeleteMedia :one
DELETE FROM media
WHERE id = $1
RETURNING *;

-- name: CountMedia :one
SELECT COUNT(*) FROM media;
