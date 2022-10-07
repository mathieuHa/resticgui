package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func TestCreateBackupPolicy(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/backuppolicy"
	const httpMethod = "POST"

	newBP, _ := json.Marshal(ent.BackupPolicy{})
	responseBody := bytes.NewBuffer(newBP)

	testInvalidFormatID(t, router, httpMethod, apiURL, "create backup policy with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "create backup policy with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

	t.Run("create backup policy with no post data", func(t *testing.T) {
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

	t.Run("create backup policy with valid post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		obj := ent.BackupPolicy{NbLastBackup: 1}
		l, _ := json.Marshal(obj)
		requestBody := bytes.NewBuffer(l)

		req, _ := http.NewRequest(httpMethod, url, requestBody)

		router.ServeHTTP(w, req)

		var response ent.BackupPolicy
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

}

func TestUpdateBackupPolicy(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/backuppolicy"
	const httpMethod = "PATCH"

	newLP, _ := json.Marshal(ent.BackupPolicy{})
	responseBody := bytes.NewBuffer(newLP)

	testInvalidFormatID(t, router, httpMethod, apiURL, "update backup policy with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "update backup policy with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

	t.Run("update backup policy with no patch data", func(t *testing.T) {
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

	t.Run("update backup policy with valid patch data", func(t *testing.T) {
		cleanDataBaseTest(t)
		ctx := context.Background()
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		obj := ent.BackupPolicy{}

		_, err = database.CreateBackupPolicyDB(obj, id, ctx)
		assert.Nil(t, err)

		obj2 := ent.BackupPolicy{}
		l, _ := json.Marshal(obj2)

		responseBody := bytes.NewBuffer(l)

		req, _ := http.NewRequest(httpMethod, url, responseBody)

		router.ServeHTTP(w, req)

		var response ent.BackupPolicy
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
