package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/auth"
	"github.com/mathieuHa/resticgui/internal/store"
)

func guarded(allowed []string, authOff bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	g := r.Group(BasePath, Guard(allowed, authOff))
	g.GET("/thing", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.POST("/thing", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/ui/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func send(r http.Handler, method, host, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://"+host+BasePath+"/thing", strings.NewReader(body))
	if body == "" {
		req = httptest.NewRequest(method, "http://"+host+BasePath+"/thing", nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestGuardRefusesCrossSiteWrites(t *testing.T) {
	r := guarded(nil, false)
	json := map[string]string{"Content-Type": "application/json"}
	cases := []struct {
		name    string
		method  string
		body    string
		headers map[string]string
		want    int
	}{
		{"the web interface", http.MethodPost, `{}`, map[string]string{"Content-Type": "application/json", "Sec-Fetch-Site": "same-origin", "Origin": "http://127.0.0.1:3001"}, http.StatusOK},
		{"curl", http.MethodPost, `{}`, json, http.StatusOK},
		{"a POST without a body", http.MethodPost, "", nil, http.StatusOK},
		{"reads are never refused", http.MethodGet, "", map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusOK},
		// fetch(url, {method: "POST", mode: "no-cors", body}) from another site
		{"text/plain from another site", http.MethodPost, `{}`, map[string]string{"Content-Type": "text/plain;charset=UTF-8", "Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example"}, http.StatusForbidden},
		{"a sibling subdomain", http.MethodPost, `{}`, map[string]string{"Content-Type": "application/json", "Sec-Fetch-Site": "same-site", "Origin": "https://app.example.com"}, http.StatusForbidden},
		{"an old browser, another origin", http.MethodPost, `{}`, map[string]string{"Content-Type": "application/json", "Origin": "https://evil.example"}, http.StatusForbidden},
		{"an old browser, behind a proxy", http.MethodPost, `{}`, map[string]string{"Content-Type": "application/json", "Origin": "https://backup.example.com", "X-Forwarded-Host": "backup.example.com"}, http.StatusOK},
		{"a form", http.MethodPost, `a=b`, map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, http.StatusUnsupportedMediaType},
		{"text/plain without Origin", http.MethodPost, `{}`, map[string]string{"Content-Type": "text/plain"}, http.StatusUnsupportedMediaType},
		{"no content type", http.MethodPost, `{}`, nil, http.StatusUnsupportedMediaType},
		{"JSON with a charset", http.MethodPost, `{}`, map[string]string{"Content-Type": "application/json; charset=utf-8"}, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, send(r, tc.method, "127.0.0.1:3001", tc.body, tc.headers).Code)
		})
	}
}

func TestGuardChecksTheHostWithoutLogin(t *testing.T) {
	open := guarded([]string{"backup.lan"}, true)
	for _, host := range []string{"127.0.0.1:3001", "[::1]:3001", "localhost:3001", "LOCALHOST", "192.168.1.20:3001", "backup.lan", "nas.localhost:3001"} {
		assert.Equal(t, http.StatusOK, send(open, http.MethodGet, host, "", nil).Code, host)
	}
	// DNS rebinding: a page on rebind.evil.example, now resolving to 127.0.0.1
	for _, host := range []string{"rebind.evil.example", "rebind.evil.example:3001", "backup.lan.evil.example"} {
		rec := send(open, http.MethodGet, host, "", nil)
		assert.Equal(t, http.StatusMisdirectedRequest, rec.Code, host)
		assert.Contains(t, rec.Body.String(), "server.allowed_hosts")
	}

	withLogin := guarded(nil, false)
	assert.Equal(t, http.StatusOK, send(withLogin, http.MethodGet, "backup.example.com", "", nil).Code,
		"the session cookie protects a server with a login")
}

func TestSecurityHeaders(t *testing.T) {
	r := guarded(nil, false)
	for _, path := range []string{BasePath + "/thing", "/ui/"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		h := rec.Header()
		assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"), path)
		assert.Equal(t, "DENY", h.Get("X-Frame-Options"), path)
		assert.Equal(t, "no-referrer", h.Get("Referrer-Policy"), path)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, BasePath+"/thing", nil))
	assert.Equal(t, apiCSP, rec.Header().Get("Content-Security-Policy"))
}

func TestLoopback(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1:3001": true, "[::1]:3001": true, "localhost:3001": true,
		"0.0.0.0:3030": false, ":3030": false, "192.168.1.2:3001": false,
	} {
		assert.Equal(t, want, loopback(addr), addr)
	}
}

func TestSessionsAreStoredAndEndedInTheDatabase(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(ctx, filepath.Join(t.TempDir(), "s.db"))
	require.NoError(t, err)
	defer st.Close()
	hash, err := auth.HashPassword("long enough password")
	require.NoError(t, err)
	guard, err := auth.New("admin", hash, []byte("k"), storeSessions{st})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	guard.Register(r)
	r.GET("/secret", guard.Require(), func(c *gin.Context) { c.Status(http.StatusOK) })
	call := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	rec := call(http.MethodPost, "/auth/login", `{"username":"admin","password":"long enough password"}`, nil)
	require.Equal(t, http.StatusNoContent, rec.Code)
	cookie := rec.Result().Cookies()[0]

	var stored string
	require.NoError(t, st.DB.QueryRowContext(ctx, "SELECT hex(token_hash) FROM sessions").Scan(&stored))
	assert.NotContains(t, strings.ToLower(stored), strings.ToLower(cookie.Value), "only the hash is stored")

	assert.Equal(t, http.StatusOK, call(http.MethodGet, "/secret", "", cookie).Code)
	call(http.MethodPost, "/auth/logout", "", cookie)
	assert.Equal(t, http.StatusUnauthorized, call(http.MethodGet, "/secret", "", cookie).Code)
	var n int
	require.NoError(t, st.DB.QueryRowContext(ctx, "SELECT count(*) FROM sessions").Scan(&n))
	assert.Equal(t, 0, n)
}
