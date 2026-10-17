package commands

import (
	"errors"
	"fmt"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/rs/zerolog/log"
)

const backupCmd = "backup"
const findCmd = "find"
const dumpCmd = "dump"
const initCmd = "init"
const restoreCmd = "restore"
const snapshotsCmd = "snapshots"
const lsCmd = "ls"
const forgetCmd = "forget"
const pruneCmd = "prune"
const statsCmd = "stats"
const checkCmd = "check"
const jsonFormat = "--json"

func getDefaultEnv() []string {
	return []string{}
}

func GetResticVersion() (string, string, error) {
	var args []string

	args = append(args, "version")
	return runRestic(args, getDefaultEnv(), "")
}

func GetResticHelp() (string, string, error) {
	log.Debug().Msg("GetResticHelp()")

	var args []string

	args = append(args, "--help")
	return runRestic(args, getDefaultEnv(), "")
}

func GetSnapshots(repository *ent.Repository) (string, string, error) {
	log.Debug().Msg("GetSnapshots()")

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, snapshotsCmd)
	args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func InitRepository(repository *ent.Repository) (string, string, error) {
	log.Debug().Msg("InitRepository()")

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, initCmd)
	args = append(args, repoFormat...)
	return runRestic(args, env, "")
}

func ListFilesSnapshot(repository *ent.Repository, version string) (string, string, error) {
	log.Debug().Msg("ListFilesSnapshot()")

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, lsCmd)
	args = append(args, repoFormat...)
	args = append(args, version)
	args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func SearchFilesSnapshot(repository *ent.Repository, version string, pattern string) (string, string, error) {
	log.Debug().Msg("SearchFilesSnapshot()")

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, findCmd)
	args = append(args, pattern)
	args = append(args, "-s")
	args = append(args, version)
	args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func DownloadFilesSnapshot(repository *ent.Repository, version string, path string, tmpPath string) (string, string, error) {
	log.Debug().Msg("DownloadFilesSnapshot()")

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, dumpCmd)
	args = append(args, "-a")
	args = append(args, "zip")
	args = append(args, version)
	args = append(args, path)
	return runRestic(args, env, tmpPath)
}

func RestoreRepository(repository *ent.Repository, path string, version string) (string, string, error) {
	log.Debug().Msg("RestoreRepository()")

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	commonArgs := getRestoreCommonArgs(path, version)
	args = append(args, repoFormat...)
	args = append(args, commonArgs...)
	// does not support json format output
	// args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func getRestoreCommonArgs(path string, version string) []string {
	var args []string
	// todo check target path exists and is writable
	args = append(args, restoreCmd)
	args = append(args, version)
	args = append(args, "--target")
	args = append(args, path)
	return args
}

func getPruneCommonArgs(r *ent.Repository) []string {
	var args []string
	if r.Edges.BackupPolicy.NbKeepLast != 0 {
		args = append(args, "--keep-last")
		args = append(args, fmt.Sprint(r.Edges.BackupPolicy.NbKeepLast))
	}
	if r.Edges.BackupPolicy.NbKeepHours != 0 {
		args = append(args, "--keep-hourly")
		args = append(args, fmt.Sprint(r.Edges.BackupPolicy.NbKeepHours))
	}
	if r.Edges.BackupPolicy.NbKeepDays != 0 {
		args = append(args, "--keep-daily")
		args = append(args, fmt.Sprint(r.Edges.BackupPolicy.NbKeepDays))
	}
	if r.Edges.BackupPolicy.NbKeepWeeks != 0 {
		args = append(args, "--keep-weekly")
		args = append(args, fmt.Sprint(r.Edges.BackupPolicy.NbKeepWeeks))
	}
	if r.Edges.BackupPolicy.NbKeepMonths != 0 {
		args = append(args, "--keep-monthly")
		args = append(args, fmt.Sprint(r.Edges.BackupPolicy.NbKeepMonths))
	}

	return args
}

func ForgetSnapshots(repository *ent.Repository, version string) (string, string, error) {
	log.Debug().Msg("ForgetSnapshots()")

	if repository == nil {
		err := errors.New("repository is nil")
		return "", "", err
	}
	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, forgetCmd)
	args = append(args, repoFormat...)
	args = append(args, version)
	return runRestic(args, env, "")
}

func ForgetSnapshotsPolicy(r *ent.Repository) (string, string, error) {
	log.Debug().Msg("ForgetSnapshotsPolicy()")

	if r == nil {
		err := errors.New("repository is nil")
		return "", "", err
	}

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(r)
	commonArgs := getPruneCommonArgs(r)
	args = append(args, forgetCmd)
	args = append(args, repoFormat...)
	args = append(args, commonArgs...)
	args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func PruneSnaphots(r *ent.Repository) (string, string, error) {
	log.Debug().Msg("PruneSnaphots()")

	if r == nil {
		err := errors.New("repository is nil")
		return "", "", err
	}

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(r)
	args = append(args, pruneCmd)
	args = append(args, repoFormat...)
	// does not support json output
	// args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func BackupRepository(repository *ent.Repository, dryRun bool) (string, string, error) {
	log.Debug().Msg("BackupRepository()")

	if repository == nil {
		err := errors.New("repository is nil")
		return "", "", err
	}

	args, env := getBackupArgs(repository, dryRun)
	return runRestic(args, env, "")
}

func getBackupArgs(repository *ent.Repository, dryRun bool) (args []string, env []string) {
	repoFormat, env, _ := getProviderData(repository)
	commonArgs := getBackupCommonArgs(*repository)
	dryRunParameter := getDryRunParameter(dryRun)

	args = append(args, repoFormat...)
	args = append(args, backupCmd)
	args = append(args, commonArgs...)
	args = append(args, dryRunParameter...)
	args = append(args, jsonFormat)
	return args, env
}

func StatsRepository(repository *ent.Repository, snapshotID string) (string, string, error) {
	log.Debug().Msg("StatsRepository()")

	if repository == nil {
		err := errors.New("repository is nil")
		return "", "", err
	}

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, statsCmd)
	if snapshotID != "" {
		args = append(args, snapshotID)
	}
	args = append(args, jsonFormat)
	return runRestic(args, env, "")
}

func getBackupCommonArgs(repository ent.Repository) []string {
	log.Debug().Msg("getBackupCommonArgs()")

	var args []string

	filesParameter := getIncludedAndExludedFiles(repository)
	maxSizeParameter := getMaxSizeParameter(repository)
	args = append(args, filesParameter...)
	args = append(args, maxSizeParameter...)
	return args
}

func getDryRunParameter(dryRun bool) []string {
	if dryRun {
		return []string{"--dry-run"}
	}
	return nil
}

func getMaxSizeParameter(repository ent.Repository) []string {
	var args []string
	// todo check size is correct
	if repository.MaxSize != "" {
		args = append(args, "--exclude-larger-than")
		args = append(args, repository.MaxSize)
	} else {
		args = []string{}
	}
	return args
}

func getIncludedAndExludedFiles(repository ent.Repository) []string {
	log.Debug().Msg("getIncludedAndExludedFiles()")

	var files []string
	var included []string
	var excluded []string
	log.Debug().Msgf("len(path): %d", len(repository.Edges.Paths))
	for _, path := range repository.Edges.Paths {
		log.Debug().Msgf("path.included: %v", path.Included)
		if path.Included { // change to abspath when front is fixed
			included = append(included, path.RelPath)
		} else {
			excluded = append(excluded, path.RelPath)
		}
	}
	files = append(files, included...)
	if len(excluded) > 0 {
		files = append(files, "--exclude")
		files = append(files, excluded...)
	}
	return files
}

func CheckIntegrity(repository *ent.Repository, full bool, pourcentage float64) (string, string, error) {
	// todo check repository string and pourcentage
	log.Debug().Msg("CheckIntegrity()")

	if repository == nil {
		err := errors.New("repository is nil")
		return "", "", err
	}

	var args []string
	var env []string

	repoFormat, env, _ := getProviderData(repository)
	args = append(args, checkCmd)
	args = append(args, repoFormat...)
	if full {
		args = append(args, "--read-data")
	}
	if pourcentage != 0 {
		args = append(args, "--read-data-subset", fmt.Sprintf("%f", pourcentage))
	}
	return runRestic(args, env, "")
}
