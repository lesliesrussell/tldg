package scan

import (
	"path/filepath"
	"sort"
	"strings"
)

// tldg-eca

// sourceExts maps source-code extensions used for entrypoint/source detection.
var sourceExts = map[string]bool{
	".go": true, ".rs": true, ".ts": true, ".tsx": true, ".js": true,
	".jsx": true, ".py": true, ".java": true, ".kt": true, ".rb": true,
	".php": true, ".cs": true, ".ex": true, ".exs": true, ".hs": true,
	".lua": true, ".c": true, ".cc": true, ".cpp": true, ".h": true,
	".hpp": true, ".zig": true, ".sh": true, ".swift": true, ".scala": true,
}

// manifestNames maps well-known manifest/lockfile basenames to categories.
var manifestNames = map[string]Category{
	"go.mod": CatManifest, "go.work": CatManifest, "go.sum": CatLockfile,
	"cargo.toml": CatManifest, "cargo.lock": CatLockfile,
	"package.json": CatManifest, "package-lock.json": CatLockfile,
	"yarn.lock": CatLockfile, "pnpm-lock.yaml": CatLockfile,
	"tsconfig.json": CatConfig, "pyproject.toml": CatManifest,
	"setup.py": CatManifest, "setup.cfg": CatManifest,
	"requirements.txt": CatManifest, "pipfile": CatManifest,
	"pipfile.lock": CatLockfile, "poetry.lock": CatLockfile,
	"build.zig": CatManifest, "build.zig.zon": CatManifest,
	"pom.xml": CatManifest, "build.gradle": CatManifest,
	"build.gradle.kts": CatManifest, "gemfile": CatManifest,
	"gemfile.lock": CatLockfile, "composer.json": CatManifest,
	"composer.lock": CatLockfile, "cmakelists.txt": CatBuild,
	"makefile": CatBuild, "taskfile.yml": CatBuild, "justfile": CatBuild,
	"mix.exs": CatManifest,
}

// Classify assigns a Category and priority (higher = more relevant) to a file
// by its relative path (spec §9.3, §9.5).
func Classify(rel string) (Category, int) {
	base := strings.ToLower(filepath.Base(rel))
	dir := strings.ToLower(filepath.Dir(rel))
	ext := strings.ToLower(filepath.Ext(rel))
	stem := strings.TrimSuffix(base, ext)

	switch {
	case strings.HasPrefix(base, "readme"):
		return CatReadme, 100
	case strings.HasPrefix(base, "license") || strings.HasPrefix(base, "licence") || strings.HasPrefix(base, "copying"):
		return CatLicense, 60
	case strings.HasPrefix(base, "changelog") || strings.HasPrefix(base, "history"):
		return CatChangelog, 70
	case strings.HasPrefix(base, "contributing"):
		return CatContrib, 55
	case strings.HasPrefix(base, "security"):
		return CatSecurity, 55
	case strings.HasPrefix(base, "code_of_conduct"):
		return CatDoc, 30
	}

	if c, ok := manifestNames[base]; ok {
		prio := 90
		if c == CatLockfile {
			prio = 50
		}
		return c, prio
	}

	// CI / deployment.
	if strings.Contains(dir, ".github/workflows") || strings.Contains(dir, ".gitlab") ||
		base == ".gitlab-ci.yml" || strings.Contains(dir, ".circleci") ||
		base == "azure-pipelines.yml" || strings.Contains(dir, ".buildkite") {
		return CatCI, 75
	}
	if base == "dockerfile" || strings.HasPrefix(base, "dockerfile") ||
		base == "docker-compose.yml" || base == "docker-compose.yaml" ||
		base == "compose.yaml" || base == "compose.yml" {
		return CatDocker, 70
	}

	// ADRs and docs.
	if strings.Contains(dir, "adr") || strings.Contains(dir, "decisions") {
		return CatADR, 65
	}
	if strings.HasPrefix(dir, "docs") || dir == "docs" {
		return CatDoc, 60
	}
	if ext == ".md" || ext == ".rst" || ext == ".adoc" || ext == ".txt" {
		return CatDoc, 40
	}

	// Tests.
	if isTest(rel, base, stem) {
		return CatTest, 45
	}

	// Entry points.
	if sourceExts[ext] {
		if isEntrypoint(rel, base, stem) {
			return CatEntrypoint, 85
		}
		return CatSource, 35
	}

	// Config-ish.
	if ext == ".yaml" || ext == ".yml" || ext == ".toml" || ext == ".json" || ext == ".ini" || ext == ".env" {
		return CatConfig, 25
	}

	return CatOther, 10
}

func isTest(rel, base, stem string) bool {
	switch {
	case strings.HasSuffix(stem, "_test"), // Go
		strings.HasSuffix(stem, ".test"),
		strings.HasSuffix(stem, ".spec"),
		strings.HasPrefix(base, "test_"): // Python
		return true
	}
	low := strings.ToLower(rel)
	return strings.Contains(low, "/tests/") || strings.Contains(low, "/test/") ||
		strings.HasPrefix(low, "tests/") || strings.HasPrefix(low, "test/")
}

func isEntrypoint(rel, base, stem string) bool {
	switch base {
	case "main.go", "main.rs", "main.py", "__main__.py", "index.ts", "index.js",
		"app.py", "server.go", "cmd.go":
		return true
	}
	if stem == "main" {
		return true
	}
	low := strings.ToLower(rel)
	return strings.HasPrefix(low, "cmd/") || strings.Contains(low, "/cmd/")
}

// sortSelected orders selected files by descending priority then path.
func sortSelected(res *Result) {
	sort.SliceStable(res.Selected, func(i, j int) bool {
		a, b := res.Selected[i], res.Selected[j]
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}
		return a.RelPath < b.RelPath
	})
}

// ByCategory returns selected files matching any of the given categories.
func (r *Result) ByCategory(cats ...Category) []File {
	set := make(map[Category]bool, len(cats))
	for _, c := range cats {
		set[c] = true
	}
	var out []File
	for _, f := range r.Selected {
		if set[f.Category] {
			out = append(out, f)
		}
	}
	return out
}

// Has reports whether any selected file matches the category.
func (r *Result) Has(cat Category) bool {
	for _, f := range r.Selected {
		if f.Category == cat {
			return true
		}
	}
	return false
}
