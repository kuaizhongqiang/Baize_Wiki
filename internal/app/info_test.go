package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInfoNoWiki(t *testing.T) {
	dir := t.TempDir()

	_, err := RunInfo(dir, false, false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wiki not found")
}

func TestRunInfoJSONOutput(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	info, err := RunInfo(outDir, false, false, true)
	require.NoError(t, err)

	infoMap, ok := info.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, infoMap["success"])
	assert.Equal(t, outDir, infoMap["path"])
	assert.Contains(t, infoMap["meta"], "page_count")
}

func TestRunInfoTreeFlag(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	result := RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	_, err := RunInfo(outDir, true, false, false)
	assert.NoError(t, err)
}
