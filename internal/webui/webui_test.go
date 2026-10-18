package webui

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	ui := fstest.MapFS{
		"index.html":               {Data: []byte("<html>app</html>")},
		"favicon.ico":              {Data: []byte("icon")},
		"_app/immutable/app.1a.js": {Data: []byte("js")},
	}
	h := Handler(ui, "/ui")
	get := func(p string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		return rec
	}

	for _, p := range []string{"/ui", "/ui/", "/ui/repositories/3", "/ui/_app", "/ui/../../etc/passwd"} {
		rec := get(p)
		assert.Equal(t, http.StatusOK, rec.Code, p)
		assert.Equal(t, "<html>app</html>", rec.Body.String(), p)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), p)
	}

	rec := get("/ui/favicon.ico")
	assert.Equal(t, "icon", rec.Body.String())
	assert.Empty(t, rec.Header().Get("Cache-Control"))

	rec = get("/ui/_app/immutable/app.1a.js")
	assert.Equal(t, "js", rec.Body.String())
	assert.Contains(t, rec.Header().Get("Cache-Control"), "immutable")
}

func TestHandlerWithoutBuild(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(fstest.MapFS{}, "/ui").ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ui/", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "not built")
}
