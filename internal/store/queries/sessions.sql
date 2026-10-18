-- name: CreateSession :exec
INSERT INTO sessions (token_hash, principal, created_at, last_seen_at) VALUES (?, ?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions WHERE token_hash = ?;

-- name: TouchSession :exec
UPDATE sessions SET last_seen_at = ? WHERE token_hash = ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = ?;

-- name: DeleteStaleSessions :execrows
-- sessions past their lifetime or idle timeout, or of another principal
DELETE FROM sessions
WHERE created_at < sqlc.arg(created_before)
   OR last_seen_at < sqlc.arg(seen_before)
   OR principal <> sqlc.arg(principal);
