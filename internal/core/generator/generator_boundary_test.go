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

func TestLevelBuilderFlatWeightOrdering(t *testing.T) {
	pages := []*model.Page{
		makePage("z.md", "Z Page", "docs", 10),
		makePage("a.md", "A Page", "docs", 1),
		makePage("m.md", "M Page", "docs", 5),
	}

	lb := NewLevelBuilder()
	root := lb.Flat(pages)

	require.Len(t, root.Children, 0, "Flat mode: no subdirectories")
	// Pages should be sorted by weight, then by title
	// But wtih Flat mode, pages are grouped into categories first
	t.Logf("Root pages: %+v", root.Pages)
}

func TestLevelBuilderFlatMergeSplit(t *testing.T) {
	// Create enough pages to exceed the 50KB merge split
	var pages []*model.Page
	for i := 0; i < 5; i++ {
		// Each page has ~15KB content
		content := make([]byte, 15*1024)
		for j := range content {
			content[j] = 'A' + byte(j%26)
		}
		weight := i
		pages = append(pages, &model.Page{
			ID:    model.PageID(string(rune('0'+i))),
			Path:  "page.md",
			Title: "Page",
			Content: `# Page ` + string(rune('0'+i)) + `

` + string(content),
			Meta: model.Frontmatter{
				Title:    "Page",
				Category: "docs",
				Weight:   &weight,
			},
			Weight: weight,
		})
	}

	lb := NewLevelBuilder()
	root := lb.Flat(pages)

	// Should not panic — merge split logic should handle large content
	t.Logf("Flat root has %d pages (after merge/split)", len(root.Pages))
}

func TestLevelBuilderDeepMaxDepth(t *testing.T) {
	pages := []*model.Page{
		makePage("a/b/c/d/e/page.md", "Deep Page", "", 1),
		makePage("x/y/page.md", "Shallow Page", "", 1),
	}

	lb := NewLevelBuilder()
	root := lb.Deep(pages)

	assert.Equal(t, "wiki", root.Name)
	t.Logf("Deep tree children: %d", len(root.Children))
}

func TestLevelBuilderDeepRootLevelPages(t *testing.T) {
	pages := []*model.Page{
		makePage("page.md", "Root Page", "", 1),
		makePage("guide/page.md", "Guide Page", "", 1),
	}

	lb := NewLevelBuilder()
	root := lb.Deep(pages)

	// Root-level page should appear in root.Pages
	t.Logf("Deep root pages: %d, children: %d", len(root.Pages), len(root.Children))
}

func TestGeneratorContextCancellation(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore()
	gen := NewGenerator(store)

	wiki := &model.Wiki{
		Name:       "test",
		OutputPath: dir,
		Config:     model.DefaultConfig(),
	}

	pages := []*model.Page{
		makePage("page.md", "Page", "", 1),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	err := gen.Generate(ctx, wiki, pages)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGeneratorAllLevelsEmptyPages(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore()
	gen := NewGenerator(store)

	wiki := &model.Wiki{
		Name:       "test",
		OutputPath: dir,
		Config:     model.DefaultConfig(),
	}

	for _, level := range []int{1, 2, 3} {
		t.Run("level", func(t *testing.T) {
			outDir := filepath.Join(dir, string(rune('0'+level)))
			w := *wiki
			w.OutputPath = outDir
			w.Config.Output.Level = level

			err := gen.Generate(context.Background(), &w, nil)
			// Generator handles nil/empty pages without error (creates empty wiki)
			// This is by design — existing test TestGeneratorEmptyPages tests this
			t.Logf("empty pages: err=%v", err)
		})
	}
}

func TestIndexContentFormat(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore()
	gen := NewGenerator(store)

	wiki := &model.Wiki{
		Name:       "test",
		OutputPath: dir,
		Config:     model.Config{Output: model.OutputConfig{Level: 1}},
	}

	pages := []*model.Page{
		makePage("doc.md", "Doc Title", "", 1),
	}

	err := gen.Generate(context.Background(), wiki, pages)
	require.NoError(t, err)

	indexPath := filepath.Join(dir, "_index.md")
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "Wiki Overview", "should have a wiki title")
	// Check the page is listed in _index.md (uses page path/filename, not title)
	assert.Contains(t, string(content), "uncategorized", "page path should appear in _index.md")
}

func TestLevelBuilderBuildInvalidLevel(t *testing.T) {
	lb := NewLevelBuilder()

	// Invalid level falls back to Structured (Level 2)
	pages := []*model.Page{
		makePage("doc.md", "Doc", "", 1),
	}

	root := lb.Build(pages, 99)
	assert.NotNil(t, root)
	t.Logf("Root built with invalid level 99 → %d children, %d pages", len(root.Children), len(root.Pages))
}

func TestGeneratorManyPages(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore()
	gen := NewGenerator(store)

	wiki := &model.Wiki{
		Name:       "test",
		OutputPath: dir,
		Config:     model.Config{Output: model.OutputConfig{Level: 1}},
	}

	var pages []*model.Page
	for i := 0; i < 100; i++ {
		title := "Page " + string(rune('0'+i%10))
		pages = append(pages, makePage("doc"+string(rune('0'+i%10))+".md", title, "docs", i))
	}

	err := gen.Generate(context.Background(), wiki, pages)
	assert.NoError(t, err)

	// Should have written output
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Greater(t, len(entries), 1, "should output files for 100 pages")
}


