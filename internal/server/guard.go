package server

import (
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Guard rejects the API requests that a web page on another site could make
// the browser send: cross-site requests that change something (CSRF), and,
// when there is no login, requests for a host name that is not this server
// (DNS rebinding, where an attacker's domain resolves to 127.0.0.1).
//
// allowedHosts are the names the server is reached by, besides IP addresses
// and localhost; only consulted when authOff.
func Guard(allowedHosts []string, authOff bool) gin.HandlerFunc {
	allowed := map[string]bool{"localhost": true}
	for _, h := range allowedHosts {
		allowed[strings.ToLower(strings.TrimSpace(h))] = true
	}
	return func(c *gin.Context) {
		r := c.Request
		if authOff && !hostAllowed(r.Host, allowed) {
			fail(c, http.StatusMisdirectedRequest, "unknown host name: add it to server.allowed_hosts, or set a password")
			return
		}
		if safeMethod(r.Method) {
			c.Next()
			return
		}
		// browsers say where a request comes from in Sec-Fetch-Site; older
		// ones only send Origin, compared with the host the client asked for
		if site := r.Header.Get("Sec-Fetch-Site"); site != "" {
			if site != "same-origin" && site != "none" {
				fail(c, http.StatusForbidden, "cross-site request refused")
				return
			}
		} else if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || !(strings.EqualFold(u.Host, r.Host) || strings.EqualFold(u.Host, r.Header.Get("X-Forwarded-Host"))) {
				fail(c, http.StatusForbidden, "cross-site request refused")
				return
			}
		}
		// a page can send text/plain or a form without asking; JSON needs
		// a preflight that the API never grants
		if r.ContentLength != 0 {
			mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if mediaType != "application/json" {
				fail(c, http.StatusUnsupportedMediaType, "the body must be application/json")
				return
			}
		}
		c.Next()
	}
}

func safeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

// hostAllowed accepts IP addresses (a rebinding attack needs a domain name),
// localhost and the configured names, with any port.
func hostAllowed(hostport string, allowed map[string]bool) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSuffix(strings.Trim(host, "[]"), "."))
	if net.ParseIP(host) != nil || allowed[host] {
		return true
	}
	return strings.HasSuffix(host, ".localhost")
}

// apiCSP keeps API responses, downloads included, from running anything
// should a browser render one.
const apiCSP = "default-src 'none'; frame-ancestors 'none'; sandbox"

// SecurityHeaders sets the headers every response carries. Pages of the web
// interface set their own Content-Security-Policy.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		if strings.HasPrefix(c.Request.URL.Path, BasePath) || c.Request.URL.Path == "/metrics" {
			h.Set("Content-Security-Policy", apiCSP)
		}
		c.Next()
	}
}
