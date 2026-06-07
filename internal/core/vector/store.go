// Package vector provides vector storage and embedding for semantic search.
package vector

import (
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// VectorResult represents a single vector search hit.
type VectorResult struct {
	PageID string  `json:"page_id"`
	Path   string  `json:"path"`
	Title  string  `json:"title"`
	Score  float64 `json:"score"` // cosine similarity
}

// VectorStore stores and retrieves vector embeddings for pages.
type VectorStore interface {
	// Store saves a vector embedding for a page with its metadata.
	Store(ctx context.Context, pageID string, path string, title string, embedding []float32) error

	// Search returns the most similar pages to the query vector.
	Search(ctx context.Context, embedding []float32, limit int) ([]VectorResult, error)

	// Close releases any resources held by the store.
	Close() error
}

// MemoryStore is an in-memory VectorStore with optional file persistence.
type MemoryStore struct {
	mu       sync.RWMutex
	vectors  map[string][]float32 // pageID → embedding
	metadata map[string]struct {  // pageID → meta
		Path  string
		Title string
	}
	dir string // persistence directory (empty = no persistence)
}

// storedVector is the serializable form for persistence.
type storedVector struct {
	PageID    string
	Embedding []float32
	Path      string
	Title     string
}

// NewMemoryStore creates an in-memory vector store.
// If dir is non-empty, it will attempt to load existing data from that directory
// and persist on Close.
func NewMemoryStore(dir string) *MemoryStore {
	s := &MemoryStore{
		vectors:  make(map[string][]float32),
		metadata: make(map[string]struct{ Path string; Title string }),
		dir:      dir,
	}
	if dir != "" {
		s.load()
	}
	return s
}

// Store saves a vector embedding for a page.
func (s *MemoryStore) Store(ctx context.Context, pageID, path, title string, embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("empty embedding for page %s", pageID)
	}

	vec := make([]float32, len(embedding))
	copy(vec, embedding)

	s.mu.Lock()
	s.vectors[pageID] = vec
	s.metadata[pageID] = struct{ Path string; Title string }{Path: path, Title: title}
	_ = ctx // context is for future extensibility (e.g. cancellation in API embedders)
	s.mu.Unlock()
	return nil
}

// Search returns the most similar pages by cosine similarity.
func (s *MemoryStore) Search(ctx context.Context, embedding []float32, limit int) ([]VectorResult, error) {
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.vectors) == 0 {
		return []VectorResult{}, nil
	}

	type scored struct {
		pageID string
		score  float64
	}

	results := make([]scored, 0, len(s.vectors))
	for pageID, vec := range s.vectors {
		sim := cosineSimilarity(embedding, vec)
		results = append(results, scored{pageID: pageID, score: sim})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if limit > len(results) {
		limit = len(results)
	}

	out := make([]VectorResult, limit)
	for i, r := range results[:limit] {
		meta := s.metadata[r.pageID]
		out[i] = VectorResult{
			PageID: r.pageID,
			Path:   meta.Path,
			Title:  meta.Title,
			Score:  r.score,
		}
	}
	_ = ctx
	return out, nil
}

// Close persists data to disk if a directory was configured.
func (s *MemoryStore) Close() error {
	if s.dir == "" {
		return nil
	}
	return s.save()
}

// cosineSimilarity computes the cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	n := min(len(a), len(b))
	var dot, normA, normB float64
	for i := 0; i < n; i++ {
		fa := float64(a[i])
		fb := float64(b[i])
		dot += fa * fb
		normA += fa * fa
		normB += fb * fb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

const vectorsFile = "vectors.gob"

func (s *MemoryStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("create vectors dir: %w", err)
	}

	path := filepath.Join(s.dir, vectorsFile)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create vectors file: %w", err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	stored := make([]storedVector, 0, len(s.vectors))
	for pageID, emb := range s.vectors {
		meta := s.metadata[pageID]
		stored = append(stored, storedVector{
			PageID:    pageID,
			Embedding: emb,
			Path:      meta.Path,
			Title:     meta.Title,
		})
	}
	return enc.Encode(stored)
}

func (s *MemoryStore) load() error {
	path := filepath.Join(s.dir, vectorsFile)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // first run, no data yet
		}
		return err
	}
	defer f.Close()

	var stored []storedVector
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&stored); err != nil {
		return fmt.Errorf("decode vectors: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sv := range stored {
		s.vectors[sv.PageID] = sv.Embedding
		s.metadata[sv.PageID] = struct{ Path string; Title string }{
			Path: sv.Path, Title: sv.Title,
		}
	}
	return nil
}

// Ensure MemoryStore implements VectorStore at compile time.
var _ VectorStore = (*MemoryStore)(nil)
