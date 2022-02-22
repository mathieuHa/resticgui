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

// todo check each provider
func getProviderData(repository Models.Repository) ([]string, []string) {
	var args []string
	var env []string

	args = append(args, "-r")
	switch repository.Provider {
	case "local":
		args = append(args, getLocalRepositoryFormat(repository.LocalProvider))
		env = append(args, getLocalEnvFormat(repository.LocalProvider))
	case "b2":
		args = append(args, getB2RepositoryFormat(repository.B2Provider))
		env = append(args, getB2EnvFormat(repository.B2Provider))
	case "aws":
		args = append(args, getAmazonS3RepositoryFormat(repository.AWSS3Provider))
		env = append(args, getAmazonS3EnvFormat(repository.AWSS3Provider))
	case "minio":
		args = append(args, getMinioS3RepositoryFormat(repository.MinioS3Provider))
		env = append(args, getMinioS3EnvFormat(repository.MinioS3Provider))
	default:
		args = append(args, getLocalRepositoryFormat(repository.LocalProvider))
		env = append(args, getLocalEnvFormat(repository.LocalProvider))
	}
	return args, env

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

func getLocalEnvFormat(local Models.LocalProvider) string {
	// todo check path exists and is writable
	// todo check repository string
	return local.RepoName
}

func getB2EnvFormat(b2 Models.B2Provider) string {
	// todo check bucketName, repository
	repoFormat := fmt.Sprintf("b2:%v:%v", b2.BucketName, b2.RepoName)
	return repoFormat
}

func getAmazonS3EnvFormat(aws Models.AWSS3Provider) string {
	// todo check bucketName
	repoFormat := fmt.Sprintf("s3:s3.amazonaws.com/%v", aws.BucketName)
	return repoFormat
}

func getMinioS3EnvFormat(minio Models.MinioS3Provider) string {
	// todo check bucketName, minioPort, minioIP, minioScheme
	repoFormat := fmt.Sprintf("s3:%v://%v:%v/%v", minio.Scheme, minio.IP, minio.Port, minio.BucketName)
	return repoFormat
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
