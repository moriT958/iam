package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func newTestServer() http.Handler {
	site := fstest.MapFS{
		"index.html":            {Data: []byte("<h1>home</h1>")},
		"posts/foo/index.html":  {Data: []byte("<h1>foo</h1>")},
		"_astro/app.abc123.css": {Data: []byte("body{}")},
		"fonts/plemol.woff2":    {Data: []byte("wOF2....")},
		"images/me.webp":        {Data: []byte("RIFF....WEBP")},
		"favicon.svg":           {Data: []byte("<svg/>")},
		"rss.xml":               {Data: []byte("<rss/>")},
	}
	return New(Config{Site: site, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
}

func do(t *testing.T, h http.Handler, method, target string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestServesFilesLikeTryFiles(t *testing.T) {
	h := newTestServer()
	tests := []struct {
		path string
		body string
	}{
		{"/", "<h1>home</h1>"},
		{"/index.html", "<h1>home</h1>"},
		{"/posts/foo/", "<h1>foo</h1>"},
		{"/posts/foo", "<h1>foo</h1>"},
		{"/rss.xml", "<rss/>"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, tt.path, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Body.String(); got != tt.body {
				t.Fatalf("body = %q, want %q", got, tt.body)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	h := newTestServer()
	for _, p := range []string{"/nope", "/posts/", "/posts/missing/", "/api/likes/foo"} {
		if rec := do(t, h, http.MethodGet, p, nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", p, rec.Code)
		}
	}
}

func TestCustom404Page(t *testing.T) {
	site := fstest.MapFS{"404.html": {Data: []byte("<h1>not found</h1>")}}
	h := New(Config{Site: site, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	rec := do(t, h, http.MethodGet, "/missing", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if got := rec.Body.String(); got != "<h1>not found</h1>" {
		t.Fatalf("body = %q", got)
	}
}

func TestPathTraversalIsContained(t *testing.T) {
	h := newTestServer()
	// ServeMux redirects to the cleaned path; following it must not reach outside the site.
	rec := do(t, h, http.MethodGet, "/posts/../../etc/passwd", nil)
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/etc/passwd" {
		t.Fatalf("Location = %q, want /etc/passwd", loc)
	}
	if rec := do(t, h, http.MethodGet, loc, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestServer()
	if rec := do(t, h, http.MethodPost, "/", nil); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestCacheControl(t *testing.T) {
	h := newTestServer()
	const immutable = "public, max-age=31536000, immutable"
	tests := []struct {
		path string
		want string
	}{
		{"/", "public, no-cache"},
		{"/posts/foo/", "public, no-cache"},
		{"/rss.xml", "public, no-cache"},
		{"/_astro/app.abc123.css", immutable},
		{"/fonts/plemol.woff2", immutable},
		{"/images/me.webp", immutable},
		{"/favicon.svg", immutable},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, tt.path, nil)
			if got := rec.Header().Get("Cache-Control"); got != tt.want {
				t.Fatalf("Cache-Control = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContentType(t *testing.T) {
	h := newTestServer()
	tests := map[string]string{
		"/":                      "text/html",
		"/_astro/app.abc123.css": "text/css",
		"/fonts/plemol.woff2":    "font/woff2",
		"/images/me.webp":        "image/webp",
		"/favicon.svg":           "image/svg+xml",
	}
	for p, want := range tests {
		got := do(t, h, http.MethodGet, p, nil).Header().Get("Content-Type")
		if !strings.HasPrefix(got, want) {
			t.Errorf("%s: Content-Type = %q, want prefix %q", p, got, want)
		}
	}
}

func TestConditionalRequest(t *testing.T) {
	h := newTestServer()
	first := do(t, h, http.MethodGet, "/", nil)
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("ETag is empty")
	}
	second := do(t, h, http.MethodGet, "/", map[string]string{"If-None-Match": etag})
	if second.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", second.Code)
	}
}

func TestHead(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodHead, "/", nil)
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("status = %d, body len = %d; want 200 and empty body", rec.Code, rec.Body.Len())
	}
}

func TestHealthz(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestAccessLogLevel(t *testing.T) {
	tests := []struct {
		status int
		level  string
	}{
		{http.StatusOK, "INFO"},
		{http.StatusNotModified, "INFO"},
		{http.StatusNotFound, "INFO"},
		{http.StatusTooManyRequests, "INFO"},
		{http.StatusInternalServerError, "ERROR"},
		{http.StatusBadGateway, "ERROR"},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			var buf strings.Builder
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			h := accessLog(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			do(t, h, http.MethodGet, "/x", nil)
			if want := `"level":"` + tt.level + `"`; !strings.Contains(buf.String(), want) {
				t.Fatalf("log = %s; want %s", buf.String(), want)
			}
		})
	}
}

func TestAccessLogSkipsHealthz(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	h := accessLog(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	do(t, h, http.MethodGet, "/healthz", nil)
	if buf.Len() != 0 {
		t.Fatalf("healthz was logged: %s", buf.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := newTestServer()
	// Every kind of response gets them: pages, assets, 404s, 405s and /healthz.
	paths := []struct {
		method, path string
	}{
		{http.MethodGet, "/"},
		{http.MethodGet, "/_astro/app.abc123.css"},
		{http.MethodGet, "/missing"},
		{http.MethodPost, "/"},
		{http.MethodGet, "/healthz"},
	}
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"X-Frame-Options":        "DENY",
	}
	for _, p := range paths {
		rec := do(t, h, p.method, p.path, nil)
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s %s: %s = %q, want %q", p.method, p.path, k, got, v)
			}
		}
	}
}
