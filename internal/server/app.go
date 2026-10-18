package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/mathieuHa/resticgui/internal/auth"
	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/legacy"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/scheduler"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
)

// Config configures Start.
type Config struct {
	// Database is the SQLite file of the store, created if missing.
	Database string
	// LegacyDatabase is a database of resticgui v0.1. When it exists and
	// Database does not, it is imported on the first start.
	LegacyDatabase string
	// KeyFile holds the secrets key, created if missing; $RESTICGUI_SECRET_KEY wins.
	KeyFile string
	// ResticBin is the restic binary, "restic" looks it up in PATH.
	ResticBin string
	// Workers is how many jobs may run at once.
	Workers int
	// Schedule runs the schedules; off is for tests and one-off runs.
	Schedule bool
	// Roots bound the paths the API accepts; empty allows any.
	Roots []string
	// Username and PasswordHash (bcrypt) protect the API; no hash leaves it open.
	Username     string
	PasswordHash string
}

// App is the running API: store, job manager and scheduler.
type App struct {
	Store   *store.Store
	Jobs    *jobs.Manager
	stopped chan struct{}
	stop    context.CancelFunc
}

// Start opens the store (importing a legacy database the first time),
// recovers interrupted jobs, starts the scheduler and mounts the API on
// router under BasePath.
func Start(ctx context.Context, cfg Config, router gin.IRouter) (*App, error) {
	roots, err := NewRoots(cfg.Roots)
	if err != nil {
		return nil, err
	}
	key, err := secrets.LoadKey(cfg.KeyFile)
	if err != nil {
		return nil, err
	}
	box, err := secrets.New(key)
	if err != nil {
		return nil, err
	}
	guard, err := auth.New(cfg.Username, cfg.PasswordHash, key)
	if err != nil {
		return nil, err
	}
	if !guard.Enabled() {
		log.Warn().Msg("no auth.password_hash configured: the API is open to anyone who can reach it")
	}

	st, err := openStore(ctx, cfg, box)
	if err != nil {
		return nil, err
	}

	runner := restic.NewRunner(cfg.ResticBin)
	if version, err := runner.CheckVersion(ctx); err != nil {
		// keep serving: settings can still be edited, and jobs report the error
		log.Error().Err(err).Msg("restic is not usable")
	} else {
		log.Info().Str("version", version).Msg("restic found")
	}

	manager := jobs.New(st, box, runner, jobs.NewBroker(), jobs.Options{Workers: cfg.Workers})
	if n, err := manager.Recover(ctx); err != nil {
		st.Close()
		return nil, err
	} else if n > 0 {
		log.Warn().Int64("jobs", n).Msg("marked jobs left unfinished by the previous run as interrupted")
	}
	sched := scheduler.New(st, manager, nil)

	runCtx, stop := context.WithCancel(context.Background())
	app := &App{Store: st, Jobs: manager, stopped: make(chan struct{}), stop: stop}
	go func() {
		defer close(app.stopped)
		if cfg.Schedule {
			_ = sched.Run(runCtx)
		}
	}()

	(&Server{Store: st, Box: box, Restic: runner, Jobs: manager, Scheduler: sched, Roots: roots, Auth: guard}).Register(router.Group(BasePath))
	return app, nil
}

func openStore(ctx context.Context, cfg Config, box *secrets.Box) (*store.Store, error) {
	_, err := os.Stat(cfg.Database)
	fresh := errors.Is(err, os.ErrNotExist)
	importLegacy := false
	if fresh && cfg.LegacyDatabase != "" {
		if _, err := os.Stat(cfg.LegacyDatabase); err == nil {
			importLegacy = true
		}
	}

	st, err := store.Open(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cfg.Database, err)
	}
	if !importLegacy {
		return st, nil
	}

	report, err := legacy.Import(ctx, cfg.LegacyDatabase, st, box)
	if err != nil {
		st.Close()
		for _, f := range []string{cfg.Database, cfg.Database + "-wal", cfg.Database + "-shm"} {
			os.Remove(f)
		}
		return nil, fmt.Errorf("import %s: %w", cfg.LegacyDatabase, err)
	}
	log.Info().Str("from", cfg.LegacyDatabase).Strs("imported", report.Imported).Msg("imported the v0.1 database")
	for _, s := range report.Skipped {
		log.Warn().Str("repository", s).Msg("legacy import: skipped")
	}
	for _, w := range report.Warnings {
		log.Warn().Msg("legacy import: " + w)
	}
	return st, nil
}

// Close stops the scheduler, interrupts running jobs and closes the store.
func (a *App) Close() error {
	a.stop()
	<-a.stopped
	a.Jobs.Close()
	return a.Store.Close()
}

// Recovery answers 500 to a panicking handler, but lets http.ErrAbortHandler
// through so that the server cuts the connection, which is how a download
// that fails midway tells the client it is incomplete.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler { //nolint:errorlint // sentinel value, compared as http does
					panic(v)
				}
				log.Error().Interface("panic", v).Str("path", c.Request.URL.Path).Msg("handler panicked")
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				}
			}
		}()
		c.Next()
	}
}

// Logger logs one line per request, without the query string, which can
// carry file paths.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if strings.HasPrefix(c.Request.URL.Path, BasePath) {
			log.Debug().Str("method", c.Request.Method).Str("path", c.Request.URL.Path).Int("status", c.Writer.Status()).Msg("request")
		}
	}
}
