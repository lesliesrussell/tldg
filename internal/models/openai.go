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

// OpenAICompatible is a Provider for OpenAI-compatible endpoints such as
// LM Studio, llama.cpp server, vLLM, and LocalAI (spec §10.1).
type OpenAICompatible struct {
	Endpoint string // base ending in /v1
	Model    string
	APIKey   string
	HTTP     *http.Client
}

func (o *OpenAICompatible) Name() string { return "openai-compatible" }

func (o *OpenAICompatible) http() *http.Client {
	if o.HTTP != nil {
		return o.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func (o *OpenAICompatible) auth(req *http.Request) {
	if o.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.APIKey)
	}
}

// Ping lists available models via GET /v1/models.
func (o *OpenAICompatible) Ping(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.Endpoint+"/models", nil)
	if err != nil {
		return nil, err
	}
	o.auth(req)
	resp, err := o.http().Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai-compatible unreachable at %s: %w", o.Endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai-compatible /models status %d", resp.StatusCode)
	}
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode models: %w", err)
	}
	names := make([]string, 0, len(body.Data))
	for _, m := range body.Data {
		names = append(names, m.ID)
	}
	return names, nil
}

// Generate calls POST /v1/chat/completions (non-streaming).
func (o *OpenAICompatible) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	msgs := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	payload := map[string]any{
		"model":       o.Model,
		"messages":    msgs,
		"temperature": req.Temperature,
		"stream":      false,
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	buf, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Endpoint+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	o.auth(httpReq)
	resp, err := o.http().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("chat/completions request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chat/completions status %d: %s", resp.StatusCode, string(b))
	}
	var body struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode chat/completions: %w", err)
	}
	if len(body.Choices) == 0 {
		return "", fmt.Errorf("chat/completions returned no choices")
	}
	return body.Choices[0].Message.Content, nil
}
