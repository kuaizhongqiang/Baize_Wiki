package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RemoteEmbedderConfig configures the RemoteEmbedder.
type RemoteEmbedderConfig struct {
	Endpoint string // e.g. "http://localhost:1234/v1/embeddings"
	Model    string // model name for the API
	APIKey   string // API key (if required)
}

// RemoteEmbedder calls an OpenAI-compatible embedding API (LM Studio, OpenAI, etc.).
type RemoteEmbedder struct {
	config RemoteEmbedderConfig
	dim    int
}

// NewRemoteEmbedder creates a RemoteEmbedder.
// A probe call is made to detect the embedding dimension.
func NewRemoteEmbedder(cfg RemoteEmbedderConfig) *RemoteEmbedder {
	e := &RemoteEmbedder{config: cfg, dim: 1024}
	if dim, err := e.probeDim(); err == nil && dim > 0 {
		e.dim = dim
	}
	return e
}

// Dim returns the embedding dimension.
func (e *RemoteEmbedder) Dim() int { return e.dim }

// Embed converts text to a vector via the remote API.
func (e *RemoteEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return e.batchEmbed(ctx, []string{text})
}

// batchEmbed sends a batch of texts to the embedding API.
func (e *RemoteEmbedder) batchEmbed(ctx context.Context, texts []string) ([]float32, error) {
	// Clean inputs
	for i, t := range texts {
		texts[i] = strings.TrimSpace(t)
	}

	endpoint := e.config.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:1234/v1/embeddings"
	}

	reqBody := map[string]any{
		"input": texts,
	}
	if e.config.Model != "" {
		reqBody["model"] = e.config.Model
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.config.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respData))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("empty response from embedding API")
	}

	// Convert float64 from JSON to float32
	embedding := result.Data[0].Embedding
	vec := make([]float32, len(embedding))
	for i, v := range embedding {
		vec[i] = float32(v)
	}

	return vec, nil
}

// probeDim gets the embedding dimension by sending a single test request.
func (e *RemoteEmbedder) probeDim() (int, error) {
	vec, err := e.Embed(context.Background(), "probe")
	if err != nil {
		return 0, err
	}
	return len(vec), nil
}
