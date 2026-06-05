package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBuildFn simulates a successful Wiki build without importing app.
var mockBuildFn RunBuildFunc = func(ctx context.Context, source, output, configPath string, level int, draft, quiet bool) BuildResult {
	return BuildResult{
		Success:    true,
		DurationMs: 100,
		Summary: BuildSummary{
			TotalFiles:  1,
			Parsed:      1,
			Pages:       1,
			Directories: 1,
		},
	}
}

var mockBuildFailFn RunBuildFunc = func(ctx context.Context, source, output, configPath string, level int, draft, quiet bool) BuildResult {
	return BuildResult{
		Success: false,
		Errors:  []string{"source not found"},
	}
}

// setupTestWiki creates a temporary Wiki directory with a few pages for testing.
func setupTestWiki(t *testing.T) string {
	t.Helper()
	wikiDir := t.TempDir()

	// Create a simple Wiki structure
	require.NoError(t, os.MkdirAll(filepath.Join(wikiDir, "guide"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "guide", "start.md"), []byte("# Getting Started\n\nWelcome!"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "guide", "advanced.md"), []byte("# Advanced\n\nDeep content."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(wikiDir, "index.md"), []byte("# Wiki Home"), 0644))

	// Create meta.json
	store := storage.NewStore()
	cfg := model.DefaultConfig()
	cfg.Output.Level = 2
	wiki := model.NewWiki("Test Wiki", "./src", wikiDir, cfg)
	wiki.PageCount = 3
	require.NoError(t, store.WriteMeta(wikiDir, wiki))

	return wikiDir
}

func TestToolWikiBuild(t *testing.T) {
	params, _ := json.Marshal(map[string]any{"level": 1})
	result, errObj := toolWikiBuild(mockBuildFn)(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.False(t, toolResult.IsError)
}

func TestToolWikiBuildInvalidSource(t *testing.T) {
	params, _ := json.Marshal(map[string]any{})
	result, errObj := toolWikiBuild(mockBuildFailFn)(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.True(t, toolResult.IsError)
}

func TestToolWikiRead(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiRead(wikiDir)

	params, _ := json.Marshal(map[string]string{"path": "guide/start.md"})
	result, errObj := handler(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.Contains(t, toolResult.Content[0].Text, "Getting Started")
}

func TestToolWikiReadNotFound(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiRead(wikiDir)

	params, _ := json.Marshal(map[string]string{"path": "nonexistent.md"})
	result, errObj := handler(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.True(t, toolResult.IsError)
	assert.Contains(t, toolResult.Content[0].Text, "ERR_PAGE_NOT_FOUND")
}

func TestToolWikiReadPathTraversal(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiRead(wikiDir)

	params, _ := json.Marshal(map[string]string{"path": "../../etc/passwd"})
	_, errObj := handler(context.Background(), params)
	require.NotNil(t, errObj)
	assert.Equal(t, ErrInvalidParams, errObj.Code)
}

func TestToolWikiList(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiList(wikiDir)

	params, _ := json.Marshal(map[string]any{"depth": 2})
	result, errObj := handler(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.Contains(t, toolResult.Content[0].Text, "guide")
	assert.Contains(t, toolResult.Content[0].Text, "start.md")
}

func TestToolWikiListDepth(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiList(wikiDir)

	params, _ := json.Marshal(map[string]any{"depth": 0})
	result, errObj := handler(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.NotContains(t, toolResult.Content[0].Text, "start.md")
}

func TestToolWikiAddNew(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiAdd(wikiDir)

	params, _ := json.Marshal(map[string]string{
		"path":    "new-page.md",
		"content": "# New Page\n\nContent",
	})
	result, errObj := handler(context.Background(), params)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.Contains(t, toolResult.Content[0].Text, "created")

	assert.FileExists(t, filepath.Join(wikiDir, "new-page.md"))
}

func TestToolWikiAddOverwrite(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiAdd(wikiDir)

	params1, _ := json.Marshal(map[string]string{
		"path":    "test.md",
		"content": "# First",
	})
	_, errObj := handler(context.Background(), params1)
	assert.Nil(t, errObj)

	params2, _ := json.Marshal(map[string]string{
		"path":    "test.md",
		"content": "# Second",
	})
	result, errObj := handler(context.Background(), params2)
	assert.Nil(t, errObj)
	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.True(t, toolResult.IsError)
}

func TestToolWikiAddOverwriteAllowed(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiAdd(wikiDir)

	params1, _ := json.Marshal(map[string]string{
		"path":    "test.md",
		"content": "# First",
	})
	_, errObj := handler(context.Background(), params1)
	assert.Nil(t, errObj)

	params2, _ := json.Marshal(map[string]any{
		"path":      "test.md",
		"content":   "# Second",
		"overwrite": true,
	})
	result, errObj := handler(context.Background(), params2)
	assert.Nil(t, errObj)
	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.False(t, toolResult.IsError)
}

func TestToolWikiStats(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiStats(wikiDir)

	result, errObj := handler(context.Background(), nil)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.Contains(t, toolResult.Content[0].Text, "page_count")
	assert.Contains(t, toolResult.Content[0].Text, "Test Wiki")
}

func TestToolWikiStatsNoWiki(t *testing.T) {
	emptyDir := t.TempDir()
	handler := toolWikiStats(emptyDir)

	result, errObj := handler(context.Background(), nil)
	assert.Nil(t, errObj)

	toolResult, ok := result.(MCPToolResult)
	require.True(t, ok)
	assert.True(t, toolResult.IsError)
}

func TestToolWikiAddPathSecurity(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiAdd(wikiDir)

	params, _ := json.Marshal(map[string]string{
		"path":    "../../evil.md",
		"content": "# evil",
	})
	_, errObj := handler(context.Background(), params)
	require.NotNil(t, errObj)
	assert.Equal(t, ErrInvalidParams, errObj.Code)
}

func TestSecureJoin(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		path    string
		wantErr bool
	}{
		{"normal path", "/tmp/wiki", "guide/page.md", false},
		{"simple traversal", "/tmp/wiki", "../outside.md", true},
		{"deep traversal", "/tmp/wiki", "a/../../outside.md", true},
		{"absolute path", "/tmp/wiki", "/etc/passwd", true},
		{"dot path", "/tmp/wiki", ".", false},
		{"empty path", "/tmp/wiki", "", false},
		{"current dir", "/tmp/wiki", "./guide/page.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := secureJoin(tt.base, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterAllTools(t *testing.T) {
	wikiDir := t.TempDir()
	srv := NewServer(newTestTransport(""))
	RegisterAllTools(srv, wikiDir, mockBuildFn)

	assert.Len(t, srv.tools, 5)
	assert.Contains(t, srv.tools, "wiki_build")
	assert.Contains(t, srv.tools, "wiki_read")
	assert.Contains(t, srv.tools, "wiki_list")
	assert.Contains(t, srv.tools, "wiki_add")
	assert.Contains(t, srv.tools, "wiki_stats")
}

func TestToolWikiAddCreatesParentDir(t *testing.T) {
	wikiDir := setupTestWiki(t)
	handler := toolWikiAdd(wikiDir)

	params, _ := json.Marshal(map[string]string{
		"path":    "deep/nested/page.md",
		"content": "# Nested",
	})
	_, errObj := handler(context.Background(), params)
	assert.Nil(t, errObj)

	assert.FileExists(t, filepath.Join(wikiDir, "deep", "nested", "page.md"))
}
