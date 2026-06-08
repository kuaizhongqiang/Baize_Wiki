// Package catalog provides the Level 2 cataloging pipeline:
// content summarization, keyword extraction, and entity extraction.
//
// It supports two backends:
//   - Local: first-paragraph extraction, zero external dependencies
//   - Remote: HTTP API call to an OpenAI-compatible provider (DeepSeek, OpenAI, etc.)
//
// Post-processing includes sensitive info filtering, dedup, and truncation.
package catalog

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

// CatalogLevel controls the depth of cataloging during Wiki build.
type CatalogLevel int

const (
	CatalogLevelNone CatalogLevel = 0 // no cataloging (raw content only)
	CatalogLevel1    CatalogLevel = 1 // Level 1: directory index + raw content (current default)
	CatalogLevel2    CatalogLevel = 2 // Level 2: per-page summary + keywords + entities
)

// Summarizer defines the interface for generating page summaries.
type Summarizer interface {
	// Summarize generates a summary, keywords, and entities for the given page content.
	Summarize(ctx context.Context, page *model.Page, lang string) (*CatalogResult, error)
}

// CatalogResult holds the output of the cataloging process for one page.
type CatalogResult struct {
	Abstract string   `json:"abstract"`
	Keywords []string `json:"keywords"`
	Entities []model.Entity `json:"entities,omitempty"`
	Warning  string   `json:"warning,omitempty"`
}

// CatalogConfig configures the cataloging pipeline.
type CatalogConfig struct {
	Level     CatalogLevel // cataloging depth
	Backend   string       // "local" or "remote"
	Lang      string       // output language (default "zh")
	MaxTokens int          // max tokens for summary (default 100)
	Endpoint  string       // remote API endpoint
	Model     string       // remote model name
	APIKey    string       // remote API key
}

// DefaultCatalogConfig returns sensible defaults.
func DefaultCatalogConfig() CatalogConfig {
	return CatalogConfig{
		Level:     CatalogLevelNone,
		Backend:   "local",
		Lang:      "zh",
		MaxTokens: 100,
	}
}

// NewSummarizer creates a Summarizer based on the config.
func NewSummarizer(cfg CatalogConfig) Summarizer {
	switch cfg.Backend {
	case "remote":
		return &RemoteSummarizer{
			endpoint: cfg.Endpoint,
			model:    cfg.Model,
			apiKey:   cfg.APIKey,
			lang:     cfg.Lang,
		}
	default:
		return &LocalSummarizer{
			lang:      cfg.Lang,
			maxTokens: cfg.MaxTokens,
		}
	}
}

// FromModelConfig converts a model.CatalogConfig to a catalog.CatalogConfig.
func FromModelConfig(cfg model.CatalogConfig, level int) CatalogConfig {
	return CatalogConfig{
		Level:     CatalogLevel(level),
		Backend:   cfg.Backend,
		Lang:      "zh",
		MaxTokens: 100,
		Endpoint:  normalizeEndpoint(cfg.Endpoint),
		Model:     cfg.Model,
		APIKey:    cfg.APIKey,
	}
}

// ContentHash returns the SHA256 hex digest of content.
func ContentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

// PostProcess applies sensitive info filtering, dedup, and truncation.
func PostProcess(result *CatalogResult, maxTokens int) {
	if result == nil {
		return
	}

	// 1. Sensitive info filtering
	result.Abstract = FilterSensitive(result.Abstract)
	for i, kw := range result.Keywords {
		result.Keywords[i] = FilterSensitive(kw)
	}
	for i, e := range result.Entities {
		result.Entities[i].Name = FilterSensitive(e.Name)
	}

	// 2. Truncate abstract
	const avgCharsPerToken = 4 // rough estimate for Chinese text
	maxChars := maxTokens * avgCharsPerToken
	if len(result.Abstract) > maxChars {
		result.Abstract = result.Abstract[:maxChars] + "..."
	}

	// 3. Dedup keywords
	seen := make(map[string]bool)
	deduped := make([]string, 0, len(result.Keywords))
	for _, kw := range result.Keywords {
		kw = strings.TrimSpace(kw)
		if kw != "" && !seen[kw] {
			seen[kw] = true
			deduped = append(deduped, kw)
		}
	}
	result.Keywords = deduped
}

// normalizeEndpoint ensures the API endpoint includes the chat completions path.
func normalizeEndpoint(endpoint string) string {
	if endpoint == "" {
		return endpoint
	}
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = strings.TrimRight(endpoint, "/") + "/chat/completions"
	}
	return endpoint
}

// detectLanguage returns a file type label for LLM prompts.
func detectLanguage(page *model.Page) string {
	if page.SourceFile == "" {
		return "text"
	}
	ext := strings.ToLower(page.SourceFile[strings.LastIndex(page.SourceFile, "."):])
	switch ext {
	case ".cs":
		return "csharp"
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".ts":
		return "javascript/typescript"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cpp", ".cxx", ".cc", ".c":
		return "c/c++"
	case ".md", ".mdx":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		return "text"
	}
}
