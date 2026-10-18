// Package auth protects the API with a single user and a signed session
// cookie. The password is stored as a bcrypt hash in the configuration.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	// SessionTTL is how long a login lasts (default 7 days).
	SessionTTL time.Duration
	key        []byte
	now        func() time.Time
}

// New returns an Auth whose sessions are signed with a key derived from
// secret. Changing the password hash invalidates every session.
func New(username, passwordHash string, secret []byte) (*Auth, error) {
	if passwordHash != "" {
		if _, err := bcrypt.Cost([]byte(passwordHash)); err != nil {
			return nil, fmt.Errorf("auth: the password hash is not a bcrypt hash (make one with: resticgui hash-password): %w", err)
		}
		if username == "" {
			return nil, errors.New("auth: a username is required with a password")
		}
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("resticgui session\x00" + username + "\x00" + passwordHash))
	return &Auth{Username: username, PasswordHash: passwordHash, SessionTTL: 7 * 24 * time.Hour, key: mac.Sum(nil), now: time.Now}, nil
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
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(a.Username)) == 1
	hash := []byte(a.PasswordHash)
	if !userOK {
		hash = dummyHash
	}
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil && userOK
}

func (a *Auth) sign(payload string) string {
	mac := hmac.New(sha256.New, a.key)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// token is "expiry-unix.signature".
func (a *Auth) token(expires time.Time) string {
	payload := strconv.FormatInt(expires.Unix(), 10)
	return payload + "." + a.sign(payload)
}

func (a *Auth) valid(token string) bool {
	payload, sig, ok := strings.Cut(token, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(a.sign(payload))) {
		return false
	}
	expires, err := strconv.ParseInt(payload, 10, 64)
	return err == nil && a.now().Unix() < expires
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
	if err := c.ShouldBindJSON(&in); err != nil || !a.check(in.Username, in.Password) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "wrong username or password"})
		return
	}
	a.setCookie(c, a.token(a.now().Add(a.SessionTTL)), int(a.SessionTTL.Seconds()))
	c.Status(http.StatusNoContent)
}

func (a *Auth) logout(c *gin.Context) {
	a.setCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (a *Auth) authenticated(c *gin.Context) bool {
	if !a.Enabled() {
		return true
	}
	cookie, err := c.Cookie(CookieName)
	return err == nil && a.valid(cookie)
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
