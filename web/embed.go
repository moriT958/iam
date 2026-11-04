// Package web embeds the statically built Astro site (web/dist).
//
// dist is produced by `pnpm build` and is not tracked by git, so the frontend
// must be built before this package (and therefore the Go binary) compiles.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist returns the built site rooted at dist/.
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
