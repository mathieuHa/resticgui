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

func TestCreatePaths(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/paths"
	const httpMethod = "POST"

	newBP, _ := json.Marshal(ent.Path{Name: "MyPath"})
	requestBody := bytes.NewBuffer(newBP)

	testInvalidFormatID(t, router, httpMethod, apiURL, "create paths with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "create paths with valid id not existing", requestBody, HTTPError{Error: errRNotFound})

	t.Run("create paths with no post data", func(t *testing.T) {
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

	t.Run("create paths with valid post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		name := "myBMName"
		obj := ent.Path{
			Name: name,
		}
		l, _ := json.Marshal(obj)
		requestBody := bytes.NewBuffer(l)

		req, _ := http.NewRequest(httpMethod, url, requestBody)

		router.ServeHTTP(w, req)

		var response ent.Path
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, response.Name, name)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

}

func TestUpdatePaths(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/paths"
	const httpMethod = "PATCH"

	var myPaths []ent.Path
	myPaths = append(myPaths, ent.Path{Name: "myBackup"})

	newLP, _ := json.Marshal(myPaths)
	requestBody := bytes.NewBuffer(newLP)

	testInvalidFormatID(t, router, httpMethod, apiURL, "update paths with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "update paths with valid id not existing", requestBody, HTTPError{Error: errRNotFound})

	t.Run("update paths with no patch data", func(t *testing.T) {
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

	t.Run("update paths with valid patch data", func(t *testing.T) {
		cleanDataBaseTest(t)
		ctx := context.Background()
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		var obj []*ent.Path
		name := "myBPName"
		absPath := "myAbsPath"
		obj = append(obj, &ent.Path{Name: name, AbsPath: absPath})

		_, err = database.CreatePathsDB(obj, id, ctx)
		assert.Nil(t, err)

		var myPaths []ent.Path
		name2 := "myBPName2"
		absPath2 := "myAbsPath2"
		myPaths = append(myPaths, ent.Path{Name: name2, AbsPath: absPath2})

		newLP, _ := json.Marshal(myPaths)
		requestBody := bytes.NewBuffer(newLP)

		req, _ := http.NewRequest(httpMethod, url, requestBody)

		router.ServeHTTP(w, req)

		var response []ent.Path
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.Equal(t, len(response), len(obj))
		//assert.Equal(t, response.Name, name)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
