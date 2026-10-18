package webui

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCSPAllowsOnlyThePageScripts(t *testing.T) {
	page := []byte("<head><script>\n  theme()\n</script><script src=\"/ui/app.js\"></script></head><body><div><script>{ start() }</script></div></body>")
	ui := fstest.MapFS{"index.html": {Data: page}, "app.js": {Data: []byte("js")}}
	h := Handler(ui, "/ui")
	for _, p := range []string{"/ui/", "/ui/repositories/3", "/ui/app.js"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		csp := rec.Header().Get("Content-Security-Policy")
		assert.Contains(t, csp, "frame-ancestors 'none'", p)
		assert.Contains(t, csp, "object-src 'none'", p)
		assert.Contains(t, csp, "'sha256-", p)
	}
	csp := CSP(page)
	sum := sha256.Sum256([]byte("\n  theme()\n"))
	assert.Contains(t, csp, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
	assert.Equal(t, 2, strings.Count(csp, "'sha256-"), "the two inline scripts, not the loaded one")
	scriptSrc, _, _ := strings.Cut(strings.Split(csp, "script-src")[1], ";")
	assert.NotContains(t, scriptSrc, "unsafe")
}
