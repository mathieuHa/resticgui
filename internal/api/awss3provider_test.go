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
// 	"gitlab.com/resticgui/resticgui/ent"
// 	"gitlab.com/resticgui/resticgui/internal/database"
// )

// func TestCreateAWSS3Provider(t *testing.T) {
// 	router := setupRouterTest(t)

// 	const apiURL = "/api/v1/repositories/%v/awss3provider"
// 	const httpMethod = "POST"

// 	newAWS, _ := json.Marshal(ent.AWSS3Provider{BucketName: "newRName"})
// 	responseBody := bytes.NewBuffer(newAWS)

// 	testInvalidFormatID(t, router, httpMethod, apiURL, "create awss3 provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
// 	testUnknownID(t, router, httpMethod, apiURL, "create awss3 provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

// 	t.Run("create awss3 provider with no post data", func(t *testing.T) {
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

// 	t.Run("create awss3 provider with valid post data", func(t *testing.T) {
// 		cleanDataBaseTest(t)
// 		rName := "myRepo"
// 		repository := ent.Repository{Name: rName}
// 		r, err := database.CreateRepositoryDB(repository, context.Background())
// 		assert.Nil(t, err)
// 		id := r.ID
// 		url := fmt.Sprintf(apiURL, id)
// 		w := httptest.NewRecorder()

// 		bucketName := "myBucket"
// 		accessKeyId := "myaccessKeyId"
// 		secretAccessKey := "mySecretAccessKey"
// 		obj := ent.AWSS3Provider{BucketName: bucketName, AccessKeyID: accessKeyId, SecretAccessKey: secretAccessKey}
// 		awss3, _ := json.Marshal(obj)
// 		responseBody := bytes.NewBuffer(awss3)

// 		req, _ := http.NewRequest(httpMethod, url, responseBody)

// 		router.ServeHTTP(w, req)

// 		var response ent.AWSS3Provider
// 		err = json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Nil(t, err)
// 		log.Debug().Msg(w.Body.String())
// 		assert.GreaterOrEqual(t, response.ID, 0)
// 		assert.Equal(t, response.BucketName, bucketName)
// 		assert.Equal(t, response.AccessKeyID, accessKeyId)
// 		assert.Equal(t, response.SecretAccessKey, secretAccessKey)
// 		assert.Equal(t, http.StatusCreated, w.Code)

// 	})

// }

// func TestUpdateAWSS3Provider(t *testing.T) {
// 	router := setupRouterTest(t)

// 	const apiURL = "/api/v1/repositories/%v/awss3provider"
// 	const httpMethod = "PATCH"

// 	newAWS, _ := json.Marshal(ent.AWSS3Provider{BucketName: "newRName"})
// 	responseBody := bytes.NewBuffer(newAWS)

// 	testInvalidFormatID(t, router, httpMethod, apiURL, "update awss3 provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
// 	testUnknownID(t, router, httpMethod, apiURL, "update awss3 provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

// 	t.Run("update awss3 provider with no patch data", func(t *testing.T) {
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

// 	t.Run("update awss3 provider with valid patch data", func(t *testing.T) {
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
// 		accessKeyId := "myaccessKeyId"
// 		secretAccessKey := "mySecretAccessKey"
// 		obj := ent.AWSS3Provider{BucketName: bucketName, AccessKeyID: accessKeyId, SecretAccessKey: secretAccessKey}

// 		_, err = database.CreateAWSS3ProviderDB(obj, id, ctx)
// 		assert.Nil(t, err)

// 		bucketName2 := "myBucket"
// 		accessKeyId2 := "myaccessKeyId"
// 		secretAccessKey2 := "mySecretAccessKey"
// 		obj2 := ent.AWSS3Provider{BucketName: bucketName2, AccessKeyID: accessKeyId2, SecretAccessKey: secretAccessKey2}
// 		awss32, _ := json.Marshal(obj2)

// 		responseBody := bytes.NewBuffer(awss32)

// 		req, _ := http.NewRequest(httpMethod, url, responseBody)

// 		router.ServeHTTP(w, req)

// 		var response ent.AWSS3Provider
// 		err = json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Nil(t, err)
// 		log.Debug().Msg(w.Body.String())
// 		assert.GreaterOrEqual(t, response.ID, 0)
// 		assert.Equal(t, response.BucketName, bucketName)
// 		assert.Equal(t, response.AccessKeyID, accessKeyId)
// 		assert.Equal(t, response.SecretAccessKey, secretAccessKey)
// 		assert.Equal(t, http.StatusOK, w.Code)

// 	})

// }
