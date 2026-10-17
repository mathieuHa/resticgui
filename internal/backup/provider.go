package backup

import (
	"errors"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/rs/zerolog/log"
)

func validProvider(repository *ent.Repository) (valid bool, err error) {
	log.Debug().Msg("validProvider()")

	if repository == nil {
		valid = false
		err = errors.New("repository is nil")
		return valid, err
	}
	log.Debug().Msgf("Provider = %v", repository.Provider)
	switch repository.Provider {
	case "localProvider":
		valid, err = validLocalProvider(repository.Edges.LocalProvider)
	case "MinioS3Provider":
		valid, err = validMinioS3Provider(repository.Edges.MinioS3Provider)
	case "AWSS3Provider":
		valid, err = validAWSS3Provider(repository.Edges.AWSS3Provider)
	case "B2Provider":
		valid, err = validB2Provider(repository.Edges.B2Provider)
	default:
		valid = false
		err = errors.New("provider does not exist")
	}
	if err != nil {
		log.Err(err).Msg("Error in validation of provider")
	}
	return valid, err
}
