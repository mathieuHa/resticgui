-- name: CreateNotifier :one
INSERT INTO notifiers (name, url, service, on_failure, on_success, on_missed, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetNotifier :one
SELECT * FROM notifiers WHERE id = ?;

-- name: ListNotifiers :many
SELECT * FROM notifiers ORDER BY name;

-- name: UpdateNotifier :one
UPDATE notifiers
SET name       = sqlc.arg(name),
    on_failure = sqlc.arg(on_failure),
    on_success = sqlc.arg(on_success),
    on_missed  = sqlc.arg(on_missed),
    enabled    = sqlc.arg(enabled)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateNotifierURL :execrows
UPDATE notifiers SET url = ?, service = ? WHERE id = ?;

-- name: DeleteNotifier :execrows
DELETE FROM notifiers WHERE id = ?;
