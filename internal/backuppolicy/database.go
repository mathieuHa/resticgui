package backuppolicy

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func CreateBackupPolicyDB(backupPolicy ent.BackupPolicy, repoID int, ctx context.Context) (b *ent.BackupPolicy, err error) {
	log.Debug().Msg("CreateBackupPolicyDB()")

	db := Database.Connect()
	defer db.Close()

	b, err = db.BackupPolicy.
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
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return b, err
}

func UpdateBackupPolicyDB(backupPolicy ent.BackupPolicy, ctx context.Context) (b *ent.BackupPolicy, err error) {
	log.Debug().Msg("updateBackupPolicyDB()")

	db := Database.Connect()
	defer db.Close()

	b, err = db.BackupPolicy.
		UpdateOneID(backupPolicy.ID).
		SetKeepAll(backupPolicy.KeepAll).
		SetName(backupPolicy.Name).
		SetNbLastBackup(backupPolicy.NbLastBackup).
		SetOneByDay(backupPolicy.OneByDay).
		SetOneByHour(backupPolicy.OneByHour).
		SetOneByMonth(backupPolicy.OneByMonth).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}
	return b, err
}
