// Package models provides LLM provider clients behind a common interface
// (spec §10.1). M0/M1 ships Ollama and an OpenAI-compatible client.
package models

import (
	"context"
	"fmt"

	"github.com/lesliesrussell/tldg/internal/config"
)

// tldg-5xh

// Message is a single chat message.
type Message struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// GenerateRequest is a synthesis request.
type GenerateRequest struct {
	Messages    []Message
	Temperature float64
	MaxTokens   int
}

// Provider is a chat-capable LLM endpoint.
type Provider interface {
	// Name identifies the provider kind.
	Name() string
	// Ping verifies reachability and returns available model identifiers.
	Ping(ctx context.Context) (models []string, err error)
	// Generate synthesizes a response from the given messages.
	Generate(ctx context.Context, req GenerateRequest) (string, error)
}

// New constructs a Provider from a ModelConfig.
func New(mc config.ModelConfig) (Provider, error) {
	switch mc.Provider {
	case "ollama":
		return &Ollama{Endpoint: mc.Endpoint, Model: mc.Model}, nil
	case "openai-compatible":
		return &OpenAICompatible{Endpoint: mc.Endpoint, Model: mc.Model}, nil
	default:
		return nil, fmt.Errorf("unsupported model provider %q", mc.Provider)
	}
}

// HasModel reports whether want appears in the available list (matching on
// exact name or name without tag).
func HasModel(available []string, want string) bool {
	base := want
	if i := indexByte(want, ':'); i >= 0 {
		base = want[:i]
	}
	for _, m := range available {
		if m == want {
			return true
		}
		mb := m
		if i := indexByte(m, ':'); i >= 0 {
			mb = m[:i]
		}
		if mb == base && (m == want || mb == base) {
			return true
		}
	}
	return false
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
