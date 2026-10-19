-- +goose Up
-- logins: the SHA-256 of each session token, so that a logout ends the
-- session on the server and a copy of the database holds no usable token
CREATE TABLE sessions (
    token_hash   BLOB    PRIMARY KEY,
    -- who the session belongs to: a hash of the username and password hash,
    -- so that changing either ends every session
    principal    TEXT    NOT NULL,
    created_at   INTEGER NOT NULL,
    last_seen_at INTEGER NOT NULL
);

-- +goose Down
DROP TABLE sessions;
