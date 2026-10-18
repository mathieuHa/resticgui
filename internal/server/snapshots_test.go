package server

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
)

func TestSnapshotFiles(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	base := "/repositories/" + itoa(repo.ID) + "/snapshots/"
	a.restic.nodes = []restic.Node{{Name: "docs", Path: "/docs", Type: "dir"}, {Name: "a.txt", Path: "/docs/a.txt", Type: "file"}}

	status, body := a.do(http.MethodGet, base+"0123abcd/files?path=/docs/", nil)
	require.Equal(t, http.StatusOK, status, string(body))
	var nodes []restic.Node
	a.decode(body, &nodes)
	assert.Equal(t, []restic.Node{{Name: "a.txt", Path: "/docs/a.txt", Type: "file"}}, nodes, "the directory itself is left out")

	status, _ = a.do(http.MethodGet, base+"latest/files", nil)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []string{"ls 0123abcd /docs", "ls latest /"}, a.restic.calls)

	for _, p := range []string{base + "not-an-id/files", base + "--help/files", base + "0123abcd/files?path=relative"} {
		status, _ = a.do(http.MethodGet, p, nil)
		assert.Equal(t, http.StatusBadRequest, status, p)
	}
	status, _ = a.do(http.MethodGet, "/repositories/999/snapshots/latest/files", nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestFindAndStats(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	a.restic.found = []restic.FindResult{{Snapshot: "abc", Matches: []restic.FindMatch{{Path: "/docs/a.txt"}}}}

	status, body := a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/snapshots/latest/find?pattern=*.txt", nil)
	require.Equal(t, http.StatusOK, status, string(body))
	var matches []restic.FindMatch
	a.decode(body, &matches)
	assert.Equal(t, []restic.FindMatch{{Path: "/docs/a.txt"}}, matches)
	status, _ = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/snapshots/latest/find", nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status, body = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/stats?mode=raw-data&snapshot=0123abcd", nil)
	require.Equal(t, http.StatusOK, status, string(body))
	assert.JSONEq(t, `{"total_size":42,"snapshots_count":3}`, string(body))
	status, _ = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/stats?mode=everything", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, []string{"find latest *.txt", "stats raw-data 0123abcd"}, a.restic.calls)
}

func TestForgetSnapshot(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	base := "/repositories/" + itoa(repo.ID) + "/snapshots/"

	status, _ := a.do(http.MethodDelete, base+"0123abcd?prune=true", nil)
	assert.Equal(t, http.StatusNoContent, status)
	status, _ = a.do(http.MethodDelete, base+"latest", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	a.restic.err = &restic.SnapshotNotFoundError{IDs: []string{"0123abce"}}
	status, _ = a.do(http.MethodDelete, base+"0123abce", nil)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, []string{"forget [0123abcd] prune=true", "forget [0123abce] prune=false"}, a.restic.calls)
}

func TestDownload(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	base := "/repositories/" + itoa(repo.ID) + "/snapshots/0123abcd/download"
	a.restic.dump = "PK\x03\x04archive"

	resp, err := http.Get(a.srv.URL + BasePath + base + "?path=/docs")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `attachment; filename=docs.zip`, resp.Header.Get("Content-Disposition"))
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"), "a backed-up page is never rendered")

	// hostile names stay one quoted or encoded parameter
	for name, want := range map[string]string{
		`a "b".txt`:                `attachment; filename="a \"b\".txt"`,
		"été.txt":                  `attachment; filename*=utf-8''%C3%A9t%C3%A9.txt`,
		"x\r\nSet-Cookie: a=b.txt": "",
	} {
		resp, err := http.Get(a.srv.URL + BasePath + base + "?archive=none&path=" + url.QueryEscape("/docs/"+name))
		require.NoError(t, err)
		resp.Body.Close()
		got := resp.Header.Get("Content-Disposition")
		if want != "" {
			assert.Equal(t, want, got, name)
		}
		assert.Empty(t, resp.Header.Get("Set-Cookie"), name)
	}
	a.restic.calls = a.restic.calls[:1]

	status, _ := a.do(http.MethodGet, base+"?path=/docs/a.txt&archive=none", nil)
	assert.Equal(t, http.StatusOK, status)
	status, _ = a.do(http.MethodGet, base+"?archive=tar", nil)
	assert.Equal(t, http.StatusOK, status)
	status, _ = a.do(http.MethodGet, base+"?archive=rar", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, []string{"dump 0123abcd /docs zip", "dump 0123abcd /docs/a.txt ", "dump 0123abcd / tar"}, a.restic.calls)

	// an error before any byte is still a JSON error
	a.restic.dump = ""
	a.restic.err = &restic.Error{Command: "dump", ExitCode: 12}
	status, body := a.do(http.MethodGet, base, nil)
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, string(body), "password")
}

func TestRestore(t *testing.T) {
	root := t.TempDir()
	a := newAPI(t, func(s *Server) { s.Roots = Roots{root} })
	sealedRepo := a.create("photos")
	path := "/repositories/" + itoa(sealedRepo.ID) + "/snapshots/0123abcd/restore"

	status, body := a.do(http.MethodPost, path, map[string]any{"target": "/etc", "include": []string{"relative"}})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "outside the allowed directories")
	assert.Contains(t, string(body), `include \"relative\" is not an absolute path`)

	target := filepath.Join(root, "restore")
	status, body = a.do(http.MethodPost, path, map[string]any{"target": target, "include": []string{"/docs"}})
	require.Equal(t, http.StatusAccepted, status, string(body))
	var job jobs.JobDTO
	a.decode(body, &job)
	assert.Equal(t, jobs.KindRestore, job.Kind)
	assert.JSONEq(t, `{"snapshot_id":"0123abcd","target":"`+target+`","include":["/docs"]}`, string(job.Params))
}

func TestRootsLimitPathsAndLocations(t *testing.T) {
	root := t.TempDir()
	a := newAPI(t, func(s *Server) { s.Roots = Roots{root} })

	status, body := a.do(http.MethodPost, "/repositories", map[string]any{
		"name": "local", "backend": "local", "location": "/var/restic", "password": password,
	})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "location: ")

	status, body = a.do(http.MethodPost, "/repositories", map[string]any{
		"name": "local", "backend": "local", "location": filepath.Join(root, "repo"), "password": password,
	})
	require.Equal(t, http.StatusCreated, status, string(body))
	var repo RepositoryDTO
	a.decode(body, &repo)

	paths := "/repositories/" + itoa(repo.ID) + "/paths"
	status, body = a.do(http.MethodPut, paths, PathsDTO{Include: []string{filepath.Join(root, "..", "escape")}})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "outside the allowed directories")
	status, _ = a.do(http.MethodPut, paths, PathsDTO{Include: []string{filepath.Join(root, "data")}})
	assert.Equal(t, http.StatusOK, status)
}

func TestBrowse(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "photos"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "notes.txt"), []byte("hi"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "escape")))
	a := newAPI(t, func(s *Server) { s.Roots = Roots{root} })

	status, body := a.do(http.MethodGet, "/fs", nil)
	require.Equal(t, http.StatusOK, status)
	var entries []Entry
	a.decode(body, &entries)
	assert.Equal(t, []Entry{{Name: root, Path: root, Dir: true}}, entries, "without a path, the roots")

	status, body = a.do(http.MethodGet, "/fs?path="+root, nil)
	require.Equal(t, http.StatusOK, status, string(body))
	a.decode(body, &entries)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name)
	}
	assert.Equal(t, []string{"escape", "photos", "notes.txt"}, names, "directories first")

	for _, p := range []string{"/etc", filepath.Join(root, "escape"), root + "/../"} {
		status, _ = a.do(http.MethodGet, "/fs?path="+p, nil)
		assert.Equal(t, http.StatusForbidden, status, p)
	}
	status, _ = a.do(http.MethodGet, "/fs?path="+filepath.Join(root, "missing"), nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestRoots(t *testing.T) {
	roots, err := NewRoots([]string{"/srv/data/", "/home"})
	require.NoError(t, err)
	assert.NoError(t, roots.Check("/srv/data"))
	assert.NoError(t, roots.Check("/home/me/Pictures"))
	assert.Error(t, roots.Check("/srv/database"), "a sibling with the same prefix")
	assert.Error(t, roots.Check("/home/../etc"))
	assert.Error(t, roots.Check("home/me"))
	assert.NoError(t, Roots(nil).Check("/anything"))
	_, err = NewRoots([]string{"relative"})
	assert.Error(t, err)
}

func TestRootsRefuseNewlinesAndLinksOutside(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "escape")))
	require.NoError(t, os.Mkdir(filepath.Join(root, "inside"), 0o755))
	require.NoError(t, os.Symlink(filepath.Join(root, "inside"), filepath.Join(root, "alias")))
	a := newAPI(t, func(s *Server) { s.Roots = Roots{root} })
	repo := a.create("photos")
	paths := "/repositories/" + itoa(repo.ID) + "/paths"

	for _, p := range []string{
		filepath.Join(root, "x") + "\n/etc", // restic reads the paths one per line
		filepath.Join(root, "escape"),
		filepath.Join(root, "escape", "not", "yet", "there"),
	} {
		status, body := a.do(http.MethodPut, paths, PathsDTO{Include: []string{p}})
		assert.Equal(t, http.StatusBadRequest, status, p)
		assert.NotContains(t, string(body), `"paths"`, p)
	}
	for _, p := range []string{filepath.Join(root, "alias"), filepath.Join(root, "alias", "new")} {
		status, body := a.do(http.MethodPut, paths, PathsDTO{Include: []string{p}})
		assert.Equal(t, http.StatusOK, status, string(body))
	}

	a.restic.snapshots = []restic.Snapshot{{ID: "abcdef12"}}
	status, body := a.do(http.MethodPost, "/repositories/"+itoa(repo.ID)+"/snapshots/abcdef12/restore",
		map[string]any{"target": filepath.Join(root, "escape", "restored")})
	assert.Equal(t, http.StatusBadRequest, status, string(body))
	assert.Contains(t, string(body), "leads outside the allowed directories")
}

func FuzzRootsCheck(f *testing.F) {
	for _, p := range []string{"/srv/data/x", "/srv/data/../etc", "/srv/database", "/srv/data\n/etc", "/srv//data/./a", "relative", "/"} {
		f.Add(p)
	}
	roots, err := NewRoots([]string{"/srv/data", "/home"})
	require.NoError(f, err)
	f.Fuzz(func(t *testing.T, p string) {
		if roots.Check(p) != nil {
			return
		}
		clean := filepath.Clean(p)
		if strings.ContainsFunc(p, unicode.IsControl) {
			t.Fatalf("%q accepted with a control character", p)
		}
		if clean != "/srv/data" && clean != "/home" && !strings.HasPrefix(clean, "/srv/data/") && !strings.HasPrefix(clean, "/home/") {
			t.Fatalf("%q accepted, cleaned to %q outside the roots", p, clean)
		}
		if roots.Check(clean) != nil {
			t.Fatalf("%q accepted but not its clean form %q", p, clean)
		}
	})
}
