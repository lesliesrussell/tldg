// Package config loads, defaults, and resolves tldg YAML configuration
// (spec §15). Credentials are never stored here; only keychain service names.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// tldg-5xh

// Config is the top-level tldg configuration (subset of spec §15.2 needed for M0/M1).
type Config struct {
	Version    int                    `yaml:"version"`
	Profile    string                 `yaml:"profile"`
	Profiles   map[string]Profile     `yaml:"profiles"`
	Models     map[string]ModelConfig `yaml:"models"`
	Embeddings map[string]ModelConfig `yaml:"embeddings"`
	Index      IndexConfig            `yaml:"index"`
	Retrieval  RetrievalConfig        `yaml:"retrieval"`
	Privacy    PrivacyConfig          `yaml:"privacy"`
	Output     OutputConfig           `yaml:"output"`
	Logging    LoggingConfig          `yaml:"logging"`

	// paths holds resolved runtime locations; not serialized.
	paths Paths `yaml:"-"`
}

// Profile binds a model and a policy name.
type Profile struct {
	Model  string `yaml:"model"`
	Policy string `yaml:"policy"`
}

// ModelConfig describes an LLM or embedding endpoint.
type ModelConfig struct {
	Provider              string  `yaml:"provider"`
	Endpoint              string  `yaml:"endpoint"`
	Model                 string  `yaml:"model"`
	ContextWindow         int     `yaml:"context_window,omitempty"`
	Temperature           float64 `yaml:"temperature,omitempty"`
	MaxOutputTokens       int     `yaml:"max_output_tokens,omitempty"`
	Dimensions            int     `yaml:"dimensions,omitempty"`
	APIKeyKeychainService string  `yaml:"api_key_keychain_service,omitempty"`
}

// IndexConfig controls file selection and index storage (spec §15.2).
type IndexConfig struct {
	Root               string   `yaml:"root"`
	SQLitePath         string   `yaml:"sqlite_path"`
	MaxFileSizeKB      int      `yaml:"max_file_size_kb"`
	MaxTotalIndexMB    int      `yaml:"max_total_index_size_mb"`
	IncludeUntracked   bool     `yaml:"include_untracked"`
	FollowSymlinks     bool     `yaml:"follow_symlinks"`
	Ignore             []string `yaml:"ignore"`
	Include            []string `yaml:"include"`
}

// RetrievalConfig controls context budgets (spec §15.2).
type RetrievalConfig struct {
	MaxChunks       int     `yaml:"max_chunks"`
	MaxChunkTokens  int     `yaml:"max_chunk_tokens"`
	LexicalWeight   float64 `yaml:"lexical_weight"`
	SemanticWeight  float64 `yaml:"semantic_weight"`
	StructuralWeight float64 `yaml:"structural_weight"`
}

// PrivacyConfig controls disclosure defaults (spec §15.2).
type PrivacyConfig struct {
	ExternalResearchDefault string   `yaml:"external_research_default"`
	ExternalCodeDefault     string   `yaml:"external_code_default"`
	PersistentExternalCache bool     `yaml:"persistent_external_cache"`
	LocalPathDisplay        string   `yaml:"local_path_display"`
	RedactPatterns          []string `yaml:"redact_patterns"`
}

// OutputConfig controls rendering defaults (spec §15.2).
type OutputConfig struct {
	Citations   string `yaml:"citations"`
	Confidence  bool   `yaml:"confidence"`
	Color       string `yaml:"color"`
	SourcePaths string `yaml:"source_paths"`
}

// LoggingConfig controls diagnostic logging (spec §15.2).
type LoggingConfig struct {
	Level                string `yaml:"level"`
	PersistPrompts       bool   `yaml:"persist_prompts"`
	PersistModelResponses bool  `yaml:"persist_model_responses"`
}

// Paths holds resolved config/data/cache locations (spec §15.1).
type Paths struct {
	Config string
	Data   string
	Cache  string
}

// Paths returns the resolved runtime paths for this config.
func (c *Config) Paths() Paths { return c.paths }

// ActiveProfile returns the selected profile, falling back to a default.
func (c *Config) ActiveProfile() Profile {
	if p, ok := c.Profiles[c.Profile]; ok {
		return p
	}
	return Profile{Model: "local-coder", Policy: "local-first"}
}

// ActiveModel returns the ModelConfig for the active (or overridden) profile.
// override, when non-empty, selects a model by name directly.
func (c *Config) ActiveModel(override string) (ModelConfig, string, error) {
	name := override
	if name == "" {
		name = c.ActiveProfile().Model
	}
	m, ok := c.Models[name]
	if !ok {
		return ModelConfig{}, name, fmt.Errorf("model %q not defined in config", name)
	}
	return m, name, nil
}

// DefaultEmbedding returns the default embedding model config, if any.
func (c *Config) DefaultEmbedding() (ModelConfig, bool) {
	m, ok := c.Embeddings["default"]
	return m, ok
}

// OSPaths resolves platform config/data/cache directories (spec §15.1).
func OSPaths() (Paths, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve config dir: %w", err)
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve cache dir: %w", err)
	}
	var dataDir string
	switch runtime.GOOS {
	case "darwin":
		// macOS: data co-located with config under Application Support.
		dataDir = cfgDir
	case "windows":
		dataDir = cfgDir
	default:
		// Linux/XDG: ~/.local/share.
		home, herr := os.UserHomeDir()
		if herr != nil {
			return Paths{}, fmt.Errorf("resolve home dir: %w", herr)
		}
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			dataDir = xdg
		} else {
			dataDir = filepath.Join(home, ".local", "share")
		}
	}
	return Paths{
		Config: filepath.Join(cfgDir, "tldg"),
		Data:   filepath.Join(dataDir, "tldg"),
		Cache:  filepath.Join(cacheDir, "tldg"),
	}, nil
}

// DefaultConfigFile returns the default config.yaml path.
func DefaultConfigFile() (string, error) {
	p, err := OSPaths()
	if err != nil {
		return "", err
	}
	return filepath.Join(p.Config, "config.yaml"), nil
}

// Load reads config from path (or the default path when path is ""). A missing
// file is not an error: built-in defaults are returned. Resolved runtime paths
// are always attached.
func Load(path string) (*Config, error) {
	p, err := OSPaths()
	if err != nil {
		return nil, err
	}
	if path == "" {
		path = filepath.Join(p.Config, "config.yaml")
	}
	cfg := Default()
	data, rerr := os.ReadFile(path)
	switch {
	case rerr == nil:
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		applyDefaults(cfg)
	case os.IsNotExist(rerr):
		// Use built-in defaults.
	default:
		return nil, fmt.Errorf("read %s: %w", path, rerr)
	}
	cfg.paths = expandPaths(p, cfg)
	return cfg, nil
}

// expandPaths expands ~ in index paths against the resolved data dir.
func expandPaths(p Paths, cfg *Config) Paths {
	home, _ := os.UserHomeDir()
	expand := func(s string) string {
		if len(s) >= 2 && s[:2] == "~/" && home != "" {
			return filepath.Join(home, s[2:])
		}
		return s
	}
	cfg.Index.Root = expand(cfg.Index.Root)
	cfg.Index.SQLitePath = expand(cfg.Index.SQLitePath)
	return p
}
