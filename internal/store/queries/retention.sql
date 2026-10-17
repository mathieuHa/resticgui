-- name: UpsertRetention :one
INSERT INTO retention (repository_id, keep_last, keep_hourly, keep_daily, keep_weekly, keep_monthly, keep_yearly, keep_within, prune)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (repository_id) DO UPDATE SET
    keep_last    = excluded.keep_last,
    keep_hourly  = excluded.keep_hourly,
    keep_daily   = excluded.keep_daily,
    keep_weekly  = excluded.keep_weekly,
    keep_monthly = excluded.keep_monthly,
    keep_yearly  = excluded.keep_yearly,
    keep_within  = excluded.keep_within,
    prune        = excluded.prune
RETURNING *;

-- name: GetRetention :one
SELECT * FROM retention WHERE repository_id = ?;
