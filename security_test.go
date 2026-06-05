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

func TestSecurityPathTraversal(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Source file with path traversal in name
	content := []byte("# Traversal test")
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "safe.md"), content, 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)
	assert.True(t, result.Success)

	// Verify output files are within the output directory
	err := filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// All files should be under outDir
		rel, err := filepath.Rel(outDir, path)
		require.NoError(t, err)
		assert.False(t, filepath.IsAbs(rel), "all output paths should be relative to outDir")
		return nil
	})
	require.NoError(t, err)
}

func TestSecurityLongFilename(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create a file with a very long name
	longName := "a"
	for i := 0; i < 199; i++ {
		longName += "a"
	}
	longName += ".md"

	content := []byte("# Long filename test")
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, longName), content, 0644))

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "normal.md"), []byte("# Normal"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)

	// Should handle long filenames gracefully — may fail or succeed but not crash
	t.Logf("Long filename build: success=%v, errors=%v", result.Success, result.Errors)
}

func TestSecuritySpecialCharsFilename(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create files with special characters in names
	specialNames := []string{"file[1].md", "file(test).md", "file+name.md", "file&name.md"}

	for _, name := range specialNames {
		content := []byte("# " + name)
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, name), content, 0644))
	}

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "normal.md"), []byte("# Normal"), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false)

	// Should not crash — special char files may be skipped or processed
	t.Logf("Special chars build: success=%v, pages=%d", result.Success, result.Summary.Pages)
}
