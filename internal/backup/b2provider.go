package backup

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func validB2Provider(provider *ent.B2Provider) (valid bool, err error) {
	log.Debug().Msg("validB2Provider()")

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
