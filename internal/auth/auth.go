// Package auth protects the API with a single user and a session cookie.
// The password is stored as a bcrypt hash in the configuration; sessions
// are random tokens whose hashes are kept in a Sessions store.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CookieName is the session cookie.
const CookieName = "resticgui_session"

// Auth checks credentials and sessions. A zero PasswordHash disables it.
type Auth struct {
	Username     string
	PasswordHash string
	// SessionTTL is how long a login lasts at most (default 7 days).
	SessionTTL time.Duration
	// IdleTimeout ends a session not used for that long (default 24 hours).
	IdleTimeout time.Duration
	sessions    Sessions
	principal   string
	limiter     *limiter
	// checking bounds the bcrypt comparisons running at once
	checking chan struct{}
	now      func() time.Time
}

// New returns an Auth that keeps its sessions in sessions. They belong to a
// principal derived from secret, the username and the password hash, so
// changing either ends every session.
func New(username, passwordHash string, secret []byte, sessions Sessions) (*Auth, error) {
	if passwordHash != "" {
		if _, err := bcrypt.Cost([]byte(passwordHash)); err != nil {
			return nil, fmt.Errorf("auth: the password hash is not a bcrypt hash (make one with: resticgui hash-password): %w", err)
		}
		if username == "" {
			return nil, errors.New("auth: a username is required with a password")
		}
		if sessions == nil {
			return nil, errors.New("auth: no session store")
		}
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("resticgui session\x00" + username + "\x00" + passwordHash))
	return &Auth{
		Username: username, PasswordHash: passwordHash,
		SessionTTL: 7 * 24 * time.Hour, IdleTimeout: 24 * time.Hour,
		sessions: sessions, principal: hex.EncodeToString(mac.Sum(nil)),
		limiter: newLimiter(), checking: make(chan struct{}, 2), now: time.Now,
	}, nil
}

// Enabled reports whether a login is required.
func (a *Auth) Enabled() bool { return a.PasswordHash != "" }

// HashPassword returns the bcrypt hash to put in the configuration.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("use a password of at least 8 characters")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

// dummyHash is compared against when the username is wrong, so that both
// failures take the same time.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not the password"), bcrypt.DefaultCost)

func (a *Auth) check(username, password string) bool {
	a.checking <- struct{}{}
	defer func() { <-a.checking }()
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(a.Username)) == 1
	hash := []byte(a.PasswordHash)
	if !userOK {
		hash = dummyHash
	}
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil && userOK
}

// newToken returns a random session token and the hash the store keeps.
func newToken() (token string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

func hashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// valid reports whether token is a live session, and records its use.
func (a *Auth) valid(ctx context.Context, token string) bool {
	if token == "" {
		return false
	}
	hash := hashToken(token)
	s, err := a.sessions.Get(ctx, hash)
	if err != nil {
		return false
	}
	now := a.now()
	if s.Principal != a.principal || now.Sub(s.Created) >= a.SessionTTL || now.Sub(s.LastSeen) >= a.IdleTimeout {
		_ = a.sessions.Delete(ctx, hash)
		return false
	}
	// a write per minute at most
	if now.Sub(s.LastSeen) >= time.Minute {
		_ = a.sessions.Touch(ctx, hash, now)
	}
	return true
}

func secure(c *gin.Context) bool {
	return c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
}

func (a *Auth) setCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: CookieName, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: secure(c), SameSite: http.SameSiteStrictMode,
	})
}

// Register adds POST /auth/login, POST /auth/logout and GET /auth/session.
func (a *Auth) Register(r gin.IRouter) {
	r.POST("/auth/login", a.login)
	r.POST("/auth/logout", a.logout)
	r.GET("/auth/session", a.session)
}

func (a *Auth) login(c *gin.Context) {
	if !a.Enabled() {
		c.Status(http.StatusNoContent)
		return
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	client := c.ClientIP()
	if wait := a.limiter.wait(client, a.now()); wait > 0 {
		c.Header("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many failed logins, try again later"})
		return
	}
	if err := c.ShouldBindJSON(&in); err != nil || !a.check(in.Username, in.Password) {
		a.limiter.fail(client, a.now())
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "wrong username or password"})
		return
	}
	a.limiter.reset(client)
	token, hash, err := newToken()
	if err == nil {
		now := a.now()
		_ = a.sessions.Prune(c, a.principal, now.Add(-a.SessionTTL), now.Add(-a.IdleTimeout))
		err = a.sessions.Create(c, hash, Session{Principal: a.principal, Created: now, LastSeen: now})
	}
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	a.setCookie(c, token, int(a.SessionTTL.Seconds()))
	c.Status(http.StatusNoContent)
}

// logout ends the session on the server, so that a copy of the cookie stops
// working too.
func (a *Auth) logout(c *gin.Context) {
	if cookie, err := c.Cookie(CookieName); err == nil && a.Enabled() {
		_ = a.sessions.Delete(c, hashToken(cookie))
	}
	a.setCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (a *Auth) authenticated(c *gin.Context) bool {
	if !a.Enabled() {
		return true
	}
	cookie, err := c.Cookie(CookieName)
	return err == nil && a.valid(c, cookie)
}

// session tells the frontend whether it must show the login form.
func (a *Auth) session(c *gin.Context) {
	out := gin.H{"auth_enabled": a.Enabled(), "authenticated": a.authenticated(c)}
	if a.Enabled() && out["authenticated"] == true {
		out["username"] = a.Username
	}
	c.JSON(http.StatusOK, out)
}

// Require rejects requests without a valid session with 401.
func (a *Auth) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.authenticated(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "login required"})
			return
		}
		c.Next()
	}
}
