package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInitGenerateFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(dir))

	err := RunInit("./docs", "./wiki", "TestWiki", false)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "baize.yaml"))

	content, err := os.ReadFile(filepath.Join(dir, "baize.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "TestWiki")
	assert.Contains(t, string(content), "./docs")
	assert.Contains(t, string(content), "./wiki")
}

func TestRunInitForce(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(dir))

	// First init
	require.NoError(t, RunInit(".", "./wiki", "First", false))

	// Second init without force should fail
	err := RunInit(".", "./wiki", "Second", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// With force should succeed
	err = RunInit(".", "./wiki", "Second", true)
	assert.NoError(t, err)
}

func TestRunInitDefaultSource(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(dir))

	err := RunInit("", "", "", false)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "baize.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `paths:`)
}
