package API

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateB2Provider(db *ent.Client, b2Provider *ent.B2Provider, ctx context.Context) (err error) {
	_, err = db.B2Provider.
		UpdateOneID(b2Provider.ID).
		SetAccountID(b2Provider.AccountID).
		SetAccountKey(b2Provider.AccountKey).
		SetBucketName(b2Provider.BucketName).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func createB2Provider(db *ent.Client, b2Provider *ent.B2Provider, repoID int, ctx context.Context) (err error) {
	_, err = db.B2Provider.
		Create().
		SetAccountID(b2Provider.AccountID).
		SetAccountKey(b2Provider.AccountKey).
		SetBucketName(b2Provider.BucketName).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func validB2Provider(provider *ent.B2Provider) (valid bool, err error) {
	if provider == nil {
		valid = false
		err = errors.New("B2Provider is nil")
		return valid, err
	}
	endpoint := "https://api.backblazeb2.com/b2api/v2/b2_authorize_account"
	client := httpClient()
	req, err := http.NewRequest("GET", endpoint, nil)
	authorization := fmt.Sprintf("%v:%v", provider.AccountID, provider.AccountKey)
	encoded := base64.StdEncoding.EncodeToString([]byte(authorization))
	req.Header.Add("Authorization", encoded)
	response, err := client.Do(req)
	if err != nil {
		log.Err(err)
		return false, err
	}
	// Close the connection to reuse it
	defer response.Body.Close()

	if response.StatusCode == 200 {
		return true, nil
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Err(err).Msgf("%v", body)
			return false, err
		}
		return false, nil
	}
}
