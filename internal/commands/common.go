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

func GetResticVersion() string {
	var args []string

	args = append(args, "version")
	return runRestic(args)
}

func GetResticHelp() string {
	var args []string

	args = append(args, "--help")
	return runRestic(args)
}

func GetSnapshots(repository Models.Repository) string {
	var args []string

	repoFormat := getProviderFormat(repository)
	args = append(args, repoFormat, snapshotsCmd)
	return runRestic(args)
}

func InitRepository(repository Models.Repository) string {
	var args []string

	repoFormat := getProviderFormat(repository)
	args = append(args, initCmd, repoFormat)
	return runRestic(args)
}

func RestoreRepository(repository Models.Repository, path string, version string) string {
	var args []string

	repoFormat := getProviderFormat(repository)
	commonArgs := getRestoreCommonArgs(path, version)
	args = append(args, repoFormat, commonArgs)
	return runRestic(args)
}

func getRestoreCommonArgs(path string, version string) string {
	// todo check path exists and is writable
	return fmt.Sprintf("%v %v --target %v", restoreCmd, version, path)
}

func ForgetSnaphots(repository Models.Repository, version string) string {
	var args []string
	repoFormat := getProviderFormat(repository)
	args = append(args, forgetCmd, repoFormat, version)
	return runRestic(args)
}

func PruneSnaphots(repository Models.Repository, version string) string {
	var args []string
	repoFormat := getProviderFormat(repository)
	args = append(args, pruneCmd, repoFormat, version)
	return runRestic(args)
}

// todo check each provider
func getProviderFormat(repository Models.Repository) string {
	switch repository.Provider {
	case "local":
		return "-r " + getLocalRepositoryFormat(repository.LocalProvider)
	case "b2":
		return "-r " + getB2RepositoryFormat(repository.B2Provider)
	case "aws":
		return "-r " + getAmazonS3RepositoryFormat(repository.AWSS3Provider)
	case "minio":
		return "-r " + getMinioS3RepositoryFormat(repository.MinioS3Provider)
	default:
		return "-r " + getLocalRepositoryFormat(repository.LocalProvider)
	}
}

func BackupRepository(repository Models.Repository, dryRun bool) string {
	var args []string
	repoFormat := getProviderFormat(repository)
	commonArgs := getBackupCommonArgs(repository, dryRun)
	args = append(args, repoFormat)
	args = append(args, commonArgs...)
	args = append(args, backupCmd)
	return runRestic(args)
}

func StatsRepository(repository Models.Repository, dryRun bool) string {
	var args []string
	repoFormat := getProviderFormat(repository)
	args = append(args, repoFormat, statsCmd)
	return runRestic(args)
}

func getBackupCommonArgs(repository Models.Repository, dryRun bool) []string {
	var args []string
	dryRunParameter := getDryRunParameter(dryRun)
	filesParameter := getIncludedAndExludedFiles(repository)
	maxSizeParameter := getMaxSizeParameter(repository)
	args = append(args, filesParameter, maxSizeParameter, dryRunParameter)
	return args
}

func getDryRunParameter(dryRun bool) string {
	if dryRun {
		return "--dry-run"
	} else {
		return ""
	}
}

func getMaxSizeParameter(repository Models.Repository) string {
	// todo check size is correct
	if repository.MaxSize != "" {
		return repository.MaxSize
	} else {
		return ""
	}
}

func getIncludedAndExludedFiles(repository Models.Repository) string {
	var files string
	var included string
	var excluded string
	for _, path := range repository.Paths {
		if path.Included {
			included += path.AbsPath
		} else {
			excluded += path.AbsPath
		}
	}
	files += fmt.Sprintf("%v", included)
	if len(excluded) > 0 {
		files += fmt.Sprintf(" -- exclude %v ", excluded)
	}
	return files
}

func getLocalRepositoryFormat(local Models.LocalProvider) string {
	// todo check path exists and is writable
	// todo check repository string
	return local.RepoName
}

func getB2RepositoryFormat(b2 Models.B2Provider) string {
	// todo check bucketName, repository
	repoFormat := fmt.Sprintf("b2:%v:%v", b2.BucketName, b2.RepoName)
	return repoFormat
}

func getAmazonS3RepositoryFormat(aws Models.AWSS3Provider) string {
	// todo check bucketName
	repoFormat := fmt.Sprintf("s3:s3.amazonaws.com/%v", aws.BucketName)
	return repoFormat
}

func getMinioS3RepositoryFormat(minio Models.MinioS3Provider) string {
	// todo check bucketName, minioPort, minioIP, minioScheme
	repoFormat := fmt.Sprintf("s3:%v://%v:%v/%v", minio.Scheme, minio.IP, minio.Port, minio.BucketName)
	return repoFormat
}

func CheckIntegrity(repository Models.Repository, full bool, pourcentage float64) string {
	// todo check repository string and pourcentage
	var args []string

	repoFormat := getProviderFormat(repository)
	args = append(args, checkCmd, repoFormat)
	if full {
		args = append(args, "--read-data")
	}
	if pourcentage != 0 {
		args = append(args, "--read-data-subset", fmt.Sprintf("%f", pourcentage))
	}
	return runRestic(args)
}
