package config

// tldg-5xh

// Default returns a Config populated with the built-in defaults from spec §15.2,
// so tldg works with no config file present.
func Default() *Config {
	return &Config{
		Version: 1,
		Profile: "default",
		Profiles: map[string]Profile{
			"default":  {Model: "local-coder", Policy: "local-first"},
			"research": {Model: "local-coder", Policy: "public-research"},
		},
		Models: map[string]ModelConfig{
			"local-coder": {
				Provider:        "ollama",
				Endpoint:        "http://127.0.0.1:11434",
				Model:           "qwen2.5-coder:7b",
				ContextWindow:   32768,
				Temperature:     0.15,
				MaxOutputTokens: 4096,
			},
			"local-compat": {
				Provider:              "openai-compatible",
				Endpoint:              "http://127.0.0.1:1234/v1",
				Model:                 "local-model",
				APIKeyKeychainService: "tldg/lm-studio",
			},
		},
		Embeddings: map[string]ModelConfig{
			"default": {
				Provider:   "ollama",
				Endpoint:   "http://127.0.0.1:11434",
				Model:      "nomic-embed-text",
				Dimensions: 768,
			},
		},
		Index: IndexConfig{
			Root:            "~/.local/share/tldg/indexes",
			SQLitePath:      "~/.local/share/tldg/tldg.db",
			MaxFileSizeKB:   512,
			MaxTotalIndexMB: 2048,
			Ignore: []string{
				".git", "node_modules", "vendor", "dist", "build",
				"target", "coverage", ".venv",
			},
			Include: []string{},
		},
		Retrieval: RetrievalConfig{
			MaxChunks:        24,
			MaxChunkTokens:   1200,
			LexicalWeight:    0.45,
			SemanticWeight:   0.35,
			StructuralWeight: 0.20,
		},
		Privacy: PrivacyConfig{
			ExternalResearchDefault: "ask",
			ExternalCodeDefault:     "never",
			PersistentExternalCache: true,
			LocalPathDisplay:        "relative",
			RedactPatterns: []string{
				`(?i)api[_-]?key\s*[:=]\s*\S+`,
				`(?i)secret\s*[:=]\s*\S+`,
				`(?i)password\s*[:=]\s*\S+`,
			},
		},
		Output: OutputConfig{
			Citations:   "inline",
			Confidence:  true,
			Color:       "auto",
			SourcePaths: "relative",
		},
		Logging: LoggingConfig{Level: "info"},
	}
}

// applyDefaults fills zero-valued fields in a loaded config with defaults so a
// partial config file still yields a usable configuration.
func applyDefaults(cfg *Config) {
	d := Default()
	if cfg.Version == 0 {
		cfg.Version = d.Version
	}
	if cfg.Profile == "" {
		cfg.Profile = d.Profile
	}
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = d.Profiles
	}
	if len(cfg.Models) == 0 {
		cfg.Models = d.Models
	}
	if len(cfg.Embeddings) == 0 {
		cfg.Embeddings = d.Embeddings
	}
	if cfg.Index.Root == "" {
		cfg.Index.Root = d.Index.Root
	}
	if cfg.Index.SQLitePath == "" {
		cfg.Index.SQLitePath = d.Index.SQLitePath
	}
	if cfg.Index.MaxFileSizeKB == 0 {
		cfg.Index.MaxFileSizeKB = d.Index.MaxFileSizeKB
	}
	if cfg.Index.MaxTotalIndexMB == 0 {
		cfg.Index.MaxTotalIndexMB = d.Index.MaxTotalIndexMB
	}
	if len(cfg.Index.Ignore) == 0 {
		cfg.Index.Ignore = d.Index.Ignore
	}
	if cfg.Retrieval.MaxChunks == 0 {
		cfg.Retrieval = d.Retrieval
	}
	if cfg.Output.Citations == "" {
		cfg.Output = d.Output
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = d.Logging.Level
	}
}
