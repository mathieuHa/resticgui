package Commands

import (
	"fmt"

	"gitlab.com/resticgui/resticgui/internal/Models"
)

var maxSizeFile string

func GetResticVersion() string {
	return runRestic("version")
}

func GetResticHelp() string {
	return runRestic("--help")
}

func GetLocalSnapshots(repository string) string {
	args := fmt.Sprintf("-r %v snapshots", repository)
	return runRestic(args)
}

func InitLocalRepository(repository string) string {

	repoFormat := getLocalRepositoryFormat(repository)
	args := fmt.Sprintf("init -r %v", repoFormat)
	return runRestic(args)
}

func BackupLocalRepository(repository Models.Repository) string {

	repoFormat := getLocalRepositoryFormat(repository.Local)
	commonArgs := getBackupCommonArgs(repository)
	args := fmt.Sprintf("-r %v backup %v", repoFormat, commonArgs)
	return runRestic(args)
}

func getBackupCommonArgs(repository Models.Repository) string {
	files := getIncludedAndExludedFiles(repository)
	maxSize := getMaxSizeParameter(repository)
	return files + maxSize

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
		files += fmt.Sprintf(" -- exclude %v", excluded)
	}
	return files
}

func InitAmazonS3Repository(bucketName string) string {
	// todo Add env var AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY and optional AWS_DEFAULT_REGION
	repoFormat := getAmazonS3RepositoryFormat(bucketName)
	args := fmt.Sprintf("init -r %v", repoFormat)
	return runRestic(args)
}

func InitMinioS3Repository(bucketName string, minioPort int, minioIP string, minioScheme string) string {
	// todo Add env var AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	// add ca cert option
	repoFormat := getMinioS3RepositoryFormat(bucketName, minioPort, minioIP, minioScheme)
	args := fmt.Sprintf("init -r %v", repoFormat)
	return runRestic(args)
}

func InitB2Repository(bucketName string, repository string) string {
	// todo Add env var AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	// add ca cert option
	repoFormat := getB2RepositoryFormat(bucketName, repository)
	args := fmt.Sprintf("init -r %v", repoFormat)
	return runRestic(args)
}

func getLocalRepositoryFormat(repository string) string {
	// todo check path exists and is writable
	// todo check repository string
	return repository
}

func getB2RepositoryFormat(bucketName string, repository string) string {
	// todo check bucketName, repository
	repoFormat := fmt.Sprintf("b2:%v:%v", bucketName, repository)
	return repoFormat
}

func getAmazonS3RepositoryFormat(bucketName string) string {
	// todo check bucketName
	repoFormat := fmt.Sprintf("s3:s3.amazonaws.com/%v", bucketName)
	return repoFormat
}

func getMinioS3RepositoryFormat(bucketName string, minioPort int, minioIP string, minioScheme string) string {
	// todo check bucketName, minioPort, minioIP, minioScheme
	repoFormat := fmt.Sprintf("s3:%v://%v:%d/%v", minioScheme, minioIP, minioPort, bucketName)
	return repoFormat
}

func BackupRepository(bucketName string, repository string) string {
	// todo Add env var AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	// todo check bucketName, repository
	// add ca cert option
	args := fmt.Sprintf("init -r b2:%v:%v", bucketName, repository)
	return runRestic(args)
}

func CheckIntegrity(repository string, full bool, pourcentage float64) string {
	// todo check repository string and pourcentage
	args := fmt.Sprintf("-r %v check", repository)
	if full {
		args += " --read-data"
	}
	if pourcentage != 0 {
		args += fmt.Sprintf(" --read-data-subset %f", pourcentage)
	}
	return runRestic(args)
}
