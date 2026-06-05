package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWritePage(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	err := s.WritePage(dir, "hello.md", "# Hello World")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "hello.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Hello World", string(content))
}

func TestWritePageCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	err := s.WritePage(dir, "guide/start.md", "# Start")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "guide", "start.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Start", string(content))
}

func TestWritePageAtomicity(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	err := s.WritePage(dir, "page.md", "content")
	require.NoError(t, err)

	// Verify no .tmp file remains
	matches, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	require.NoError(t, err)
	assert.Empty(t, matches, "no .tmp files should remain after successful write")
}

func TestWritePageOverwrite(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	err := s.WritePage(dir, "page.md", "version 1")
	require.NoError(t, err)

	err = s.WritePage(dir, "page.md", "version 2")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "page.md"))
	require.NoError(t, err)
	assert.Equal(t, "version 2", string(content))
}

func TestWriteMeta(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	wiki := &model.Wiki{
		ID:   "test-wiki",
		Name: "Test Wiki",
	}

	err := s.WriteMeta(dir, wiki)
	require.NoError(t, err)

	// Verify file exists
	metaPath := filepath.Join(dir, ".baize", "meta.json")
	assert.FileExists(t, metaPath)

	// Verify content is valid JSON
	data, err := os.ReadFile(metaPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"id": "test-wiki"`)
	assert.Contains(t, string(data), `"name": "Test Wiki"`)

	// Verify no .tmp files
	matches, err := filepath.Glob(filepath.Join(dir, ".baize", "*.tmp"))
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestReadMeta(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	wiki := &model.Wiki{
		ID:        "test-wiki",
		Name:      "Test Wiki",
		Version:   3,
		PageCount: 10,
	}

	err := s.WriteMeta(dir, wiki)
	require.NoError(t, err)

	read, err := s.ReadMeta(dir)
	require.NoError(t, err)
	assert.Equal(t, "test-wiki", read.ID)
	assert.Equal(t, "Test Wiki", read.Name)
	assert.Equal(t, 3, read.Version)
	assert.Equal(t, 10, read.PageCount)
}

func TestReadMetaNotFound(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()

	_, err := s.ReadMeta(dir)
	assert.ErrorIs(t, err, model.ErrWikiNotFound)
}
