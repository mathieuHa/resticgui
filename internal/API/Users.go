package API

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func QueryUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]
	log.Debug().Msgf("ID = %v", userId)

	if !IsValidUUID(userId) {
		log.Info().Msgf("UUID is not valid")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("UUID is not valid")
		if err != nil {
			log.Err(err)
			return
		} else {
			return
		}
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, "id = ?", userId)
	log.Debug().Err(result.Error).Msg("Error")
	log.Debug().Msg(strconv.FormatInt(result.RowsAffected, 10))
	log.Debug().Msgf("%t", errors.Is(result.Error, gorm.ErrRecordNotFound))

	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Err(err)
		return
	}
	//w.WriteHeader(http.StatusOK)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var user Models.User

	for key, value := range vars {
		log.Debug().Msgf("%v : %v", key, value)
	}
	db := Database.Connect()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Info().Msgf("Error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Error reading body")
		if err != nil {
			log.Err(err)
			return
		} else {
			return
		}
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Info().Msgf("Can't match body to user: %v", err)
		err := json.NewEncoder(w).Encode("Can't match body to user")
		w.WriteHeader(http.StatusBadRequest)
		if err != nil {
			log.Err(err)
			return
		} else {
			return
		}
	}

	result := db.Create(&user) // pass pointer of data to Create
	log.Debug().Err(result.Error).Msg("Error")
	log.Debug().Msg(strconv.FormatInt(result.RowsAffected, 10))
	err = json.NewEncoder(w).Encode(user)
	w.WriteHeader(http.StatusCreated)
	if err != nil {
		log.Err(err)
		return
	} else {
		return

	}
}

func EditUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
