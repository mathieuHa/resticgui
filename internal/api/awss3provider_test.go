package api

import (
	"bytes"
	"encoding/json"
	"testing"

	"gitlab.com/resticgui/resticgui/ent"
)

func TestCreateAWSS3Provider(t *testing.T) {
	router := setupRouterTest(t)

	const apiURL = "/api/v1/repositories/%v/awss3provider"
	const httpMethod = "POST"

	newAWS, _ := json.Marshal(ent.AWSS3Provider{BucketName: "newRName"})
	responseBody := bytes.NewBuffer(newAWS)

	testInvalidFormatID(t, router, httpMethod, apiURL, "valid", nil, HTTPError{Error: errIDNotValid})
	testUnknownID(t, router, httpMethod, apiURL, "", responseBody, HTTPError{Error: errRNotFound})

	// t.Run("create repository with no post data", func(t *testing.T) {
	// 	cleanDataBaseTest(t)
	// 	url := apiURL
	// 	w := httptest.NewRecorder()
	// 	req, _ := http.NewRequest(httpMethod, url, nil)
	// 	router.ServeHTTP(w, req)

	// 	var err HTTPError
	// 	err.Error = errInvalidRequest
	// 	errString, errBind := json.Marshal(err)

	// 	assert.Nil(t, errBind)
	// 	assert.Equal(t, http.StatusBadRequest, w.Code)
	// 	assert.JSONEq(t, string(errString), w.Body.String())
	// })

	// t.Run("create repository with valid post data", func(t *testing.T) {
	// 	cleanDataBaseTest(t)
	// 	url := apiURL
	// 	w := httptest.NewRecorder()

	// 	rName := "myRepo"
	// 	r, _ := json.Marshal(ent.Repository{Name: rName})
	// 	responseBody := bytes.NewBuffer(r)

	// 	req, _ := http.NewRequest(httpMethod, url, responseBody)

	// 	router.ServeHTTP(w, req)

	// 	var repositoryReceived ent.Repository
	// 	err := json.Unmarshal(w.Body.Bytes(), &repositoryReceived)
	// 	assert.Nil(t, err)
	// 	assert.GreaterOrEqual(t, repositoryReceived.ID, 0)
	// 	assert.Equal(t, repositoryReceived.Name, rName)
	// 	assert.Equal(t, http.StatusCreated, w.Code)
	// })

}
