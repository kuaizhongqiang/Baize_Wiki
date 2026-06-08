package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMDXFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.mdx")
	content := `---
title: MDX Doc
---

import SomeComponent from './components'

# MDX Title

MDX content with <CustomComponent />
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	file := model.FileInfo{
		Path:      "doc.mdx",
		AbsPath:   path,
		Extension: ".mdx",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	assert.Equal(t, "MDX Doc", page.Title)
	assert.Contains(t, page.Content, "MDX Title")
}

func TestParseCodeFile(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		extension string
		content   string
	}{
		{"go file", "main.go", ".go", "package main\nfunc main() {}"},
		{"python file", "app.py", ".py", "def hello():\n    pass"},
		{"js file", "index.js", ".js", "function foo() { return 1; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0644))

			file := model.FileInfo{
				Path:      tt.filename,
				AbsPath:   path,
				Extension: tt.extension,
			}

			page, warn := Parse(file)
			assert.Empty(t, warn)
			// Title should be filename without extension
			expectedTitle := strings.TrimSuffix(tt.filename, tt.extension)
			assert.Equal(t, expectedTitle, page.Title)
			assert.Equal(t, tt.content, page.Content)
		})
	}
}

func TestParseOnlyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fm.md")

	// File with only frontmatter, no body
	content := "---\ntitle: Just FM\n---\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	file := model.FileInfo{
		Path:      "fm.md",
		AbsPath:   path,
		Extension: ".md",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	assert.Equal(t, "Just FM", page.Title)
}

func TestParseContentWithSpecialChars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "special.md")
	content := "# 你好世界 🌍\n\nEmoji test: 🎉🚀👋\n\nSpecial: ñáéíóú\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	file := model.FileInfo{
		Path:      "special.md",
		AbsPath:   path,
		Extension: ".md",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	// Without frontmatter title, fallback behavior depends on section extraction
	assert.NotEmpty(t, page.Title)
	assert.Contains(t, page.Content, "你好世界")
	assert.Contains(t, page.Content, "🎉🚀")
}

func TestParseBatchPartialFailure(t *testing.T) {
	dir := t.TempDir()

	files := []model.FileInfo{
		{Path: "good.md", AbsPath: createFile(t, dir, "good.md", "# Good"), Extension: ".md"},
		{Path: "bad.md", AbsPath: filepath.Join(dir, "nonexistent.md"), Extension: ".md"},
		{Path: "also-good.md", AbsPath: createFile(t, dir, "also-good.md", "# Also Good"), Extension: ".md"},
	}

	pages, warnings := ParseBatch(context.Background(), files)

	// The bad file should fail but good files should still be parsed
	assert.GreaterOrEqual(t, len(pages), 2, "good files should still be parsed")
	assert.NotEmpty(t, warnings, "should contain warning for bad file")
}

func TestParseBatchConcurrencyLimit(t *testing.T) {
	dir := t.TempDir()
	count := 100

	var files []model.FileInfo
	for i := 0; i < count; i++ {
		name := "doc.go"
		files = append(files, model.FileInfo{
			Path:      name,
			AbsPath:   createFile(t, dir, name, "# Doc"),
			Extension: ".go",
		})
	}

	// Should not panic or hang
	pages, warnings := ParseBatch(context.Background(), files)
	assert.Len(t, pages, count)
	assert.Empty(t, warnings)
}

func TestFrontmatterWithoutTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notitle.md")
	content := "---\ntags: [test]\n---\n\n# Hello\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	file := model.FileInfo{
		Path:      "notitle.md",
		AbsPath:   path,
		Extension: ".md",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	// Frontmatter has no title — should fallback to first h1 or filename
	assert.NotEmpty(t, page.Title, "title should fallback to h1 or filename when frontmatter has no title")
}

func TestParseMarkdownHeadingGaps(t *testing.T) {
	// h1 → h3 (skipping h2)
	content := `# Level 1

### Level 3 directly

More content.

## Level 2
`
	sections := parseMarkdown(content)

	// Should handle heading level gaps gracefully
	t.Logf("Sections: %+v", sections)
	// At minimum should not panic
	assert.NotPanics(t, func() { parseMarkdown(content) })
}

func TestParseMarkdownManyHeadings(t *testing.T) {
	var lines []string
	lines = append(lines, "# Title")
	for i := 0; i < 50; i++ {
		lines = append(lines, "")
		lines = append(lines, "## Section ", string(rune('A'+i%26)), string(rune('0'+i/10)))
	}
	content := strings.Join(lines, "\n")

	// Should not panic with many headings
	sections := parseMarkdown(content)
	assert.NotNil(t, sections)
	t.Logf("Generated %d top-level sections", len(sections))
}

func TestExtractFrontmatterNoClosing(t *testing.T) {
	content := "---\ntitle: Broken\nno closing delimiter\n"
	fm, body, warn := extractFrontmatter(content)

	assert.Empty(t, fm.Title, "no closing --- means no valid frontmatter")
	assert.Equal(t, content, body, "full content treated as body")
	assert.Empty(t, warn)
}
