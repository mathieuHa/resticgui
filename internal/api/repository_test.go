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

	"github.com/gin-gonic/gin"
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

func testInvalidFormatID(t *testing.T, router *gin.Engine, httpMethod string, apiURL string, name string, body *bytes.Buffer, httpError HTTPError) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check invalid id
		w := httptest.NewRecorder()
		url := fmt.Sprintf(apiURL, IDNotValid)
		var req *http.Request
		if body == nil {
			req, _ = http.NewRequest(httpMethod, url, nil)
		} else {
			req, _ = http.NewRequest(httpMethod, url, body)
		}
		router.ServeHTTP(w, req)

		errString, errBind := json.Marshal(httpError)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})
}

func testUnknownID(t *testing.T, router *gin.Engine, httpMethod string, apiURL string, name string, body *bytes.Buffer, httpError HTTPError) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check invalid id
		url := fmt.Sprintf(apiURL, IDUnknown)
		w := httptest.NewRecorder()
		var req *http.Request
		if body == nil {
			req, _ = http.NewRequest(httpMethod, url, nil)
		} else {
			req, _ = http.NewRequest(httpMethod, url, body)
		}
		router.ServeHTTP(w, req)

		errString, errBind := json.Marshal(httpError)

		assert.Nil(t, errBind)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, string(errString), w.Body.String())
	})
}

func TestGetRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v"
	const httpMethod = "GET"

	testInvalidFormatID(t, router, httpMethod, apiURL, "GET repository Invalid ID Format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "GET repository Unknown ID", nil, HTTPError{Error: errRNotFound})

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
		req, err := http.NewRequest(httpMethod, url, nil)
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
	const httpMethod = "POST"

	t.Run("create repository with no post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		url := apiURL
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(httpMethod, url, nil)
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

		req, _ := http.NewRequest(httpMethod, url, responseBody)

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
	const httpMethod = "PATCH"

	newRName := "newRepo"
	newR, _ := json.Marshal(ent.Repository{Name: newRName})
	responseBody := bytes.NewBuffer(newR)

	testInvalidFormatID(t, router, httpMethod, apiURL, "", responseBody, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "", responseBody, HTTPError{Error: errRNotFound})

	t.Run("update repository with valid id but empty post data", func(t *testing.T) {
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

		req, _ := http.NewRequest(httpMethod, url, responseBody)

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
	const httpMethod = "DELETE"

	testInvalidFormatID(t, router, httpMethod, apiURL, "", nil, HTTPError{Error: errIDNotValid})

	t.Run("delete repository with valid id format non existing", func(t *testing.T) {
		cleanDataBaseTest(t)
		// Check Unknown ID
		url := fmt.Sprintf(apiURL, IDUnknown)
		w := httptest.NewRecorder()
		req, err := http.NewRequest(httpMethod, url, nil)
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
		req, err := http.NewRequest(httpMethod, url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

}

func TestLogsRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/logs"
	const httpMethod = "GET"

	testInvalidFormatID(t, router, httpMethod, apiURL, "get logs with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "get logs with valid id not existing", nil, HTTPError{Error: errRNotFound})

	t.Run("get logs repository with valid id", func(t *testing.T) {
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
		req, _ := http.NewRequest(httpMethod, url, nil)
		router.ServeHTTP(w, req)

		var logsReceived []ent.Logs
		err = json.Unmarshal(w.Body.Bytes(), &logsReceived)
		assert.Nil(t, err)
		assert.Len(t, logsReceived, 2)
		assert.Equal(t, logsReceived[1].Operation, "operation1")
		assert.Equal(t, logsReceived[0].Operation, "operation2")
		assert.Equal(t, logsReceived[1].Stdout, "stdout1")
		assert.Equal(t, logsReceived[0].Stdout, "stdout2")
		assert.Equal(t, logsReceived[1].Stderr, "stderr1")
		assert.Equal(t, logsReceived[0].Stderr, "stderr2")
		assert.Equal(t, logsReceived[1].Error, "")
		assert.Equal(t, logsReceived[0].Error, errors.New("error2").Error())
		assert.Equal(t, http.StatusOK, w.Code)
	})

}

func TestStatsRepository(t *testing.T) {
	// todo
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/stats"
	const httpMethod = "GET"

	testInvalidFormatID(t, router, httpMethod, apiURL, "get stats with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "get stats with valid id not existing", nil, HTTPError{Error: errRNotFound})

	t.Run("get stats repository with valid id", func(t *testing.T) {
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

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		//assert.Equal(t, "", w.Body.String())
	})

}

func TestCheckRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/check"
	const httpMethod = "GET"

	testInvalidFormatID(t, router, httpMethod, apiURL, "get check with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "get check with valid id not existing", nil, HTTPError{Error: errRNotFound})

	t.Run("get check repository with valid id", func(t *testing.T) {
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

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// assert.Equal(t, "", w.Body.String())
	})

}

func TestGetSnapshotsRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/snapshots"
	const httpMethod = "GET"

	testInvalidFormatID(t, router, httpMethod, apiURL, "get snapshots with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "get snapshots with valid id not existing", nil, HTTPError{Error: errRNotFound})

	t.Run("get snapshots repository with valid id", func(t *testing.T) {
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

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// assert.Equal(t, "", w.Body.String())
	})

}

func TestValidRepository(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/valid"
	const httpMethod = "GET"

	testInvalidFormatID(t, router, httpMethod, apiURL, "get valid with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "get valid with valid id not existing", nil, HTTPError{Error: errRNotFound})

	t.Run("get valid repository with valid id", func(t *testing.T) {
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

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Body.String())
	})

}
