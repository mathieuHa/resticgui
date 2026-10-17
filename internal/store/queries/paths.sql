-- name: ListPaths :many
SELECT * FROM paths WHERE repository_id = ? ORDER BY kind, path;

-- name: AddPath :one
INSERT INTO paths (repository_id, path, kind) VALUES (?, ?, ?) RETURNING *;

-- name: DeletePaths :exec
DELETE FROM paths WHERE repository_id = ?;
