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

func TestE2EBuildLevel1(t *testing.T) {
	// Prepare test source
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc1.md"), []byte("---\ntitle: Doc 1\ncategory: guide\n---\n\n# Doc 1\n\nContent"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc2.md"), []byte("---\ntitle: Doc 2\ncategory: api\n---\n\n# Doc 2\n\nContent"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)

	assert.True(t, result.Success)
	assert.Equal(t, 2, result.Summary.Pages)

	assert.FileExists(t, filepath.Join(outDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outDir, "guide.md"))
	assert.FileExists(t, filepath.Join(outDir, "api.md"))
	assert.FileExists(t, filepath.Join(outDir, ".baize", "meta.json"))
}

func TestE2EBuildLevel2(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "guide"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "api"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "start.md"), []byte("---\ntitle: Start\n---\n\n# Start\n\nGuide content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "api", "ref.md"), []byte("---\ntitle: API Ref\n---\n\n# API\n\nAPI content"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 2, false, false)

	assert.True(t, result.Success)
	assert.Equal(t, 2, result.Summary.Pages)

	assert.FileExists(t, filepath.Join(outDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outDir, "guide", "_index.md"))
	assert.FileExists(t, filepath.Join(outDir, "guide", "start.md"))
	assert.FileExists(t, filepath.Join(outDir, "api", "_index.md"))
	assert.FileExists(t, filepath.Join(outDir, "api", "ref.md"))
}

func TestE2EBuildLevel3(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "guide", "advanced"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "start.md"), []byte("---\ntitle: Start\n---\n\n# Start"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "advanced", "features.md"), []byte("---\ntitle: Features\n---\n\n# Features"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 3, false, false)

	assert.True(t, result.Success)
	assert.Equal(t, 2, result.Summary.Pages)

	assert.FileExists(t, filepath.Join(outDir, "guide", "_index.md"))
	assert.FileExists(t, filepath.Join(outDir, "guide", "start.md"))
	assert.FileExists(t, filepath.Join(outDir, "guide", "advanced", "_index.md"))
	assert.FileExists(t, filepath.Join(outDir, "guide", "advanced", "features.md"))
}

func TestE2EBuildBinarySkipped(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Valid doc"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "image.png"), []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)

	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Summary.Pages)
	assert.FileExists(t, filepath.Join(outDir, "uncategorized.md"))
}

func TestE2EBuildEmptySource(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	assert.False(t, result.Success)
	assert.Contains(t, result.Errors[0], "no valid files")
}

func TestE2EBuildInvalidSource(t *testing.T) {
	outDir := t.TempDir()

	result := app.RunBuild(context.Background(), "/nonexistent/path", outDir, "", 1, false, false)
	assert.False(t, result.Success)
	assert.Contains(t, result.Errors[0], "does not exist")
}

func TestE2EInit(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(dir))

	err := app.RunInit("./docs", "./wiki", "TestWiki", false)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "baize.yaml"))

	content, err := os.ReadFile(filepath.Join(dir, "baize.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "TestWiki")
	assert.Contains(t, string(content), "./docs")
}

func TestE2EInitForce(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(dir))

	// First init
	require.NoError(t, app.RunInit(".", "./wiki", "First", false))

	// Second init without force should fail
	err := app.RunInit(".", "./wiki", "Second", false)
	assert.Error(t, err)

	// With force should succeed
	err = app.RunInit(".", "./wiki", "Second", true)
	assert.NoError(t, err)
}

func TestE2EBuildJSONOutput(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, true)
	assert.True(t, result.Success)
	assert.Greater(t, result.DurationMs, int64(0))
	assert.Equal(t, 1, result.Summary.TotalFiles)
}

func TestE2EBuildWithIndexContent(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("---\ntitle: My Doc\n---\n\n# My Doc\n\nHello world"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	assert.True(t, result.Success)

	content, err := os.ReadFile(filepath.Join(outDir, "_index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Wiki Overview")

	metaContent, err := os.ReadFile(filepath.Join(outDir, ".baize", "meta.json"))
	require.NoError(t, err)
	assert.Contains(t, string(metaContent), `"page_count": 1`)
}

func TestE2EBuildExistingWikiIdempotent(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Stable"), 0644))

	r1 := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	r2 := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)

	assert.True(t, r1.Success)
	assert.True(t, r2.Success)
	assert.Equal(t, r1.Summary.Pages, r2.Summary.Pages)
}

func TestE2EInfo(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))
	_ = app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, true)

	// Verify meta.json exists and is readable
	assert.FileExists(t, filepath.Join(outDir, ".baize", "meta.json"))
}
