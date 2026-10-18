package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRouter(t *testing.T, a *Auth) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	a.Register(r)
	protected := r.Group("", a.Require())
	protected.GET("/secret", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func request(r http.Handler, method, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == CookieName {
			return c
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func newAuth(t *testing.T, password string) *Auth {
	t.Helper()
	hash, err := HashPassword(password)
	require.NoError(t, err)
	a, err := New("admin", hash, []byte("secret key"))
	require.NoError(t, err)
	return a
}

func TestLoginFlow(t *testing.T) {
	a := newAuth(t, "correct horse")
	r := newRouter(t, a)

	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "").Code)
	assert.JSONEq(t, `{"auth_enabled":true,"authenticated":false}`, request(r, http.MethodGet, "/auth/session", "").Body.String())

	for _, body := range []string{`{"username":"admin","password":"wrong"}`, `{"username":"root","password":"correct horse"}`, `not json`} {
		assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodPost, "/auth/login", body).Code, body)
	}

	rec := request(r, http.MethodPost, "/auth/login", `{"username":"admin","password":"correct horse"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	cookie := sessionCookie(t, rec)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.False(t, cookie.Secure, "plain HTTP")

	assert.Equal(t, http.StatusOK, request(r, http.MethodGet, "/secret", "", cookie).Code)
	assert.JSONEq(t, `{"auth_enabled":true,"authenticated":true,"username":"admin"}`, request(r, http.MethodGet, "/auth/session", "", cookie).Body.String())

	rec = request(r, http.MethodPost, "/auth/logout", "", cookie)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, -1, sessionCookie(t, rec).MaxAge)
}

func TestSessionsExpireAndAreBoundToThePassword(t *testing.T) {
	a := newAuth(t, "correct horse")
	r := newRouter(t, a)
	cookie := sessionCookie(t, request(r, http.MethodPost, "/auth/login", `{"username":"admin","password":"correct horse"}`))

	tampered := *cookie
	tampered.Value = "9999999999." + strings.SplitN(cookie.Value, ".", 2)[1]
	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "", &tampered).Code, "a longer expiry breaks the signature")

	a.now = func() time.Time { return time.Now().Add(8 * 24 * time.Hour) }
	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "", cookie).Code, "expired")
	a.now = time.Now

	other := newAuth(t, "a new password")
	assert.Equal(t, http.StatusUnauthorized, request(newRouter(t, other), http.MethodGet, "/secret", "", cookie).Code,
		"changing the password ends existing sessions")
}

func TestSecureCookieBehindHTTPS(t *testing.T) {
	a := newAuth(t, "correct horse")
	r := newRouter(t, a)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"admin","password":"correct horse"}`))
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.True(t, sessionCookie(t, rec).Secure)
}

func TestDisabled(t *testing.T) {
	a, err := New("", "", []byte("k"))
	require.NoError(t, err)
	r := newRouter(t, a)
	assert.Equal(t, http.StatusOK, request(r, http.MethodGet, "/secret", "").Code)
	assert.JSONEq(t, `{"auth_enabled":false,"authenticated":true}`, request(r, http.MethodGet, "/auth/session", "").Body.String())
}

func TestConfigErrors(t *testing.T) {
	_, err := New("admin", "plaintext-password", []byte("k"))
	assert.ErrorContains(t, err, "hash-password")
	hash, err := HashPassword("long enough")
	require.NoError(t, err)
	_, err = New("", hash, []byte("k"))
	assert.Error(t, err)
	_, err = HashPassword("short")
	assert.Error(t, err)
}
