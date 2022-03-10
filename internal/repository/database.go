package repository

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/repository"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func GetRepositoriesDB(ctx context.Context) (repositories []*ent.Repository, err error) {

	db := Database.Connect()
	defer db.Close()

	repositories, err = db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		All(ctx)

	if err != nil {
		log.Err(err).Msgf("error")
	}

	return repositories, err
}

func GetRepositoryByIDDB(id int, ctx context.Context) (r *ent.Repository, err error) {

	db := Database.Connect()
	defer db.Close()

	r, err = db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		Where(repository.ID(id)).
		Only(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errNotFound)
	}
	return r, err
}

func CreateRepositoryDB(repo ent.Repository, ctx context.Context) (r *ent.Repository, err error) {
	db := Database.Connect()
	defer db.Close()

	if r.Provider == "" {
		r.Provider = "localProvider"
	}

	// create repository
	r, err = db.Repository.
		Create().
		SetName(repo.Name).
		SetActive(repo.Active).
		SetRunning(false).
		SetPassword(repo.Password).
		SetProvider(repo.Provider).
		SetMaxSize(repo.MaxSize).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
		return
	}

	return r, err
}
