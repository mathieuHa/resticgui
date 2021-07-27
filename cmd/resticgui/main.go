package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/API"
	"gitlab.com/resticgui/resticgui/internal/Config"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Logs"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	server *http.Server
)

func main() {
	Config.InitConfig()
	Logs.Initlogs()



	log.Info().Msg("Start main")

	resticInit()
	//Run the cron scheduler

	router := API.SetupRouter()

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
