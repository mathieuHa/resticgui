package main

import (
	"context"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/API"
	"gitlab.com/resticgui/resticgui/internal/Config"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Logs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var (
	server *http.Server
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
	router := gin.Default()
	//grp := router.Group("/")
	//{
	//	grp.StaticFS("", http.Dir("dist/front"))
	//}
	router.Use(static.Serve("/", static.LocalFile("./web/dist", true)))

	// API routes
	router.GET("/test/restic", testRestic)
	router.GET("/test/shutdown", shutdown)
	grpUsers := router.Group("/api/v1/users/")
	{
		grpUsers.GET("", API.QueryUsers)
		grpUsers.GET(":id", API.GetUsers)
		grpUsers.POST("", API.CreateUser)
		grpUsers.PATCH(":id", API.EditUser)
		grpUsers.DELETE(":id", API.RemoveUser)
	}

	grpBackups := router.Group("/api/v1/backups")
	{
		grpBackups.GET("", API.QueryBackups)
		grpBackups.GET("{id}", API.GetBackup)
		grpBackups.POST("", API.CreateBackup)
		grpBackups.PATCH("{id}", API.EditBackup)
		grpBackups.DELETE("{id}", API.RemoveBackup)
	}

	// front route
	//singlePageApp := spaHandler{staticPath: "front", indexPath: "index.html"}
	//router.PathPrefix("/").Handler(singlePageApp)
	//router.NoRoute(func(c *gin.Context){
	//	c.File("./web/dist/index.html")
	//})

	// setting server information
	server = &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	Database.Migrate()

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Err(err).Msg("Error running server")
		} else {
			log.Info().Msg("Start Server listen at 127.0.0.1:8080")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err := server.Shutdown(ctx)
	if err != nil {
		log.Error().Msg("error shutting down")
		return
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info().Msg("shutting down")
	os.Exit(0)
}

func testRestic(c *gin.Context) {
	runRestic()
}

func shutdown(c *gin.Context) {
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Err(err)
		return
	}
}
