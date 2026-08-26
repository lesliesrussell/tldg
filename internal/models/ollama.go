package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// tldg-5xh

// Ollama is a Provider backed by an Ollama server (spec §10.1).
type Ollama struct {
	Endpoint string
	Model    string
	// HTTP allows overriding the client (e.g. in tests). Defaults to a 5m client.
	HTTP *http.Client
}

func (o *Ollama) Name() string { return "ollama" }

func (o *Ollama) http() *http.Client {
	if o.HTTP != nil {
		return o.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

// Ping lists available models via GET /api/tags.
func (o *Ollama) Ping(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.Endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.http().Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama unreachable at %s: %w", o.Endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama /api/tags status %d", resp.StatusCode)
	}
	var body struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode ollama tags: %w", err)
	}
	names := make([]string, 0, len(body.Models))
	for _, m := range body.Models {
		names = append(names, m.Name)
	}
	return names, nil
}

// Generate calls POST /api/chat with streaming disabled.
func (o *Ollama) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	msgs := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	payload := map[string]any{
		"model":    o.Model,
		"messages": msgs,
		"stream":   false,
		"options": map[string]any{
			"temperature": req.Temperature,
		},
	}
	if req.MaxTokens > 0 {
		payload["options"].(map[string]any)["num_predict"] = req.MaxTokens
	}
	buf, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Endpoint+"/api/chat", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := o.http().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ollama chat request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama chat status %d: %s", resp.StatusCode, string(b))
	}
	var body struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode ollama chat: %w", err)
	}
	return body.Message.Content, nil
}
