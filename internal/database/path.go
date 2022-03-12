package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func UpdatePathsDB(paths []*ent.Path, repoID int, ctx context.Context) (p []*ent.Path, err error) {
	log.Debug().Msg("updatePaths()")

	db := Connect()
	defer db.Close()

	// Remove old path edges
	_, err = db.Repository.
		UpdateOneID(repoID).RemovePaths(db.Repository.Query().QueryPaths().AllX(ctx)...).Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
		return nil, err
	}

	// Update path edges
	if len(paths) > 0 {
		for _, p := range paths {
			// ID has not been provided if p.ID == 0 -> new path
			path, err := db.Path.
				Create().
				SetAbsPath(p.AbsPath).
				SetIsDir(p.IsDir).
				SetExists(p.Exists).
				SetIncluded(p.Included).
				SetName(p.Name).
				SetRelPath(p.RelPath).
				Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errUpdate)
				return nil, err
			}
			_, err = db.Repository.
				UpdateOneID(repoID).AddPaths(path).Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errUpdate)
				return nil, err
			}
		}
		p, err = QueryPathsByRepositoryDB(repoID, ctx)
	} else {
		p = []*ent.Path{}
	}

	return p, nil
}

func CreatePathsDB(paths []*ent.Path, repoID int, ctx context.Context) (p []*ent.Path, err error) {
	log.Debug().Msg("createPaths()")

	db := Connect()
	defer db.Close()

	if len(paths) > 0 {
		for _, p := range paths {
			path, err := db.Path.
				Create().
				SetAbsPath(p.AbsPath).
				SetIsDir(p.IsDir).
				SetExists(p.Exists).
				SetIncluded(p.Included).
				SetName(p.Name).
				SetRelPath(p.RelPath).
				Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errCreate)
				return nil, err
			}
			_, err = db.Repository.
				UpdateOneID(repoID).AddPaths(path).Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errUpdate)
				return nil, err
			}
		}
		p, err = QueryPathsByRepositoryDB(repoID, ctx)
	} else {
		p = []*ent.Path{}
	}
	return p, err
}
