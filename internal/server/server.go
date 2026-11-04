// Package server wires the HTTP handlers of the iam application.
package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"
)

var zeroTime time.Time

// New returns the application handler. site is the built Astro site.
func New(site fs.FS, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.Handle("/", newStaticHandler(site))

	return recoverer(logger, accessLog(logger, mux))
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}
