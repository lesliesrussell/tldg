package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leslierussell/tldg/internal/scan"
)

// tldg-eca

func TestBuildBundleBudgetAndCitations(t *testing.T) {
	dir := t.TempDir()
	var files []scan.File
	for _, name := range []string{"README.md", "main.go", "util.go"} {
		p := filepath.Join(dir, name)
		os.WriteFile(p, []byte("line one\nline two\nline three\n"), 0o644)
		files = append(files, scan.File{RelPath: name, AbsPath: p, Category: scan.CatSource})
	}
	b := BuildBundle(files, BundleOptions{MaxChunks: 2, MaxChunkTokens: 1200, Ref: "HEAD"})
	if len(b.Snippets) != 2 {
		t.Fatalf("expected budget of 2 snippets, got %d", len(b.Snippets))
	}
	// Citations must be well-formed and evidence IDs stable/sequential.
	if b.Snippets[0].Evidence.ID != "ev_001" || b.Snippets[1].Evidence.ID != "ev_002" {
		t.Errorf("unexpected evidence ids: %s, %s", b.Snippets[0].Evidence.ID, b.Snippets[1].Evidence.ID)
	}
	if b.Snippets[0].Evidence.Citation != "local:README.md:1-3" {
		t.Errorf("citation = %q", b.Snippets[0].Evidence.Citation)
	}
	// The produced bundle validates its own citations.
	text := "see [" + b.Snippets[0].Evidence.Citation + "]"
	if unknown := Validate(text, b.Evidences()); len(unknown) != 0 {
		t.Errorf("own citation flagged unknown: %v", unknown)
	}
}

func TestBuildBundleSkipsEmpty(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.txt")
	os.WriteFile(empty, []byte("   \n"), 0o644)
	files := []scan.File{{RelPath: "empty.txt", AbsPath: empty, Category: scan.CatDoc}}
	b := BuildBundle(files, BundleOptions{MaxChunks: 5})
	if len(b.Snippets) != 0 {
		t.Errorf("expected empty file skipped, got %d snippets", len(b.Snippets))
	}
}
