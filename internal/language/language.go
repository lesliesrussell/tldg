// Package language detects language ecosystems and extracts direct manifest
// dependencies (spec §9.4, §9.8). Detection is evidence-driven and may report a
// polyglot repository.
package language

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/leslierussell/tldg/internal/repo"
	"github.com/leslierussell/tldg/internal/scan"
)

// tldg-eca

// extLang maps a source extension to a language name.
var extLang = map[string]string{
	".go": "Go", ".rs": "Rust", ".ts": "TypeScript", ".tsx": "TypeScript",
	".js": "JavaScript", ".jsx": "JavaScript", ".py": "Python",
	".java": "Java", ".kt": "Kotlin", ".rb": "Ruby", ".php": "PHP",
	".cs": "C#", ".ex": "Elixir", ".exs": "Elixir", ".hs": "Haskell",
	".lua": "Lua", ".c": "C", ".cc": "C++", ".cpp": "C++", ".hpp": "C++",
	".zig": "Zig", ".sh": "Shell", ".swift": "Swift", ".scala": "Scala",
}

// Detect analyzes selected files and returns detected languages (primary first)
// and direct manifest dependencies. root is the repository root.
func Detect(root string, files []scan.File) ([]repo.Language, []repo.Dependency) {
	counts := map[string]int{}
	manifests := map[string][]string{} // language -> manifest paths
	var deps []repo.Dependency

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.RelPath))
		if lang, ok := extLang[ext]; ok {
			counts[lang]++
		}
		base := strings.ToLower(filepath.Base(f.RelPath))
		switch base {
		case "go.mod":
			manifests["Go"] = append(manifests["Go"], f.RelPath)
			deps = append(deps, parseGoMod(f.AbsPath, f.RelPath)...)
		case "package.json":
			manifests["JavaScript"] = append(manifests["JavaScript"], f.RelPath)
			deps = append(deps, parsePackageJSON(f.AbsPath, f.RelPath)...)
		case "cargo.toml":
			manifests["Rust"] = append(manifests["Rust"], f.RelPath)
			deps = append(deps, parseCargoToml(f.AbsPath, f.RelPath)...)
		case "requirements.txt":
			manifests["Python"] = append(manifests["Python"], f.RelPath)
			deps = append(deps, parseRequirements(f.AbsPath, f.RelPath)...)
		case "pyproject.toml":
			manifests["Python"] = append(manifests["Python"], f.RelPath)
		}
	}

	// Rank languages: manifest presence boosts, otherwise by file count.
	type lc struct {
		name  string
		files int
	}
	var ranked []lc
	seen := map[string]bool{}
	for name, n := range counts {
		ranked = append(ranked, lc{name, n})
		seen[name] = true
	}
	for name := range manifests {
		if !seen[name] {
			ranked = append(ranked, lc{name, 0})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].files != ranked[j].files {
			return ranked[i].files > ranked[j].files
		}
		return ranked[i].name < ranked[j].name
	})

	var langs []repo.Language
	for i, r := range ranked {
		langs = append(langs, repo.Language{
			Name:      r.name,
			Files:     r.files,
			Manifests: manifests[r.name],
			Primary:   i == 0,
		})
	}
	return langs, deps
}

func readFile(abs string) (string, bool) {
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", false
	}
	return string(b), true
}
