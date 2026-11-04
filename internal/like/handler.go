package like

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
)

// Handler serves the like API:
//
//	GET    /api/likes          counters of every post
//	GET    /api/likes/{slug}   counter of one post
//	POST   /api/likes/{slug}   add a like
//	DELETE /api/likes/{slug}   remove a like
type Handler struct {
	store   Store
	slugs   Slugs
	limiter *Limiter
	logger  *slog.Logger
}

// NewHandler returns a Handler. limiter may be nil to disable rate limiting.
func NewHandler(store Store, slugs Slugs, limiter *Limiter, logger *slog.Logger) *Handler {
	return &Handler{store: store, slugs: slugs, limiter: limiter, logger: logger}
}

// Register mounts the API on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/likes", h.list)
	mux.HandleFunc("GET /api/likes/{slug}", h.get)
	mux.HandleFunc("POST /api/likes/{slug}", h.mutate(h.store.Increment))
	mux.HandleFunc("DELETE /api/likes/{slug}", h.mutate(h.store.Decrement))
}

type countResponse struct {
	Count int64 `json:"count"`
}

type countsResponse struct {
	Counts map[string]int64 `json:"counts"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	stored, err := h.store.Counts(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	// Report every known post, so clients never have to treat "missing" specially.
	counts := make(map[string]int64, len(h.slugs))
	for slug := range h.slugs {
		counts[slug] = stored[slug]
	}
	writeJSON(w, http.StatusOK, countsResponse{Counts: counts})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !h.slugs.Has(slug) {
		writeError(w, http.StatusNotFound, "unknown post")
		return
	}
	n, err := h.store.Count(r.Context(), slug)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, countResponse{Count: n})
}

func (h *Handler) mutate(op func(ctx context.Context, slug string) (int64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if !h.slugs.Has(slug) {
			writeError(w, http.StatusNotFound, "unknown post")
			return
		}
		if h.limiter != nil && !h.limiter.Allow(clientIP(r)) {
			w.Header().Set("Retry-After", strconv.Itoa(5))
			writeError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		n, err := op(r.Context(), slug)
		if err != nil {
			h.fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, countResponse{Count: n})
	}
}

func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "like store failed",
		slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Any("error", err))
	writeError(w, http.StatusInternalServerError, "internal server error")
}

// clientIP returns the visitor's address. The app is reachable only through
// Cloudflare -> cloudflared -> Traefik (NetworkPolicy blocks everything else),
// and Cloudflare overwrites CF-Connecting-IP, so the header can be trusted.
// Without it (local development) the peer address is used.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Counters change at any time and Cloudflare caches by the origin's Cache-Control.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
