package vector

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreAndSearch(t *testing.T) {
	s := NewMemoryStore("")
	ctx := context.Background()

	emb1 := []float32{1, 0, 0}
	emb2 := []float32{0, 1, 0}

	require.NoError(t, s.Store(ctx, "page1", "doc/a.md", "Doc A", emb1))
	require.NoError(t, s.Store(ctx, "page2", "doc/b.md", "Doc B", emb2))

	// Search for something similar to emb1
	results, err := s.Search(ctx, []float32{0.9, 0.1, 0}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "page1", results[0].PageID, "page1 should be most similar")
	assert.Greater(t, results[0].Score, results[1].Score)
}

func TestSearchOrderByCosine(t *testing.T) {
	s := NewMemoryStore("")
	ctx := context.Background()

	embSimilar := []float32{1, 0, 0}
	embSomewhat := []float32{0.7, 0.7, 0}
	embUnrelated := []float32{0, 1, 0}

	require.NoError(t, s.Store(ctx, "sim", "doc/sim.md", "Similar", embSimilar))
	require.NoError(t, s.Store(ctx, "some", "doc/some.md", "Somewhat", embSomewhat))
	require.NoError(t, s.Store(ctx, "unr", "doc/unr.md", "Unrelated", embUnrelated))

	// Query vector close to embSimilar
	results, err := s.Search(ctx, []float32{0.95, 0.05, 0}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Verify order is correct
	assert.Equal(t, "sim", results[0].PageID)
	assert.Equal(t, "some", results[1].PageID)
	assert.Equal(t, "unr", results[2].PageID)
}

func TestSearchEmpty(t *testing.T) {
	s := NewMemoryStore("")
	ctx := context.Background()

	results, err := s.Search(ctx, []float32{1, 0, 0}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchLimit(t *testing.T) {
	s := NewMemoryStore("")
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		emb := []float32{float32(i) / 5, 1 - float32(i)/5, 0}
		pageID := fmt.Sprintf("p%d", i)
		require.NoError(t, s.Store(ctx, pageID, "doc/p.md", "P", emb))
	}

	results, err := s.Search(ctx, []float32{1, 0, 0}, 3)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	vectorsDir := filepath.Join(dir, "vectors")

	// Create and populate
	s1 := NewMemoryStore(vectorsDir)
	ctx := context.Background()
	require.NoError(t, s1.Store(ctx, "p1", "doc/a.md", "Doc A", []float32{1, 0, 0}))
	require.NoError(t, s1.Store(ctx, "p2", "doc/b.md", "Doc B", []float32{0, 1, 0}))
	require.NoError(t, s1.Close())

	// Reopen and verify data persists
	s2 := NewMemoryStore(vectorsDir)
	defer s2.Close()

	results, err := s2.Search(ctx, []float32{1, 0, 0}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "p1", results[0].PageID)
}

func TestPersistenceEmptyStore(t *testing.T) {
	dir := t.TempDir()
	vectorsDir := filepath.Join(dir, "vectors")

	s1 := NewMemoryStore(vectorsDir)
	require.NoError(t, s1.Close())

	// Reopen — should not fail
	s2 := NewMemoryStore(vectorsDir)
	defer s2.Close()

	results, err := s2.Search(context.Background(), []float32{1, 0, 0}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    []float32
		b    []float32
		want float64
	}{
		{"identical", []float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"orthogonal", []float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{"opposite", []float32{1, 0, 0}, []float32{-1, 0, 0}, -1.0},
		{"partial", []float32{1, 1, 0}, []float32{1, 0, 0}, 0.707106},
		{"zero vector", []float32{0, 0, 0}, []float32{1, 0, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestStoreEmptyEmbedding(t *testing.T) {
	s := NewMemoryStore("")
	err := s.Store(context.Background(), "p1", "doc/a.md", "Doc A", []float32{})
	assert.Error(t, err)
}
