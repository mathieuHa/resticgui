package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

var (
	LOG_FILE string
	server   *http.Server
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	path := filepath.Join(h.staticPath, r.URL.Path)

	// check whether a file exists at the given path
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {

	resticInit()

	// define router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/test/restic", testRestic).Methods("GET")
	router.HandleFunc("/test/shutdown", shutdown).Methods("GET")

	router.HandleFunc("/api/users", queryUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", getUsers).Methods("GET")
	router.HandleFunc("/api/users", createUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", editUser).Methods("PATCH")
	router.HandleFunc("/api/users/{id}", removeUser).Methods("DELETE")

	router.HandleFunc("/api/backup", queryBackups).Methods("GET")
	router.HandleFunc("/api/backup/{id}", getBackup).Methods("GET")
	router.HandleFunc("/api/backup", createBackup).Methods("POST")
	router.HandleFunc("/api/backup/{id}", editBackup).Methods("PATCH")
	router.HandleFunc("/api/backup/{id}", removeBackup).Methods("DELETE")

	// front route
	singlePageApp := spaHandler{staticPath: "front", indexPath: "index.html"}
	router.PathPrefix("/").Handler(singlePageApp)

	// setting server information
	server = &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func queryUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func editUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func removeUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func queryBackups(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func createBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func editBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func removeBackup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func testRestic(w http.ResponseWriter, r *http.Request) {
	runRestic()
	w.WriteHeader(http.StatusOK)
}

func shutdown(w http.ResponseWriter, r *http.Request) {
	server.Shutdown(context.Background())
}
