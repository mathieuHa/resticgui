-- name: CreatePlan :one
INSERT INTO plans (repository_id, name, tag, tags, options) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = ?;

-- name: ListPlans :many
SELECT * FROM plans WHERE repository_id = ? ORDER BY id;

-- name: UpdatePlan :one
UPDATE plans SET name = sqlc.arg(name), tags = sqlc.arg(tags), options = sqlc.arg(options)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePlan :execrows
DELETE FROM plans WHERE id = ?;

-- name: ListPlanPaths :many
SELECT * FROM plan_paths WHERE plan_id = ? ORDER BY kind, path;

-- name: AddPlanPath :exec
INSERT INTO plan_paths (plan_id, path, kind) VALUES (?, ?, ?);

-- name: DeletePlanPaths :exec
DELETE FROM plan_paths WHERE plan_id = ?;

-- name: UpsertPlanSchedule :exec
INSERT INTO plan_schedules (plan_id, job, cron, enabled)
VALUES (?, ?, ?, ?)
ON CONFLICT (plan_id, job) DO UPDATE SET cron = excluded.cron, enabled = excluded.enabled;

-- name: ListPlanSchedules :many
SELECT * FROM plan_schedules WHERE plan_id = ? ORDER BY job;

-- name: DeletePlanSchedules :exec
DELETE FROM plan_schedules WHERE plan_id = ?;

-- name: ListEnabledPlanSchedules :many
SELECT p.repository_id, s.plan_id, s.job, s.cron FROM plan_schedules s JOIN plans p ON p.id = s.plan_id
WHERE s.enabled
ORDER BY p.repository_id, s.plan_id, s.job;

-- name: UpsertPlanRetention :exec
INSERT INTO plan_retention (plan_id, keep_last, keep_hourly, keep_daily, keep_weekly, keep_monthly, keep_yearly, keep_within, prune)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (plan_id) DO UPDATE SET
    keep_last    = excluded.keep_last,
    keep_hourly  = excluded.keep_hourly,
    keep_daily   = excluded.keep_daily,
    keep_weekly  = excluded.keep_weekly,
    keep_monthly = excluded.keep_monthly,
    keep_yearly  = excluded.keep_yearly,
    keep_within  = excluded.keep_within,
    prune        = excluded.prune;

-- name: GetPlanRetention :one
SELECT * FROM plan_retention WHERE plan_id = ?;

-- name: DeletePlanRetention :exec
DELETE FROM plan_retention WHERE plan_id = ?;

-- name: LastPlanBackups :many
-- the latest finished backup of each plan, for missed-backup alerts
SELECT * FROM jobs
WHERE id IN (
    SELECT max(id) FROM jobs
    WHERE kind = 'backup' AND plan_id IS NOT NULL AND status NOT IN ('queued', 'running')
    GROUP BY plan_id
)
ORDER BY plan_id;
