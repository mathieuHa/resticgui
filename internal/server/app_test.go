package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
