package API

import "net/http"

func QueryBackups(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func GetBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func CreateBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func EditBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func RemoveBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
