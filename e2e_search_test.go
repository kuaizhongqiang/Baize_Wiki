package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/app"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EBuildWithIndex(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc1.md"), []byte("---\ntitle: Doc 1\ncategory: guide\ntags: [go, test]\n---\n\n# Doc 1\n\nContent about Go programming."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc2.md"), []byte("---\ntitle: Doc 2\ncategory: api\ntags: [python]\n---\n\n# Doc 2\n\nPython data science content."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)

	assert.True(t, result.Success)
	assert.Equal(t, 2, result.Summary.Pages)

	// Verify index exists
	indexPath := filepath.Join(outDir, ".baize", "index.bleve")
	assert.DirExists(t, indexPath)

	// Verify we can search
	idx, err := index.NewIndex(indexPath)
	require.NoError(t, err)
	defer idx.Close()

	results, err := idx.Search(context.Background(), "Go programming", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "should find results for 'Go programming'")
}

func TestE2ESearchAfterBuild(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "hello.md"), []byte("---\ntitle: Hello World\n---\n\n# Hello\n\nThis is a test document about searching."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Use the app's search
	searchResult := app.RunSearch(context.Background(), outDir, "searching", index.SearchOpts{Limit: 10})
	assert.True(t, searchResult.Success)
	assert.GreaterOrEqual(t, searchResult.Total, 1)
}

func TestE2ESearchEmptySource(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Build with empty source (will fail, no index)
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	assert.False(t, result.Success)

	// Try searching without index via RunSearch
	searchResult := app.RunSearch(context.Background(), outDir, "test", index.SearchOpts{Limit: 10})
	assert.Equal(t, 0, searchResult.Total, "no results from empty source")
}

func TestE2EScanAllWithCode(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create a .go file
	goContent := `// Package calculator provides math operations.
// It supports addition and subtraction.
package calculator

func Add(a, b int) int {
    return a + b
}
`
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "calc.go"), []byte(goContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "readme.md"), []byte("# Calculator\n\nSimple calculator."), 0644))

	// Build with scan-all
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, true)
	require.True(t, result.Success)

	// Index should contain both files
	indexPath := filepath.Join(outDir, ".baize", "index.bleve")
	idx, err := index.NewIndex(indexPath)
	require.NoError(t, err)
	defer idx.Close()

	results, err := idx.Search(context.Background(), "calculator", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "should find calculator content")
}

func TestE2EScanAllSkippedWithoutFlag(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "code.go"), []byte("package main\nfunc main() {}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "readme.md"), []byte("# Readme\n\nContent"), 0644))

	// Without scan-all, only .md files should be indexed
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Only 1 page (the .md file) should be parsed
	assert.Equal(t, 1, result.Summary.Pages)
}
