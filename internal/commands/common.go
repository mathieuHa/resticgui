package commands

import (
	"fmt"

	"gitlab.com/resticgui/resticgui/internal/Models"
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

func GetResticVersion() string {
	var args []string

	args = append(args, "version")
	return runRestic(args, getDefaultEnv())
}

func GetResticHelp() string {
	var args []string

	args = append(args, "--help")
	return runRestic(args, getDefaultEnv())
}

func GetSnapshots(repository Models.Repository) string {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, snapshotsCmd)
	return runRestic(args, env)
}

func InitRepository(repository Models.Repository) string {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, initCmd)
	args = append(args, repoFormat...)
	return runRestic(args, env)
}

func RestoreRepository(repository Models.Repository, path string, version string) string {
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

func ForgetSnaphots(repository Models.Repository, version string) string {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, forgetCmd)
	args = append(args, repoFormat...)
	args = append(args, version)
	return runRestic(args, env)
}

func PruneSnaphots(repository Models.Repository, version string) string {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, pruneCmd)
	args = append(args, repoFormat...)
	args = append(args, version)
	return runRestic(args, env)
}

func BackupRepository(repository Models.Repository, dryRun bool) string {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	commonArgs := getBackupCommonArgs(repository, dryRun)
	args = append(args, repoFormat...)
	args = append(args, commonArgs...)
	args = append(args, backupCmd)
	return runRestic(args, env)
}

func StatsRepository(repository Models.Repository, dryRun bool) string {
	var args []string
	var env []string

	repoFormat, env := getProviderData(repository)
	args = append(args, repoFormat...)
	args = append(args, statsCmd)
	return runRestic(args, env)
}

func getBackupCommonArgs(repository Models.Repository, dryRun bool) []string {
	var args []string

	dryRunParameter := getDryRunParameter(dryRun)
	filesParameter := getIncludedAndExludedFiles(repository)
	maxSizeParameter := getMaxSizeParameter(repository)
	args = append(args, filesParameter...)
	args = append(args, maxSizeParameter...)
	args = append(args, dryRunParameter)
	return args
}

func getDryRunParameter(dryRun bool) string {
	if dryRun {
		return "--dry-run"
	} else {
		return ""
	}
}

func getMaxSizeParameter(repository Models.Repository) []string {
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

func getIncludedAndExludedFiles(repository Models.Repository) []string {
	var files []string
	var included []string
	var excluded []string
	for _, path := range repository.Paths {
		if path.Included {
			included = append(included, path.AbsPath)
		} else {
			excluded = append(excluded, path.AbsPath)
		}
	}
	files = append(files, included...)
	if len(excluded) > 0 {
		files = append(files, "--exclude")
		files = append(files, excluded...)
	}
	return files
}

func CheckIntegrity(repository Models.Repository, full bool, pourcentage float64) string {
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
