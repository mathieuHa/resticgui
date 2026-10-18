-- name: UpsertHealthcheck :exec
INSERT INTO healthchecks (repository_id, url, host) VALUES (?, ?, ?)
ON CONFLICT (repository_id) DO UPDATE SET url = excluded.url, host = excluded.host;

-- name: GetHealthcheck :one
SELECT * FROM healthchecks WHERE repository_id = ?;

-- name: ListHealthchecks :many
SELECT * FROM healthchecks;

-- name: DeleteHealthcheck :execrows
DELETE FROM healthchecks WHERE repository_id = ?;

-- name: CreateHook :one
INSERT INTO hooks (repository_id, event, kind, command, url, host, timeout_seconds, position)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListHooks :many
SELECT * FROM hooks WHERE repository_id = ? ORDER BY event, position, id;

-- name: ListAllHooks :many
SELECT * FROM hooks ORDER BY repository_id, event, position, id;

-- name: DeleteHooks :exec
DELETE FROM hooks WHERE repository_id = ?;

-- name: UpsertRepositoryStats :exec
INSERT INTO repository_stats (repository_id, snapshots, total_size) VALUES (?, ?, ?)
ON CONFLICT (repository_id) DO UPDATE
SET snapshots = excluded.snapshots, total_size = excluded.total_size, updated_at = CURRENT_TIMESTAMP;

-- name: ListRepositoryStats :many
SELECT * FROM repository_stats;
