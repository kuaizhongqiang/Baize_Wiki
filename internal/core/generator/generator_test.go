package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makePage(path, title, category string, weight int) *model.Page {
	w := weight
	return &model.Page{
		ID:    model.PageID(path),
		Path:  path,
		Title: title,
		Content: `# ` + title + `

Content for ` + title,
		Meta: model.Frontmatter{
			Title:    title,
			Category: category,
			Weight:   &w,
		},
		Weight: weight,
	}
}

func TestLevelBuilderFlat(t *testing.T) {
	pages := []*model.Page{
		makePage("guide/getting-started.md", "Getting Started", "guide", 1),
		makePage("guide/advanced.md", "Advanced Usage", "guide", 2),
		makePage("api/reference.md", "API Reference", "api", 1),
	}

	builder := NewLevelBuilder()
	root := builder.Flat(pages)

	assert.Equal(t, "wiki", root.Name)
	assert.Len(t, root.Pages, 2)

	// Check page refs exist for both categories
	refTitles := make([]string, len(root.Pages))
	for i, p := range root.Pages {
		refTitles[i] = p.Title
	}
	assert.Contains(t, refTitles, "guide")
	assert.Contains(t, refTitles, "api")
}

func TestLevelBuilderFlatSinglePage(t *testing.T) {
	pages := []*model.Page{
		makePage("doc.md", "Doc", "", 1),
	}

	builder := NewLevelBuilder()
	root := builder.Flat(pages)

	assert.Len(t, root.Pages, 1)
	assert.Equal(t, "uncategorized", root.Pages[0].Title)
}

func TestLevelBuilderFlatEmpty(t *testing.T) {
	builder := NewLevelBuilder()
	root := builder.Flat(nil)

	assert.Empty(t, root.Pages)
	assert.Empty(t, root.Children)
}

func TestLevelBuilderStructured(t *testing.T) {
	pages := []*model.Page{
		makePage("guide/getting-started.md", "Getting Started", "", 0),
		makePage("guide/advanced.md", "Advanced Usage", "", 0),
		makePage("api/reference.md", "API Reference", "", 0),
	}

	builder := NewLevelBuilder()
	root := builder.Structured(pages)

	require.Len(t, root.Children, 2)
	assert.Equal(t, "api", root.Children[0].Name)
	assert.Equal(t, "guide", root.Children[1].Name)

	assert.Len(t, root.Children[0].Pages, 1)
	assert.Len(t, root.Children[1].Pages, 2)
}

func TestLevelBuilderDeep(t *testing.T) {
	pages := []*model.Page{
		makePage("guide/getting-started.md", "Getting Started", "", 0),
		makePage("guide/advanced/features.md", "Features", "", 0),
		makePage("api/reference/v1/endpoints.md", "Endpoints", "", 0),
	}

	builder := NewLevelBuilder()
	root := builder.Deep(pages)

	require.Len(t, root.Children, 2)

	// Find guide dir
	var guide, api *DirNode
	for _, c := range root.Children {
		if c.Name == "guide" {
			guide = c
		}
		if c.Name == "api" {
			api = c
		}
	}
	require.NotNil(t, guide)
	require.NotNil(t, api)

	// Guide should have 2 pages (one at root, one nested)
	if assert.Len(t, guide.Pages, 1) {
		assert.Equal(t, "Getting Started", guide.Pages[0].Title)
	}
	assert.Len(t, guide.Children, 1)

	// API should have nested reference/v1 structure
	assert.Len(t, api.Children, 1)
	assert.Equal(t, "reference", api.Children[0].Name)
}

func TestGeneratorGenerate(t *testing.T) {
	outputDir := t.TempDir()

	pages := []*model.Page{
		makePage("guide/getting-started.md", "Getting Started", "", 0),
		makePage("guide/advanced.md", "Advanced Usage", "", 0),
		makePage("api/reference.md", "API Reference", "", 0),
	}

	cfg := model.DefaultConfig()
	cfg.Output.Level = 2
	wiki := model.NewWiki("Test Wiki", "./src", outputDir, cfg)

	store := storage.NewStore()
	gen := NewGenerator(store)

	err := gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	// Verify output structure
	assert.FileExists(t, filepath.Join(outputDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "getting-started.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "advanced.md"))
	assert.FileExists(t, filepath.Join(outputDir, "api", "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "api", "reference.md"))
	assert.FileExists(t, filepath.Join(outputDir, ".baize", "meta.json"))

	// Verify meta.json content
	w, err := store.ReadMeta(outputDir)
	require.NoError(t, err)
	assert.Equal(t, "Test Wiki", w.Name)
	assert.Equal(t, 3, w.PageCount)
}

func TestGeneratorLevel1(t *testing.T) {
	outputDir := t.TempDir()

	pages := []*model.Page{
		makePage("guide/start.md", "Start", "guide", 1),
		makePage("api/ref.md", "Reference", "api", 1),
	}

	cfg := model.DefaultConfig()
	cfg.Output.Level = 1
	wiki := model.NewWiki("Flat Wiki", "./src", outputDir, cfg)

	store := storage.NewStore()
	gen := NewGenerator(store)

	err := gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	// Level 1: flat output, only _index.md and merged pages
	assert.FileExists(t, filepath.Join(outputDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide.md"))
	assert.FileExists(t, filepath.Join(outputDir, "api.md"))
	assert.NoDirExists(t, filepath.Join(outputDir, "guide"))
	assert.NoDirExists(t, filepath.Join(outputDir, "api"))
}

func TestGeneratorLevel3(t *testing.T) {
	outputDir := t.TempDir()

	pages := []*model.Page{
		makePage("guide/getting-started.md", "Getting Started", "", 0),
		makePage("guide/advanced/features.md", "Features", "", 0),
	}

	cfg := model.DefaultConfig()
	cfg.Output.Level = 3
	wiki := model.NewWiki("Deep Wiki", "./src", outputDir, cfg)

	store := storage.NewStore()
	gen := NewGenerator(store)

	err := gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "getting-started.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "advanced", "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "guide", "advanced", "features.md"))
}

func TestGeneratorEmptyPages(t *testing.T) {
	outputDir := t.TempDir()

	cfg := model.DefaultConfig()
	wiki := model.NewWiki("Empty", "./src", outputDir, cfg)

	store := storage.NewStore()
	gen := NewGenerator(store)

	err := gen.Generate(context.Background(), wiki, nil)
	require.NoError(t, err)

	// Should still create _index.md and meta.json
	assert.FileExists(t, filepath.Join(outputDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, ".baize", "meta.json"))

	w, err := store.ReadMeta(outputDir)
	require.NoError(t, err)
	assert.Equal(t, 0, w.PageCount)
}

func TestGeneratorIdempotent(t *testing.T) {
	outputDir := t.TempDir()

	pages := []*model.Page{
		makePage("doc.md", "Doc", "", 0),
	}

	cfg := model.DefaultConfig()
	wiki := model.NewWiki("Test", "./src", outputDir, cfg)

	store := storage.NewStore()
	gen := NewGenerator(store)

	// Run twice
	err := gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	err = gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "_index.md"))
	assert.FileExists(t, filepath.Join(outputDir, "doc.md"))
}

func TestIndexContent(t *testing.T) {
	outputDir := t.TempDir()

	pages := []*model.Page{
		makePage("guide/start.md", "Start Guide", "", 0),
		makePage("api/reference.md", "API Ref", "", 0),
	}

	cfg := model.DefaultConfig()
	cfg.Output.Level = 2
	wiki := model.NewWiki("Test", "./src", outputDir, cfg)

	store := storage.NewStore()
	gen := NewGenerator(store)

	err := gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	// Read root _index.md
	content, err := os.ReadFile(filepath.Join(outputDir, "_index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Wiki Overview")
	assert.Contains(t, string(content), "api/")
	assert.Contains(t, string(content), "guide/")

	// Read guide _index.md
	guideContent, err := os.ReadFile(filepath.Join(outputDir, "guide", "_index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "guide")
	assert.Contains(t, string(guideContent), "Start Guide")
}

func TestLevelBuilderBuild(t *testing.T) {
	pages := []*model.Page{makePage("test.md", "Test", "", 0)}
	builder := NewLevelBuilder()

	root1 := builder.Build(pages, 1)
	assert.Equal(t, "wiki", root1.Name)
	assert.Len(t, root1.Pages, 1)

	root2 := builder.Build(pages, 2)
	assert.Len(t, root2.Pages, 0)

	root3 := builder.Build(pages, 3)
	assert.Len(t, root3.Pages, 1)
}
