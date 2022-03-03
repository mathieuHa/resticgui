package API

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func createBackupPolicy(db *ent.Client, backupPolicy ent.BackupPolicy, repoID int, ctx context.Context) {
	_, err := db.BackupPolicy.
		Create().
		SetKeepAll(backupPolicy.KeepAll).
		SetName(backupPolicy.Name).
		SetNbLastBackup(backupPolicy.NbLastBackup).
		SetOneByDay(backupPolicy.OneByDay).
		SetOneByHour(backupPolicy.OneByHour).
		SetOneByMonth(backupPolicy.OneByMonth).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
}

func updateBackupPolicy(db *ent.Client, backupPolicy ent.BackupPolicy, ctx context.Context) {
	_, err := db.BackupPolicy.
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
		// return
	}

}
