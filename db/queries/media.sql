-- name: CreateMedia :one
INSERT INTO media (filename, mime_type, size, url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMediaByID :one
SELECT * FROM media
WHERE id = $1;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteMedia :one
DELETE FROM media
WHERE id = $1
RETURNING *;

-- name: CountMedia :one
SELECT COUNT(*) FROM media;
