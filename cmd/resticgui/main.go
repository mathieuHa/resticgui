package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.com/resticgui/resticgui/internal/API"
	"gitlab.com/resticgui/resticgui/internal/Config"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Logs"
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
	serverIP := viper.GetString("server.ip")
	serverPort := viper.GetInt("server.port")
	serverAddr := fmt.Sprintf("%s:%d", serverIP, serverPort)

	server = &http.Server{
		Handler:      router,
		Addr:         serverAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	Database.Migrate()

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Err(err).Msg("Error running server")
		} else {
			log.Info().Msgf("Start Server listen at %v", serverAddr)
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
