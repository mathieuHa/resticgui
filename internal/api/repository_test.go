package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func TestQueryRepositories(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories"

	t.Run("query all repository, empty database", func(t *testing.T) {
		cleanDataBaseTest(t)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", apiURL, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})
}

func TestGetRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v"

	t.Run("get repository with invalid id format", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check invalid id
		url := fmt.Sprintf(apiURL, IDNotValid)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		var err HTTPError
		err.Error = errIDNotValid
		errString, errBind := json.Marshal(err)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("get repository with valid id format non existing", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check Unknown ID
		url := fmt.Sprintf(apiURL, IDUnknown)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		var err HTTPError
		err.Error = errRNotFound
		errString, errBind := json.Marshal(err)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("get repository with existing id", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check Valid ID get
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID

		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		var repositoryReceived ent.Repository
		err = json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
		assert.Nil(t, err)
		assert.Equal(t, repositoryReceived.ID, id)
		assert.Equal(t, repositoryReceived.Name, rName)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCreateRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories"

	t.Run("create repository with no post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		url := apiURL
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", url, nil)
		router.ServeHTTP(w, req)

		var err HTTPError
		err.Error = errInvalidRequest
		errString, errBind := json.Marshal(err)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("create repository with valid post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		url := apiURL
		w := httptest.NewRecorder()

		rName := "myRepo"
		r, _ := json.Marshal(ent.Repository{Name: rName})
		responseBody := bytes.NewBuffer(r)

		req, _ := http.NewRequest("POST", url, responseBody)

		router.ServeHTTP(w, req)

		var repositoryReceived ent.Repository
		err := json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, repositoryReceived.ID, 0)
		assert.Equal(t, repositoryReceived.Name, rName)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

}

func TestUpdateRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v"

	t.Run("update repository with invalid id format", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check invalid id
		url := fmt.Sprintf(apiURL, IDNotValid)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", url, nil)
		router.ServeHTTP(w, req)

		var err HTTPError
		err.Error = errIDNotValid
		errString, errBind := json.Marshal(err)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("update repository with not existing id", func(t *testing.T) {
		cleanDataBaseTest(t)
		url := fmt.Sprintf(apiURL, IDUnknown)
		w := httptest.NewRecorder()
		rName := "myRepo"
		r, _ := json.Marshal(ent.Repository{Name: rName})
		responseBody := bytes.NewBuffer(r)

		req, _ := http.NewRequest("PATCH", url, responseBody)
		router.ServeHTTP(w, req)

		var err HTTPError
		err.Error = errRNotFound
		errString, errBind := json.Marshal(err)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("update repository with valid id but empty post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)

		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		req, _ := http.NewRequest("PATCH", url, nil)

		router.ServeHTTP(w, req)

		var errHTTP HTTPError
		errHTTP.Error = errInvalidRequest
		errString, errBind := json.Marshal(errHTTP)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("update repository with valid id and valid post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		newRName := "newRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)

		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		newR, _ := json.Marshal(ent.Repository{Name: newRName})
		responseBody := bytes.NewBuffer(newR)

		req, _ := http.NewRequest("PATCH", url, responseBody)

		router.ServeHTTP(w, req)

		var repositoryReceived ent.Repository
		err = json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
		assert.Nil(t, err)
		assert.Equal(t, repositoryReceived.ID, r.ID)
		assert.Equal(t, repositoryReceived.Name, newRName)
		assert.Equal(t, http.StatusOK, w.Code)
	})

}

func TestRemoveRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v"

	t.Run("delete repository with invalid id format", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check invalid id
		url := fmt.Sprintf(apiURL, IDNotValid)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		var errHTTP HTTPError
		errHTTP.Error = errIDNotValid
		errString, errBind := json.Marshal(errHTTP)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("delete repository with valid id format non existing", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check Unknown ID
		url := fmt.Sprintf(apiURL, IDUnknown)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete repository with existing id", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check Valid ID get
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID

		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

}

func TestLogsRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/logs"

	t.Run("get logs repository with invalid id format", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check invalid id
		url := fmt.Sprintf(apiURL, IDNotValid)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		var errHTTP HTTPError
		errHTTP.Error = errIDNotValid
		errString, errBind := json.Marshal(errHTTP)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("get logs repository with valid id format non existing", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check Unknown ID
		url := fmt.Sprintf(apiURL, IDUnknown)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		var err HTTPError
		err.Error = errRNotFound
		errString, errBind := json.Marshal(err)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})

	t.Run("get logs repository with valid id", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)

		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})

	t.Run("get logs repository with valid id and some logs", func(t *testing.T) {
		cleanDataBaseTest(t)
		ctx := context.Background()
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, ctx)
		assert.Nil(t, err)
		database.LogCommand(r.ID, "operation1", "stdout1", "stderr1", nil, ctx)
		database.LogCommand(r.ID, "operation2", "stdout2", "stderr2", errors.New("error2"), ctx)

		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		var logsReceived []ent.Logs
		err = json.Unmarshal(w.Body.Bytes(), &logsReceived)
		assert.Nil(t, err)
		assert.Len(t, logsReceived, 2)
		assert.Equal(t, logsReceived[0].Operation, "operation1")
		assert.Equal(t, logsReceived[1].Operation, "operation2")
		assert.Equal(t, logsReceived[0].Stdout, "stdout1")
		assert.Equal(t, logsReceived[1].Stdout, "stdout2")
		assert.Equal(t, logsReceived[0].Stderr, "stderr1")
		assert.Equal(t, logsReceived[1].Stderr, "stderr2")
		assert.Equal(t, logsReceived[0].Error, "")
		assert.Equal(t, logsReceived[1].Error, errors.New("error2").Error())
		assert.Equal(t, http.StatusOK, w.Code)
	})

}
