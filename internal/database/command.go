package database

import (
	"context"

	"github.com/rs/zerolog/log"
)

func LogCommand(id int, operation string, stdout string, stderr string, err error, ctx context.Context) {

	var errStr string
	if err == nil {
		errStr = ""
	} else {
		errStr = err.Error()
		log.Err(err).Msg(operation)
	}
	_, errDB := db.Logs.
		Create().
		SetOperation(operation).
		SetStdout(stdout).
		SetStderr(stderr).
		SetError(errStr).
		SetRepositoryID(id).
		Save(ctx)

	if errDB != nil {
		log.Err(errDB)
	}
}
