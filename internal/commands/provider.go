package commands

import (
	"fmt"

	"gitlab.com/resticgui/resticgui/internal/Models"
)

// todo check each provider
func getProviderData(repository Models.Repository) ([]string, []string) {
	var args []string
	var env []string

	args = append(args, "-r")
	env = append(env, repository.Password)
	switch repository.Provider {
	case "local":
		args = append(args, getLocalRepositoryFormat(repository.LocalProvider))
	case "b2":
		args = append(args, getB2RepositoryFormat(repository.B2Provider))
		env = append(args, getB2EnvFormat(repository.B2Provider)...)
	case "aws":
		args = append(args, getAmazonS3RepositoryFormat(repository.AWSS3Provider))
		env = append(args, getAmazonS3EnvFormat(repository.AWSS3Provider)...)
	case "minio":
		args = append(args, getMinioS3RepositoryFormat(repository.MinioS3Provider))
		env = append(args, getMinioS3EnvFormat(repository.MinioS3Provider)...)
	default:
		args = append(args, getLocalRepositoryFormat(repository.LocalProvider))
	}
	return args, env

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

func getB2EnvFormat(b2 Models.B2Provider) []string {
	var env []string

	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", b2.AccountID))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", b2.AccountKey))

	return env
}

func getAmazonS3EnvFormat(aws Models.AWSS3Provider) []string {
	var env []string

	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", aws.AccessKeyID))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", aws.SecretAccessKey))
	if aws.DefaultRegion != "" {
		env = append(env, fmt.Sprintf("AWS_DEFAULT_REGION=%v", aws.DefaultRegion))
	}

	return env
}

func getMinioS3EnvFormat(minio Models.MinioS3Provider) []string {
	var env []string

	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", minio.AccessKeyID))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", minio.SecretAccessKey))

	return env
}
