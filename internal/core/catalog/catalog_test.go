package catalog

import (
	"context"
	"strings"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
)

func TestLocalSummarize(t *testing.T) {
	s := &LocalSummarizer{lang: "zh", maxTokens: 100}

	page := &model.Page{
		Title:   "Getting Started",
		Content: "# Introduction\n\nThis is the first paragraph with important information about the system.\n\n## Details\n\nThis should not be included in the first paragraph extraction.",
	}
	result, err := s.Summarize(context.Background(), page, "zh")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Abstract)
	assert.Contains(t, result.Abstract, "first paragraph")
	assert.NotEmpty(t, result.Keywords)
	assert.NotEmpty(t, result.Entities)
}

func TestLocalSummarizeWithFrontmatterDescription(t *testing.T) {
	s := &LocalSummarizer{lang: "zh", maxTokens: 100}

	page := &model.Page{
		Title:   "API Reference",
		Content: "Some content here.",
		Meta: model.Frontmatter{
			Description: "Custom description from frontmatter.",
		},
	}
	result, err := s.Summarize(context.Background(), page, "zh")

	assert.NoError(t, err)
	assert.Equal(t, "Custom description from frontmatter.", result.Abstract)
}

func TestLocalSummarizeEmptyContent(t *testing.T) {
	s := &LocalSummarizer{lang: "zh", maxTokens: 100}
	page := &model.Page{Title: "Empty", Content: ""}

	result, err := s.Summarize(context.Background(), page, "zh")
	assert.NoError(t, err)
	assert.Empty(t, result.Abstract)
}

func TestExtractFirstParagraph(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple paragraph",
			content: "This is a simple paragraph.\n\nMore text here.",
			want:    "This is a simple paragraph.",
		},
		{
			name:    "skip headings",
			content: "# Title\n\nThis is after the heading.\n\n## Subtitle\nMore text.",
			want:    "This is after the heading.",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFirstParagraph(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterSensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "redact api key",
			input: "API key: sk-abc123def456ghi789",
			want:  "[REDACTED]",
		},
		{
			name:  "redact internal IP",
			input: "Server at 10.0.1.5 is running",
			want:  "Server at [REDACTED] is running",
		},
		{
			name:  "no sensitive info",
			input: "This is a normal summary about architecture.",
			want:  "This is a normal summary about architecture.",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterSensitive(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostProcess(t *testing.T) {
	result := &CatalogResult{
		Abstract: "This is a test summary with secret: sk-test1234567890 inside. " + strings.Repeat("x", 500),
		Keywords: []string{"  test  ", "DUPLICATE", "duplicate", "", "  ", "valid"},
		Entities: []model.Entity{
			{Name: "ClassA", Type: "class", Role: "defined"},
		},
	}

	PostProcess(result, 100)

	// Sensitive info filtered
	assert.NotContains(t, result.Abstract, "sk-test1234567890")
	assert.Contains(t, result.Abstract, "[REDACTED]")

	// Truncated to ~400 chars (100 tokens * 4)
	assert.True(t, len(result.Abstract) <= 420, "abstract too long: %d", len(result.Abstract))

	// Dedup keywords
	assert.Contains(t, result.Keywords, "valid")
	assert.NotContains(t, result.Keywords, "")
	assert.NotContains(t, result.Keywords, "  ")
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		sourceFile string
		want       string
	}{
		{"file.cs", "csharp"},
		{"file.go", "go"},
		{"file.py", "python"},
		{"file.ts", "javascript/typescript"},
		{"file.rs", "rust"},
		{"file.md", "markdown"},
		{"file.json", "json"},
		{"file.yaml", "yaml"},
		{"unknown.xyz", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			page := &model.Page{SourceFile: tt.sourceFile}
			got := detectLanguage(page)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSummarizer(t *testing.T) {
	cfg := DefaultCatalogConfig()

	localCfg := cfg
	localCfg.Backend = "local"
	assert.IsType(t, &LocalSummarizer{}, NewSummarizer(localCfg))

	remoteCfg := cfg
	remoteCfg.Backend = "remote"
	assert.IsType(t, &RemoteSummarizer{}, NewSummarizer(remoteCfg))
}
