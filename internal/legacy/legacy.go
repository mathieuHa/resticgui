// Package legacy imports a database written by the ent-based ResticGUI
// (v0.1.x) into the new store.
package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// Report says what an import did.
type Report struct {
	Imported []string
	// Skipped repositories and why, e.g. "unconfigured: no storage backend configured".
	Skipped []string
	// Warnings about data that was imported partially, e.g. an invalid schedule.
	Warnings []string
	// LogsIgnored counts command log rows, which are not carried over.
	LogsIgnored int
}

type legacyRepository struct {
	id                  int64
	name, provider      string
	password, maxSize   string
	active, pruneActive bool
}

// Import reads the legacy database at path (opened read-only) and writes
// every repository it can convert into dst, in one transaction: either all
// convertible repositories are imported or none are.
func Import(ctx context.Context, path string, dst *store.Store, box *secrets.Box) (Report, error) {
	var report Report
	src, err := sql.Open("sqlite", "file:"+path+"?"+url.Values{"mode": {"ro"}}.Encode())
	if err != nil {
		return report, err
	}
	defer src.Close()
	if err := src.PingContext(ctx); err != nil {
		return report, fmt.Errorf("legacy: open %s: %w", path, err)
	}

	repos, err := readRepositories(ctx, src)
	if err != nil {
		return report, err
	}
	if err := src.QueryRowContext(ctx, "SELECT count(*) FROM `Logs`").Scan(&report.LogsIgnored); err != nil {
		return report, fmt.Errorf("legacy: count logs: %w", err)
	}

	err = dst.InTx(ctx, func(q *db.Queries) error {
		for _, r := range repos {
			skip, err := importRepository(ctx, src, q, box, r, &report)
			if err != nil {
				return fmt.Errorf("legacy: repository %q: %w", r.name, err)
			}
			if skip != "" {
				report.Skipped = append(report.Skipped, r.name+": "+skip)
				continue
			}
			report.Imported = append(report.Imported, r.name)
		}
		return nil
	})
	if err != nil {
		return Report{}, err
	}
	return report, nil
}

func readRepositories(ctx context.Context, src *sql.DB) ([]legacyRepository, error) {
	rows, err := src.QueryContext(ctx, "SELECT id, name, provider, password, max_size, active, prune_active FROM `repositories` ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("legacy: read repositories (is this a ResticGUI v0.1 database?): %w", err)
	}
	defer rows.Close()
	var repos []legacyRepository
	for rows.Next() {
		var r legacyRepository
		var provider, password, maxSize sql.NullString
		if err := rows.Scan(&r.id, &r.name, &provider, &password, &maxSize, &r.active, &r.pruneActive); err != nil {
			return nil, err
		}
		r.provider, r.password, r.maxSize = provider.String, password.String, maxSize.String
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

// importRepository returns a non-empty reason when the repository is skipped.
func importRepository(ctx context.Context, src *sql.DB, q *db.Queries, box *secrets.Box, r legacyRepository, report *Report) (string, error) {
	if r.password == "" {
		return "no password set", nil
	}
	backend, err := readBackend(ctx, src, r)
	if err != nil {
		return "", err
	}
	if backend == nil {
		return "no storage backend configured", nil
	}

	backend.secrets.Password = r.password
	sealed, err := store.SealSecrets(box, backend.secrets)
	if err != nil {
		return "", err
	}
	options, err := store.EncodeOptions(backend.options)
	if err != nil {
		return "", err
	}
	// initialized_at is left empty: v0.1 set its "initiated" flag even when
	// restic init failed, so the new code checks the repository itself.
	created, err := q.CreateRepository(ctx, db.CreateRepositoryParams{
		Name: r.name, Backend: backend.kind, Location: backend.location, Options: options,
		Secrets: sealed, Compression: "auto", ExcludeLargerThan: r.maxSize,
	})
	if err != nil {
		return "", err
	}

	if err := importPaths(ctx, src, q, r.id, created.ID); err != nil {
		return "", err
	}
	if err := importRetention(ctx, src, q, r.id, created.ID); err != nil {
		return "", err
	}
	schedules := []struct {
		table, column, job string
		enabled            bool
	}{
		{"BackupSchedule", "repository_backup_schedule", "backup", r.active},
		// the v0.1 prune schedule ran forget with the policy, then prune
		{"PruneSchedule", "repository_prune_schedule", "forget", r.pruneActive},
	}
	for _, s := range schedules {
		spec, found, err := readSchedule(ctx, src, s.table, s.column, r.id)
		if err != nil {
			return "", err
		}
		if !found {
			continue
		}
		if spec.err != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s: %s schedule not imported: %v", r.name, s.job, spec.err))
			continue
		}
		if _, err := q.UpsertSchedule(ctx, db.UpsertScheduleParams{
			RepositoryID: created.ID, Job: s.job, Cron: spec.cron, Enabled: s.enabled,
		}); err != nil {
			return "", err
		}
	}
	return "", nil
}

type backend struct {
	kind, location string
	options        store.RepositoryOptions
	secrets        store.RepositorySecrets
}

// readBackend mirrors how v0.1 built the restic repository string and
// environment (internal/commands/provider.go). It returns nil when the
// repository has no provider row.
func readBackend(ctx context.Context, src *sql.DB, r legacyRepository) (*backend, error) {
	row := func(query string, dest ...any) (bool, error) {
		err := src.QueryRowContext(ctx, query, r.id).Scan(dest...)
		if err == sql.ErrNoRows {
			return false, nil
		}
		return err == nil, err
	}

	switch r.provider {
	case "", "localProvider":
		var path string
		found, err := row("SELECT repo_path FROM `LocalProvider` WHERE repository_local_provider = ?", &path)
		if !found || err != nil {
			return nil, err
		}
		return &backend{kind: "local", location: path}, nil

	case "AWSS3Provider":
		var bucket, keyID, secret string
		var region sql.NullString
		found, err := row("SELECT bucket_name, access_key_id, secret_access_key, default_region FROM `AWSS3Provider` WHERE repository_awss3provider = ?",
			&bucket, &keyID, &secret, &region)
		if !found || err != nil {
			return nil, err
		}
		b := &backend{
			kind: "s3", location: "s3:s3.amazonaws.com/" + bucket,
			secrets: store.RepositorySecrets{Env: map[string]string{"AWS_ACCESS_KEY_ID": keyID, "AWS_SECRET_ACCESS_KEY": secret}},
		}
		if region.String != "" {
			b.options.Env = map[string]string{"AWS_DEFAULT_REGION": region.String}
		}
		return b, nil

	case "B2Provider":
		var bucket, repoName, accountID, accountKey string
		found, err := row("SELECT bucket_name, repo_name, account_id, account_key FROM `B2Provider` WHERE repository_b2provider = ?",
			&bucket, &repoName, &accountID, &accountKey)
		if !found || err != nil {
			return nil, err
		}
		return &backend{
			kind: "b2", location: fmt.Sprintf("b2:%s:%s", bucket, repoName),
			secrets: store.RepositorySecrets{Env: map[string]string{"B2_ACCOUNT_ID": accountID, "B2_ACCOUNT_KEY": accountKey}},
		}, nil

	case "MinioS3Provider":
		var bucket, host, port, scheme, keyID, secret, ca string
		found, err := row("SELECT bucket_name, hostname, port, scheme, access_key_id, secret_access_key, certificate_authority FROM `MinioS3Provider` WHERE repository_minio_s3provider = ?",
			&bucket, &host, &port, &scheme, &keyID, &secret, &ca)
		if !found || err != nil {
			return nil, err
		}
		return &backend{
			kind: "s3", location: fmt.Sprintf("s3:%s://%s:%s/%s", scheme, host, port, bucket),
			options: store.RepositoryOptions{CACert: ca},
			secrets: store.RepositorySecrets{Env: map[string]string{"AWS_ACCESS_KEY_ID": keyID, "AWS_SECRET_ACCESS_KEY": secret}},
		}, nil
	}
	return nil, fmt.Errorf("unknown provider %q", r.provider)
}

func importPaths(ctx context.Context, src *sql.DB, q *db.Queries, legacyID, newID int64) error {
	rows, err := src.QueryContext(ctx, "SELECT abs_path, rel_path, included FROM `paths` WHERE repository_paths = ? ORDER BY id", legacyID)
	if err != nil {
		return err
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var absPath, relPath string
		var included bool
		if err := rows.Scan(&absPath, &relPath, &included); err != nil {
			return err
		}
		// v0.1 passed rel_path to restic (abs_path was only displayed)
		path := relPath
		if path == "" {
			path = absPath
		}
		kind := "exclude"
		if included {
			kind = "include"
		}
		if path == "" || seen[kind+path] {
			continue
		}
		seen[kind+path] = true
		if _, err := q.AddPath(ctx, db.AddPathParams{RepositoryID: newID, Path: path, Kind: kind}); err != nil {
			return err
		}
	}
	return rows.Err()
}

func importRetention(ctx context.Context, src *sql.DB, q *db.Queries, legacyID, newID int64) error {
	var p db.UpsertRetentionParams
	err := src.QueryRowContext(ctx,
		"SELECT nb_keep_last, nb_keep_hours, nb_keep_days, nb_keep_weeks, nb_keep_months FROM `BackupPolicy` WHERE repository_backup_policy = ?", legacyID).
		Scan(&p.KeepLast, &p.KeepHourly, &p.KeepDaily, &p.KeepWeekly, &p.KeepMonthly)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	// keep_all was stored but never used by v0.1, so it is not carried over.
	p.RepositoryID, p.Prune = newID, true
	_, err = q.UpsertRetention(ctx, p)
	return err
}

type scheduleSpec struct {
	cron string
	err  error
}

func readSchedule(ctx context.Context, src *sql.DB, table, column string, legacyID int64) (scheduleSpec, bool, error) {
	var s legacySchedule
	var hour, day, interval, cronSyntax sql.NullString
	err := src.QueryRowContext(ctx, "SELECT choice, every_day, every_week, every_month, every_monday, every_tuesday, every_wednesday, every_thursday, every_friday, every_saturday, every_sunday, hour, day, interval, cron_syntax FROM `"+table+"` WHERE "+column+" = ?", legacyID).
		Scan(&s.choice, &s.everyDay, &s.everyWeek, &s.everyMonth, &s.weekdays[1], &s.weekdays[2], &s.weekdays[3], &s.weekdays[4], &s.weekdays[5], &s.weekdays[6], &s.weekdays[0],
			&hour, &day, &interval, &cronSyntax)
	if err == sql.ErrNoRows {
		return scheduleSpec{}, false, nil
	}
	if err != nil {
		return scheduleSpec{}, false, err
	}
	s.hour, s.day, s.interval, s.cronSyntax = hour.String, day.String, interval.String, cronSyntax.String
	spec, err := s.cron()
	return scheduleSpec{cron: spec, err: err}, true, nil
}

type legacySchedule struct {
	choice                          string
	everyDay, everyWeek, everyMonth bool
	weekdays                        [7]bool // index 0 is Sunday, as in cron
	hour, day, interval, cronSyntax string
}

// cron converts a v0.1 schedule into a robfig/cron spec, following v0.1's
// getScheduleStringSyntax, plus the "monthly" choice it never handled.
func (s legacySchedule) cron() (string, error) {
	var spec string
	switch s.choice {
	case "cron":
		spec = strings.TrimSpace(s.cronSyntax)
	case "interval":
		d, err := parseInterval(s.interval)
		if err != nil {
			return "", err
		}
		spec = "@every " + d.String()
	case "weekly":
		at, err := time.Parse("15:04", s.hour)
		if err != nil {
			return "", fmt.Errorf("invalid hour %q", s.hour)
		}
		var days []string
		// v0.1 listed Monday to Saturday, then Sunday
		for _, d := range []int{1, 2, 3, 4, 5, 6, 0} {
			if s.weekdays[d] {
				days = append(days, fmt.Sprint(d))
			}
		}
		if len(days) == 0 {
			return "", fmt.Errorf("weekly schedule without any day")
		}
		spec = fmt.Sprintf("%d %d * * %s", at.Minute(), at.Hour(), strings.Join(days, ","))
	case "monthly", "regular":
		at, err := time.Parse("15:04", s.hour)
		if err != nil {
			return "", fmt.Errorf("invalid hour %q", s.hour)
		}
		switch {
		case s.everyDay:
			spec = fmt.Sprintf("%d %d * * *", at.Minute(), at.Hour())
		case s.everyWeek:
			spec = fmt.Sprintf("%d %d * * %s", at.Minute(), at.Hour(), s.day)
		case s.everyMonth:
			spec = fmt.Sprintf("%d %d %s * *", at.Minute(), at.Hour(), s.day)
		default:
			return "", fmt.Errorf("%s schedule without a period", s.choice)
		}
	default:
		return "", fmt.Errorf("unknown schedule type %q", s.choice)
	}
	if _, err := cron.ParseStandard(spec); err != nil {
		return "", fmt.Errorf("invalid cron spec %q: %w", spec, err)
	}
	return spec, nil
}

// parseInterval accepts Go durations and the "1h30" form the v0.1 UI
// suggested, which time.ParseDuration rejects.
func parseInterval(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	d, err := time.ParseDuration(s)
	if err != nil && s != "" && strings.ContainsAny(s[len(s)-1:], "0123456789") && strings.Contains(s, "h") {
		d, err = time.ParseDuration(s + "m")
	}
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid interval %q", s)
	}
	return d, nil
}
