package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

const backupCmd = "backup"
const initCmd = "init"
const restoreCmd = "restore"
const snapshotsCmd = "snapshots"
const forgetCmd = "forget"
const pruneCmd = "prune"
const statsCmd = "stats"
const checkCmd = "check"

func getDefaultEnv() []string {
	return []string{}
}

func GetResticVersion() (string, string, error) {
	var args []string

	args = append(args, "version")
	return runRestic(args, getDefaultEnv())
}

func GetResticHelp() (string, string, error) {
	var args []string

	args = append(args, "--help")
	return runRestic(args, getDefaultEnv())
}

func GetSnapshots(repository ent.Repository) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, snapshotsCmd)
	return runRestic(args, env)
}

func InitRepository(repository ent.Repository) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, initCmd)
	args = append(args, repoFormat...)
	return runRestic(args, env)
}

func RestoreRepository(repository ent.Repository, path string, version string) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	commonArgs := getRestoreCommonArgs(path, version)
	args = append(args, repoFormat...)
	args = append(args, commonArgs...)
	return runRestic(args, env)
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

func ForgetSnaphots(repository ent.Repository, version string) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, forgetCmd)
	args = append(args, repoFormat...)
	args = append(args, version)
	return runRestic(args, env)
}

func PruneSnaphots(repository ent.Repository, version string) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, pruneCmd)
	args = append(args, repoFormat...)
	args = append(args, version)
	return runRestic(args, env)
}

func BackupRepository(repository ent.Repository, dryRun bool) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	commonArgs := getBackupCommonArgs(repository)
	dryRunParameter := getDryRunParameter(dryRun)

	args = append(args, repoFormat...)
	args = append(args, backupCmd)
	args = append(args, commonArgs...)
	args = append(args, dryRunParameter)

	return runRestic(args, env)
}

func StatsRepository(repository ent.Repository, dryRun bool) (string, string, error) {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, statsCmd)
	return runRestic(args, env)
}

func getBackupCommonArgs(repository ent.Repository) []string {
	var args []string

	filesParameter := getIncludedAndExludedFiles(repository)
	maxSizeParameter := getMaxSizeParameter(repository)
	args = append(args, filesParameter...)
	args = append(args, maxSizeParameter...)
	return args
}

func getDryRunParameter(dryRun bool) string {
	if dryRun {
		return "--dry-run"
	} else {
		return ""
	}
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
	log.Debug().Msgf("len(path): %v", repository.Edges.Paths)
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

func CheckIntegrity(repository ent.Repository, full bool, pourcentage float64) (string, string, error) {
	// todo check repository string and pourcentage
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, checkCmd)
	args = append(args, repoFormat...)
	if full {
		args = append(args, "--read-data")
	}
	if pourcentage != 0 {
		args = append(args, "--read-data-subset", fmt.Sprintf("%f", pourcentage))
	}
	return runRestic(args, env)
}
