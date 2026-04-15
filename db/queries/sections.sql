-- name: CreateSectionCounter :one
INSERT INTO section_counters (site_id, current_value)
VALUES ($1, 0)
RETURNING site_id, current_value;

-- name: GetSectionCounter :one
SELECT site_id, current_value FROM section_counters WHERE site_id = $1;

-- name: GetNextSectionID :one
INSERT INTO section_counters (site_id, current_value)
VALUES ($1, 1)
ON CONFLICT (site_id) DO UPDATE
SET current_value = section_counters.current_value + 1
RETURNING current_value AS id;
