package like

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
)

// slugPattern is deliberately strict: slugs come from file names under
// web/src/content/posts and end up in the URL path and the database.
var slugPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

// Slugs is the set of posts that can be liked.
type Slugs map[string]struct{}

// Has reports whether slug is a known post.
func (s Slugs) Has(slug string) bool {
	_, ok := s[slug]
	return ok
}

// LoadSlugs reads the manifest that Astro emits at build time
// (web/src/pages/posts.json.ts), so the server accepts likes only for real posts.
func LoadSlugs(fsys fs.FS, name string) (Slugs, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	var manifest struct {
		Slugs []string `json:"slugs"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	slugs := make(Slugs, len(manifest.Slugs))
	for _, slug := range manifest.Slugs {
		if !slugPattern.MatchString(slug) {
			return nil, fmt.Errorf("%s: invalid slug %q", name, slug)
		}
		slugs[slug] = struct{}{}
	}
	return slugs, nil
}
