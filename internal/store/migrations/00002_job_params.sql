-- +goose Up
-- what a job runs on, as JSON: for restore {"snapshot_id", "target", "include"}
ALTER TABLE jobs ADD COLUMN params TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(params));

-- +goose Down
ALTER TABLE jobs DROP COLUMN params;
