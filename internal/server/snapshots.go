package server

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

var snapshotPattern = regexp.MustCompile(`^([0-9a-f]{8,64}|latest)$`)

// snapshotTarget resolves the repository and snapshot of a request, and
// answers the error itself when it cannot.
func (s *Server) snapshotTarget(c *gin.Context) (db.Repository, *restic.Repo, string, bool) {
	id, ok := repositoryID(c)
	if !ok {
		return db.Repository{}, nil, "", false
	}
	snapshot := c.Param("snapshot")
	if !snapshotPattern.MatchString(snapshot) {
		fail(c, http.StatusBadRequest, "invalid snapshot id")
		return db.Repository{}, nil, "", false
	}
	r, err := s.Store.GetRepository(c.Request.Context(), id)
	if err != nil {
		notFound(c, err, "repository")
		return db.Repository{}, nil, "", false
	}
	target, err := s.resticRepo(r)
	if err != nil {
		internal(c, err)
		return db.Repository{}, nil, "", false
	}
	return r, target, snapshot, true
}

// snapshotPath reads the path query parameter: an absolute path inside the
// snapshot, "/" when missing.
func snapshotPath(c *gin.Context) (string, bool) {
	p := c.DefaultQuery("path", "/")
	if !strings.HasPrefix(p, "/") || strings.ContainsRune(p, 0) {
		fail(c, http.StatusBadRequest, "path is an absolute path inside the snapshot")
		return "", false
	}
	return path.Clean(p), true
}

func (s *Server) listFiles(c *gin.Context) {
	_, target, snapshot, ok := s.snapshotTarget(c)
	if !ok {
		return
	}
	dir, ok := snapshotPath(c)
	if !ok {
		return
	}
	nodes, err := s.Restic.Ls(c.Request.Context(), target, snapshot, dir)
	if err != nil {
		resticFailure(c, err)
		return
	}
	// restic also lists dir itself
	out := make([]restic.Node, 0, len(nodes))
	for _, n := range nodes {
		if n.Path != dir {
			out = append(out, n)
		}
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) findFiles(c *gin.Context) {
	_, target, snapshot, ok := s.snapshotTarget(c)
	if !ok {
		return
	}
	pattern := c.Query("pattern")
	if strings.TrimSpace(pattern) == "" {
		fail(c, http.StatusBadRequest, "pattern is required, such as *.jpg")
		return
	}
	results, err := s.Restic.Find(c.Request.Context(), target, snapshot, pattern)
	if err != nil {
		resticFailure(c, err)
		return
	}
	matches := []restic.FindMatch{}
	for _, r := range results {
		matches = append(matches, r.Matches...)
	}
	c.JSON(http.StatusOK, matches)
}

var statsModes = map[string]bool{"restore-size": true, "files-by-contents": true, "raw-data": true, "blobs-per-file": true}

func (s *Server) stats(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	mode := c.DefaultQuery("mode", "restore-size")
	if !statsModes[mode] {
		fail(c, http.StatusBadRequest, "mode is one of restore-size, files-by-contents, raw-data, blobs-per-file")
		return
	}
	snapshot := c.Query("snapshot")
	if snapshot != "" && !snapshotPattern.MatchString(snapshot) {
		fail(c, http.StatusBadRequest, "invalid snapshot id")
		return
	}
	r, err := s.Store.GetRepository(c.Request.Context(), id)
	if err != nil {
		notFound(c, err, "repository")
		return
	}
	target, err := s.resticRepo(r)
	if err != nil {
		internal(c, err)
		return
	}
	stats, err := s.Restic.Stats(c.Request.Context(), target, mode, snapshot)
	if err != nil {
		resticFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (s *Server) forgetSnapshot(c *gin.Context) {
	_, target, snapshot, ok := s.snapshotTarget(c)
	if !ok {
		return
	}
	if snapshot == "latest" {
		fail(c, http.StatusBadRequest, "name the snapshot to forget by its id")
		return
	}
	err := s.Restic.ForgetSnapshots(c.Request.Context(), target, []string{snapshot}, c.Query("prune") == "true")
	var missing *restic.SnapshotNotFoundError
	if errors.As(err, &missing) {
		fail(c, http.StatusNotFound, "snapshot not found")
		return
	}
	if err != nil {
		resticFailure(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) restore(c *gin.Context) {
	r, _, snapshot, ok := s.snapshotTarget(c)
	if !ok {
		return
	}
	var in struct {
		Target  string   `json:"target"`
		Include []string `json:"include"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	var errs []error
	if err := s.Roots.Check(in.Target); err != nil {
		errs = append(errs, fmt.Errorf("target: %w", err))
	}
	for _, p := range in.Include {
		if !strings.HasPrefix(p, "/") {
			errs = append(errs, fmt.Errorf("include %q is not an absolute path inside the snapshot", p))
		}
	}
	if err := errors.Join(errs...); err != nil {
		invalid(c, err)
		return
	}
	job, err := s.Jobs.EnqueueRestore(c.Request.Context(), r.ID, jobs.RestoreParams{SnapshotID: snapshot, Target: in.Target, Include: in.Include})
	switch {
	case errors.Is(err, jobs.ErrClosed):
		fail(c, http.StatusServiceUnavailable, "shutting down")
	case err != nil:
		internal(c, err)
	default:
		c.JSON(http.StatusAccepted, jobs.NewJobDTO(job))
	}
}

// lazyWriter sets the download headers on the first byte, so that an error
// before any output can still be answered as JSON.
type lazyWriter struct {
	c        *gin.Context
	filename string
	started  bool
}

func (w *lazyWriter) Write(p []byte) (int, error) {
	if !w.started {
		w.started = true
		w.c.Header("Content-Type", "application/octet-stream")
		w.c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", w.filename))
		w.c.Status(http.StatusOK)
	}
	return w.c.Writer.Write(p)
}

func (s *Server) download(c *gin.Context) {
	r, target, snapshot, ok := s.snapshotTarget(c)
	if !ok {
		return
	}
	p, ok := snapshotPath(c)
	if !ok {
		return
	}
	archive := c.DefaultQuery("archive", "zip")
	if archive != "zip" && archive != "tar" && archive != "none" {
		fail(c, http.StatusBadRequest, "archive is zip, tar, or none for a single file")
		return
	}
	name := path.Base(p)
	if name == "/" {
		name = r.Name
	}
	if archive != "none" {
		name += "." + archive
	} else {
		archive = ""
	}
	w := &lazyWriter{c: c, filename: name}
	err := s.Restic.Dump(c.Request.Context(), target, snapshot, p, archive, w)
	if err == nil && !w.started {
		// an empty file
		w.Write(nil)
	}
	if err != nil {
		if w.started {
			// too late for a status: cut the transfer so the client sees it failed
			_ = c.Error(err)
			panic(http.ErrAbortHandler)
		}
		resticFailure(c, err)
	}
}
