package index

import (
	"context"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestPage(path, title, content string, tags []string) *model.Page {
	return &model.Page{
		ID:      model.PageID(path),
		Path:    path,
		Title:   title,
		Content: content,
		Tags:    tags,
		Meta: model.Frontmatter{
			Title: title,
			Tags:  tags,
		},
	}
}

func TestBuildAndSearch(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx.Close()

	pages := []*model.Page{
		makeTestPage("guide/start.md", "快速开始", "欢迎使用 Baize Wiki。这是一个文档工具。", []string{"guide"}),
		makeTestPage("api/reference.md", "API 参考", "REST API 接口文档，包含认证和授权。", []string{"api", "reference"}),
	}

	err = idx.Build(context.Background(), pages)
	require.NoError(t, err)

	results, err := idx.Search(context.Background(), "API", SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Contains(t, results[0].Path, "api")
}

func TestSearchNoResults(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx.Close()

	pages := []*model.Page{
		makeTestPage("doc.md", "Doc", "some content", nil),
	}
	err = idx.Build(context.Background(), pages)
	require.NoError(t, err)

	results, err := idx.Search(context.Background(), "nonexistent_keyword_xyz", SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchTagFilter(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx.Close()

	pages := []*model.Page{
		makeTestPage("guide.md", "Guide", "setup guide", []string{"guide"}),
		makeTestPage("api.md", "API Ref", "api documentation", []string{"api"}),
	}
	err = idx.Build(context.Background(), pages)
	require.NoError(t, err)

	results, err := idx.Search(context.Background(), "", SearchOpts{Tags: []string{"api"}, Limit: 10})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "api.md", results[0].Path)
}

func TestSearchLimit(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx.Close()

	pages := make([]*model.Page, 5)
	for i := 0; i < 5; i++ {
		pages[i] = makeTestPage(
			string(rune('A'+i))+".md",
			"Page", "content about searchable text", nil)
	}
	err = idx.Build(context.Background(), pages)
	require.NoError(t, err)

	results, err := idx.Search(context.Background(), "searchable", SearchOpts{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestReBuild(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)

	pages1 := []*model.Page{
		makeTestPage("a.md", "A", "content a", nil),
	}
	err = idx.Build(context.Background(), pages1)
	require.NoError(t, err)
	idx.Close()

	// Re-open and add new data (incremental — old docs persist)
	idx2, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx2.Close()

	pages2 := []*model.Page{
		makeTestPage("b.md", "B", "content b", nil),
	}
	err = idx2.Build(context.Background(), pages2)
	require.NoError(t, err)

	// Both old and new docs should be searchable (incremental)
	resultsB, err := idx2.Search(context.Background(), "b", SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, resultsB)

	resultsA, err := idx2.Search(context.Background(), "content a", SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, resultsA, "old docs should persist (incremental)")
}

func TestNewIndexCreates(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	assert.NotNil(t, idx)
	idx.Close()
}

func TestNewIndexOpensExisting(t *testing.T) {
	dir := t.TempDir()
	idx1, err := NewIndex(dir)
	require.NoError(t, err)
	idx1.Close()

	idx2, err := NewIndex(dir)
	require.NoError(t, err)
	assert.NotNil(t, idx2)
	idx2.Close()
}

func TestSearchWithSnippet(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx.Close()

	longContent := "The quick brown fox jumps over the lazy dog. " +
		"This is a test document with enough content to generate a meaningful snippet. " +
		"We need to verify that the snippet functionality works correctly."
	pages := []*model.Page{
		makeTestPage("test.md", "Test", longContent, nil),
	}
	err = idx.Build(context.Background(), pages)
	require.NoError(t, err)

	results, err := idx.Search(context.Background(), "fox", SearchOpts{Limit: 10})
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.NotEmpty(t, results[0].Snippet)
}

func TestSearchEmptyQuery(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewIndex(dir)
	require.NoError(t, err)
	defer idx.Close()

	pages := []*model.Page{
		makeTestPage("doc.md", "Doc", "content", nil),
	}
	err = idx.Build(context.Background(), pages)
	require.NoError(t, err)

	results, err := idx.Search(context.Background(), "", SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}
