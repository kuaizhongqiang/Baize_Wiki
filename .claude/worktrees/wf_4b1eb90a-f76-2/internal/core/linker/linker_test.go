package linker

import (
	"context"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkExactMatch(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "guide/data-model.md", Title: "数据模型"},
		{ID: "p2", Path: "guide/api.md", Title: "API 参考",
			Links: []model.Link{
				{SourceID: "p2", TargetPath: "guide/data-model.md"},
			}},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Len(t, pages[1].Links, 1)
	assert.Equal(t, "p1", pages[1].Links[0].TargetID, "should match by exact path")
	assert.Equal(t, model.LinkInternal, pages[1].Links[0].Type)
}

func TestLinkTitleMatch(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "guide/data-model.md", Title: "数据模型"},
		{ID: "p2", Path: "guide/api.md", Title: "API 参考"},
	}

	// Page with [[link]] in content
	pages[1].Links = []model.Link{
		{SourceID: "p2", TargetPath: "数据模型"},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Equal(t, "p1", pages[1].Links[0].TargetID, "should match by title")
}

func TestLinkFileNameMatch(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "guide/data-model.md", Title: "数据模型规范"},
		{ID: "p2", Path: "guide/api.md", Title: "API"},
	}

	pages[1].Links = []model.Link{
		{SourceID: "p2", TargetPath: "data-model"},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Equal(t, "p1", pages[1].Links[0].TargetID, "should match by filename")
}

func TestLinkDangling(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "guide/api.md", Title: "API"},
	}

	pages[0].Links = []model.Link{
		{SourceID: "p1", TargetPath: "不存在的页面"},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Empty(t, pages[0].Links[0].TargetID, "dangling link should have empty TargetID")
	assert.Equal(t, "不存在的页面", pages[0].Links[0].TargetPath)
}

func TestBacklinks(t *testing.T) {
	pageA := &model.Page{ID: "a", Path: "doc/a.md", Title: "Page A"}
	pageB := &model.Page{ID: "b", Path: "doc/b.md", Title: "Page B",
		Links: []model.Link{
			{SourceID: "b", TargetPath: "doc/a.md"},
		}}
	pageC := &model.Page{ID: "c", Path: "doc/c.md", Title: "Page C",
		Links: []model.Link{
			{SourceID: "c", TargetPath: "doc/a.md"},
		}}

	pages := []*model.Page{pageA, pageB, pageC}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Len(t, pageA.Backlinks, 2, "page A should have 2 backlinks")
	assert.Len(t, pageB.Backlinks, 0, "page B should have 0 backlinks")
}

func TestLinkEmptyPages(t *testing.T) {
	l := New()
	err := l.Link(context.Background(), []*model.Page{})
	require.NoError(t, err)
}

func TestExternalLink(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "doc/page.md", Title: "Page",
			Links: []model.Link{
				{SourceID: "p1", TargetPath: "https://example.com"},
				{SourceID: "p1", TargetPath: "http://test.org"},
			}},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Equal(t, model.LinkExternal, pages[0].Links[0].Type)
	assert.Equal(t, model.LinkExternal, pages[0].Links[1].Type)
	assert.Empty(t, pages[0].Links[0].TargetID, "external links have no TargetID")
}

func TestResourceLink(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "doc/page.md", Title: "Page",
			Links: []model.Link{
				{SourceID: "p1", TargetPath: "./image.png"},
				{SourceID: "p1", TargetPath: "assets/style.css"},
			}},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Equal(t, model.LinkResource, pages[0].Links[0].Type)
	assert.Equal(t, model.LinkResource, pages[0].Links[1].Type)
}

func TestAnchorLink(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "doc/page.md", Title: "Page",
			Links: []model.Link{
				{SourceID: "p1", TargetPath: "#installation"},
			}},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Equal(t, model.LinkInternal, pages[0].Links[0].Type)
	assert.Empty(t, pages[0].Links[0].TargetID, "anchor links have no TargetID")
}

func TestLinkExtractsFromContent(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "doc/a.md", Title: "Page A", Content: "Hello [[Page B]] world."},
		{ID: "p2", Path: "doc/b.md", Title: "Page B", Content: "Response."},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	assert.Len(t, pages[0].Links, 1, "should extract link from content")
	assert.Equal(t, "p2", pages[0].Links[0].TargetID)
}

func TestLinkCaseSensitivity(t *testing.T) {
	pages := []*model.Page{
		{ID: "p1", Path: "doc/Data-Model.md", Title: "Data Model"},
		{ID: "p2", Path: "doc/api.md", Title: "API",
			Links: []model.Link{
				{SourceID: "p2", TargetPath: "doc/data-model.md"},
			}},
	}

	l := New()
	err := l.Link(context.Background(), pages)
	require.NoError(t, err)

	// Case-sensitive match should fail — "data-model.md" != "Data-Model.md"
	assert.Empty(t, pages[1].Links[0].TargetID, "case-sensitive mismatch should be dangling")
}
