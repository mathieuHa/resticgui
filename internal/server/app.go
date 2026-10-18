package server

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/scheduler"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
)

// Config configures Start.
type Config struct {
	// Database is the SQLite file of the new store, created if missing.
	Database string
	// KeyFile holds the secrets key, created if missing; $RESTICGUI_SECRET_KEY wins.
	KeyFile string
	// ResticBin is the restic binary, "restic" looks it up in PATH.
	ResticBin string
	// Workers is how many jobs may run at once.
	Workers int
	// Schedule runs the schedules of the new store. Leave it off while the
	// old scheduler runs on an imported copy of the same repositories, or
	// every scheduled job would run twice.
	Schedule bool
}

// App is the running /api/v2 stack: store, job manager and scheduler.
type App struct {
	Store   *store.Store
	Jobs    *jobs.Manager
	stopped chan struct{}
	stop    context.CancelFunc
}

// Start opens the store, recovers interrupted jobs, starts the scheduler and
// mounts the API on router under BasePath.
func Start(ctx context.Context, cfg Config, router gin.IRouter) (*App, error) {
	key, err := secrets.LoadKey(cfg.KeyFile)
	if err != nil {
		return nil, err
	}
	box, err := secrets.New(key)
	if err != nil {
		return nil, err
	}
	st, err := store.Open(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cfg.Database, err)
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

	(&Server{Store: st, Box: box, Restic: runner, Jobs: manager, Scheduler: sched}).Register(router.Group(BasePath))
	return app, nil
}

// Close stops the scheduler, interrupts running jobs and closes the store.
func (a *App) Close() error {
	a.stop()
	<-a.stopped
	a.Jobs.Close()
	return a.Store.Close()
}
