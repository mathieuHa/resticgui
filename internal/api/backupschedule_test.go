package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/mathieuHa/resticgui/ent/backupschedule"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestCreateBackupSchedule(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/backupschedule"
	const httpMethod = "POST"

	newLP, _ := json.Marshal(ent.BackupSchedule{Choice: backupschedule.ChoiceCron})
	responseBody := bytes.NewBuffer(newLP)

	testInvalidFormatID(t, router, httpMethod, apiURL, "create backup schedule with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "create backup schedule with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

	t.Run("create backup schedule with no post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(httpMethod, url, nil)
		router.ServeHTTP(w, req)

		var httpError HTTPError
		httpError.Error = errInvalidRequest
		errString, errBind := json.Marshal(httpError)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("create backup schedule with valid post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		choice := backupschedule.Choice("cron")
		obj := ent.BackupSchedule{
			Choice: choice,
		}
		l, _ := json.Marshal(obj)
		requestBody := bytes.NewBuffer(l)

		req, _ := http.NewRequest(httpMethod, url, requestBody)

		router.ServeHTTP(w, req)

		var response ent.BackupSchedule
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, response.Choice, choice)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

}

func TestUpdateBackupSchedule(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/backupschedule"
	const httpMethod = "PATCH"

	newLP, _ := json.Marshal(ent.BackupSchedule{Choice: backupschedule.Choice("cron")})
	responseBody := bytes.NewBuffer(newLP)

	testInvalidFormatID(t, router, httpMethod, apiURL, "update backup schedule with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "update backup schedule with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

	t.Run("update backup schedule with no patch data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(httpMethod, url, nil)
		router.ServeHTTP(w, req)

		var httpError HTTPError
		httpError.Error = errInvalidRequest
		errString, errBind := json.Marshal(httpError)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("update backup schedule with valid patch data", func(t *testing.T) {
		cleanDataBaseTest(t)
		ctx := context.Background()
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		choice := backupschedule.Choice("cron")
		obj := ent.BackupSchedule{
			Choice: backupschedule.Choice(choice),
		}

		_, err = database.CreateBackupScheduleDB(obj, id, ctx)
		assert.Nil(t, err)

		choice2 := backupschedule.Choice("interval")
		obj2 := ent.BackupSchedule{
			Choice: backupschedule.Choice(choice2),
		}
		l, _ := json.Marshal(obj2)

		requestBody := bytes.NewBuffer(l)

		req, _ := http.NewRequest(httpMethod, url, requestBody)

		router.ServeHTTP(w, req)

		var response ent.BackupSchedule
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, response.Choice, choice2)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
