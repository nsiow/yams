package ui

import (
	"io/fs"
	"net/http"
)

// Handler returns an http.Handler that serves the embedded UI under /ui/.
// If the binary was built without UI support, all requests return a JSON error.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return noUIHandler()
	}

	// Check if the embedded FS actually has files
	entries, err := fs.ReadDir(sub, ".")
	if err != nil || len(entries) == 0 {
		return noUIHandler()
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.StripPrefix("/ui/", spaHandler(fileServer, sub))
}

// spaHandler wraps a file server with SPA fallback: if a file is not found, serve index.html
func spaHandler(fileServer http.Handler, root fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to open the requested file
		path := r.URL.Path
		if path == "" || path == "/" {
			path = "index.html"
		}

		_, err := fs.Stat(root, path)
		if err != nil {
			// File not found — serve index.html for SPA routing
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}

// noUIHandler returns a handler that responds with a JSON error for all requests
func noUIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "binary was not built with UI support"}` + "\n"))
	})
}
