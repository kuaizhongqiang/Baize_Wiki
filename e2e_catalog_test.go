package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ECatalogLevel2Build(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create test source files
	content := []byte("---\ntitle: Test Page\ntags: [test]\n---\n\n# Test\n\nThis is test content for cataloging.")
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "guide"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "test.md"), content, 0644))

	_ = os.WriteFile(filepath.Join(srcDir, "baize.yaml"), []byte("name: test\noutput:\n  dir: "+outDir+"\n  level: 2\n"), 0644)

	result := app.RunBuildWithOpts(context.Background(), srcDir, outDir, filepath.Join(srcDir, "baize.yaml"), 2, false, false, false, 2, false, "", "")

	assert.True(t, result.Success)
	assert.Greater(t, result.Summary.Pages, 0)
	assert.FileExists(t, filepath.Join(outDir, "guide", "test.md"))
}

func TestE2ECatalogKnowledgeGraph(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create pages with entities for graph testing
	content1 := []byte("# Page A\n\nThis file defines a core utility.\n")
	content2 := []byte("# Page B\n\nThis file uses the utility from Page A.\n")
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "core"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "app"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "core", "a.md"), content1, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "app", "b.md"), content2, 0644))

	result := app.RunBuildWithOpts(context.Background(), srcDir, outDir, "", 2, false, false, false, 2, false, "", "")
	assert.True(t, result.Success)

	// Verify graph.json exists
	graphPath := filepath.Join(outDir, ".baize", "graph.json")
	assert.FileExists(t, graphPath)
}

func TestE2ECatalogIncrementalBuild(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	content := []byte("# Initial\n\nInitial content.\n")
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "page.md"), content, 0644))

	// First build
	result1 := app.RunBuildWithOpts(context.Background(), srcDir, outDir, "", 2, false, false, false, 2, false, "", "")
	assert.True(t, result1.Success)

	// Incremental build with no changes
	result2 := app.RunBuildWithOpts(context.Background(), srcDir, outDir, "", 2, false, false, false, 2, true, "", "")
	assert.True(t, result2.Success)

	// Modify file
	content2 := []byte("# Modified\n\nModified content.\n")
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "page.md"), content2, 0644))

	// Incremental build with changes
	result3 := app.RunBuildWithOpts(context.Background(), srcDir, outDir, "", 2, false, false, false, 2, true, "", "")
	assert.True(t, result3.Success)
}

func TestE2ECatalogLargeProject(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Generate 100 test files
	for i := 0; i < 100; i++ {
		cat := "category-" + string(rune('A'+i%5))
		dir := filepath.Join(srcDir, cat)
		_ = os.MkdirAll(dir, 0755)
		content := []byte("---\ntitle: Page " + string(rune('0'+i)) + "\n---\n\n# Page\n\nContent " + string(rune('0'+i)))
		_ = os.WriteFile(filepath.Join(dir, "page-"+string(rune('0'+i))+".md"), content, 0644)
	}

	result := app.RunBuildWithOpts(context.Background(), srcDir, outDir, "", 2, false, false, false, 2, false, "", "")
	assert.True(t, result.Success)
	assert.Greater(t, result.Summary.Pages, 0)
}

func TestE2ECatalogWithScanAll(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	codeContent := []byte("using System;\nnamespace Test { class Program { static void Main() {} } }")
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "program.cs"), codeContent, 0644))

	result := app.RunBuildWithOpts(context.Background(), srcDir, outDir, "", 2, false, false, true, 2, false, "", "")
	assert.True(t, result.Success)
	assert.Greater(t, result.Summary.Pages, 0)
}
