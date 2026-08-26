package evidence

import (
	"testing"
	"time"
)

// tldg-5xh

func TestCitation(t *testing.T) {
	cases := []struct {
		scheme, ref    string
		start, end     int
		want           string
	}{
		{"local", "README.md", 10, 44, "local:README.md:10-44"},
		{"local", "go.mod", 0, 0, "local:go.mod"},
		{"local", "main.go", 5, 5, "local:main.go:5"},
		{"git", "abc1234", 0, 0, "git:abc1234"},
	}
	for _, c := range cases {
		if got := Citation(c.scheme, c.ref, c.start, c.end); got != c.want {
			t.Errorf("Citation(%q,%q,%d,%d) = %q, want %q", c.scheme, c.ref, c.start, c.end, got, c.want)
		}
	}
}

func TestLocalFileCitation(t *testing.T) {
	e := LocalFile("ev_1", "cmd/tldg/main.go", "HEAD", 1, 114, time.Now())
	if e.Citation != "local:cmd/tldg/main.go:1-114" {
		t.Errorf("citation = %q", e.Citation)
	}
	if e.Kind != KindLocalFile || e.URI != "file:cmd/tldg/main.go" {
		t.Errorf("unexpected evidence: %+v", e)
	}
}

func TestExtractAndValidate(t *testing.T) {
	bundle := []Evidence{
		{Citation: "local:README.md:1-10"},
		{Citation: "git:abc1234"},
	}
	text := "The entry point is here [local:README.md:1-10] and it changed in [git:abc1234]. " +
		"But this one is fake [local:GHOST.go:1-2]."
	got := ExtractCitations(text)
	if len(got) != 3 {
		t.Fatalf("extracted %d citations, want 3: %v", len(got), got)
	}
	unknown := Validate(text, bundle)
	if len(unknown) != 1 || unknown[0] != "local:GHOST.go:1-2" {
		t.Errorf("unknown = %v, want [local:GHOST.go:1-2]", unknown)
	}
}
