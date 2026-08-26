package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesliesrussell/tldg/internal/config"
)

// tldg-eca

// writeFixture creates a minimal Go project in a fresh temp dir (outside any
// Git worktree) and returns its path.
func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":            "module example.com/demo\n\ngo 1.26\n\nrequire github.com/spf13/cobra v1.8.0\n",
		"README.md":         "# demo\n\nA tiny example CLI fixture.\n",
		"cmd/demo/main.go":  "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hi\") }\n",
		"main_test.go":      "package demo\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\n",
	}
	for rel, content := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestRunSummaryOfflineProfile(t *testing.T) {
	dir := writeFixture(t)
	cfg := config.Default()
	cfg.Index.Root = t.TempDir() // isolate index output

	res, err := RunSummary(context.Background(), cfg, SummaryOptions{
		Target:  dir,
		Offline: true,
	})
	if err != nil {
		t.Fatalf("RunSummary: %v", err)
	}
	if res.Profile == nil {
		t.Fatal("expected profile attached")
	}
	if got := res.Profile.PrimaryLanguage(); got != "Go" {
		t.Errorf("primary language = %q, want Go", got)
	}
	if !res.Profile.HasTests {
		t.Error("expected HasTests true")
	}
	if len(res.Profile.Dependencies) == 0 {
		t.Error("expected at least one dependency")
	}
	if res.Profile.SelectedFiles == 0 {
		t.Error("expected selected files > 0")
	}
	if len(res.Evidence) == 0 {
		t.Error("expected evidence records")
	}
	// Offline answer must carry the skipped-synthesis warning.
	foundWarn := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "offline") {
			foundWarn = true
		}
	}
	if !foundWarn {
		t.Errorf("expected offline warning, got %v", res.Warnings)
	}
}

func TestRunSummaryNoIndex(t *testing.T) {
	dir := writeFixture(t)
	cfg := config.Default()
	cfg.Index.Root = t.TempDir()

	res, err := RunSummary(context.Background(), cfg, SummaryOptions{
		Target:  dir,
		Offline: true,
		NoIndex: true,
	})
	if err != nil {
		t.Fatalf("RunSummary: %v", err)
	}
	// With --no-index no index directory should be created.
	entries, _ := os.ReadDir(cfg.Index.Root)
	if len(entries) != 0 {
		t.Errorf("expected no index output with --no-index, got %d entries", len(entries))
	}
	_ = res
}
