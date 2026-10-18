package jobs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mathieuHa/resticgui/internal/store/db"
)

// JobDTO is a job as the API returns it.
type JobDTO struct {
	ID           int64           `json:"id"`
	RepositoryID int64           `json:"repository_id"`
	Kind         Kind            `json:"kind"`
	Trigger      Trigger         `json:"trigger"`
	Status       Status          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	StartedAt    *time.Time      `json:"started_at"`
	FinishedAt   *time.Time      `json:"finished_at"`
	SnapshotID   string          `json:"snapshot_id,omitempty"`
	Summary      json.RawMessage `json:"summary"`
	Error        string          `json:"error,omitempty"`
	ExitCode     *int64          `json:"exit_code,omitempty"`
	LogExcerpt   string          `json:"log_excerpt,omitempty"`
}

func NewJobDTO(j db.Job) JobDTO {
	return JobDTO{
		ID: j.ID, RepositoryID: j.RepositoryID, Kind: Kind(j.Kind), Trigger: Trigger(j.Trigger), Status: Status(j.Status),
		CreatedAt: j.CreatedAt, StartedAt: j.StartedAt, FinishedAt: j.FinishedAt,
		SnapshotID: j.SnapshotID, Summary: json.RawMessage(j.Summary), Error: j.Error, ExitCode: j.ExitCode, LogExcerpt: j.LogExcerpt,
	}
}

// Handler serves the jobs API.
type Handler struct {
	Manager *Manager
	// Heartbeat is how often /events sends a comment line so that proxies
	// keep the connection open (default 15s).
	Heartbeat time.Duration
}

// Register adds the routes:
//
//	POST /repositories/:id/backup  start a backup now (202 with the job)
//	GET  /jobs?repository=ID       jobs of a repository, newest first
//	GET  /jobs/:id                 one job
//	POST /jobs/:id/cancel          cancel a queued or running job (202)
//	GET  /events                   Server-Sent Events: "job" and "progress"
func (h *Handler) Register(r gin.IRouter) {
	r.POST("/repositories/:id/backup", h.backup)
	r.GET("/jobs", h.list)
	r.GET("/jobs/:id", h.get)
	r.POST("/jobs/:id/cancel", h.cancel)
	r.GET("/events", h.stream)
}

func pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func (h *Handler) backup(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	job, err := h.Manager.Enqueue(c.Request.Context(), id, KindBackup, TriggerManual)
	switch {
	case errors.Is(err, ErrRepositoryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
	case errors.Is(err, ErrClosed):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "shutting down"})
	case err != nil:
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start the backup"})
	default:
		c.JSON(http.StatusAccepted, NewJobDTO(job))
	}
}

func (h *Handler) list(c *gin.Context) {
	repoID, err := strconv.ParseInt(c.Query("repository"), 10, 64)
	if err != nil || repoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "the repository query parameter is required"})
		return
	}
	limit, offset := int64(50), int64(0)
	if v := c.Query("limit"); v != "" {
		if limit, err = strconv.ParseInt(v, 10, 64); err != nil || limit < 1 || limit > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 500"})
			return
		}
	}
	if v := c.Query("offset"); v != "" {
		if offset, err = strconv.ParseInt(v, 10, 64); err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
	}
	jobs, err := h.Manager.store.ListJobsForRepository(c.Request.Context(), db.ListJobsForRepositoryParams{RepositoryID: repoID, Limit: limit, Offset: offset})
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list jobs"})
		return
	}
	out := make([]JobDTO, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, NewJobDTO(j))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	job, err := h.Manager.store.GetJob(c.Request.Context(), id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
	case err != nil:
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read the job"})
	default:
		c.JSON(http.StatusOK, NewJobDTO(job))
	}
}

func (h *Handler) cancel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.Manager.Cancel(c.Request.Context(), id)
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
	case errors.Is(err, ErrFinished):
		c.JSON(http.StatusConflict, gin.H{"error": "the job already finished"})
	case err != nil:
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not cancel the job"})
	default:
		c.Status(http.StatusAccepted)
	}
}

func (h *Handler) stream(c *gin.Context) {
	heartbeat := h.Heartbeat
	if heartbeat <= 0 {
		heartbeat = 15 * time.Second
	}
	events, unsubscribe := h.Manager.events.Subscribe(256)
	defer unsubscribe()

	w := c.Writer
	// the stream outlives the server's WriteTimeout
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	// tell EventSource how long to wait before reconnecting
	fmt.Fprint(w, "retry: 3000\n\n")
	w.Flush()

	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprint(w, ": ping\n\n")
			w.Flush()
		case e, ok := <-events:
			if !ok {
				// too slow to keep up: the client reconnects and reloads
				return
			}
			data, err := json.Marshal(e)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, data)
			w.Flush()
		}
	}
}
