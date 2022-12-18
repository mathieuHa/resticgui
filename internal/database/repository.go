package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/logs"
	"gitlab.com/resticgui/resticgui/ent/path"
	"gitlab.com/resticgui/resticgui/ent/repository"
)

func GetRepositoriesDB(ctx context.Context) (repositories []*ent.Repository, err error) {

	repositories, err = db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithPruneSchedule().
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
	log.Debug().Msg("GetRepositoryByIDDB()")

	r, err = db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithPruneSchedule().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		Where(repository.ID(id)).
		Only(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", err)
	}
	return r, err
}

func CreateRepositoryDB(repo ent.Repository, ctx context.Context) (r *ent.Repository, err error) {
	log.Debug().Msg("CreateRepositoryDB()")

	if repo.Provider == "" {
		repo.Provider = "localProvider"
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
	log.Debug().Msg("UpdateRepositoryDB()")

	if repo.Provider == "" {
		repo.Provider = "localProvider"
	}
	// update repository
	r, err = db.Repository.
		UpdateOneID(id).
		SetName(repo.Name).
		SetActive(repo.Active).
		SetRunning(false).
		SetPassword(repo.Password).
		SetProvider(repo.Provider).
		SetMaxSize(repo.MaxSize).
		SetRepositorySize(repo.RepositorySize).
		SetRepositoryFileCount(repo.RepositoryFileCount).
		SetSnapshotCount(repo.SnapshotCount).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
		return
	}

	return r, err
}

func SetRepositoryInitiatedDB(id int, ctx context.Context) (r *ent.Repository, err error) {
	log.Debug().Msg("SetRepositoryInitiatedDB()")

	// update repository
	r, err = db.Repository.
		UpdateOneID(id).
		SetInitiated(true).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
		return r, err
	}

	return r, err
}

func SetRepositoryRunningDB(id int, running bool, ctx context.Context) (r *ent.Repository, err error) {
	log.Debug().Msg("SetRepositoryRunningDB()")

	// update repository set running
	r, err = db.Repository.
		UpdateOneID(id).
		SetRunning(running).
		Save(ctx)

	log.Debug().Msgf("running %v", r.Running)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
		return r, err
	}

	return r, err
}

func SetRepositoryValidDB(value bool, id int, ctx context.Context) (r *ent.Repository, err error) {
	log.Debug().Msg("SetRepositoryValidDB()")

	// update repository
	r, err = db.Repository.
		UpdateOneID(id).
		SetValid(value).
		Save(ctx)

	if err != nil {
		log.Err(err)
		return r, err
	}

	return r, err
}

func DeleteRepositoryDB(id int, ctx context.Context) (err error) {

	// delete repository
	err = db.Repository.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return err
}

func QueryLogsByRepositoryDB(id int, limit int, page int, ctx context.Context) (l []*ent.Logs, err error) {

	offset := page * limit
	// query logs by repository
	l, err = db.Logs.Query().Where(logs.HasRepositoryWith(repository.ID(id))).Order(ent.Desc(logs.FieldStartTime)).Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", "error Query Logs")
	}

	return l, err
}

func QueryPathsByRepositoryDB(id int, ctx context.Context) (p []*ent.Path, err error) {

	// query path by repository
	p, err = db.Path.Query().Where(path.HasRepositoryWith(repository.ID(id))).All(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return p, err
}
