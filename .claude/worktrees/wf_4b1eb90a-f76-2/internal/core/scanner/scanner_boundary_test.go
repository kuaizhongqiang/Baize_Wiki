package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanFileNoExtension(t *testing.T) {
	dir := t.TempDir()

	// Files without extension (like Makefile)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:\n\techo hi"), 0644))
	// Normal .md should still be found
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Doc"), 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{
		ScanAll: true, // Phase 2+: need ScanAll to pick up non-.md files
	})
	require.NoError(t, err)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	assert.Contains(t, paths, "doc.md", ".md files should be scanned")
}

func TestScanAllMode(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Doc"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "note.txt"), []byte("plain text"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), 0644))

	// Phase 1: default (ScanAll=false) — only .md/.mdx
	files, err := Scan(context.Background(), dir, ScanConfig{ScanAll: false})
	require.NoError(t, err)
	assert.Len(t, files, 1, "Phase 1: only .md files")
	assert.Equal(t, "doc.md", files[0].Path)

	// Phase 2+: ScanAll=true — all text files
	files, err = Scan(context.Background(), dir, ScanConfig{ScanAll: true})
	require.NoError(t, err)
	assert.Len(t, files, 3, "ScanAll: all text files")
}

func TestScanMaxSizeExact(t *testing.T) {
	dir := t.TempDir()

	// 100 bytes of valid UTF-8 text (non-null to avoid binary detection)
	data := make([]byte, 100)
	for i := range data {
		data[i] = 'A'
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "exact.md"), data, 0644))

	// Scanner uses `info.Size() > cfg.MaxSize` → 100 > 100 = false → file passes
	files, err := Scan(context.Background(), dir, ScanConfig{MaxSize: 100})
	require.NoError(t, err)
	require.Len(t, files, 1, "file exactly at MaxSize should be included (100 > 100 = false)")
	assert.Equal(t, "exact.md", files[0].Path)
}

func TestScanBinaryAtBoundary(t *testing.T) {
	dir := t.TempDir()

	// 512 bytes of valid UTF-8, then a null byte at position 513
	data := make([]byte, 513)
	for i := 0; i < 512; i++ {
		data[i] = 'A'
	}
	data[512] = 0x00 // null byte at byte 513 — outside the 512-byte check window

	require.NoError(t, os.WriteFile(filepath.Join(dir, "boundary.md"), data, 0644))

	// Phase 1: .md extension — should be scanned and pass binary check
	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	require.Len(t, files, 1, "file with null after 512-byte window should pass")
}

func TestScanUnreadableFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "readable.md"), []byte("# Readable"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "unreadable.md"), []byte("# Secret"), 0000))

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)

	// Should still work, unreadable file should be skipped gracefully
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	assert.Contains(t, paths, "readable.md")
}

func TestScanIgnoreNegation(t *testing.T) {
	dir := t.TempDir()

	ignoreContent := "*.md\n!keep.md\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".baizeignore"), []byte(ignoreContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skip.md"), []byte("skip"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "keep.md"), []byte("keep"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doc.txt"), []byte("txt"), 0644))

	// Phase 1: only .md — use ScanAll to test full ignore logic
	files, err := Scan(context.Background(), dir, ScanConfig{ScanAll: true})
	require.NoError(t, err)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	assert.Contains(t, paths, "keep.md", "negated pattern should override ignore")
	assert.Contains(t, paths, "doc.txt", "non-.md files should pass through")
	assert.NotContains(t, paths, "skip.md", "*.md should be ignored")
}

func TestScanDeepDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create 10-level deep directory with a .md file at the bottom
	current := dir
	for i := 0; i < 10; i++ {
		current = filepath.Join(current, "sub")
		require.NoError(t, os.Mkdir(current, 0755))
	}
	require.NoError(t, os.WriteFile(filepath.Join(current, "deep.md"), []byte("# Deep"), 0644))

	// Also put a .md at the root
	require.NoError(t, os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644))

	files, err := Scan(context.Background(), dir, ScanConfig{})
	require.NoError(t, err)
	assert.Len(t, files, 2, "should find both root and deeply nested files")
}

func TestRuleMatcherDoubleStar(t *testing.T) {
	rm := NewRuleMatcher([]string{"**/*.md"})

	assert.True(t, rm.Match("readme.md", false))
	assert.True(t, rm.Match("docs/readme.md", false))
	assert.True(t, rm.Match("a/b/c/doc.md", false))
	assert.False(t, rm.Match("main.go", false))
}

func TestRuleMatcherNegateDir(t *testing.T) {
	rm := NewRuleMatcher([]string{
		"build/",
		"!build/keep/",
	})

	assert.True(t, rm.Match("build/output.txt", false))
	assert.False(t, rm.Match("build/keep/doc.md", false))
	assert.False(t, rm.Match("src/main.go", false))
}

func TestRuleMatcherLeadingSlash(t *testing.T) {
	rm := NewRuleMatcher([]string{"/build"})

	// Current impl strips leading slash, so `/build` matches any `build` at any depth
	// Phase 1 limitation — see issue #9
	assert.True(t, rm.Match("build", false))
	assert.True(t, rm.Match("src/build", false), "leading slash stripped, matches any depth (Phase 1 limitation)")
}
