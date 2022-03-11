package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func updateValidProvider(repository ent.Repository, valid bool) (err error) {
	log.Debug().Msg("updateValidProvider()")
	ctx := context.Background()

	db := Database.Connect()
	defer db.Close()

	switch repository.Provider {
	case "localProvider":
		if repository.Edges.LocalProvider != nil {
			_, err = db.LocalProvider.UpdateOneID(repository.Edges.LocalProvider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	case "minioS3Provider":
		if repository.Edges.MinioS3Provider != nil {
			_, err = db.MinioS3Provider.UpdateOneID(repository.Edges.MinioS3Provider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	case "AWSS3Provider":
		if repository.Edges.AWSS3Provider != nil {
			_, err = db.AWSS3Provider.UpdateOneID(repository.Edges.AWSS3Provider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	case "B2Provider":
		if repository.Edges.B2Provider != nil {
			_, err = db.B2Provider.UpdateOneID(repository.Edges.B2Provider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	default:
		err = errors.New("provider does not exist")
	}
	if err != nil {
		log.Err(err).Msg("Could not update valid for repository")
	}

	return err
}
