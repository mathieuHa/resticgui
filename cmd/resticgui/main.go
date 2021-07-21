package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/API"
	"gitlab.com/resticgui/resticgui/internal/Config"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Logs"
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
	Config.InitConfig()
	Logs.Initlogs()

	log.Info().Msg("Start main")

	resticInit()

	// define router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/test/restic", testRestic).Methods("GET")
	router.HandleFunc("/test/shutdown", shutdown).Methods("GET")

	router.HandleFunc("/api/users", API.QueryUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", API.GetUsers).Methods("GET")
	router.HandleFunc("/api/users", API.CreateUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", API.EditUser).Methods("PATCH")
	router.HandleFunc("/api/users/{id}", API.RemoveUser).Methods("DELETE")

	router.HandleFunc("/api/backup", API.QueryBackups).Methods("GET")
	router.HandleFunc("/api/backup/{id}", API.GetBackup).Methods("GET")
	router.HandleFunc("/api/backup", API.CreateBackup).Methods("POST")
	router.HandleFunc("/api/backup/{id}", API.EditBackup).Methods("PATCH")
	router.HandleFunc("/api/backup/{id}", API.RemoveBackup).Methods("DELETE")

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

	Database.Migrate()

	log.Info().Msg("Start Server listen at 127.0.0.1:8080")
	log.Fatal().Err(server.ListenAndServe())
}

func testRestic(w http.ResponseWriter, r *http.Request) {
	runRestic()
	w.WriteHeader(http.StatusOK)
}

func shutdown(w http.ResponseWriter, r *http.Request) {
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Err(err)
		return
	}
}
