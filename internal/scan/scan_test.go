package scan

import (
	"os"
	"path/filepath"
	"testing"
)

// tldg-eca

func TestClassify(t *testing.T) {
	cases := []struct {
		path string
		want Category
	}{
		{"README.md", CatReadme},
		{"LICENSE", CatLicense},
		{"CHANGELOG.md", CatChangelog},
		{"go.mod", CatManifest},
		{"go.sum", CatLockfile},
		{".github/workflows/ci.yml", CatCI},
		{"Dockerfile", CatDocker},
		{"docs/adr/0001-x.md", CatADR},
		{"cmd/tldg/main.go", CatEntrypoint},
		{"internal/foo/foo_test.go", CatTest},
		{"internal/foo/foo.go", CatSource},
		{"config.yaml", CatConfig},
	}
	for _, c := range cases {
		if got, _ := Classify(c.path); got != c.want {
			t.Errorf("Classify(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestSelectFilesFiltersBinaryAndOversize(t *testing.T) {
	dir := t.TempDir()
	write := func(rel string, data []byte) {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, data, 0o644)
	}
	write("README.md", []byte("# hi\n"))
	write("bin.dat", []byte{0x00, 0x01, 0x02, 0x00}) // NUL → binary
	write("big.txt", make([]byte, 4096))              // 4KB
	write("vendor/dep.go", []byte("package dep\n"))   // ignored segment

	res := SelectFiles(dir, []string{"README.md", "bin.dat", "big.txt", "vendor/dep.go"}, Options{
		IgnoreNames:   []string{"vendor"},
		MaxFileSizeKB: 2, // 2KB cap → big.txt skipped
	})
	if len(res.Selected) != 1 || res.Selected[0].RelPath != "README.md" {
		t.Fatalf("expected only README.md selected, got %+v", res.Selected)
	}
	if !res.Has(CatReadme) {
		t.Error("expected readme category present")
	}
}
