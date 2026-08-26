package language

import (
	"encoding/json"
	"strings"

	"github.com/lesliesrussell/tldg/internal/repo"
)

// tldg-eca
// Best-effort, direct-dependency manifest parsers (spec §9.8). Lockfile-resolved
// versions and transitive graphs are out of scope for milestone 1.

// parseGoMod extracts require directives from a go.mod file.
func parseGoMod(abs, rel string) []repo.Dependency {
	content, ok := readFile(abs)
	if !ok {
		return nil
	}
	var deps []repo.Dependency
	inBlock := false
	for _, line := range strings.Split(content, "\n") {
		l := strings.TrimSpace(line)
		if i := strings.Index(l, "//"); i >= 0 {
			l = strings.TrimSpace(l[:i])
		}
		switch {
		case strings.HasPrefix(l, "require ("):
			inBlock = true
			continue
		case l == ")":
			inBlock = false
			continue
		case strings.HasPrefix(l, "require "):
			l = strings.TrimSpace(strings.TrimPrefix(l, "require"))
		case !inBlock:
			continue
		}
		fields := strings.Fields(l)
		if len(fields) >= 2 && strings.Contains(fields[0], ".") {
			deps = append(deps, repo.Dependency{
				Name: fields[0], Version: fields[1], Ecosystem: "go", Manifest: rel,
			})
		}
	}
	return deps
}

// parsePackageJSON extracts dependencies and devDependencies.
func parsePackageJSON(abs, rel string) []repo.Dependency {
	content, ok := readFile(abs)
	if !ok {
		return nil
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return nil
	}
	var deps []repo.Dependency
	for name, ver := range pkg.Dependencies {
		deps = append(deps, repo.Dependency{Name: name, Version: ver, Ecosystem: "npm", Manifest: rel})
	}
	for name, ver := range pkg.DevDependencies {
		deps = append(deps, repo.Dependency{Name: name, Version: ver, Ecosystem: "npm", Manifest: rel})
	}
	return deps
}

// parseCargoToml extracts the [dependencies] table (simple form).
func parseCargoToml(abs, rel string) []repo.Dependency {
	content, ok := readFile(abs)
	if !ok {
		return nil
	}
	var deps []repo.Dependency
	inDeps := false
	for _, line := range strings.Split(content, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "[") {
			inDeps = l == "[dependencies]"
			continue
		}
		if !inDeps || l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		if i := strings.Index(l, "="); i > 0 {
			name := strings.TrimSpace(l[:i])
			ver := strings.Trim(strings.TrimSpace(l[i+1:]), `"`)
			deps = append(deps, repo.Dependency{Name: name, Version: ver, Ecosystem: "crates", Manifest: rel})
		}
	}
	return deps
}

// parseRequirements extracts top-level pins from a requirements.txt file.
func parseRequirements(abs, rel string) []repo.Dependency {
	content, ok := readFile(abs)
	if !ok {
		return nil
	}
	var deps []repo.Dependency
	for _, line := range strings.Split(content, "\n") {
		l := strings.TrimSpace(line)
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "-") {
			continue
		}
		name, ver := l, ""
		for _, sep := range []string{"==", ">=", "<=", "~=", ">", "<"} {
			if i := strings.Index(l, sep); i > 0 {
				name = strings.TrimSpace(l[:i])
				ver = strings.TrimSpace(l[i:])
				break
			}
		}
		deps = append(deps, repo.Dependency{Name: name, Version: ver, Ecosystem: "pypi", Manifest: rel})
	}
	return deps
}
