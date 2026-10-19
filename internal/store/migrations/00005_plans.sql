-- +goose Up
-- what to back up, separate from where: a repository has several plans, each
-- with its paths, options, schedules and retention, deduplicated together
CREATE TABLE plans (
    id            INTEGER  PRIMARY KEY,
    repository_id INTEGER  NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    name          TEXT     NOT NULL CHECK (name <> ''),
    -- put on every snapshot of the plan, and the only snapshots its
    -- retention applies to
    tag           TEXT     NOT NULL UNIQUE CHECK (tag <> ''),
    -- more tags put on its snapshots, as a JSON list
    tags          TEXT     NOT NULL DEFAULT '[]' CHECK (json_valid(tags)),
    -- restic backup options as JSON, see store.PlanOptions
    options       TEXT     NOT NULL DEFAULT '{}' CHECK (json_valid(options)),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (repository_id, name)
);

CREATE TABLE plan_paths (
    id      INTEGER PRIMARY KEY,
    plan_id INTEGER NOT NULL REFERENCES plans (id) ON DELETE CASCADE,
    path    TEXT    NOT NULL CHECK (path <> ''),
    kind    TEXT    NOT NULL CHECK (kind IN ('include', 'exclude')),
    UNIQUE (plan_id, path, kind)
);

CREATE TABLE plan_schedules (
    plan_id INTEGER NOT NULL REFERENCES plans (id) ON DELETE CASCADE,
    -- forget applies the plan's retention, then prunes when it says so
    job     TEXT    NOT NULL CHECK (job IN ('backup', 'forget')),
    cron    TEXT    NOT NULL CHECK (cron <> ''),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (plan_id, job)
);

CREATE TABLE plan_retention (
    plan_id      INTEGER PRIMARY KEY REFERENCES plans (id) ON DELETE CASCADE,
    keep_last    INTEGER NOT NULL DEFAULT 0 CHECK (keep_last >= 0),
    keep_hourly  INTEGER NOT NULL DEFAULT 0 CHECK (keep_hourly >= 0),
    keep_daily   INTEGER NOT NULL DEFAULT 0 CHECK (keep_daily >= 0),
    keep_weekly  INTEGER NOT NULL DEFAULT 0 CHECK (keep_weekly >= 0),
    keep_monthly INTEGER NOT NULL DEFAULT 0 CHECK (keep_monthly >= 0),
    keep_yearly  INTEGER NOT NULL DEFAULT 0 CHECK (keep_yearly >= 0),
    keep_within  TEXT    NOT NULL DEFAULT '',
    prune        BOOLEAN NOT NULL DEFAULT TRUE
);

-- the plan a backup or forget job ran for; kept when the plan is deleted,
-- so that its history stays
ALTER TABLE jobs ADD COLUMN plan_id INTEGER;
CREATE INDEX jobs_plan_id ON jobs (plan_id, id DESC) WHERE plan_id IS NOT NULL;

-- each repository gets one plan with what it had. It keeps the tag of the
-- snapshots made so far, so that its retention goes on applying to them.
INSERT INTO plans (id, repository_id, name, tag, options)
SELECT id, id, 'Default', 'resticgui:' || id,
       json_object('compression', compression, 'exclude_larger_than', exclude_larger_than)
FROM repositories;
INSERT INTO plan_paths (plan_id, path, kind) SELECT repository_id, path, kind FROM paths;
INSERT INTO plan_schedules (plan_id, job, cron, enabled)
SELECT repository_id, job, cron, enabled FROM schedules WHERE job IN ('backup', 'forget');
DELETE FROM schedules WHERE job IN ('backup', 'forget');
INSERT INTO plan_retention (plan_id, keep_last, keep_hourly, keep_daily, keep_weekly, keep_monthly, keep_yearly, keep_within, prune)
SELECT repository_id, keep_last, keep_hourly, keep_daily, keep_weekly, keep_monthly, keep_yearly, keep_within, prune FROM retention;
UPDATE jobs SET plan_id = repository_id WHERE kind IN ('backup', 'forget');

DROP TABLE paths;
DROP TABLE retention;
ALTER TABLE repositories DROP COLUMN compression;
ALTER TABLE repositories DROP COLUMN exclude_larger_than;

-- +goose Down
-- each repository gets back the settings of its first plan; other plans are
-- lost, their snapshots stay in the repository
ALTER TABLE repositories ADD COLUMN compression TEXT NOT NULL DEFAULT 'auto' CHECK (compression IN ('auto', 'off', 'max'));
ALTER TABLE repositories ADD COLUMN exclude_larger_than TEXT NOT NULL DEFAULT '';

CREATE TABLE paths (
    id            INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    path          TEXT    NOT NULL CHECK (path <> ''),
    kind          TEXT    NOT NULL CHECK (kind IN ('include', 'exclude')),
    UNIQUE (repository_id, path, kind)
);

CREATE TABLE retention (
    repository_id INTEGER PRIMARY KEY REFERENCES repositories (id) ON DELETE CASCADE,
    keep_last     INTEGER NOT NULL DEFAULT 0 CHECK (keep_last >= 0),
    keep_hourly   INTEGER NOT NULL DEFAULT 0 CHECK (keep_hourly >= 0),
    keep_daily    INTEGER NOT NULL DEFAULT 0 CHECK (keep_daily >= 0),
    keep_weekly   INTEGER NOT NULL DEFAULT 0 CHECK (keep_weekly >= 0),
    keep_monthly  INTEGER NOT NULL DEFAULT 0 CHECK (keep_monthly >= 0),
    keep_yearly   INTEGER NOT NULL DEFAULT 0 CHECK (keep_yearly >= 0),
    keep_within   TEXT    NOT NULL DEFAULT '',
    prune         BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TEMP TABLE first_plans AS
SELECT min(id) AS id, repository_id FROM plans GROUP BY repository_id;

UPDATE repositories SET
    compression = coalesce((SELECT json_extract(p.options, '$.compression') FROM plans p JOIN first_plans f ON f.id = p.id
                            WHERE f.repository_id = repositories.id), 'auto'),
    exclude_larger_than = coalesce((SELECT json_extract(p.options, '$.exclude_larger_than') FROM plans p JOIN first_plans f ON f.id = p.id
                                    WHERE f.repository_id = repositories.id), '');
INSERT INTO paths (repository_id, path, kind)
SELECT f.repository_id, pp.path, pp.kind FROM plan_paths pp JOIN first_plans f ON f.id = pp.plan_id;
INSERT INTO schedules (repository_id, job, cron, enabled)
SELECT f.repository_id, s.job, s.cron, s.enabled FROM plan_schedules s JOIN first_plans f ON f.id = s.plan_id;
INSERT INTO retention (repository_id, keep_last, keep_hourly, keep_daily, keep_weekly, keep_monthly, keep_yearly, keep_within, prune)
SELECT f.repository_id, r.keep_last, r.keep_hourly, r.keep_daily, r.keep_weekly, r.keep_monthly, r.keep_yearly, r.keep_within, r.prune
FROM plan_retention r JOIN first_plans f ON f.id = r.plan_id;
DROP TABLE first_plans;

DROP INDEX jobs_plan_id;
ALTER TABLE jobs DROP COLUMN plan_id;
DROP TABLE plan_retention;
DROP TABLE plan_schedules;
DROP TABLE plan_paths;
DROP TABLE plans;
