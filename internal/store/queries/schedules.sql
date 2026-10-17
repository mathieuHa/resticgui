-- name: UpsertSchedule :one
INSERT INTO schedules (repository_id, job, cron, enabled)
VALUES (?, ?, ?, ?)
ON CONFLICT (repository_id, job) DO UPDATE SET cron = excluded.cron, enabled = excluded.enabled
RETURNING *;

-- name: ListSchedules :many
SELECT * FROM schedules WHERE repository_id = ? ORDER BY job;

-- name: ListEnabledSchedules :many
SELECT * FROM schedules WHERE enabled ORDER BY repository_id, job;

-- name: DeleteSchedule :execrows
DELETE FROM schedules WHERE repository_id = ? AND job = ?;
