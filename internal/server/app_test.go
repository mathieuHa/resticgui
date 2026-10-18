package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/auth"
	"github.com/mathieuHa/resticgui/internal/secrets"
)

func TestStartAndClose(t *testing.T) {
	t.Setenv(secrets.KeyEnv, "")
	os.Unsetenv(secrets.KeyEnv)
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	r := gin.New()
	app, err := Start(context.Background(), Config{
		Database:  filepath.Join(dir, "v2.db"),
		KeyFile:   filepath.Join(dir, "resticgui.key"),
		ResticBin: filepath.Join(dir, "no-restic-here"), // a missing restic does not stop the API
		Schedule:  true,
	}, r)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, BasePath+"/repositories", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `[]`, rec.Body.String())
	info, err := os.Stat(filepath.Join(dir, "resticgui.key"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	require.NoError(t, app.Close())
}

func startApp(t *testing.T, cfg Config) (*gin.Engine, *App, error) {
	t.Helper()
	t.Setenv(secrets.KeyEnv, "")
	os.Unsetenv(secrets.KeyEnv)
	gin.SetMode(gin.TestMode)
	if cfg.ResticBin == "" {
		cfg.ResticBin = filepath.Join(t.TempDir(), "no-restic-here")
	}
	r := gin.New()
	app, err := Start(context.Background(), cfg, r)
	return r, app, err
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	data, err := os.ReadFile(from)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(to, data, 0o600))
}

func TestStartImportsTheLegacyDatabaseOnce(t *testing.T) {
	dir := t.TempDir()
	legacyDB := filepath.Join(dir, "resticgui.db")
	copyFile(t, "../legacy/testdata/legacy-v0.1.10.db", legacyDB)
	cfg := Config{
		Database:       filepath.Join(dir, "resticgui.sqlite"),
		LegacyDatabase: legacyDB,
		KeyFile:        filepath.Join(dir, "resticgui.key"),
	}

	r, app, err := startApp(t, cfg)
	require.NoError(t, err)
	var repos []RepositoryDTO
	require.NoError(t, json.Unmarshal(get(r, BasePath+"/repositories").Body.Bytes(), &repos))
	require.NotEmpty(t, repos)
	imported := len(repos)
	// remove one, which a second import would bring back
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, BasePath+"/repositories/"+itoa(repos[0].ID), nil))
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	require.NoError(t, app.Close())

	r, app, err = startApp(t, cfg)
	require.NoError(t, err)
	defer app.Close()
	require.NoError(t, json.Unmarshal(get(r, BasePath+"/repositories").Body.Bytes(), &repos))
	assert.Len(t, repos, imported-1, "the legacy database is only imported into a new store")
}

func TestStartFailedImportLeavesNoStore(t *testing.T) {
	dir := t.TempDir()
	notLegacy := filepath.Join(dir, "resticgui.db")
	require.NoError(t, os.WriteFile(notLegacy, []byte("not a database"), 0o600))
	cfg := Config{
		Database:       filepath.Join(dir, "resticgui.sqlite"),
		LegacyDatabase: notLegacy,
		KeyFile:        filepath.Join(dir, "resticgui.key"),
	}
	_, _, err := startApp(t, cfg)
	require.Error(t, err)
	_, statErr := os.Stat(cfg.Database)
	assert.ErrorIs(t, statErr, os.ErrNotExist, "the next start must try the import again")
}

func TestStartWithAuth(t *testing.T) {
	dir := t.TempDir()
	hash, err := auth.HashPassword("correct horse")
	require.NoError(t, err)
	r, app, err := startApp(t, Config{
		Database: filepath.Join(dir, "resticgui.sqlite"),
		KeyFile:  filepath.Join(dir, "resticgui.key"),
		Username: "admin", PasswordHash: hash,
		Roots: []string{dir},
	})
	require.NoError(t, err)
	defer app.Close()
	assert.Equal(t, http.StatusUnauthorized, get(r, BasePath+"/repositories").Code)
	assert.Equal(t, http.StatusOK, get(r, BasePath+"/auth/session").Code)
}

func TestStartRejectsBadConfig(t *testing.T) {
	dir := t.TempDir()
	base := Config{Database: filepath.Join(dir, "resticgui.sqlite"), KeyFile: filepath.Join(dir, "resticgui.key")}

	relative := base
	relative.Roots = []string{"data"}
	_, _, err := startApp(t, relative)
	assert.Error(t, err)

	badHash := base
	badHash.Username, badHash.PasswordHash = "admin", "plaintext"
	_, _, err = startApp(t, badHash)
	assert.Error(t, err)
}

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Recovery())
	r.GET("/boom", func(*gin.Context) { panic("boom") })
	r.GET("/abort", func(*gin.Context) { panic(http.ErrAbortHandler) })

	rec := get(r, "/boom")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":"internal error"}`, rec.Body.String())

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() { get(r, "/abort") },
		"the server must cut the connection of an aborted download")
}
