-- name: CreateRepository :one
INSERT INTO repositories (name, backend, location, options, secrets, host)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetRepository :one
SELECT * FROM repositories WHERE id = ?;

-- name: GetRepositoryByName :one
SELECT * FROM repositories WHERE name = ?;

-- name: ListRepositories :many
SELECT * FROM repositories ORDER BY name;

-- name: UpdateRepository :one
UPDATE repositories
SET name                = sqlc.arg(name),
    backend             = sqlc.arg(backend),
    location            = sqlc.arg(location),
    options             = sqlc.arg(options),
    host                = sqlc.arg(host),
    updated_at          = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateRepositorySecrets :execrows
UPDATE repositories SET secrets = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: MarkRepositoryInitialized :exec
UPDATE repositories SET initialized_at = CURRENT_TIMESTAMP WHERE id = ? AND initialized_at IS NULL;

-- name: DeleteRepository :execrows
DELETE FROM repositories WHERE id = ?;
