package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Roots are the directories resticgui may back up, restore into and browse.
// Empty means the whole filesystem.
type Roots []string

// NewRoots cleans the configured roots, which must be absolute.
func NewRoots(dirs []string) (Roots, error) {
	var roots Roots
	for _, d := range dirs {
		if !filepath.IsAbs(d) {
			return nil, fmt.Errorf("path root %q is not absolute", d)
		}
		roots = append(roots, filepath.Clean(d))
	}
	return roots, nil
}

// Check returns an error unless p is an absolute path inside a root.
func (r Roots) Check(p string) error {
	if !filepath.IsAbs(p) {
		return fmt.Errorf("%q is not an absolute path", p)
	}
	if !r.contains(filepath.Clean(p)) {
		return fmt.Errorf("%q is outside the allowed directories (%s)", p, strings.Join(r, ", "))
	}
	return nil
}

func (r Roots) contains(p string) bool {
	if len(r) == 0 {
		return true
	}
	for _, root := range r {
		rel, err := filepath.Rel(root, p)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// checkResolved also follows symbolic links, so that a link inside a root
// cannot expose a directory outside of it.
func (r Roots) checkResolved(p string) error {
	if err := r.Check(p); err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return err
	}
	if !r.contains(resolved) {
		return fmt.Errorf("%q leads outside the allowed directories", p)
	}
	return nil
}

// Entry is one item of a directory listing.
type Entry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Dir     bool      `json:"dir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mtime"`
}

// browse lists a local directory, to pick backup paths and restore targets.
// Without a path it lists the roots.
func (s *Server) browse(c *gin.Context) {
	dir := c.Query("path")
	if dir == "" {
		if len(s.Roots) == 0 {
			dir = string(filepath.Separator)
		} else {
			out := make([]Entry, 0, len(s.Roots))
			for _, root := range s.Roots {
				out = append(out, Entry{Name: root, Path: root, Dir: true})
			}
			c.JSON(http.StatusOK, out)
			return
		}
	}
	if err := s.Roots.checkResolved(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fail(c, http.StatusNotFound, "no such directory")
			return
		}
		fail(c, http.StatusForbidden, err.Error())
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			fail(c, http.StatusNotFound, "no such directory")
		case errors.Is(err, os.ErrPermission):
			fail(c, http.StatusForbidden, "permission denied")
		default:
			fail(c, http.StatusBadRequest, "not a readable directory")
		}
		return
	}
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue // removed while listing
		}
		full := filepath.Join(dir, e.Name())
		isDir := e.IsDir()
		if e.Type()&os.ModeSymlink != 0 {
			if target, err := os.Stat(full); err == nil {
				isDir = target.IsDir()
			}
		}
		out = append(out, Entry{Name: e.Name(), Path: full, Dir: isDir, Size: info.Size(), ModTime: info.ModTime()})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Dir && !out[j].Dir })
	c.JSON(http.StatusOK, out)
}
