package server

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/mathieuHa/resticgui/internal/auth"
)

// TestOpenAPIMatchesRoutes keeps the document and the router in step: every
// route is documented and every documented operation exists.
func TestOpenAPIMatchesRoutes(t *testing.T) {
	var doc struct {
		Paths      map[string]map[string]any `yaml:"paths"`
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(openAPI, &doc))

	documented := map[string]bool{}
	for path, item := range doc.Paths {
		for method := range item {
			if method != "parameters" {
				documented[strings.ToUpper(method)+" "+path] = true
			}
		}
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	a, err := auth.New("", "", []byte("k"), nil)
	require.NoError(t, err)
	(&Server{Auth: a}).Register(r.Group(BasePath))
	param := regexp.MustCompile(`:(\w+)`)
	routed := map[string]bool{}
	for _, route := range r.Routes() {
		path := param.ReplaceAllString(strings.TrimPrefix(route.Path, BasePath), "{$1}")
		routed[route.Method+" "+path] = true
	}

	require.Greater(t, len(routed), 20)
	assert.Equal(t, keys(routed), keys(documented))

	// every $ref points at something defined
	for _, ref := range regexp.MustCompile(`#/components/schemas/(\w+)`).FindAllStringSubmatch(string(openAPI), -1) {
		assert.Contains(t, doc.Components.Schemas, ref[1])
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestOpenAPIIsServed(t *testing.T) {
	a := newAPI(t)
	status, body := a.do(http.MethodGet, "/openapi.yaml", nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(body), "openapi: 3.1.0")
}

func TestAuthGuardsTheAPI(t *testing.T) {
	hash, err := auth.HashPassword("long enough password")
	require.NoError(t, err)
	guard, err := auth.New("admin", hash, []byte("k"), auth.NewMemorySessions())
	require.NoError(t, err)
	a := newAPI(t, func(s *Server) { s.Auth = guard })

	for _, p := range []string{"/repositories", "/backends", "/fs", "/jobs?repository=1", "/events"} {
		status, _ := a.do(http.MethodGet, p, nil)
		assert.Equal(t, http.StatusUnauthorized, status, p)
	}
	status, _ := a.do(http.MethodGet, "/openapi.yaml", nil)
	assert.Equal(t, http.StatusOK, status)
	status, _ = a.do(http.MethodGet, "/version", nil)
	assert.Equal(t, http.StatusOK, status)
	status, body := a.do(http.MethodGet, "/auth/session", nil)
	assert.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{"auth_enabled":true,"authenticated":false}`, string(body))
}

func TestVersion(t *testing.T) {
	a := newAPI(t, func(s *Server) { s.Version = "1.2.3" })
	status, body := a.do(http.MethodGet, "/version", nil)
	assert.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{"version":"1.2.3"}`, string(body))
}
