package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func TestQueryRepositoriesRoute(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories"

	t.Run("query all repository, empty database", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", apiURL, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})
}

func TestGetRepositoriesRoute(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v"

	t.Run("get repository with invalid id format", func(t *testing.T) {
		// Check invalid id
		id := "aze"
		url := fmt.Sprintf(apiURL, id)
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
		// Check Unknown ID
		id := "1280898"
		url := fmt.Sprintf(apiURL, id)
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
		// Check Valid ID get
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID

		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		var repositoryReceived ent.Repository
		err = json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
		assert.Nil(t, err)
		assert.Equal(t, repositoryReceived.ID, id)
		assert.Equal(t, repositoryReceived.Name, rName)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRemoveRepositoriesRoute(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v"

	t.Run("delete repository with invalid id format", func(t *testing.T) {
		// Check invalid id
		id := "aze"
		url := fmt.Sprintf(apiURL, id)
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
		// Check Unknown ID
		id := "1280898"
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", url, nil)
		assert.Nil(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete repository with existing id", func(t *testing.T) {
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
