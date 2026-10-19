-- name: CreateJob :one
INSERT INTO jobs (repository_id, plan_id, kind, trigger, params) VALUES (?, ?, ?, ?, ?) RETURNING *;

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

-- name: LastJobs :many
-- the latest finished job of each kind for each repository, and the latest
-- successful one, for the metrics endpoint and missed-backup alerts
SELECT j.* FROM jobs j
WHERE j.id IN (
    SELECT max(id) FROM jobs WHERE status NOT IN ('queued', 'running') GROUP BY repository_id, kind
    UNION
    SELECT max(id) FROM jobs WHERE status = 'succeeded' GROUP BY repository_id, kind
)
ORDER BY j.repository_id, j.kind, j.id;
