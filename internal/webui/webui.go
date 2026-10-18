// Package webui serves the frontend embedded in the binary.
package webui

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
)

//go:embed ui/*
var assets embed.FS

// UI is the embedded frontend
var UI, _ = fs.Sub(assets, "ui")

// Fallback is the page SvelteKit renders every route from; see front/svelte.config.js.
const Fallback = "index.html"

// Handler serves files from ui under prefix. A path that is not a file gets
// the fallback page, so that a reload or a bookmark of a frontend route such
// as /ui/repositories/3 opens that route.
func Handler(ui fs.FS, prefix string) http.Handler {
	files := http.FileServerFS(ui)
	page, _ := fs.ReadFile(ui, Fallback)
	csp := CSP(page)
	return http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", csp)
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if info, err := fs.Stat(ui, name); name == "" || err != nil || info.IsDir() {
			serveFallback(w, r, ui)
			return
		}
		if strings.HasPrefix(name, "_app/immutable/") {
			// file names carry a content hash
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	}))
}

// inlineScript matches the scripts written in a page rather than loaded.
var inlineScript = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// CSP returns the Content-Security-Policy of the interface: its own files
// only, plus the inline scripts of page (SvelteKit's start-up and the theme
// script), allowed by their hashes. Svelte sets style attributes, hence
// 'unsafe-inline' for styles only.
func CSP(page []byte) string {
	scripts := []string{"'self'"}
	for _, m := range inlineScript.FindAllSubmatch(page, -1) {
		sum := sha256.Sum256(m[1])
		scripts = append(scripts, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
	}
	return strings.Join([]string{
		"default-src 'self'",
		"script-src " + strings.Join(scripts, " "),
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data:",
		"font-src 'self' data:",
		"connect-src 'self'",
		"object-src 'none'",
		"base-uri 'none'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")
}

func serveFallback(w http.ResponseWriter, r *http.Request, ui fs.FS) {
	page, err := fs.ReadFile(ui, Fallback)
	if err != nil {
		http.Error(w, "the web interface was not built into this binary (make buildweb)", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if r.Method != http.MethodHead {
		_, _ = w.Write(page)
	}
}
