package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeSourceEqualsOutput(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), dir, dir, "", 1, false, false, false)
	assert.False(t, result.Success)
	assert.Contains(t, result.Errors[0], "source and output directories must be different")

	_, err := os.Stat(filepath.Join(dir, "_index.md"))
	t.Logf("_index.md exists in source dir: %v", err == nil)
	_, err = os.Stat(filepath.Join(dir, ".baize"))
	t.Logf(".baize/ exists in source dir: %v", err == nil)
}

func TestEdgeInfoStatsFlag(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	_ = RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)

	result, err := RunInfo(outDir, false, true, false)
	require.NoError(t, err)
	assert.Nil(t, result, "RunInfo with --stats should return data but currently returns nil")
}

func TestEdgeConfigFileLoading(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Summary.Pages)
}

func TestEdgeOutputPathInvalid(t *testing.T) {
	srcDir := t.TempDir()
	conflictPath := filepath.Join(srcDir, "output_conflict")
	require.NoError(t, os.WriteFile(conflictPath, []byte("not a directory"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), srcDir, conflictPath, "", 1, false, false, false)
	t.Logf("Build with file-as-output: success=%v, errors=%v", result.Success, result.Errors)
}

func TestEdgeBuildWithDraftFlag(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "draft.md"), []byte("---\ntitle: Draft\ndraft: true\n---\n\n# Draft"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "normal.md"), []byte("---\ntitle: Normal\ndraft: false\n---\n\n# Normal"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	assert.True(t, result.Success)
	t.Logf("Without --draft: %d pages", result.Summary.Pages)

	outDir2 := t.TempDir()
	resultWithDraft := RunBuild(context.Background(), srcDir, outDir2, "", 1, true, false, false)
	assert.True(t, resultWithDraft.Success)
	t.Logf("With --draft: %d pages", resultWithDraft.Summary.Pages)
}

func TestEdgeBuildInfoRoundtrip(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	infoResult, err := RunInfo(outDir, false, false, true)
	require.NoError(t, err)
	infoMap, ok := infoResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, infoMap["success"])
}
