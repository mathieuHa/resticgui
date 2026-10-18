package auth

import (
	"net/http"
	"net/http/httptest"
	"strconv"
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

func newAuth(t *testing.T, password string, sessions ...Sessions) *Auth {
	t.Helper()
	hash, err := HashPassword(password)
	require.NoError(t, err)
	var store Sessions = NewMemorySessions()
	if len(sessions) > 0 {
		store = sessions[0]
	}
	a, err := New("admin", hash, []byte("secret key"), store)
	require.NoError(t, err)
	return a
}

func login(t *testing.T, r http.Handler) *http.Cookie {
	t.Helper()
	return sessionCookie(t, request(r, http.MethodPost, "/auth/login", `{"username":"admin","password":"correct horse"}`))
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
	store := NewMemorySessions()
	a := newAuth(t, "correct horse", store)
	r := newRouter(t, a)
	cookie := login(t, r)
	assert.Len(t, cookie.Value, 43, "32 random bytes")
	assert.NotEqual(t, cookie.Value, login(t, r).Value)

	forged := *cookie
	forged.Value = "9999999999.c2lnbmF0dXJl"
	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "", &forged).Code, "an unknown token")

	start := time.Now()
	at := func(d time.Duration) { a.now = func() time.Time { return start.Add(d) } }
	// used every 20 hours, the session outlives the idle timeout...
	for d := 20 * time.Hour; d < 7*24*time.Hour; d += 20 * time.Hour {
		at(d)
		assert.Equal(t, http.StatusOK, request(r, http.MethodGet, "/secret", "", cookie).Code, d)
	}
	// ...but not its lifetime
	at(7*24*time.Hour + time.Minute)
	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "", cookie).Code, "past the lifetime")

	a.now = time.Now
	idle := login(t, r)
	at(25 * time.Hour)
	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "", idle).Code, "idle for a day")

	a.now = time.Now
	cookie = login(t, r)
	other := newAuth(t, "a new password", store)
	assert.Equal(t, http.StatusUnauthorized, request(newRouter(t, other), http.MethodGet, "/secret", "", cookie).Code,
		"changing the password ends existing sessions")
}

func TestLogoutEndsTheSessionOnTheServer(t *testing.T) {
	store := NewMemorySessions()
	a := newAuth(t, "correct horse", store)
	r := newRouter(t, a)
	cookie := login(t, r)
	stolen := *cookie
	require.Equal(t, 1, store.Len())

	request(r, http.MethodPost, "/auth/logout", "", cookie)
	assert.Equal(t, 0, store.Len())
	assert.Equal(t, http.StatusUnauthorized, request(r, http.MethodGet, "/secret", "", &stolen).Code,
		"a copy of the cookie stops working")
}

func TestFailedLoginsAreLimited(t *testing.T) {
	a := newAuth(t, "correct horse")
	r := newRouter(t, a)
	from := func(ip, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = ip + ":40000"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	wrong := `{"username":"admin","password":"guess"}`
	right := `{"username":"admin","password":"correct horse"}`
	for i := range maxFailures {
		assert.Equal(t, http.StatusUnauthorized, from("192.0.2.1", wrong).Code, i)
	}
	rec := from("192.0.2.1", right)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "even the right password waits")
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.Equal(t, http.StatusNoContent, from("192.0.2.2", right).Code, "other clients are not blocked")

	start := time.Now()
	a.now = func() time.Time { return start.Add(window) }
	assert.Equal(t, http.StatusNoContent, from("192.0.2.1", right).Code, "after the window")
}

func TestLimiterMemoryIsBounded(t *testing.T) {
	l := newLimiter()
	now := time.Now()
	for i := range maxClients + 10 {
		l.fail(strconv.Itoa(i), now)
	}
	assert.LessOrEqual(t, len(l.failures), maxClients)
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
	a, err := New("", "", []byte("k"), nil)
	require.NoError(t, err)
	r := newRouter(t, a)
	assert.Equal(t, http.StatusOK, request(r, http.MethodGet, "/secret", "").Code)
	assert.JSONEq(t, `{"auth_enabled":false,"authenticated":true}`, request(r, http.MethodGet, "/auth/session", "").Body.String())
}

func TestConfigErrors(t *testing.T) {
	_, err := New("admin", "plaintext-password", []byte("k"), NewMemorySessions())
	assert.ErrorContains(t, err, "hash-password")
	hash, err := HashPassword("long enough")
	require.NoError(t, err)
	_, err = New("", hash, []byte("k"), NewMemorySessions())
	assert.Error(t, err)
	_, err = New("admin", hash, []byte("k"), nil)
	assert.Error(t, err)
	_, err = HashPassword("short")
	assert.Error(t, err)
}
