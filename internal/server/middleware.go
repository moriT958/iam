package server

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

func accessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health checks run every few seconds and only add noise.
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		// Client errors (e.g. bots probing for /wp-login.php) are routine on a
		// public site, so only server errors are raised above INFO.
		level := slog.LevelInfo
		if rec.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		logger.Log(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Int("bytes", rec.bytes),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
		)
	})
}

func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic", slog.Any("panic", v), slog.String("path", r.URL.Path))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// securityHeaders adds the response headers that harden every response, API and
// static files alike. The site has no cookies of its own to protect and embeds
// nothing, so these are safe defaults rather than a tuned policy.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		// Browsers must trust the Content-Type we send instead of guessing.
		h.Set("X-Content-Type-Options", "nosniff")
		// Do not leak full URLs of this site to other sites.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Nobody has a reason to frame this site (clickjacking).
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}
