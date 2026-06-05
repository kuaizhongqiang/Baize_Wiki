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

func TestE2EBuildNonMDInSource(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "notes.txt"), []byte("plain text"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "app.go"), []byte("package main"), 0644))

	// Phase 1: should only pick up .md files
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Summary.TotalFiles, "Phase 1: only .md files")
}

func TestE2EBuildNestedSource(t *testing.T) {
	srcDir := t.TempDir()

	// Create nested structure
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "a", "b", "c"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a", "b", "c", "deep.md"), []byte("# Deep"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "root.md"), []byte("# Root"), 0644))

	for _, level := range []int{1, 2, 3} {
		t.Run("level", func(t *testing.T) {
			outDir := t.TempDir()
			result := app.RunBuild(context.Background(), srcDir, outDir, "", level, false, false)
			assert.True(t, result.Success, "level %d build should succeed", level)
		})
	}
}

func TestE2EInitNoSourceArg(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(dir))

	// RunInit with all empty strings should use defaults
	err := app.RunInit("", "", "", false)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "baize.yaml"))
}

func TestE2EInfoOnInvalidDir(t *testing.T) {
	dir := t.TempDir()

	_, err := app.RunInfo(dir, false, false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wiki not found")
}

func TestE2EBuildWithFrontmatterNoTitle(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// File with frontmatter but no title
	content := "---\ntags: [test]\n---\n\n# Heading Title\n"
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "notitle.md"), []byte(content), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	assert.True(t, result.Success)
}

func TestE2EBuildWithFrontmatterDraft(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "published.md"), []byte("---\ntitle: Published\ndraft: false\n---\n\n# Published"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "draft.md"), []byte("---\ntitle: Draft\ndraft: true\n---\n\n# Draft"), 0644))

	// Default: draft pages not filtered out (Phase 1 just includes everything)
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.Summary.Pages, "Phase 1 does not filter drafts")
}
