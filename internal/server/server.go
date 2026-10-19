// Package server is the /api/v2 HTTP API, backed by the store, the job
// manager and the scheduler.
package server

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mathieuHa/resticgui/internal/auth"
	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/notify"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
)

// BasePath is where Register is expected to be mounted.
const BasePath = "/api/v2"

// Restic is the part of *restic.Runner the API calls directly, for reads
// that answer at once rather than as jobs.
type Restic interface {
	Snapshots(ctx context.Context, repo *restic.Repo, filter restic.SnapshotFilter) ([]restic.Snapshot, error)
	Ls(ctx context.Context, repo *restic.Repo, snapshotID, dir string) ([]restic.Node, error)
	Find(ctx context.Context, repo *restic.Repo, snapshotID, pattern string) ([]restic.FindResult, error)
	Stats(ctx context.Context, repo *restic.Repo, mode, snapshotID string) (*restic.Stats, error)
	Forget(ctx context.Context, repo *restic.Repo, opts restic.ForgetOptions) ([]restic.ForgetGroup, error)
	ForgetSnapshots(ctx context.Context, repo *restic.Repo, ids []string, prune bool) error
	Dump(ctx context.Context, repo *restic.Repo, snapshotID, path, archive string, w io.Writer) error
	Diff(ctx context.Context, repo *restic.Repo, from, to string) (*restic.Diff, error)
	Exists(ctx context.Context, repo *restic.Repo) (bool, error)
}

// Reloader is told when schedules change; *scheduler.Scheduler implements it.
type Reloader interface {
	Reload()
}

// Server holds what the handlers need.
type Server struct {
	Store     *store.Store
	Box       *secrets.Box
	Restic    Restic
	Jobs      *jobs.Manager
	Scheduler Reloader
	// Roots limits backup paths, local repositories, restore targets and
	// the folder browser; empty allows the whole filesystem.
	Roots Roots
	// Auth guards every route but its own and the OpenAPI document; nil
	// leaves the API open.
	Auth *auth.Auth
	// Version is the ResticGUI version, "dev" for a build that is not a release.
	Version string
	// Notify sends the test messages of notifiers.
	Notify *notify.Service
	// HookCommands allows command hooks (hooks.commands in the configuration).
	HookCommands bool
	// Metrics says whether /metrics is served, for GET /settings.
	Metrics bool
}

//go:embed openapi.yaml
var openAPI []byte

// Register adds the routes under r, which is usually router.Group(BasePath).
func (s *Server) Register(public gin.IRouter) {
	public.GET("/openapi.yaml", func(c *gin.Context) { c.Data(http.StatusOK, "application/yaml", openAPI) })
	public.GET("/version", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"version": s.Version}) })
	r := public
	if s.Auth != nil {
		s.Auth.Register(public)
		r = public.Group("", s.Auth.Require())
	}

	r.GET("/backends", s.listBackends)
	r.GET("/settings", s.settings)

	r.GET("/notifiers", s.listNotifiers)
	r.POST("/notifiers", s.createNotifier)
	r.PUT("/notifiers/:id", s.updateNotifier)
	r.DELETE("/notifiers/:id", s.deleteNotifier)
	r.POST("/notifiers/:id/test", s.testNotifier)

	r.GET("/repositories", s.listRepositories)
	r.POST("/repositories", s.createRepository)
	r.POST("/repositories/probe", s.probe)
	r.GET("/repositories/:id", s.getRepository)
	r.PUT("/repositories/:id", s.updateRepository)
	r.DELETE("/repositories/:id", s.deleteRepository)
	r.POST("/repositories/:id/plans", s.createPlan)
	r.PUT("/repositories/:id/plans/:plan", s.updatePlan)
	r.DELETE("/repositories/:id/plans/:plan", s.deletePlan)
	r.POST("/repositories/:id/plans/:plan/retention/preview", s.previewRetention)
	r.PUT("/repositories/:id/schedules/:job", s.putSchedule)
	r.DELETE("/repositories/:id/schedules/:job", s.deleteSchedule)
	r.PUT("/repositories/:id/healthcheck", s.putHealthcheck)
	r.DELETE("/repositories/:id/healthcheck", s.deleteHealthcheck)
	r.PUT("/repositories/:id/hooks", s.putHooks)
	r.POST("/repositories/:id/jobs", s.startJob)
	r.POST("/repositories/:id/password", s.changePassword)
	r.GET("/repositories/:id/snapshots", s.listSnapshots)
	r.DELETE("/repositories/:id/snapshots/:snapshot", s.forgetSnapshot)
	r.GET("/repositories/:id/snapshots/:snapshot/files", s.listFiles)
	r.GET("/repositories/:id/snapshots/:snapshot/find", s.findFiles)
	r.GET("/repositories/:id/snapshots/:snapshot/download", s.download)
	r.GET("/repositories/:id/snapshots/:snapshot/diff", s.diff)
	r.POST("/repositories/:id/snapshots/:snapshot/restore", s.restore)
	r.GET("/repositories/:id/stats", s.stats)
	r.GET("/fs", s.browse)

	(&jobs.Handler{Manager: s.Jobs}).Register(r)
}

// apiError is the body of every error response.
type apiError struct {
	Error string `json:"error"`
	// Details lists validation problems, one per line of the error.
	Details []string `json:"details,omitempty"`
}

func fail(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, apiError{Error: msg})
}

func invalid(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, apiError{Error: "invalid request", Details: strings.Split(err.Error(), "\n")})
}

// internal logs err through gin and answers 500 without details.
func internal(c *gin.Context, err error) {
	_ = c.Error(err)
	fail(c, http.StatusInternalServerError, "internal error")
}

func repositoryID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "invalid repository id")
		return 0, false
	}
	return id, true
}

// notFound answers 404 for a missing row and 500 for any other error.
func notFound(c *gin.Context, err error, what string) {
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, what+" not found")
		return
	}
	internal(c, err)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
