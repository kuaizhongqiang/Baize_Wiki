package parser

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantFM  bool
		wantOK  bool
	}{
		{
			name: "full frontmatter",
			content: `---
title: Test Page
description: A test
tags: [go, test]
weight: 1
category: docs
---
Content here`,
			wantFM: true,
			wantOK: true,
		},
		{
			name:    "no frontmatter",
			content: "# Just a heading\n\nSome content",
			wantFM:  false,
			wantOK:  true,
		},
		{
			name: "minimal frontmatter",
			content: `---
title: Minimal
---
Body`,
			wantFM: true,
			wantOK: true,
		},
		{
			name: "invalid yaml",
			content: `---
title: "unclosed
---
Body`,
			wantFM: true,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, warn := extractFrontmatter(tt.content)

			if !tt.wantFM {
				assert.Empty(t, fm.Title)
				assert.Equal(t, tt.content, body)
				return
			}

			if tt.wantOK {
				assert.Empty(t, warn)
				assert.NotEmpty(t, fm.Title)
			} else {
				assert.NotEmpty(t, warn)
				assert.Equal(t, tt.content, body) // returns original content on error
			}
		})
	}
}

func TestParseMarkdown(t *testing.T) {
	content := `# Title

Some intro text.

## Section 1

Content of section 1.

### Subsection 1.1

Deep content.

## Section 2

Content of section 2.
`

	sections := parseMarkdown(content)
	require.Len(t, sections, 1, "should have one top-level heading")
	assert.Equal(t, "Title", sections[0].Title)
	assert.Equal(t, 1, sections[0].Level)

	children := sections[0].Children
	if assert.Len(t, children, 2) {
		assert.Equal(t, "Section 1", children[0].Title)
		assert.Equal(t, 2, children[0].Level)

		if assert.Len(t, children[0].Children, 1) {
			assert.Equal(t, "Subsection 1.1", children[0].Children[0].Title)
			assert.Equal(t, 3, children[0].Children[0].Level)
		}

		assert.Equal(t, "Section 2", children[1].Title)
	}
}

func TestParseMarkdownNoHeadings(t *testing.T) {
	sections := parseMarkdown("Just a paragraph of text with no headings.")
	assert.Empty(t, sections)
}

func TestParseMarkdownFlatHeadings(t *testing.T) {
	content := `# One

## A

## B

## C
`
	sections := parseMarkdown(content)
	require.Len(t, sections, 1)
	assert.Equal(t, "One", sections[0].Title)

	children := sections[0].Children
	if assert.Len(t, children, 3) {
		assert.Equal(t, "A", children[0].Title)
		assert.Equal(t, "B", children[1].Title)
		assert.Equal(t, "C", children[2].Title)
	}
}

func TestParseMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `---
title: Test MD
tags: [test]
---

# Hello

Some content.

## World

More content.
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	file := model.FileInfo{
		Path:      "test.md",
		AbsPath:   path,
		Extension: ".md",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	assert.Equal(t, "Test MD", page.Title)
	assert.Equal(t, []string{"test"}, page.Tags)
	assert.Contains(t, page.Content, "Hello")
	assert.Len(t, page.Sections, 1)
}

func TestParsePlainText(t *testing.T) {
	dir := t.TempDir()

	// .txt file
	txtPath := filepath.Join(dir, "notes.txt")
	require.NoError(t, os.WriteFile(txtPath, []byte("Some plain text content"), 0644))

	file := model.FileInfo{
		Path:      "notes.txt",
		AbsPath:   txtPath,
		Extension: ".txt",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	assert.Equal(t, "notes", page.Title)
	assert.Equal(t, "Some plain text content", page.Content)
	assert.Empty(t, page.Sections)
}

func TestParseEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	require.NoError(t, os.WriteFile(path, []byte(""), 0644))

	file := model.FileInfo{
		Path:      "empty.md",
		AbsPath:   path,
		Extension: ".md",
	}

	page, warn := Parse(file)
	assert.Empty(t, warn)
	assert.Equal(t, "empty", page.Title)
}

func TestParseBatch(t *testing.T) {
	dir := t.TempDir()

	// Create multiple files
	files := []model.FileInfo{
		{Path: "a.md", AbsPath: createFile(t, dir, "a.md", "# Page A"), Extension: ".md"},
		{Path: "b.md", AbsPath: createFile(t, dir, "b.md", "# Page B"), Extension: ".md"},
		{Path: "c.txt", AbsPath: createFile(t, dir, "c.txt", "text C"), Extension: ".txt"},
	}

	pages, warnings := ParseBatch(context.Background(), files)
	assert.Empty(t, warnings)
	require.Len(t, pages, 3)

	titles := make([]string, len(pages))
	for i, p := range pages {
		titles[i] = p.Title
	}
	assert.Contains(t, titles, "Page A")
	assert.Contains(t, titles, "Page B")
	assert.Contains(t, titles, "c")
}

func TestParseBatchContextCancel(t *testing.T) {
	dir := t.TempDir()
	files := []model.FileInfo{
		{Path: "a.md", AbsPath: createFile(t, dir, "a.md", "# A"), Extension: ".md"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pages, _ := ParseBatch(ctx, files)
	assert.Empty(t, pages)
}

func TestFrontmatterWithCustomFields(t *testing.T) {
	content := `---
title: Custom
version: 2
status: active
---

Body
`
	fm, body, warn := extractFrontmatter(content)
	assert.Empty(t, warn)
	assert.Equal(t, "Custom", fm.Title)
	assert.Equal(t, "Body", body)
	assert.NotNil(t, fm.Custom)
	assert.Equal(t, 2, fm.Custom["version"])
	assert.Equal(t, "active", fm.Custom["status"])
}

func TestSectionIDGeneration(t *testing.T) {
	assert.Equal(t, "hello-world", generateSectionID("Hello World"))
	assert.Equal(t, "go-test", generateSectionID("Go.Test"))
	assert.Equal(t, "a-b-c", generateSectionID("a/b/c"))
	assert.Equal(t, "already-kebab", generateSectionID("already-kebab"))
}

// Helper to create a file and return its absolute path.
func createFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}
