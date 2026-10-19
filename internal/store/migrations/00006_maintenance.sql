-- +goose Up
-- the maintenance tools run as jobs: SQLite cannot change a CHECK, so the
-- table is made again with the new kinds, keeping its columns in order
CREATE TABLE jobs_new (
    id            INTEGER  PRIMARY KEY,
    repository_id INTEGER  NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    kind          TEXT     NOT NULL CHECK (kind IN ('init', 'backup', 'forget', 'prune', 'check', 'restore', 'unlock',
                                                    'repair_index', 'repair_snapshots', 'migrate', 'change_password')),
    trigger       TEXT     NOT NULL CHECK (trigger IN ('manual', 'schedule')),
    status        TEXT     NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled', 'interrupted')),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at    DATETIME,
    finished_at   DATETIME,
    snapshot_id   TEXT     NOT NULL DEFAULT '',
    summary       TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(summary)),
    error         TEXT     NOT NULL DEFAULT '',
    exit_code     INTEGER,
    log_excerpt   TEXT     NOT NULL DEFAULT '',
    params        TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(params)),
    plan_id       INTEGER
);
INSERT INTO jobs_new SELECT id, repository_id, kind, trigger, status, created_at, started_at, finished_at, snapshot_id,
                            summary, error, exit_code, log_excerpt, params, plan_id FROM jobs;
DROP TABLE jobs;
ALTER TABLE jobs_new RENAME TO jobs;
CREATE INDEX jobs_repository_id ON jobs (repository_id, id DESC);
CREATE INDEX jobs_unfinished ON jobs (status) WHERE status IN ('queued', 'running');
CREATE INDEX jobs_plan_id ON jobs (plan_id, id DESC) WHERE plan_id IS NOT NULL;

-- how much data a scheduled check reads, as restic's --read-data-subset
-- ("10%", "1/5", "500M"); empty reads none
ALTER TABLE schedules ADD COLUMN read_data_subset TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE schedules DROP COLUMN read_data_subset;

CREATE TABLE jobs_old (
    id            INTEGER  PRIMARY KEY,
    repository_id INTEGER  NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    kind          TEXT     NOT NULL CHECK (kind IN ('init', 'backup', 'forget', 'prune', 'check', 'restore', 'unlock')),
    trigger       TEXT     NOT NULL CHECK (trigger IN ('manual', 'schedule')),
    status        TEXT     NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled', 'interrupted')),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at    DATETIME,
    finished_at   DATETIME,
    snapshot_id   TEXT     NOT NULL DEFAULT '',
    summary       TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(summary)),
    error         TEXT     NOT NULL DEFAULT '',
    exit_code     INTEGER,
    log_excerpt   TEXT     NOT NULL DEFAULT '',
    params        TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(params)),
    plan_id       INTEGER
);
-- the runs of the maintenance tools are lost
INSERT INTO jobs_old SELECT id, repository_id, kind, trigger, status, created_at, started_at, finished_at, snapshot_id,
                            summary, error, exit_code, log_excerpt, params, plan_id FROM jobs
WHERE kind IN ('init', 'backup', 'forget', 'prune', 'check', 'restore', 'unlock');
DROP TABLE jobs;
ALTER TABLE jobs_old RENAME TO jobs;
CREATE INDEX jobs_repository_id ON jobs (repository_id, id DESC);
CREATE INDEX jobs_unfinished ON jobs (status) WHERE status IN ('queued', 'running');
CREATE INDEX jobs_plan_id ON jobs (plan_id, id DESC) WHERE plan_id IS NOT NULL;
