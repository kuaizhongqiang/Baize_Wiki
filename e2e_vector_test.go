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

func TestE2EBuildWithVector(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("---\ntitle: Vector Search\n---\n\n# Vector Search\n\nContent about vector search."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Vector index is built only when features.vector=true in config
	// By default it's false, so vectors dir shouldn't exist
	vecDir := filepath.Join(outDir, ".baize", "vectors")
	_, err := os.Stat(vecDir)
	assert.True(t, os.IsNotExist(err), "vectors dir should not exist when vector feature is disabled")
}

func TestE2ESearchSemantic(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "hello.md"), []byte("---\ntitle: Hello World\n---\n\n# Hello\n\nThis is a test document about searching."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Semantic search without vectors (no vectors built) should fall back to BM25
	searchResult := app.RunSemanticSearch(context.Background(), outDir, "searching", index.SearchOpts{Limit: 10}, 0.5)
	assert.True(t, searchResult.Success)
	assert.GreaterOrEqual(t, searchResult.Total, 1)
}

func TestE2ESearchCompatibility(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "test.md"), []byte("---\ntitle: Test\n---\n\n# Test\n\nContent for compatibility check."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Regular search (no --semantic) — Phase 3 compatible
	regularResult := app.RunSearch(context.Background(), outDir, "compatibility", index.SearchOpts{Limit: 10})
	assert.True(t, regularResult.Success)
	assert.GreaterOrEqual(t, regularResult.Total, 1)
}

func TestE2ESemanticWithHybridWeight(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "data.md"), []byte("---\ntitle: Data\n---\n\n# Data\n\nContent about data processing."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Test with different alpha values
	alpha50 := app.RunSemanticSearch(context.Background(), outDir, "data", index.SearchOpts{Limit: 10}, 0.5)
	assert.True(t, alpha50.Success)

	alphaPure := app.RunSemanticSearch(context.Background(), outDir, "data", index.SearchOpts{Limit: 10}, 1.0)
	assert.True(t, alphaPure.Success)
}
