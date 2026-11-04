package like

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func newTestHandler(t *testing.T, limiter *Limiter) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	slugs := Slugs{"hello": {}, "world": {}}
	NewHandler(openTestStore(t), slugs, limiter, slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return mux
}

func call(h http.Handler, method, path, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if ip != "" {
		req.Header.Set("CF-Connecting-IP", ip)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestLikeAndUnlike(t *testing.T) {
	h := newTestHandler(t, nil)

	if got := decode[countResponse](t, call(h, "GET", "/api/likes/hello", "")); got.Count != 0 {
		t.Fatalf("initial count = %d, want 0", got.Count)
	}
	for want := int64(1); want <= 2; want++ {
		rec := call(h, "POST", "/api/likes/hello", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("POST status = %d", rec.Code)
		}
		if got := decode[countResponse](t, rec); got.Count != want {
			t.Fatalf("POST count = %d, want %d", got.Count, want)
		}
	}
	rec := call(h, "DELETE", "/api/likes/hello", "")
	if got := decode[countResponse](t, rec); rec.Code != http.StatusOK || got.Count != 1 {
		t.Fatalf("DELETE = %d / %d, want 200 / 1", rec.Code, got.Count)
	}
	if got := decode[countResponse](t, call(h, "GET", "/api/likes/hello", "")); got.Count != 1 {
		t.Fatalf("count after delete = %d, want 1", got.Count)
	}
}

func TestUnlikeNeverGoesNegative(t *testing.T) {
	h := newTestHandler(t, nil)
	rec := call(h, "DELETE", "/api/likes/hello", "")
	if got := decode[countResponse](t, rec); rec.Code != http.StatusOK || got.Count != 0 {
		t.Fatalf("DELETE on unliked = %d / %d, want 200 / 0", rec.Code, got.Count)
	}
}

func TestUnknownSlug(t *testing.T) {
	h := newTestHandler(t, nil)
	for _, method := range []string{"GET", "POST", "DELETE"} {
		if rec := call(h, method, "/api/likes/nope", ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s unknown slug: status = %d, want 404", method, rec.Code)
		}
	}
	// Slugs that are not even well-formed must not reach the database either.
	for _, p := range []string{"/api/likes/..%2Fetc", "/api/likes/a%20b", "/api/likes/hello;drop"} {
		if rec := call(h, "POST", p, ""); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: status = %d, want 404", p, rec.Code)
		}
	}
}

func TestListIncludesEveryPost(t *testing.T) {
	h := newTestHandler(t, nil)
	call(h, "POST", "/api/likes/hello", "")
	call(h, "POST", "/api/likes/hello", "")

	rec := call(h, "GET", "/api/likes", "")
	got := decode[countsResponse](t, rec)
	if rec.Code != http.StatusOK || len(got.Counts) != 2 || got.Counts["hello"] != 2 || got.Counts["world"] != 0 {
		t.Fatalf("list = %d %v; want 200 map[hello:2 world:0]", rec.Code, got.Counts)
	}
}

func TestResponsesAreNotCacheable(t *testing.T) {
	h := newTestHandler(t, nil)
	for _, tt := range []struct{ method, path string }{
		{"GET", "/api/likes"},
		{"GET", "/api/likes/hello"},
		{"POST", "/api/likes/hello"},
		{"GET", "/api/likes/nope"},
	} {
		rec := call(h, tt.method, tt.path, "")
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s %s: Cache-Control = %q, want no-store", tt.method, tt.path, got)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Errorf("%s %s: Content-Type = %q", tt.method, tt.path, ct)
		}
	}
}

func TestRateLimitPerClient(t *testing.T) {
	// Burst of 2 and a refill so slow it never happens during the test.
	h := newTestHandler(t, NewLimiter(rate.Limit(0.0001), 2))

	for i := range 2 {
		if rec := call(h, "POST", "/api/likes/hello", "203.0.113.1"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rec.Code)
		}
	}
	rec := call(h, "POST", "/api/likes/hello", "203.0.113.1")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("Retry-After is empty")
	}
	// DELETE shares the bucket with POST.
	if rec := call(h, "DELETE", "/api/likes/hello", "203.0.113.1"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("DELETE status = %d, want 429", rec.Code)
	}
	// Another visitor is unaffected.
	if rec := call(h, "POST", "/api/likes/hello", "203.0.113.2"); rec.Code != http.StatusOK {
		t.Fatalf("other client: status = %d, want 200", rec.Code)
	}
	// Reads are never limited.
	if rec := call(h, "GET", "/api/likes/hello", "203.0.113.1"); rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", rec.Code)
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.5:4242"
	if got := clientIP(req); got != "10.0.0.5" {
		t.Errorf("without header: %q, want 10.0.0.5", got)
	}
	req.Header.Set("CF-Connecting-IP", "198.51.100.7")
	if got := clientIP(req); got != "198.51.100.7" {
		t.Errorf("with header: %q, want 198.51.100.7", got)
	}
}
