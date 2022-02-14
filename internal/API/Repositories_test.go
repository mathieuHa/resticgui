package API

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog/log"
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

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/repositories", nil)
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

	// Check invalid id
	id := "aze"
	url := fmt.Sprintf("/api/v1/repositories/%s", id)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)

	b := "{\n    \"error\": \"id not valid\"\n}"

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, b, w.Body.String())

	// Check Unknown ID
	id = "1280898"
	url = fmt.Sprintf("/api/v1/repositories/%s", id)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)

	b = "{\n    \"error\": \"repository not found\"\n}"

	// Check Valid ID get
	repositoryName := "myRepo"
	repository := Models.Repository{Name: repositoryName}
	result := db.Create(&repository)
	assert.Nil(t, result.Error)
	id2 := repository.ID
	log.Debug().Msgf("%v", id2)

	url = fmt.Sprintf("/api/v1/repositories/%v", id2)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)

	log.Debug().Msg(w.Body.String())
	var repositoryReceived Models.Repository
	err := json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
	log.Debug().Err(err)
	assert.Nil(t, err)
	assert.Equal(t, repositoryReceived.ID, id2)
	assert.Equal(t, repositoryReceived.Name, repositoryName)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRemoveRepositoriesRoute(t *testing.T) {
	Utils.TestConfig()
	Database.Migrate()
	db := Database.Connect()
	router := SetupRouter()
	Utils.CleanDatabaseTest(db, t, "repositories")

	// Check invalid id
	id := "aze"
	url := fmt.Sprintf("/api/v1/repositories/%s", id)
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
	log.Debug().Msgf("%v", id2)

	url = fmt.Sprintf("/api/v1/repositories/%v", id2)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
