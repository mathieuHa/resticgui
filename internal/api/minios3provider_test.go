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

func TestCreateMinioS3Provider(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/minios3provider"
	const httpMethod = "POST"

	newMS3, _ := json.Marshal(ent.MinioS3Provider{BucketName: "newRName"})
	responseBody := bytes.NewBuffer(newMS3)

	testInvalidFormatID(t, router, httpMethod, apiURL, "create ms3 provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "create ms3 provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

	t.Run("create ms3 provider with no post data", func(t *testing.T) {
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

	t.Run("create ms3 provider with valid post data", func(t *testing.T) {
		cleanDataBaseTest(t)
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		bucketName := "myBucket"
		accessKeyID := "myaccessKeyID"
		secretAccessKey := "mysecretAccessKey"
		obj := ent.MinioS3Provider{
			BucketName:      bucketName,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}
		ms3, _ := json.Marshal(obj)
		requestBody := bytes.NewBuffer(ms3)

		req, _ := http.NewRequest(httpMethod, url, requestBody)

		router.ServeHTTP(w, req)

		var response ent.MinioS3Provider
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, response.BucketName, bucketName)
		assert.Equal(t, response.AccessKeyID, accessKeyID)
		assert.Equal(t, response.SecretAccessKey, secretAccessKey)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

}

func TestUpdateMinioS3Provider(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/minios3provider"
	const httpMethod = "PATCH"

	newMS3, _ := json.Marshal(ent.MinioS3Provider{BucketName: "newRName"})
	responseBody := bytes.NewBuffer(newMS3)

	testInvalidFormatID(t, router, httpMethod, apiURL, "update ms3 provider with invalid id format", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "update ms3 provider with valid id not existing", responseBody, HTTPError{Error: errRNotFound})

	t.Run("update ms3 provider with no patch data", func(t *testing.T) {
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

	t.Run("update ms3 provider with valid patch data", func(t *testing.T) {
		cleanDataBaseTest(t)
		ctx := context.Background()
		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		url := fmt.Sprintf(apiURL, id)
		w := httptest.NewRecorder()

		bucketName := "myBucket"
		accessKeyID := "myaccessKeyID"
		secretAccessKey := "mysecretAccessKey"
		obj := ent.MinioS3Provider{
			BucketName:      bucketName,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}

		_, err = database.CreateMinioS3ProviderDB(obj, id, ctx)
		assert.Nil(t, err)

		bucketName2 := "myBucket2"
		accessKeyID2 := "myaccessKeyID"
		secretAccessKey2 := "mysecretAccessKey"
		obj2 := ent.MinioS3Provider{
			BucketName:      bucketName2,
			AccessKeyID:     accessKeyID2,
			SecretAccessKey: secretAccessKey2,
		}
		ms3, _ := json.Marshal(obj2)

		responseBody := bytes.NewBuffer(ms3)

		req, _ := http.NewRequest(httpMethod, url, responseBody)

		router.ServeHTTP(w, req)

		var response ent.MinioS3Provider
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		log.Debug().Msg(w.Body.String())
		assert.GreaterOrEqual(t, response.ID, 0)
		assert.Equal(t, response.BucketName, bucketName2)
		assert.Equal(t, response.AccessKeyID, accessKeyID2)
		assert.Equal(t, response.SecretAccessKey, secretAccessKey2)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
