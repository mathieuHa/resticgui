package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	ta "github.com/stretchr/testify/assert"
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

		b := "{\n    \"error\": \"id not valid\"\n}"

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, b, w.Body.String())
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

		ta.Nil(t, errBind)
		ta.Equal(t, http.StatusNotFound, w.Code)
		ta.JSONEq(t, string(errString), w.Body.String())
	})

	// t.Run("valid ID", func(t *testing.T) {
	// 	// Check Valid ID get
	// 	repositoryName := "myRepo"
	// 	repository := Models.Repository{Name: repositoryName}
	// 	result := db.Create(&repository)
	// 	assert.Nil(t, result.Error)
	// 	id := repository.ID

	// 	url := fmt.Sprintf(apiURL, id)
	// 	w := httptest.NewRecorder()
	// 	req, _ := http.NewRequest("GET", url, nil)
	// 	router.ServeHTTP(w, req)

	// 	var repositoryReceived Models.Repository
	// 	err := json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
	// 	assert.Nil(t, err)
	// 	assert.Equal(t, repositoryReceived.ID, id)
	// 	assert.Equal(t, repositoryReceived.Name, repositoryName)
	// 	assert.Equal(t, http.StatusOK, w.Code)
	// })
}

// func TestRemoveRepositoriesRoute(t *testing.T) {
// 	Utils.TestConfig()
// 	Database.Migrate()
// 	db := Database.Connect()
// 	router := SetupRouter()
// 	Utils.CleanDatabaseTest(db, t, "repositories")

// 	const apiURL = "/api/v1/repositories/%v"

// 	// Check invalid id
// 	id := "aze"
// 	url := fmt.Sprintf(apiURL, id)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("DELETE", url, nil)
// 	router.ServeHTTP(w, req)

// 	b := "{\n    \"error\": \"id not valid\"\n}"

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Equal(t, b, w.Body.String())

// 	// Check Valid ID get
// 	repositoryName := "myRepo"
// 	repository := Models.Repository{Name: repositoryName}
// 	result := db.Create(&repository)
// 	assert.Nil(t, result.Error)
// 	id2 := repository.ID

// 	url = fmt.Sprintf(apiURL, id2)
// 	w = httptest.NewRecorder()
// 	req, _ = http.NewRequest("DELETE", url, nil)
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusNoContent, w.Code)
// }
