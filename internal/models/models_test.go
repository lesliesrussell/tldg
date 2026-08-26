package models

import (
	"testing"

	"github.com/leslierussell/tldg/internal/config"
)

// tldg-5xh

func TestHasModel(t *testing.T) {
	avail := []string{"qwen2.5-coder:7b", "nomic-embed-text:latest", "llama3.1:8b"}
	cases := map[string]bool{
		"qwen2.5-coder:7b": true,  // exact
		"nomic-embed-text": true,  // base name matches ":latest"
		"llama3.1:8b":      true,  // exact
		"mistral":          false, // absent
		"qwen2.5-coder":    true,  // base match
	}
	for want, expect := range cases {
		if got := HasModel(avail, want); got != expect {
			t.Errorf("HasModel(%q) = %v, want %v", want, got, expect)
		}
	}
}

func TestNewUnsupported(t *testing.T) {
	if _, err := New(config.ModelConfig{Provider: "bogus"}); err == nil {
		t.Error("expected error for unsupported provider")
	}
}
