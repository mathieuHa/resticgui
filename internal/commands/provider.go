package commands

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

// todo check each provider
func getProviderData(repository *ent.Repository) (args []string, env []string, err error) {
	log.Debug().Msg("getProviderData()")

	if repository == nil {
		log.Debug().Msg("repository == nil")
		return args, env, errors.New("test")
	} else {
		log.Debug().Msg("repository != nil")
	}

	if repository.Password == "" {
		log.Debug().Msg("password == ''")
		return args, env, errors.New("test")
	}

	args = append(args, "-r")
	envResticPass := fmt.Sprintf("RESTIC_PASSWORD=%v", repository.Password)
	env = append(env, envResticPass)
	switch repository.Provider {
	case "localProvider":
		log.Debug().Msg("repository.Provider == localProvider")
		args = append(args, getLocalRepositoryFormat(*repository.Edges.LocalProvider))
	case "B2Provider":
		log.Debug().Msg("repository.Provider == B2Provider")
		args = append(args, getB2RepositoryFormat(*repository.Edges.B2Provider))
		env = append(args, getB2EnvFormat(*repository.Edges.B2Provider)...)
	case "AWSS3Provider":
		log.Debug().Msg("repository.Provider == AWSS3Provider")
		args = append(args, getAmazonS3RepositoryFormat(*repository.Edges.AWSS3Provider))
		env = append(args, getAmazonS3EnvFormat(*repository.Edges.AWSS3Provider)...)
	case "MinioS3Provider":
		log.Debug().Msg("repository.Provider == MinioS3Provider")
		args = append(args, getMinioS3RepositoryFormat(*repository.Edges.MinioS3Provider))
		env = append(args, getMinioS3EnvFormat(*repository.Edges.MinioS3Provider)...)
	default:
		log.Warn().Msg("No Provider Found")
		args = append(args, getLocalRepositoryFormat(*repository.Edges.LocalProvider))
	}
	return args, env, err

}

func getLocalRepositoryFormat(local ent.LocalProvider) string {
	log.Debug().Msg("getLocalRepositoryFormat()")
	// todo check path exists and is writable
	// todo check repository string
	return local.RepoPath
}

func getB2RepositoryFormat(b2 ent.B2Provider) string {
	log.Debug().Msg("getB2RepositoryFormat()")

	// todo check bucketName, repository
	repoFormat := fmt.Sprintf("b2:%v:%v", b2.BucketName, b2.RepoName)
	return repoFormat
}

func getAmazonS3RepositoryFormat(aws ent.AWSS3Provider) string {
	// todo check bucketName
	repoFormat := fmt.Sprintf("s3:s3.amazonaws.com/%v", aws.BucketName)
	return repoFormat
}

func getMinioS3RepositoryFormat(minio ent.MinioS3Provider) string {
	// todo check bucketName, minioPort, minioIP, minioScheme
	repoFormat := fmt.Sprintf("s3:%v://%v:%v/%v", minio.Scheme, minio.IP, minio.Port, minio.BucketName)
	return repoFormat
}

func getB2EnvFormat(b2 ent.B2Provider) []string {
	var env []string

	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", b2.AccountID))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", b2.AccountKey))

	return env
}

func getAmazonS3EnvFormat(aws ent.AWSS3Provider) []string {
	var env []string

	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", aws.AccessKeyID))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", aws.SecretAccessKey))
	if aws.DefaultRegion != "" {
		env = append(env, fmt.Sprintf("AWS_DEFAULT_REGION=%v", aws.DefaultRegion))
	}

	return env
}

func getMinioS3EnvFormat(minio ent.MinioS3Provider) []string {
	var env []string

	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", minio.AccessKeyID))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", minio.SecretAccessKey))

	return env
}
