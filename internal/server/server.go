// Package server wires the HTTP handlers of the iam application.
package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/moriT958/iam/internal/like"
)

var zeroTime time.Time

// Config holds the dependencies of the application handler.
type Config struct {
	// Site is the built Astro site (web/dist).
	Site fs.FS
	// Likes serves /api/likes. Optional: the API is not mounted when nil.
	Likes *like.Handler
	// Logger receives access logs and errors.
	Logger *slog.Logger
}

// New returns the application handler.
func New(cfg Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	if cfg.Likes != nil {
		cfg.Likes.Register(mux)
	}
	mux.Handle("/", newStaticHandler(cfg.Site))

	return recoverer(cfg.Logger, accessLog(cfg.Logger, securityHeaders(mux)))
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}
