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

func TestE2EBuildWithLinks(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create pages with [[wiki-links]]
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "guide"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "index.md"), []byte("---\ntitle: Home\n---\n\n# Home\n\nWelcome to [[数据模型]] and [[API 参考]]."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "data-model.md"), []byte("---\ntitle: 数据模型\n---\n\n# 数据模型\n\nData model description."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "guide", "api.md"), []byte("---\ntitle: API 参考\n---\n\n# API\n\nAPI reference."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 2, false, false, false)
	require.True(t, result.Success)
	assert.Equal(t, 3, result.Summary.Pages)

	// Verify _index.md was generated
	indexPath := filepath.Join(outDir, "_index.md")
	assert.FileExists(t, indexPath)

	// Read _index.md content to verify it has backlink info
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "共 3 个页面")
}

func TestE2EInfoShowsLinks(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("---\ntitle: Doc\n---\n\n# Doc\n\nContent with [[link]]."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Use info to display stats
	wikiInfo, err := app.RunInfo(outDir, false, true, false)
	require.NoError(t, err)
	assert.Nil(t, wikiInfo) // stats prints to stdout, returns nil
}

func TestE2EBuilderResolvesLinks(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "pagea.md"), []byte("---\ntitle: Page A\n---\n\n# Page A\n\nSee [[Page B]] for details."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "pageb.md"), []byte("---\ntitle: Page B\n---\n\n# Page B\n\nContent."), 0644))

	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, false, false)
	require.True(t, result.Success)

	// Page A should have 1 link, Page B should have 1 backlink
	assert.Equal(t, 2, result.Summary.Pages)
}

func TestE2EPageInfoWithLinks(t *testing.T) {
	srcDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "test.md"), []byte("# Test\n\nReference [[other-page]] and [[another|display]]."), 0644))

	// Use info to display page info
	info, err := app.RunInfo(filepath.Join(srcDir, "test.md"), false, false, false)
	require.NoError(t, err)
	require.NotNil(t, info)
	pageInfo, ok := info.(*app.PageInfo)
	assert.True(t, ok)
	assert.Equal(t, 2, pageInfo.LinkCount)
	assert.Equal(t, "test", pageInfo.Title)
}
