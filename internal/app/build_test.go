package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunBuildSuccess(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Test Page"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)

	assert.True(t, result.Success, "build should succeed")
	assert.Equal(t, 1, result.Summary.TotalFiles)
	assert.Equal(t, 1, result.Summary.Pages)
	assert.Greater(t, result.DurationMs, int64(0))
	assert.Empty(t, result.Errors)
}

func TestRunBuildInvalidSource(t *testing.T) {
	outDir := t.TempDir()

	result := RunBuild(context.Background(), "/nonexistent/path", outDir, "", 1, false, false, false)

	assert.False(t, result.Success)
	assert.Contains(t, result.Errors[0], "does not exist")
}

func TestRunBuildEmptySource(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)

	assert.False(t, result.Success)
	assert.Contains(t, result.Errors[0], "no valid files")
}

func TestRunBuildJSONOutput(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, true, false)

	assert.True(t, result.Success)
	assert.Greater(t, result.DurationMs, int64(0))
	assert.Equal(t, 1, result.Summary.TotalFiles)
}

func TestRunBuildWithFrontmatter(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	content := "---\ntitle: My Page\n---\n\n# My Page\n\nHello world"
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "page.md"), []byte(content), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)

	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Summary.Pages)
}

func TestRunBuildMultipleLevels(t *testing.T) {
	srcDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "guide"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "start.md"), []byte("# Start"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "api.md"), []byte("# API"), 0644))

	for _, level := range []int{1, 2, 3} {
		t.Run("level", func(t *testing.T) {
			outDir := t.TempDir()
			result := RunBuild(context.Background(), srcDir, outDir, "", level, false, false, false)
			assert.True(t, result.Success, "level %d should succeed", level)
			assert.Equal(t, 2, result.Summary.TotalFiles)
		})
	}
}
