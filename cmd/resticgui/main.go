package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mathieuHa/resticgui/internal/api"
	"github.com/mathieuHa/resticgui/internal/backup"
	"github.com/mathieuHa/resticgui/internal/commands"
	Config "github.com/mathieuHa/resticgui/internal/config"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/mathieuHa/resticgui/internal/legacy"
	Logs "github.com/mathieuHa/resticgui/internal/logs"
	"github.com/mathieuHa/resticgui/internal/server"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	httpServer *http.Server
)
var cr *cron.Cron

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate-legacy" {
		os.Exit(legacy.Main(os.Args[2:], os.Stdout, os.Stderr))
	}

	Config.InitConfig()
	Logs.Initlogs()

	log.Info().Msg("Start main")

	//Run the cron scheduler
	commands.ResticInit()

	database.Migrate()
	defer database.Close()

	router := api.SetupRouter()

	// The new API runs next to the old one until the new frontend uses it.
	viper.SetDefault("api_v2.enabled", false)
	viper.SetDefault("api_v2.database", "resticgui-v2.db")
	viper.SetDefault("api_v2.key_file", "resticgui.key")
	viper.SetDefault("api_v2.workers", 2)
	viper.SetDefault("api_v2.schedule", false)
	viper.SetDefault("restic.bin", "restic")
	var v2 *server.App
	if viper.GetBool("api_v2.enabled") {
		var err error
		v2, err = server.Start(context.Background(), server.Config{
			Database:  viper.GetString("api_v2.database"),
			KeyFile:   viper.GetString("api_v2.key_file"),
			ResticBin: viper.GetString("restic.bin"),
			Workers:   viper.GetInt("api_v2.workers"),
			Schedule:  viper.GetBool("api_v2.schedule"),
		}, router)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot start the v2 API")
		}
		log.Info().Msg("v2 API enabled under " + server.BasePath)
	}

	// setting server information
	serverIP := viper.GetString("server.ip")
	serverDisable := viper.GetBool("server.disable")
	serverPort := viper.GetInt("server.port")
	serverAddr := fmt.Sprintf("%s:%d", serverIP, serverPort)

	httpServer = &http.Server{
		Handler:      router,
		Addr:         serverAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if !serverDisable {
		// Run our server in a goroutine so that it doesn't block.
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				log.Err(err).Msg("Error running server")
			} else {
				log.Info().Msgf("Start Server listen at %v", serverAddr)
			}
		}()
	}

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Remove current running backups
	backup.RemoveAllScheduledBackup(cr)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err := httpServer.Shutdown(ctx)
	if v2 != nil {
		if err := v2.Close(); err != nil {
			log.Error().Err(err).Msg("closing the v2 API")
		}
	}
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
