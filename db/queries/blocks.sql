-- name: CreateBlockCounter :one
INSERT INTO block_counters (site_id, current_value)
VALUES ($1, 2)
RETURNING site_id, current_value;

-- name: GetNextBlockID :one
INSERT INTO block_counters (site_id, current_value)
VALUES ($1, 1)
ON CONFLICT (site_id) DO UPDATE
SET current_value = block_counters.current_value + 1
RETURNING current_value AS id;
