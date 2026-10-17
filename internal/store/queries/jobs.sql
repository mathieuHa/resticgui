-- name: CreateJob :one
INSERT INTO jobs (repository_id, kind, trigger) VALUES (?, ?, ?) RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = ?;

-- name: StartJob :execrows
UPDATE jobs SET status = 'running', started_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'queued';

-- name: FinishJob :execrows
UPDATE jobs
SET status      = sqlc.arg(status),
    finished_at = CURRENT_TIMESTAMP,
    snapshot_id = sqlc.arg(snapshot_id),
    summary     = sqlc.arg(summary),
    error       = sqlc.arg(error),
    exit_code   = sqlc.arg(exit_code),
    log_excerpt = sqlc.arg(log_excerpt)
WHERE id = sqlc.arg(id) AND status IN ('queued', 'running');

-- name: ListJobsForRepository :many
SELECT * FROM jobs WHERE repository_id = ? ORDER BY id DESC LIMIT ? OFFSET ?;

-- name: MarkUnfinishedJobsInterrupted :execrows
UPDATE jobs
SET status = 'interrupted', finished_at = CURRENT_TIMESTAMP
WHERE status IN ('queued', 'running');
