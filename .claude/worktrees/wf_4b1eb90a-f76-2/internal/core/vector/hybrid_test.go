package vector

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/index"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupHybridTest creates a bleve index, vector store, and embedder
// with known test data for hybrid search testing.
func setupHybridTest(t *testing.T) (*index.Index, VectorStore, Embedder, func()) {
	t.Helper()

	// Create bleve index
	indexPath := filepath.Join(t.TempDir(), "test.bleve")
	idx, err := index.NewIndex(indexPath)
	require.NoError(t, err)

	// Index some test pages
	pages := []*model.Page{
		{ID: model.PageID("doc/go.md"), Path: "doc/go.md", Title: "Go Programming", Content: "Go is a programming language for building efficient software."},
		{ID: model.PageID("doc/python.md"), Path: "doc/python.md", Title: "Python Programming", Content: "Python is great for data science and machine learning."},
		{ID: model.PageID("doc/rust.md"), Path: "doc/rust.md", Title: "Rust Programming", Content: "Rust focuses on safety and performance."},
	}
	require.NoError(t, idx.Build(context.Background(), pages))

	// Create vector store and embedder, then embed pages
	store := NewMemoryStore("")
	embedder := NewLocalEmbedder(256)
	ctx := context.Background()

	for _, p := range pages {
		emb, err := embedder.Embed(ctx, p.Title+"\n"+p.Content)
		require.NoError(t, err)
		require.NoError(t, store.Store(ctx, p.ID, p.Path, p.Title, emb))
	}

	cleanup := func() {
		idx.Close()
		store.Close()
	}

	return idx, store, embedder, cleanup
}

func TestHybridAlphaOne(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 1.0) // pure BM25
	results, err := hybrid.Search(context.Background(), "programming", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results, "should find results")
}

func TestHybridAlphaZero(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 0.0) // pure vector
	results, err := hybrid.Search(context.Background(), "programming", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results, "should find results")
}

func TestHybridAlphaHalf(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 0.5) // balanced
	results, err := hybrid.Search(context.Background(), "programming", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results, "hybrid search should find results")
}

func TestHybridEmptyQuery(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 0.5)
	results, err := hybrid.Search(context.Background(), "", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results, "empty query should match all via vector search")
}

func TestHybridLimit(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 0.5)
	results, err := hybrid.Search(context.Background(), "programming", index.SearchOpts{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, results, 2, "should respect limit")
}

func TestHybridScoreNormalization(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 1.0) // pure BM25
	results, err := hybrid.Search(context.Background(), "programming", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// With alpha=1, BM25Score should be set and VecScore should be 0
	for _, r := range results {
		assert.GreaterOrEqual(t, r.BM25Score, 0.0)
	}
}

func TestHybridHasDetailedScores(t *testing.T) {
	idx, store, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	hybrid := NewHybridSearch(idx, store, embedder, 0.5)
	results, err := hybrid.Search(context.Background(), "programming", index.SearchOpts{Limit: 10})
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Check that the result has the detailed scores
	assert.GreaterOrEqual(t, results[0].BM25Score, 0.0)
	assert.GreaterOrEqual(t, results[0].VecScore, 0.0)
}

func TestNewHybridSearchDefaultAlpha(t *testing.T) {
	idx, _, embedder, cleanup := setupHybridTest(t)
	defer cleanup()

	// Alpha outside [0,1] should default to 0.5
	h := NewHybridSearch(idx, NewMemoryStore(""), embedder, -1)
	assert.Equal(t, 0.5, h.alpha)

	h2 := NewHybridSearch(idx, NewMemoryStore(""), embedder, 2.0)
	assert.Equal(t, 0.5, h2.alpha)
}
