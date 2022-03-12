package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/logs"
	"gitlab.com/resticgui/resticgui/ent/repository"
)

func GetRepositoriesDB(ctx context.Context) (repositories []*ent.Repository, err error) {

	db := Connect()
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

	db := Connect()
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
		log.Err(err).Msgf("error: %v", errCreate)
	}
	return r, err
}

func CreateRepositoryDB(repo ent.Repository, ctx context.Context) (r *ent.Repository, err error) {
	db := Connect()
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

func UpdateRepositoryDB(repo ent.Repository, id int, ctx context.Context) (r *ent.Repository, err error) {
	db := Connect()
	defer db.Close()

	// create repository
	r, err = db.Repository.
		UpdateOneID(id).
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

func DeleteRepositoryDB(id int, ctx context.Context) (err error) {
	db := Connect()
	defer db.Close()

	// create repository
	err = db.Repository.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return err
}

func QuerylogsByRepositoryDB(id int, ctx context.Context) (l []*ent.Logs, err error) {
	db := Connect()
	defer db.Close()

	// create repository
	l, err = db.Logs.Query().Where(logs.HasRepositoryWith(repository.ID(id))).All(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return l, err
}
