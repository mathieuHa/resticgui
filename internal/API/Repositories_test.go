package API

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
)

func TestQueryRepositoriesRoute(t *testing.T) {
	Utils.TestConfig()
	Database.Migrate()
	db := Database.Connect()
	router := SetupRouter()
	Utils.CleanDatabaseTest(db, t, "repositories")

	const apiURL = "/api/v1/repositories/%v"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", apiURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestGetRepositoriesRoute(t *testing.T) {
	Utils.TestConfig()
	Database.Migrate()
	db := Database.Connect()
	router := SetupRouter()
	Utils.CleanDatabaseTest(db, t, "repositories")

	const apiURL = "/api/v1/repositories/%v"

	t.Run("invalid id", func(t *testing.T) {
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

	t.Run("invalid id", func(t *testing.T) {
		// Check Unknown ID
		id := "1280898"
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		b := "{\n    \"error\": \"repository not found\"\n}"

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, b, w.Body.String())
	})

	t.Run("valid ID", func(t *testing.T) {
		// Check Valid ID get
		repositoryName := "myRepo"
		repository := Models.Repository{Name: repositoryName}
		result := db.Create(&repository)
		assert.Nil(t, result.Error)
		id := repository.ID

		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)

		var repositoryReceived Models.Repository
		err := json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
		assert.Nil(t, err)
		assert.Equal(t, repositoryReceived.ID, id)
		assert.Equal(t, repositoryReceived.Name, repositoryName)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRemoveRepositoriesRoute(t *testing.T) {
	Utils.TestConfig()
	Database.Migrate()
	db := Database.Connect()
	router := SetupRouter()
	Utils.CleanDatabaseTest(db, t, "repositories")

	const apiURL = "/api/v1/repositories/%v"

	// Check invalid id
	id := "aze"
	url := fmt.Sprintf(apiURL, id)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", url, nil)
	router.ServeHTTP(w, req)

	b := "{\n    \"error\": \"id not valid\"\n}"

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, b, w.Body.String())

	// Check Valid ID get
	repositoryName := "myRepo"
	repository := Models.Repository{Name: repositoryName}
	result := db.Create(&repository)
	assert.Nil(t, result.Error)
	id2 := repository.ID

	url = fmt.Sprintf(apiURL, id2)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
