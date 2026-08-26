package config

import "testing"

// tldg-5xh

func TestDefaultValid(t *testing.T) {
	cfg := Default()
	if probs := Validate(cfg); len(probs) != 0 {
		t.Fatalf("default config should validate, got: %v", probs)
	}
	if _, _, err := cfg.ActiveModel(""); err != nil {
		t.Errorf("active model: %v", err)
	}
}

func TestValidateBadModel(t *testing.T) {
	cfg := Default()
	cfg.Models["local-coder"] = ModelConfig{Provider: "bogus", Endpoint: "not a url", Model: ""}
	probs := Validate(cfg)
	if len(probs) < 2 {
		t.Fatalf("expected provider+endpoint+model problems, got %d: %v", len(probs), probs)
	}
}

func TestValidateProfileReference(t *testing.T) {
	cfg := Default()
	cfg.Profiles["default"] = Profile{Model: "missing-model", Policy: "local-first"}
	probs := Validate(cfg)
	found := false
	for _, p := range probs {
		if p.Field == "profiles.default.model" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected undefined-model problem, got: %v", probs)
	}
}

func TestValidateVersion(t *testing.T) {
	cfg := Default()
	cfg.Version = 2
	probs := Validate(cfg)
	if len(probs) == 0 {
		t.Error("expected version problem")
	}
}

func TestActiveModelOverride(t *testing.T) {
	cfg := Default()
	m, name, err := cfg.ActiveModel("local-compat")
	if err != nil {
		t.Fatalf("override: %v", err)
	}
	if name != "local-compat" || m.Provider != "openai-compatible" {
		t.Errorf("unexpected model %s: %+v", name, m)
	}
}
