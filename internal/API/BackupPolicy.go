package API

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func createBackupPolicy(db *ent.Client, backupPolicy *ent.BackupPolicy, repoID int, ctx context.Context) (err error) {
	log.Debug().Msg("createBackupPolicy()")
	log.Debug().Msgf("test %v", &backupPolicy)
	_, err = db.BackupPolicy.
		Create().
		SetKeepAll(backupPolicy.KeepAll).
		SetName(backupPolicy.Name).
		SetNbLastBackup(backupPolicy.NbLastBackup).
		SetOneByDay(backupPolicy.OneByDay).
		SetOneByHour(backupPolicy.OneByHour).
		SetOneByMonth(backupPolicy.OneByMonth).
		SetRepositoryID(repoID).
		Save(ctx)
	log.Debug().Msg("after db")

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func updateBackupPolicy(db *ent.Client, backupPolicy *ent.BackupPolicy, ctx context.Context) (err error) {
	log.Debug().Msg("updateBackupPolicy()")
	_, err = db.BackupPolicy.
		UpdateOneID(backupPolicy.ID).
		SetKeepAll(backupPolicy.KeepAll).
		SetName(backupPolicy.Name).
		SetNbLastBackup(backupPolicy.NbLastBackup).
		SetOneByDay(backupPolicy.OneByDay).
		SetOneByHour(backupPolicy.OneByHour).
		SetOneByMonth(backupPolicy.OneByMonth).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
	}
	return err

}
