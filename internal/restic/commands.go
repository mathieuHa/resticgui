package restic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// MinVersion is the oldest restic release ResticGUI supports: it introduced
// the exit codes 10, 11 and 12 and the JSON messages decoded here.
const MinVersion = "0.17.0"

var versionPattern = regexp.MustCompile(`restic (\d+\.\d+\.\d+)`)

// Version returns the restic version, e.g. "0.18.1".
func (r *Runner) Version(ctx context.Context) (string, error) {
	var out bytes.Buffer
	if err := r.run(ctx, nil, call{command: "version", args: []string{"--json"}, stdout: &out}); err == nil {
		var v struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(out.Bytes(), &v) == nil && v.Version != "" {
			return v.Version, nil
		}
	}
	// restic before 0.17 has no "version --json"
	out.Reset()
	if err := r.run(ctx, nil, call{command: "version", stdout: &out}); err != nil {
		return "", err
	}
	m := versionPattern.FindStringSubmatch(out.String())
	if m == nil {
		return "", fmt.Errorf("restic: cannot parse version from %q", strings.TrimSpace(out.String()))
	}
	return m[1], nil
}

// CheckVersion returns an error when restic is missing or older than MinVersion.
func (r *Runner) CheckVersion(ctx context.Context) (string, error) {
	v, err := r.Version(ctx)
	if err != nil {
		return "", err
	}
	if compareVersions(v, MinVersion) < 0 {
		return v, fmt.Errorf("restic %s is too old, ResticGUI needs %s or later", v, MinVersion)
	}
	return v, nil
}

func compareVersions(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Init creates the repository.
func (r *Runner) Init(ctx context.Context, repo *Repo) error {
	return r.run(ctx, repo, call{command: "init", args: []string{"--json"}, stdout: io.Discard})
}

// Exists reports whether a repository is initialized at repo.Location. It
// fails with ErrWrongPassword when the repository exists but the password
// does not open it.
func (r *Runner) Exists(ctx context.Context, repo *Repo) (bool, error) {
	err := r.run(ctx, repo, call{command: "cat", args: []string{"config", "--json"}, stdout: io.Discard})
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrRepositoryNotFound):
		return false, nil
	}
	return false, err
}

// Unlock removes stale locks.
func (r *Runner) Unlock(ctx context.Context, repo *Repo) error {
	return r.run(ctx, repo, call{command: "unlock", args: []string{"--json"}, stdout: io.Discard})
}

// SnapshotFilter narrows Snapshots and Forget to some snapshots.
type SnapshotFilter struct {
	Host string
	// Tags keeps snapshots having all of these tags.
	Tags []string
}

func (f SnapshotFilter) args() []string {
	var args []string
	if f.Host != "" {
		args = append(args, "--host", f.Host)
	}
	if len(f.Tags) > 0 {
		args = append(args, "--tag", strings.Join(f.Tags, ","))
	}
	return args
}

// Snapshots lists snapshots, oldest first.
func (r *Runner) Snapshots(ctx context.Context, repo *Repo, filter SnapshotFilter) ([]Snapshot, error) {
	var out bytes.Buffer
	args := append([]string{"--json"}, filter.args()...)
	if err := r.run(ctx, repo, call{command: "snapshots", args: args, stdout: &out}); err != nil {
		return nil, err
	}
	snapshots := []Snapshot{}
	if err := json.Unmarshal(out.Bytes(), &snapshots); err != nil {
		return nil, fmt.Errorf("restic snapshots: decode output: %w", err)
	}
	return snapshots, nil
}

// BackupOptions configures Backup.
type BackupOptions struct {
	// Paths to back up, passed verbatim through a file (no globbing).
	Paths []string
	// Excludes are restic exclude patterns: "*" and "?" are wildcards.
	Excludes          []string
	Tags              []string
	Host              string
	ExcludeLargerThan string
	// Compression is "auto", "off" or "max"; empty keeps restic's default.
	Compression string
	// OneFileSystem does not cross into other file systems under the paths.
	OneFileSystem bool
	// ExcludeCaches skips folders holding a CACHEDIR.TAG file.
	ExcludeCaches bool
	// ExcludeIfPresent skips folders holding a file of one of these names.
	ExcludeIfPresent []string
	// LimitUpload and LimitDownload cap the transfer rate in KiB/s; 0 for none.
	LimitUpload   int
	LimitDownload int
	DryRun        bool
}

// Backup runs a backup and calls onProgress for each status message.
// When some files could not be read, it returns the result together with an
// error matching ErrIncomplete: the snapshot exists but is partial.
func (r *Runner) Backup(ctx context.Context, repo *Repo, opts BackupOptions, onProgress func(BackupStatus)) (*BackupResult, error) {
	if len(opts.Paths) == 0 {
		return nil, errors.New("restic backup: no paths to back up")
	}
	dir, err := os.MkdirTemp("", "resticgui-backup-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	filesFrom := filepath.Join(dir, "files-from")
	if err := os.WriteFile(filesFrom, []byte(strings.Join(opts.Paths, "\n")+"\n"), 0o600); err != nil {
		return nil, err
	}
	args := []string{"--json", "--files-from-verbatim", filesFrom}
	if len(opts.Excludes) > 0 {
		excludeFile := filepath.Join(dir, "exclude")
		if err := os.WriteFile(excludeFile, []byte(strings.Join(opts.Excludes, "\n")+"\n"), 0o600); err != nil {
			return nil, err
		}
		args = append(args, "--exclude-file", excludeFile)
	}
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}
	if opts.Host != "" {
		args = append(args, "--host", opts.Host)
	}
	if opts.ExcludeLargerThan != "" {
		args = append(args, "--exclude-larger-than", opts.ExcludeLargerThan)
	}
	if opts.Compression != "" {
		args = append(args, "--compression", opts.Compression)
	}
	if opts.OneFileSystem {
		args = append(args, "--one-file-system")
	}
	if opts.ExcludeCaches {
		args = append(args, "--exclude-caches")
	}
	for _, name := range opts.ExcludeIfPresent {
		// one argument, so that a name starting with "-" stays a name
		args = append(args, "--exclude-if-present="+name)
	}
	if opts.LimitUpload > 0 {
		args = append(args, "--limit-upload", strconv.Itoa(opts.LimitUpload))
	}
	if opts.LimitDownload > 0 {
		args = append(args, "--limit-download", strconv.Itoa(opts.LimitDownload))
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	result := &BackupResult{}
	var decodeErr error
	err = r.run(ctx, repo, call{command: "backup", args: args, onLine: func(line []byte) {
		switch messageType(line) {
		case "status":
			var s BackupStatus
			if json.Unmarshal(line, &s) == nil && onProgress != nil {
				onProgress(s)
			}
		case "summary":
			var s BackupSummary
			if err := json.Unmarshal(line, &s); err != nil {
				decodeErr = err
				return
			}
			result.Summary = &s
		case "error":
			var e struct {
				Error  BackupError `json:"error"`
				During string      `json:"during"`
				Item   string      `json:"item"`
			}
			if json.Unmarshal(line, &e) == nil {
				e.Error.During, e.Error.Item = e.During, e.Item
				result.Errors = append(result.Errors, e.Error)
			}
		}
	}})
	if err != nil {
		if errors.Is(err, ErrIncomplete) && result.Summary != nil {
			return result, err
		}
		return nil, err
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("restic backup: decode summary: %w", decodeErr)
	}
	if result.Summary == nil {
		return nil, errors.New("restic backup: no summary in output")
	}
	return result, nil
}

// Policy is a retention policy for Forget.
type Policy struct {
	KeepLast, KeepHourly, KeepDaily, KeepWeekly, KeepMonthly, KeepYearly int
	// KeepWithin is a restic duration such as "1y6m" or "30d".
	KeepWithin string
}

// IsEmpty reports whether the policy would keep nothing explicitly, in which
// case restic refuses to forget anything.
func (p Policy) IsEmpty() bool {
	return p == Policy{}
}

func (p Policy) args() []string {
	var args []string
	for _, rule := range []struct {
		flag string
		n    int
	}{
		{"--keep-last", p.KeepLast}, {"--keep-hourly", p.KeepHourly}, {"--keep-daily", p.KeepDaily},
		{"--keep-weekly", p.KeepWeekly}, {"--keep-monthly", p.KeepMonthly}, {"--keep-yearly", p.KeepYearly},
	} {
		if rule.n > 0 {
			args = append(args, rule.flag, strconv.Itoa(rule.n))
		}
	}
	if p.KeepWithin != "" {
		args = append(args, "--keep-within", p.KeepWithin)
	}
	return args
}

// ForgetOptions configures Forget.
type ForgetOptions struct {
	Policy Policy
	// Filter limits which snapshots the policy applies to.
	Filter SnapshotFilter
	// GroupBy is restic's --group-by, e.g. "host,tags". Empty keeps restic's
	// default ("host,paths"), which starts a new group whenever paths change.
	GroupBy string
	// Prune removes the data no longer referenced once snapshots are forgotten.
	Prune  bool
	DryRun bool
}

// ErrEmptyPolicy is returned by Forget for a policy without any rule.
var ErrEmptyPolicy = errors.New("restic forget: the retention policy has no rule")

// Forget applies a retention policy and returns what was kept and removed.
func (r *Runner) Forget(ctx context.Context, repo *Repo, opts ForgetOptions) ([]ForgetGroup, error) {
	if opts.Policy.IsEmpty() {
		return nil, ErrEmptyPolicy
	}
	args := append([]string{"--json"}, opts.Policy.args()...)
	args = append(args, opts.Filter.args()...)
	if opts.GroupBy != "" {
		args = append(args, "--group-by", opts.GroupBy)
	}
	if opts.Prune {
		args = append(args, "--prune")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	var out bytes.Buffer
	if err := r.run(ctx, repo, call{command: "forget", args: args, stdout: &out}); err != nil {
		return nil, err
	}
	// with --prune, restic prints prune's text report after the JSON
	groups := []ForgetGroup{}
	if err := json.NewDecoder(&out).Decode(&groups); err != nil && err != io.EOF {
		return nil, fmt.Errorf("restic forget: decode output: %w", err)
	}
	return groups, nil
}

// ForgetSnapshots removes the given snapshots.
func (r *Runner) ForgetSnapshots(ctx context.Context, repo *Repo, ids []string, prune bool) error {
	if len(ids) == 0 {
		return errors.New("restic forget: no snapshot given")
	}
	args := []string{"--json"}
	if prune {
		args = append(args, "--prune")
	}
	args = append(args, "--")
	args = append(args, ids...)
	// restic exits 0 when an ID matches no snapshot, after printing
	// `Ignoring "ID": no matching ID found for prefix "ID"`.
	var ignored []string
	err := r.run(ctx, repo, call{command: "forget", args: args, stdout: io.Discard, onStderr: func(line string) {
		if m := ignoredSnapshot.FindStringSubmatch(line); m != nil {
			ignored = append(ignored, m[1])
		}
	}})
	if err == nil && len(ignored) > 0 {
		return &SnapshotNotFoundError{IDs: ignored}
	}
	return err
}

var ignoredSnapshot = regexp.MustCompile(`^Ignoring "(.*)": no matching ID found`)

// SnapshotNotFoundError lists the IDs ForgetSnapshots could not find. The
// other snapshots were forgotten.
type SnapshotNotFoundError struct {
	IDs []string
}

func (e *SnapshotNotFoundError) Error() string {
	return "restic forget: no snapshot matches " + strings.Join(e.IDs, ", ")
}

// Prune removes unreferenced data.
func (r *Runner) Prune(ctx context.Context, repo *Repo) error {
	return r.run(ctx, repo, call{command: "prune", stdout: io.Discard})
}

// CheckOptions configures Check.
type CheckOptions struct {
	// ReadDataSubset reads part of the data, e.g. "10%" or "1/5"; "100%" reads everything.
	ReadDataSubset string
}

// Check verifies the repository. Errors found are reported in the summary
// and as an error.
func (r *Runner) Check(ctx context.Context, repo *Repo, opts CheckOptions) (*CheckSummary, error) {
	args := []string{"--json"}
	if opts.ReadDataSubset != "" {
		args = append(args, "--read-data-subset", opts.ReadDataSubset)
	}
	var summary *CheckSummary
	err := r.run(ctx, repo, call{command: "check", args: args, onLine: func(line []byte) {
		if messageType(line) == "summary" {
			var s CheckSummary
			if json.Unmarshal(line, &s) == nil {
				summary = &s
			}
		}
	}})
	return summary, err
}

// RestoreOptions configures Restore.
type RestoreOptions struct {
	// SnapshotID is a snapshot ID or "latest".
	SnapshotID string
	Target     string
	// Include restricts the restore to these paths inside the snapshot.
	Include []string
	DryRun  bool
}

// Restore restores a snapshot and calls onProgress for each status message.
func (r *Runner) Restore(ctx context.Context, repo *Repo, opts RestoreOptions, onProgress func(RestoreStatus)) (*RestoreSummary, error) {
	if opts.SnapshotID == "" || opts.Target == "" {
		return nil, errors.New("restic restore: snapshot and target are required")
	}
	args := []string{"--json", "--target", opts.Target}
	for _, p := range opts.Include {
		args = append(args, "--include", p)
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--", opts.SnapshotID)

	var summary *RestoreSummary
	err := r.run(ctx, repo, call{command: "restore", args: args, onLine: func(line []byte) {
		var s RestoreStatus
		switch messageType(line) {
		case "status":
			if json.Unmarshal(line, &s) == nil && onProgress != nil {
				onProgress(s)
			}
		case "summary":
			if json.Unmarshal(line, &s) == nil {
				summary = &s
			}
		}
	}})
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, errors.New("restic restore: no summary in output")
	}
	return summary, nil
}

// Ls lists the entries of dir in a snapshot (the whole snapshot when dir is
// empty).
func (r *Runner) Ls(ctx context.Context, repo *Repo, snapshotID, dir string) ([]Node, error) {
	args := []string{"--json", "--", snapshotID}
	if dir != "" {
		args = append(args, dir)
	}
	nodes := []Node{}
	err := r.run(ctx, repo, call{command: "ls", args: args, onLine: func(line []byte) {
		// the first line describes the snapshot itself
		if messageType(line) == "node" || bytes.Contains(line, []byte(`"struct_type":"node"`)) {
			var n Node
			if json.Unmarshal(line, &n) == nil {
				nodes = append(nodes, n)
			}
		}
	}})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// Find searches files matching pattern in one snapshot ("latest" works).
func (r *Runner) Find(ctx context.Context, repo *Repo, snapshotID, pattern string) ([]FindResult, error) {
	var out bytes.Buffer
	args := []string{"--json", "--snapshot", snapshotID, "--", pattern}
	if err := r.run(ctx, repo, call{command: "find", args: args, stdout: &out}); err != nil {
		return nil, err
	}
	results := []FindResult{}
	if err := json.Unmarshal(out.Bytes(), &results); err != nil {
		return nil, fmt.Errorf("restic find: decode output: %w", err)
	}
	return results, nil
}

// Stats returns repository statistics. Mode is restic's --mode
// ("restore-size", "files-by-contents", "raw-data", "blobs-per-file");
// snapshotID limits it to one snapshot when set.
func (r *Runner) Stats(ctx context.Context, repo *Repo, mode, snapshotID string) (*Stats, error) {
	args := []string{"--json"}
	if mode != "" {
		args = append(args, "--mode", mode)
	}
	if snapshotID != "" {
		args = append(args, "--", snapshotID)
	}
	var out bytes.Buffer
	if err := r.run(ctx, repo, call{command: "stats", args: args, stdout: &out}); err != nil {
		return nil, err
	}
	var s Stats
	if err := json.Unmarshal(out.Bytes(), &s); err != nil {
		return nil, fmt.Errorf("restic stats: decode output: %w", err)
	}
	return &s, nil
}

// Diff lists what changed from snapshot from to snapshot to.
func (r *Runner) Diff(ctx context.Context, repo *Repo, from, to string) (*Diff, error) {
	d := &Diff{Changes: []DiffChange{}}
	var decodeErr error
	err := r.run(ctx, repo, call{command: "diff", args: []string{"--json", "--", from, to}, onLine: func(line []byte) {
		var err error
		switch messageType(line) {
		case "change":
			var c DiffChange
			if err = json.Unmarshal(line, &c); err == nil {
				d.Changes = append(d.Changes, c)
			}
		case "statistics":
			err = json.Unmarshal(line, &d.Stats)
		}
		if err != nil && decodeErr == nil {
			decodeErr = err
		}
	}})
	if err != nil {
		return nil, err
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("restic diff: decode output: %w", decodeErr)
	}
	return d, nil
}

// Dump writes a file, or a directory as an archive ("tar" or "zip"), from a
// snapshot to w.
func (r *Runner) Dump(ctx context.Context, repo *Repo, snapshotID, path, archive string, w io.Writer) error {
	args := []string{}
	if archive != "" {
		args = append(args, "--archive", archive)
	}
	args = append(args, "--", snapshotID, path)
	return r.run(ctx, repo, call{command: "dump", args: args, stdout: w})
}

func messageType(line []byte) string {
	var m struct {
		Type string `json:"message_type"`
	}
	if len(line) == 0 || line[0] != '{' || json.Unmarshal(line, &m) != nil {
		return ""
	}
	return m.Type
}
