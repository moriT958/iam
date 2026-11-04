package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

const (
	cacheImmutable  = "public, max-age=31536000, immutable"
	cacheRevalidate = "public, no-cache"
)

// immutablePrefixes are paths whose file names are content-hashed or never
// change, so browsers and Cloudflare may cache them forever.
var immutablePrefixes = []string{"/_astro/", "/fonts/", "/images/", "/favicon."}

func init() {
	// The platform mime database is not guaranteed to exist in a distroless image.
	_ = mime.AddExtensionType(".woff2", "font/woff2")
	_ = mime.AddExtensionType(".ico", "image/x-icon")
	_ = mime.AddExtensionType(".webp", "image/webp")
	_ = mime.AddExtensionType(".svg", "image/svg+xml")
	_ = mime.AddExtensionType(".xml", "application/xml")
}

// staticHandler serves a built Astro site. It resolves requests like nginx's
// `try_files $uri $uri/ =404`: a file is served as is, a directory serves its
// index.html, and anything else is a 404 (using 404.html when present).
type staticHandler struct {
	root fs.FS
}

func newStaticHandler(root fs.FS) http.Handler {
	return &staticHandler{root: root}
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// path.Clean on a rooted path removes any "..", so the result cannot escape root.
	urlPath := path.Clean("/" + r.URL.Path)
	name, ok := h.resolve(urlPath)
	if !ok {
		h.notFound(w, r)
		return
	}
	h.serveFile(w, r, urlPath, name, http.StatusOK)
}

func (h *staticHandler) resolve(urlPath string) (string, bool) {
	name := strings.TrimPrefix(urlPath, "/")
	if name == "" {
		name = "."
	}
	info, err := fs.Stat(h.root, name)
	if err != nil {
		return "", false
	}
	if !info.IsDir() {
		return name, true
	}
	index := path.Join(name, "index.html")
	if info, err := fs.Stat(h.root, index); err == nil && !info.IsDir() {
		return index, true
	}
	return "", false
}

func (h *staticHandler) notFound(w http.ResponseWriter, r *http.Request) {
	if _, err := fs.Stat(h.root, "404.html"); err != nil {
		http.NotFound(w, r)
		return
	}
	h.serveFile(w, r, "/404.html", "404.html", http.StatusNotFound)
}

func (h *staticHandler) serveFile(w http.ResponseWriter, r *http.Request, urlPath, name string, status int) {
	data, err := fs.ReadFile(h.root, name)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	hdr := w.Header()
	hdr.Set("Cache-Control", cacheControl(urlPath))
	if status == http.StatusOK {
		sum := sha256.Sum256(data)
		hdr.Set("ETag", `"`+hex.EncodeToString(sum[:8])+`"`)
		// ServeContent handles If-None-Match, Range, HEAD and Content-Type.
		http.ServeContent(w, r, name, zeroTime, bytes.NewReader(data))
		return
	}

	if ct := mime.TypeByExtension(path.Ext(name)); ct != "" {
		hdr.Set("Content-Type", ct)
	}
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func cacheControl(urlPath string) string {
	for _, p := range immutablePrefixes {
		if strings.HasPrefix(urlPath, p) {
			return cacheImmutable
		}
	}
	return cacheRevalidate
}
