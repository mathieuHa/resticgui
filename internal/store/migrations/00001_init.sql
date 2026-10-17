-- +goose Up
CREATE TABLE repositories (
    id                  INTEGER  PRIMARY KEY,
    name                TEXT     NOT NULL UNIQUE CHECK (name <> ''),
    -- backend decides which keys options and secrets carry
    backend             TEXT     NOT NULL CHECK (backend IN ('local', 's3', 'b2', 'rest', 'sftp', 'rclone', 'azure', 'gs')),
    -- the restic repository string, passed as RESTIC_REPOSITORY
    location            TEXT     NOT NULL CHECK (location <> ''),
    -- non-secret settings as JSON, see store.RepositoryOptions
    options             TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(options)),
    -- encrypted JSON, see store.RepositorySecrets; never returned by the API
    secrets             BLOB     NOT NULL,
    host                TEXT     NOT NULL DEFAULT '',
    compression         TEXT     NOT NULL DEFAULT 'auto' CHECK (compression IN ('auto', 'off', 'max')),
    exclude_larger_than TEXT     NOT NULL DEFAULT '',
    initialized_at      DATETIME,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE paths (
    id            INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    path          TEXT    NOT NULL CHECK (path <> ''),
    kind          TEXT    NOT NULL CHECK (kind IN ('include', 'exclude')),
    UNIQUE (repository_id, path, kind)
);

CREATE TABLE schedules (
    id            INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    -- forget applies the retention policy, then prunes when retention.prune is set
    job           TEXT    NOT NULL CHECK (job IN ('backup', 'forget', 'check')),
    -- a robfig/cron spec: 5 fields or a descriptor such as "@every 6h"
    cron          TEXT    NOT NULL CHECK (cron <> ''),
    enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (repository_id, job)
);

CREATE TABLE retention (
    repository_id INTEGER PRIMARY KEY REFERENCES repositories (id) ON DELETE CASCADE,
    keep_last     INTEGER NOT NULL DEFAULT 0 CHECK (keep_last >= 0),
    keep_hourly   INTEGER NOT NULL DEFAULT 0 CHECK (keep_hourly >= 0),
    keep_daily    INTEGER NOT NULL DEFAULT 0 CHECK (keep_daily >= 0),
    keep_weekly   INTEGER NOT NULL DEFAULT 0 CHECK (keep_weekly >= 0),
    keep_monthly  INTEGER NOT NULL DEFAULT 0 CHECK (keep_monthly >= 0),
    keep_yearly   INTEGER NOT NULL DEFAULT 0 CHECK (keep_yearly >= 0),
    -- a restic duration such as "2y5m7d3h", empty when unused
    keep_within   TEXT    NOT NULL DEFAULT '',
    prune         BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE jobs (
    id            INTEGER  PRIMARY KEY,
    repository_id INTEGER  NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    kind          TEXT     NOT NULL CHECK (kind IN ('init', 'backup', 'forget', 'prune', 'check', 'restore', 'unlock')),
    trigger       TEXT     NOT NULL CHECK (trigger IN ('manual', 'schedule')),
    status        TEXT     NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled', 'interrupted')),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at    DATETIME,
    finished_at   DATETIME,
    snapshot_id   TEXT     NOT NULL DEFAULT '',
    -- the restic summary message as JSON
    summary       TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(summary)),
    error         TEXT     NOT NULL DEFAULT '',
    exit_code     INTEGER,
    -- last lines of stderr, bounded by the job runner
    log_excerpt   TEXT     NOT NULL DEFAULT ''
);

CREATE INDEX jobs_repository_id ON jobs (repository_id, id DESC);
CREATE INDEX jobs_unfinished ON jobs (status) WHERE status IN ('queued', 'running');

-- +goose Down
DROP TABLE jobs;
DROP TABLE retention;
DROP TABLE schedules;
DROP TABLE paths;
DROP TABLE repositories;
