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
	"net/http"
	"strconv"
)

func QueryUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var users Models.User

	for key, value := range vars {
		log.Debug().Msgf("%v : %v", key, value)
	}
	db := Database.Connect()
	result := db.Find(&users, "id = ?", "1b74413f-f3b8-409f-ac47-e8c062e3472a")
	log.Debug().Err(result.Error)
	log.Debug().Msg(strconv.FormatInt(result.RowsAffected, 10))
	log.Debug().Msgf("%t", errors.Is(result.Error, gorm.ErrRecordNotFound))

	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Err(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	for key, value := range vars {
		log.Debug().Msgf("%v : %v", key, value)
	}
	db := Database.Connect()

	user := Models.User{}
	user.ID = uuid.NewString()
	user.Firstname = ""

	db.Create(&user) // pass pointer of data to Create

	w.WriteHeader(http.StatusCreated)
}

func EditUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
