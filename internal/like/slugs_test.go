package like

import (
	"testing"
	"testing/fstest"
)

func TestLoadSlugs(t *testing.T) {
	fsys := fstest.MapFS{"posts.json": {Data: []byte(`{"slugs":["hello-world","react_useactionstate","a1"]}`)}}
	got, err := LoadSlugs(fsys, "posts.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || !got.Has("hello-world") || !got.Has("a1") || got.Has("other") {
		t.Fatalf("slugs = %v", got)
	}
}

func TestLoadSlugsErrors(t *testing.T) {
	tests := map[string]fstest.MapFS{
		"missing file": {},
		"invalid json": {"posts.json": {Data: []byte(`{`)}},
		"bad slug":     {"posts.json": {Data: []byte(`{"slugs":["ok","../x"]}`)}},
		"empty slug":   {"posts.json": {Data: []byte(`{"slugs":[""]}`)}},
		"leading dash": {"posts.json": {Data: []byte(`{"slugs":["-x"]}`)}},
		"has a space":  {"posts.json": {Data: []byte(`{"slugs":["a b"]}`)}},
		"non-ascii":    {"posts.json": {Data: []byte(`{"slugs":["日本語"]}`)}},
	}
	for name, fsys := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadSlugs(fsys, "posts.json"); err == nil {
				t.Fatal("want error, got nil")
			}
		})
	}
}
