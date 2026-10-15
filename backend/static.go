package main

import (
	"net/http"
	"os"
	"path/filepath"
)

// registerStatic mounts the built frontend (web/) at "/" with SPA fallback.
// More specific routes (/api/*, /health) are already registered on the mux and
// take precedence over this catch-all handler.
func (a *app) registerStatic(mux *http.ServeMux) {
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		if exe, err := os.Executable(); err == nil {
			webDir = filepath.Join(filepath.Dir(exe), "..", "web")
		}
	}
	if webDir == "" {
		return
	}
	if st, err := os.Stat(webDir); err != nil || !st.IsDir() {
		return
	}

	fileServer := http.FileServer(http.Dir(webDir))
	indexPath := filepath.Join(webDir, "index.html")

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			clean := filepath.Clean(r.URL.Path)
			p := filepath.Join(webDir, filepath.FromSlash(clean))
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// SPA fallback: unknown paths serve index.html for client-side routing.
		http.ServeFile(w, r, indexPath)
	}))
}
