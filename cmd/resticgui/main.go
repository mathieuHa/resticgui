package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/mathieuHa/resticgui/internal/auth"
	Config "github.com/mathieuHa/resticgui/internal/config"
	"github.com/mathieuHa/resticgui/internal/legacy"
	Logs "github.com/mathieuHa/resticgui/internal/logs"
	"github.com/mathieuHa/resticgui/internal/server"
	"github.com/mathieuHa/resticgui/internal/webui"
)

// version is set at build time with -ldflags "-X main.version=1.2.3".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version":
			fmt.Println(version)
			return
		case "migrate-legacy":
			os.Exit(legacy.Main(os.Args[2:], os.Stdout, os.Stderr))
		case "hash-password":
			os.Exit(hashPassword())
		}
	}

	Config.InitConfig()
	Logs.Initlogs()
	log.Info().Str("version", version).Msg("resticgui starting")

	viper.SetDefault("server.ip", "127.0.0.1")
	viper.SetDefault("server.port", 3001)
	viper.SetDefault("store.path", "resticgui.sqlite")
	viper.SetDefault("store.key_file", "resticgui.key")
	viper.SetDefault("store.legacy_path", "resticgui.db")
	viper.SetDefault("restic.bin", "restic")
	viper.SetDefault("jobs.workers", 2)
	viper.SetDefault("jobs.schedule", true)
	viper.SetDefault("hooks.commands", false)
	viper.SetDefault("metrics.enabled", false)

	router := gin.New()
	router.Use(server.Logger(), server.Recovery())
	ui := gin.WrapH(webui.Handler(webui.UI, "/ui"))
	router.GET("/ui", ui)
	router.GET("/ui/*path", ui)
	router.HEAD("/ui/*path", ui)
	router.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	app, err := server.Start(context.Background(), server.Config{
		Database:       viper.GetString("store.path"),
		LegacyDatabase: viper.GetString("store.legacy_path"),
		KeyFile:        viper.GetString("store.key_file"),
		ResticBin:      viper.GetString("restic.bin"),
		Workers:        viper.GetInt("jobs.workers"),
		Schedule:       viper.GetBool("jobs.schedule"),
		Roots:          viper.GetStringSlice("paths.roots"),
		Username:       viper.GetString("auth.username"),
		PasswordHash:   viper.GetString("auth.password_hash"),
		Version:        version,
		HookCommands:   viper.GetBool("hooks.commands"),
		Metrics:        viper.GetBool("metrics.enabled"),
		MetricsToken:   viper.GetString("metrics.token"),
	}, router)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start")
	}

	addr := fmt.Sprintf("%s:%d", viper.GetString("server.ip"), viper.GetInt("server.port"))
	httpServer := &http.Server{
		Handler:           router,
		Addr:              addr,
		ReadHeaderTimeout: 10 * time.Second,
		// the event stream and downloads lift the write deadline themselves
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		log.Info().Msgf("listening on http://%s", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server stopped")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()
	stop()
	log.Info().Msg("shutting down")

	shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdown); err != nil {
		log.Error().Err(err).Msg("closing connections")
	}
	// interrupts running jobs, which are marked so on the next start
	if err := app.Close(); err != nil {
		log.Error().Err(err).Msg("closing the store")
	}
}

// hashPassword reads a password on stdin and prints its bcrypt hash, the
// value of auth.password_hash.
func hashPassword() int {
	fmt.Fprint(os.Stderr, "Password: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		fmt.Fprintln(os.Stderr, "hash-password: no password read on stdin")
		return 1
	}
	hash, err := auth.HashPassword(strings.TrimRight(line, "\r\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nhash-password:", err)
		return 1
	}
	fmt.Fprintln(os.Stderr)
	fmt.Println(hash)
	return 0
}
