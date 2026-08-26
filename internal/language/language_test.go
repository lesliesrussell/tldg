package language

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesliesrussell/tldg/internal/scan"
)

// tldg-eca

func TestDetectGoWithDeps(t *testing.T) {
	dir := t.TempDir()
	gomod := "module example.com/x\n\ngo 1.26\n\nrequire (\n\tgithub.com/spf13/cobra v1.8.0\n\tgopkg.in/yaml.v3 v3.0.1 // indirect\n)\n"
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644)

	files := []scan.File{
		{RelPath: "go.mod", AbsPath: filepath.Join(dir, "go.mod"), Category: scan.CatManifest},
		{RelPath: "main.go", AbsPath: filepath.Join(dir, "main.go"), Category: scan.CatEntrypoint},
		{RelPath: "util.go", AbsPath: filepath.Join(dir, "util.go"), Category: scan.CatSource},
	}
	langs, deps := Detect(dir, files)
	if len(langs) == 0 || langs[0].Name != "Go" || !langs[0].Primary {
		t.Fatalf("expected primary Go, got %+v", langs)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 go deps, got %d: %+v", len(deps), deps)
	}
	if deps[0].Ecosystem != "go" || deps[0].Name != "github.com/spf13/cobra" {
		t.Errorf("unexpected dep: %+v", deps[0])
	}
}

func TestDetectPolyglot(t *testing.T) {
	dir := t.TempDir()
	files := []scan.File{
		{RelPath: "a.go", AbsPath: filepath.Join(dir, "a.go")},
		{RelPath: "b.go", AbsPath: filepath.Join(dir, "b.go")},
		{RelPath: "c.py", AbsPath: filepath.Join(dir, "c.py")},
	}
	langs, _ := Detect(dir, files)
	if len(langs) != 2 {
		t.Fatalf("expected 2 languages, got %+v", langs)
	}
	if langs[0].Name != "Go" { // more files → primary
		t.Errorf("expected Go primary, got %q", langs[0].Name)
	}
}
