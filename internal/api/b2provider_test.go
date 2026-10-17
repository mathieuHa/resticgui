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

// func TestCreateB2Provider(t *testing.T) {
// 	router := setupRouterTest(t)

// 	const apiURL = "/api/v1/repositories/%v/b2provider"
// 	const httpMethod = "POST"

// 	newB2, _ := json.Marshal(ent.B2Provider{BucketName: "newRName"})
// 	responseBody := bytes.NewBuffer(newB2)

// 	testInvalidFormatID(t, router, httpMethod, apiURL, "create b2 provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
// 	testUnknownID(t, router, httpMethod, apiURL, "create b2 provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

// 	t.Run("create b2 provider with no post data", func(t *testing.T) {
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

// 	t.Run("create b2 provider with valid post data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()

// 		bucketName := "myBucket"
// 		repoName := "myRepo"
// 		accountKey := "myAccountKey"
// 		accountID := "myAccountID"
// 		obj := ent.B2Provider{
// 			BucketName: bucketName,
// 			AccountKey: accountKey,
// 			AccountID:  accountID,
// 			RepoName:   repoName,
// 		}
// 		b2, _ := json.Marshal(obj)
// 		requestBody := bytes.NewBuffer(b2)

// 		req, _ := http.NewRequest(httpMethod, url, requestBody)

// 		router.ServeHTTP(w, req)

// 		var response ent.B2Provider
// 		err = json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Nil(t, err)
// 		log.Debug().Msg(w.Body.String())
// 		assert.GreaterOrEqual(t, response.ID, 0)
// 		assert.Equal(t, response.BucketName, bucketName)
// 		assert.Equal(t, response.AccountID, accountID)
// 		assert.Equal(t, response.AccountKey, accountKey)
// 		assert.Equal(t, http.StatusCreated, w.Code)

// 	})

// }

// func TestUpdateB2Provider(t *testing.T) {
// 	router := setupRouterTest(t)

// 	const apiURL = "/api/v1/repositories/%v/b2provider"
// 	const httpMethod = "PATCH"

// 	newB2, _ := json.Marshal(ent.B2Provider{BucketName: "newRName"})
// 	responseBody := bytes.NewBuffer(newB2)

// 	testInvalidFormatID(t, router, httpMethod, apiURL, "update b2 provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
// 	testUnknownID(t, router, httpMethod, apiURL, "update b2 provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

// 	t.Run("update b2 provider with no patch data", func(t *testing.T) {
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

// 	t.Run("update b2 provider with valid patch data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		ctx := context.Background()
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()

// 		bucketName := "myBucket"
// 		accountKey := "myAccountKey"
// 		accountID := "myAccountID"
// 		repoName := "myRepo"
// 		obj := ent.B2Provider{
// 			BucketName: bucketName,
// 			AccountKey: accountKey,
// 			AccountID:  accountID,
// 			RepoName:   repoName,
// 		}

// 		_, err = database.CreateB2ProviderDB(obj, id, ctx)
// 		assert.Nil(t, err)

// 		bucketName2 := "myBucket2"
// 		accountKey2 := "myAccountKey2"
// 		accountID2 := "myAccountID2"
// 		repoName2 := "myRepo2"

// 		obj2 := ent.B2Provider{
// 			BucketName: bucketName2,
// 			AccountKey: accountKey2,
// 			AccountID:  accountID2,
// 			RepoName:   repoName2,
// 		}
// 		b2, _ := json.Marshal(obj2)

// 		responseBody := bytes.NewBuffer(b2)

// 		req, _ := http.NewRequest(httpMethod, url, responseBody)

// 		router.ServeHTTP(w, req)

// 		var response ent.B2Provider
// 		err = json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Nil(t, err)
// 		log.Debug().Msg(w.Body.String())
// 		assert.GreaterOrEqual(t, response.ID, 0)
// 		assert.Equal(t, response.BucketName, bucketName2)
// 		assert.Equal(t, response.AccountID, accountID2)
// 		assert.Equal(t, response.AccountKey, accountKey2)
// 		assert.Equal(t, response.RepoName, repoName2)
// 		assert.Equal(t, http.StatusOK, w.Code)

// 	})

// }
