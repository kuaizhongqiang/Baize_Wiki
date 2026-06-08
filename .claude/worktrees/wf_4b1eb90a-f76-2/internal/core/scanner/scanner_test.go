package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanEmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestScanMixedFiles(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.mdx"), []byte("# MDX"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "note.txt"), []byte("plain text"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), 0644))

	// Binary file with null byte
	require.NoError(t, os.WriteFile(filepath.Join(dir, "image.png"), []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 0644))

	// Phase 1: only .md/.mdx files
	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	assert.Len(t, files, 2)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	assert.Contains(t, paths, "doc.md")
	assert.Contains(t, paths, "doc.mdx")
}

func TestScanMaxSize(t *testing.T) {
	dir := t.TempDir()

	largeContent := make([]byte, 200)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "large.md"), largeContent, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "small.md"), []byte("hi"), 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{MaxSize: 100})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "small.md", files[0].Path)
}

func TestScanSymlink(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	require.NoError(t, os.Mkdir(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "real.md"), []byte("content"), 0644))

	// Create symlink
	linkPath := filepath.Join(dir, "link.md")
	if err := os.Symlink(filepath.Join(subDir, "real.md"), linkPath); err != nil {
		t.Skip("symlink creation not supported on this platform")
	}

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)

	// Only the real file should be returned, not the symlink
	assert.Len(t, files, 1)
	assert.Equal(t, filepath.Join("sub", "real.md"), files[0].Path)
}

func TestScanWithIgnoreFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "keep.md"), []byte("keep"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret.md"), []byte("secret"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".baizeignore"), []byte("secret.md\n"), 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)

	require.Len(t, files, 1)
	assert.Equal(t, "keep.md", files[0].Path)
}

func TestScanIgnoreDotDir(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("[core]"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.md"), []byte("doc"), 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "doc.md", files[0].Path)
}

func TestScanNestedDirectories(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "guide"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "api"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "guide", "start.md"), []byte("# Start"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "api", "reference.md"), []byte("# API"), 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	require.Len(t, files, 2)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	assert.Contains(t, paths, filepath.Join("guide", "start.md"))
	assert.Contains(t, paths, filepath.Join("api", "reference.md"))
}

func TestScanContextCancellation(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 10; i++ {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# test"), 0644))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	files, err := Scan(ctx, dir, ScanConfig{})
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, files)
}

func TestIsBinaryData(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		binary bool
	}{
		{"empty data", []byte{}, false},
		{"plain text", []byte("hello world"), false},
		{"markdown", []byte("# Title\n\nContent"), false},
		{"null byte", []byte{0x00, 0x01, 0x02}, true},
		{"invalid utf8", []byte{0xFF, 0xFE, 0x00}, true},
		{"png header", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, true},
		{"cjk text", []byte("你好世界"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.binary, isBinaryData(tt.data))
		})
	}
}

func TestRuleMatcherBuiltin(t *testing.T) {
	rm := NewBuiltinMatcher()
	assert.True(t, rm.HasBuiltinOnly())

	assert.True(t, rm.Match(".git/config", false))
	assert.True(t, rm.Match("node_modules/pkg/index.js", false))
	assert.True(t, rm.Match("dist/bundle.js", false))
	assert.False(t, rm.Match("src/main.go", false))
	assert.False(t, rm.Match("docs/guide.md", false))
}

func TestRuleMatcherCustom(t *testing.T) {
	rm := NewRuleMatcher([]string{
		"*.log",
		"!important.log",
		"temp/",
	})

	assert.True(t, rm.Match("error.log", false))
	assert.False(t, rm.Match("important.log", false))
	assert.True(t, rm.Match("temp/cache.txt", false))
	assert.False(t, rm.Match("docs/readme.md", false))

	// Directory pattern should only match directories
	assert.True(t, rm.Match("temp", true))
	assert.False(t, rm.Match("temp", false)) // single 'temp' file (no dir) with dir-only pattern
}

func TestRuleMatcherFromFile(t *testing.T) {
	dir := t.TempDir()
	ignoreFile := filepath.Join(dir, ".baizeignore")
	require.NoError(t, os.WriteFile(ignoreFile, []byte("*.secret\n# comment\n*.tmp\n"), 0644))

	rm, err := NewRuleMatcherFromFile(ignoreFile)
	require.NoError(t, err)

	assert.True(t, rm.Match("data.secret", false))
	assert.True(t, rm.Match("cache.tmp", false))
	assert.False(t, rm.Match("doc.md", false))
}

func TestRuleMatcherLine(t *testing.T) {
	rm := NewRuleMatcher(nil) // empty

	// Test that empty and comment lines are ignored
	rm.addPattern("")
	rm.addPattern("# comment")
	rm.addPattern("*.md")

	assert.False(t, rm.Match("main.go", false))
	assert.True(t, rm.Match("readme.md", false))
}

func TestFileInfoHasCorrectFields(t *testing.T) {
	dir := t.TempDir()
	content := []byte("# Test")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.md"), content, 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "test.md", f.Path)
	assert.Equal(t, ".md", f.Extension)
	assert.Equal(t, int64(len(content)), f.Size)
	assert.NotEmpty(t, f.AbsPath)
}
