-- +goose Up
-- where to send alerts, for every repository: a Shoutrrr URL such as
-- ntfy://ntfy.sh/topic or discord://token@id
CREATE TABLE notifiers (
    id         INTEGER  PRIMARY KEY,
    name       TEXT     NOT NULL UNIQUE CHECK (name <> ''),
    -- the encrypted URL, which carries tokens; never returned by the API
    url        BLOB     NOT NULL,
    -- the URL scheme, such as ntfy or discord, shown instead of the URL
    service    TEXT     NOT NULL,
    on_failure BOOLEAN  NOT NULL DEFAULT TRUE,
    on_success BOOLEAN  NOT NULL DEFAULT FALSE,
    -- a scheduled backup that did not succeed on time
    on_missed  BOOLEAN  NOT NULL DEFAULT TRUE,
    enabled    BOOLEAN  NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- a Healthchecks.io (or self-hosted) check pinged by every backup
CREATE TABLE healthchecks (
    repository_id INTEGER PRIMARY KEY REFERENCES repositories (id) ON DELETE CASCADE,
    -- the encrypted ping URL; never returned by the API
    url           BLOB    NOT NULL,
    host          TEXT    NOT NULL
);

-- commands or webhooks run around backups
CREATE TABLE hooks (
    id              INTEGER PRIMARY KEY,
    repository_id   INTEGER NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    -- before: a failure stops the backup; success and failure: after it
    event           TEXT    NOT NULL CHECK (event IN ('before', 'success', 'failure')),
    kind            TEXT    NOT NULL CHECK (kind IN ('command', 'webhook')),
    -- the shell command, empty for a webhook
    command         TEXT    NOT NULL DEFAULT '',
    -- the encrypted webhook URL, empty for a command; never returned by the API
    url             BLOB,
    -- the webhook's host, shown instead of its URL
    host            TEXT    NOT NULL DEFAULT '',
    timeout_seconds INTEGER NOT NULL DEFAULT 300 CHECK (timeout_seconds > 0),
    position        INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX hooks_repository_id ON hooks (repository_id, event, position);

-- the last snapshot count and size, refreshed after the jobs that change
-- them, for the metrics endpoint
CREATE TABLE repository_stats (
    repository_id INTEGER  PRIMARY KEY REFERENCES repositories (id) ON DELETE CASCADE,
    snapshots     INTEGER  NOT NULL,
    total_size    INTEGER  NOT NULL,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE repository_stats;
DROP INDEX hooks_repository_id;
DROP TABLE hooks;
DROP TABLE healthchecks;
DROP TABLE notifiers;
