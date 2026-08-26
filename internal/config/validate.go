package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// tldg-5xh

// Problem is a single validation finding.
type Problem struct {
	Field   string
	Message string
}

func (p Problem) String() string { return fmt.Sprintf("%s: %s", p.Field, p.Message) }

// supportedSourceKinds lists source identifiers recognized by the config
// schema (spec §15.2). Unknown entries are flagged by Validate.
var supportedSourceKinds = map[string]bool{
	"npm": true, "pypi": true, "crates": true, "pkg-go-dev": true,
	"maven": true, "rubygems": true, "hackernews": true, "reddit": true,
}

// Validate checks a resolved config for schema and consistency problems
// (spec §15.3 subset for M0/M1). It returns nil when the config is valid.
func Validate(cfg *Config) []Problem {
	var probs []Problem

	if cfg.Version != 1 {
		probs = append(probs, Problem{"version", fmt.Sprintf("unsupported schema version %d (expected 1)", cfg.Version)})
	}

	// Active profile must reference a defined model.
	if _, ok := cfg.Profiles[cfg.Profile]; !ok && cfg.Profile != "" {
		probs = append(probs, Problem{"profile", fmt.Sprintf("active profile %q not defined", cfg.Profile)})
	}
	for name, prof := range cfg.Profiles {
		if prof.Model != "" {
			if _, ok := cfg.Models[prof.Model]; !ok {
				probs = append(probs, Problem{"profiles." + name + ".model", fmt.Sprintf("references undefined model %q", prof.Model)})
			}
		}
	}

	// Model endpoints must be syntactically valid URLs with a known provider.
	for name, m := range cfg.Models {
		probs = append(probs, validateEndpoint("models."+name, m)...)
	}
	for name, m := range cfg.Embeddings {
		probs = append(probs, validateEndpoint("embeddings."+name, m)...)
	}

	// Index paths must be writable (parent dir creatable).
	if cfg.Index.SQLitePath != "" {
		if p := checkWritable("index.sqlite_path", filepath.Dir(cfg.Index.SQLitePath)); p != nil {
			probs = append(probs, *p)
		}
	}
	if cfg.Index.Root != "" {
		if p := checkWritable("index.root", cfg.Index.Root); p != nil {
			probs = append(probs, *p)
		}
	}

	// Ignore patterns must be valid glob patterns.
	for i, pat := range cfg.Index.Ignore {
		if _, err := filepath.Match(pat, "x"); err != nil {
			probs = append(probs, Problem{fmt.Sprintf("index.ignore[%d]", i), fmt.Sprintf("invalid pattern %q: %v", pat, err)})
		}
	}

	return probs
}

func validateEndpoint(field string, m ModelConfig) []Problem {
	var probs []Problem
	switch m.Provider {
	case "ollama", "openai-compatible":
	case "":
		probs = append(probs, Problem{field + ".provider", "missing provider"})
	default:
		probs = append(probs, Problem{field + ".provider", fmt.Sprintf("unsupported provider %q", m.Provider)})
	}
	if m.Endpoint == "" {
		probs = append(probs, Problem{field + ".endpoint", "missing endpoint"})
	} else if u, err := url.Parse(m.Endpoint); err != nil || u.Scheme == "" || u.Host == "" {
		probs = append(probs, Problem{field + ".endpoint", fmt.Sprintf("invalid endpoint URL %q", m.Endpoint)})
	}
	if m.Model == "" {
		probs = append(probs, Problem{field + ".model", "missing model name"})
	}
	return probs
}

// checkWritable verifies dir exists-and-is-writable or can be created.
func checkWritable(field, dir string) *Problem {
	if dir == "" {
		return nil
	}
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return &Problem{field, fmt.Sprintf("%s is not a directory", dir)}
		}
		// Probe writability.
		probe := filepath.Join(dir, ".tldg-write-probe")
		if f, err := os.Create(probe); err != nil {
			return &Problem{field, fmt.Sprintf("%s not writable: %v", dir, err)}
		} else {
			f.Close()
			os.Remove(probe)
		}
		return nil
	}
	// Does not exist: ensure an ancestor exists so it is creatable.
	parent := dir
	for {
		up := filepath.Dir(parent)
		if up == parent {
			break
		}
		if _, err := os.Stat(up); err == nil {
			return nil
		}
		parent = up
	}
	return nil
}

// ValidateSourceKinds flags unsupported package/community source identifiers.
func ValidateSourceKinds(kinds []string) []Problem {
	var probs []Problem
	for _, k := range kinds {
		if !supportedSourceKinds[strings.ToLower(k)] {
			probs = append(probs, Problem{"sources", fmt.Sprintf("unsupported source %q", k)})
		}
	}
	return probs
}
