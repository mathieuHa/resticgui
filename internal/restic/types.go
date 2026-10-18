package restic

import "time"

// The types below decode restic's --json output (restic 0.17 and later).

// Snapshot is one entry of "restic snapshots --json".
type Snapshot struct {
	ID             string           `json:"id"`
	ShortID        string           `json:"short_id"`
	Time           time.Time        `json:"time"`
	Parent         string           `json:"parent,omitempty"`
	Tree           string           `json:"tree"`
	Paths          []string         `json:"paths"`
	Hostname       string           `json:"hostname"`
	Username       string           `json:"username"`
	Tags           []string         `json:"tags,omitempty"`
	ProgramVersion string           `json:"program_version,omitempty"`
	Summary        *SnapshotSummary `json:"summary,omitempty"`
}

// SnapshotSummary is stored in snapshots made by restic 0.17 and later.
type SnapshotSummary struct {
	BackupStart         time.Time `json:"backup_start"`
	BackupEnd           time.Time `json:"backup_end"`
	FilesNew            uint64    `json:"files_new"`
	FilesChanged        uint64    `json:"files_changed"`
	FilesUnmodified     uint64    `json:"files_unmodified"`
	DirsNew             uint64    `json:"dirs_new"`
	DirsChanged         uint64    `json:"dirs_changed"`
	DirsUnmodified      uint64    `json:"dirs_unmodified"`
	DataBlobs           int       `json:"data_blobs"`
	TreeBlobs           int       `json:"tree_blobs"`
	DataAdded           uint64    `json:"data_added"`
	DataAddedPacked     uint64    `json:"data_added_packed"`
	TotalFilesProcessed uint64    `json:"total_files_processed"`
	TotalBytesProcessed uint64    `json:"total_bytes_processed"`
}

// BackupStatus is a progress message of "restic backup --json".
type BackupStatus struct {
	SecondsElapsed   uint64   `json:"seconds_elapsed"`
	SecondsRemaining uint64   `json:"seconds_remaining"`
	PercentDone      float64  `json:"percent_done"`
	TotalFiles       uint64   `json:"total_files"`
	FilesDone        uint64   `json:"files_done"`
	TotalBytes       uint64   `json:"total_bytes"`
	BytesDone        uint64   `json:"bytes_done"`
	ErrorCount       uint     `json:"error_count"`
	CurrentFiles     []string `json:"current_files,omitempty"`
}

// BackupSummary is the final message of "restic backup --json".
type BackupSummary struct {
	SnapshotSummary
	TotalDuration float64 `json:"total_duration"`
	// SnapshotID is empty for a dry run.
	SnapshotID string `json:"snapshot_id"`
}

// BackupError is a file restic could not back up.
type BackupError struct {
	Message string `json:"message"`
	During  string `json:"during"`
	Item    string `json:"item"`
}

// BackupResult is what Backup returns: the summary, plus the files that
// could not be read when the error is ErrIncomplete.
type BackupResult struct {
	Summary *BackupSummary
	Errors  []BackupError
}

// RestoreStatus is a progress message of "restic restore --json".
type RestoreStatus struct {
	SecondsElapsed   uint64  `json:"seconds_elapsed"`
	SecondsRemaining uint64  `json:"seconds_remaining"`
	PercentDone      float64 `json:"percent_done"`
	TotalFiles       uint64  `json:"total_files"`
	FilesRestored    uint64  `json:"files_restored"`
	FilesSkipped     uint64  `json:"files_skipped"`
	TotalBytes       uint64  `json:"total_bytes"`
	BytesRestored    uint64  `json:"bytes_restored"`
	BytesSkipped     uint64  `json:"bytes_skipped"`
}

// RestoreSummary is the final message of "restic restore --json".
type RestoreSummary = RestoreStatus

// ForgetGroup is one group of "restic forget --json".
type ForgetGroup struct {
	Host    string       `json:"host"`
	Tags    []string     `json:"tags"`
	Paths   []string     `json:"paths"`
	Keep    []Snapshot   `json:"keep"`
	Remove  []Snapshot   `json:"remove"`
	Reasons []KeepReason `json:"reasons"`
}

// KeepReason says which rules kept a snapshot.
type KeepReason struct {
	Snapshot Snapshot `json:"snapshot"`
	Matches  []string `json:"matches"`
}

// Node is a file or directory listed by "restic ls --json".
type Node struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	Size        uint64    `json:"size"`
	Mode        uint32    `json:"mode"`
	Permissions string    `json:"permissions"`
	MTime       time.Time `json:"mtime"`
	UID         uint32    `json:"uid"`
	GID         uint32    `json:"gid"`
}

// FindMatch is a file found by "restic find --json".
type FindMatch struct {
	Path        string    `json:"path"`
	Type        string    `json:"type"`
	Size        uint64    `json:"size"`
	Permissions string    `json:"permissions"`
	MTime       time.Time `json:"mtime"`
	User        string    `json:"user"`
	Group       string    `json:"group"`
}

// FindResult groups the matches of one snapshot.
type FindResult struct {
	Snapshot string      `json:"snapshot"`
	Hits     int         `json:"hits"`
	Matches  []FindMatch `json:"matches"`
}

// Stats is the output of "restic stats --json".
type Stats struct {
	TotalSize              uint64  `json:"total_size"`
	TotalUncompressedSize  uint64  `json:"total_uncompressed_size,omitempty"`
	CompressionRatio       float64 `json:"compression_ratio,omitempty"`
	CompressionSpaceSaving float64 `json:"compression_space_saving,omitempty"`
	TotalFileCount         uint64  `json:"total_file_count,omitempty"`
	TotalBlobCount         uint64  `json:"total_blob_count,omitempty"`
	SnapshotsCount         int     `json:"snapshots_count"`
}

// CheckSummary is the output of "restic check --json".
type CheckSummary struct {
	NumErrors          int      `json:"num_errors"`
	BrokenPacks        []string `json:"broken_packs"`
	SuggestRepairIndex bool     `json:"suggest_repair_index"`
	SuggestPrune       bool     `json:"suggest_prune"`
}

// DiffChange is one path that differs between two snapshots. Modifier is
// "+" (added), "-" (removed), "M" (content changed), "T" (type changed) or
// "U" (metadata changed, only with metadata). Directories end with "/".
type DiffChange struct {
	Path     string `json:"path"`
	Modifier string `json:"modifier"`
}

// DiffCount is what one side of a diff holds that the other does not.
type DiffCount struct {
	Files int    `json:"files"`
	Dirs  int    `json:"dirs"`
	Bytes uint64 `json:"bytes"`
}

// DiffStats is the statistics message closing "restic diff --json".
type DiffStats struct {
	SourceSnapshot string    `json:"source_snapshot"`
	TargetSnapshot string    `json:"target_snapshot"`
	ChangedFiles   int       `json:"changed_files"`
	Added          DiffCount `json:"added"`
	Removed        DiffCount `json:"removed"`
}

// Diff is what changed from one snapshot to another.
type Diff struct {
	Changes []DiffChange `json:"changes"`
	Stats   DiffStats    `json:"stats"`
}
