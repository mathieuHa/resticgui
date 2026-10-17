package api

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/mathieuHa/resticgui/ent"
// 	"github.com/mathieuHa/resticgui/internal/database"
// )

// func TestCreateLocalProvider(t *testing.T) {
// 	router := setupRouterTest(t)

// 	const apiURL = "/api/v1/repositories/%v/localprovider"
// 	const httpMethod = "POST"

// 	newLP, _ := json.Marshal(ent.LocalProvider{RepoPath: "newRPath"})
// 	responseBody := bytes.NewBuffer(newLP)

// 	testInvalidFormatID(t, router, httpMethod, apiURL, "create l provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
// 	testUnknownID(t, router, httpMethod, apiURL, "create l provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

// 	t.Run("create l provider with no post data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()
// 		req, _ := http.NewRequest(httpMethod, url, nil)
// 		router.ServeHTTP(w, req)

// 		var httpError HTTPError
// 		httpError.Error = errInvalidRequest
// 		errString, errBind := json.Marshal(httpError)

// 		assert.Nil(t, errBind)
// 		assert.Equal(t, http.StatusBadRequest, w.Code)
// 		assert.JSONEq(t, string(errString), w.Body.String())
// 	})

// 	t.Run("create l provider with valid post data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()

// 		repoPath := "myRepoPath"
// 		obj := ent.LocalProvider{
// 			RepoPath: repoPath,
// 		}
// 		l, _ := json.Marshal(obj)
// 		requestBody := bytes.NewBuffer(l)

// 		req, _ := http.NewRequest(httpMethod, url, requestBody)

// 		router.ServeHTTP(w, req)

// 		var response ent.LocalProvider
// 		err = json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Nil(t, err)
// 		log.Debug().Msg(w.Body.String())
// 		assert.GreaterOrEqual(t, response.ID, 0)
// 		assert.Equal(t, response.RepoPath, repoPath)
// 		assert.Equal(t, http.StatusCreated, w.Code)

// 	})

// }

// func TestUpdateLocalProvider(t *testing.T) {
// 	router := setupRouterTest(t)

// 	const apiURL = "/api/v1/repositories/%v/localprovider"
// 	const httpMethod = "PATCH"

// 	newLP, _ := json.Marshal(ent.LocalProvider{RepoPath: "repoPath"})
// 	responseBody := bytes.NewBuffer(newLP)

// 	testInvalidFormatID(t, router, httpMethod, apiURL, "update l provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
// 	testUnknownID(t, router, httpMethod, apiURL, "update l provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

// 	t.Run("update l provider with no patch data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()
// 		req, _ := http.NewRequest(httpMethod, url, nil)
// 		router.ServeHTTP(w, req)

// 		var httpError HTTPError
// 		httpError.Error = errInvalidRequest
// 		errString, errBind := json.Marshal(httpError)

// 		assert.Nil(t, errBind)
// 		assert.Equal(t, http.StatusBadRequest, w.Code)
// 		assert.JSONEq(t, string(errString), w.Body.String())
// 	})

// 	t.Run("update l provider with valid patch data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		ctx := context.Background()
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()

// 		repoPath := "myRepoPath"
// 		obj := ent.LocalProvider{
// 			RepoPath: repoPath,
// 		}

// 		_, err = database.CreateLocalProviderDB(obj, id, ctx)
// 		assert.Nil(t, err)

// 		repoPath2 := "myRepoPath2"
// 		obj2 := ent.LocalProvider{
// 			RepoPath: repoPath2,
// 		}
// 		l, _ := json.Marshal(obj2)

// 		responseBody := bytes.NewBuffer(l)

// 		req, _ := http.NewRequest(httpMethod, url, responseBody)

// 		router.ServeHTTP(w, req)

// 		var response ent.LocalProvider
// 		err = json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Nil(t, err)
// 		log.Debug().Msg(w.Body.String())
// 		assert.GreaterOrEqual(t, response.ID, 0)
// 		assert.Equal(t, response.RepoPath, repoPath2)
// 		assert.Equal(t, http.StatusOK, w.Code)
// 	})
// }
