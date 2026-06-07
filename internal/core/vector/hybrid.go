package vector

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/index"
)

// HybridSearch combines BM25 full-text search with vector semantic search.
// The two result sets are merged with normalized scores.
//
// Score normalization is critical because BM25 scores (0~∞) and cosine
// similarity (0~1) are on different scales. Without normalization, BM25
// would dominate the combined score.
//
//	final_score = α * normalized_bm25 + (1-α) * normalized_vec
//	normalized_bm25 = bm25_score / max_bm25_in_results  →  [0, 1]
//	normalized_vec  = cosine_similarity                  →  [0, 1]
type HybridSearch struct {
	bm25    *index.Index
	vectors VectorStore
	embed   Embedder
	alpha   float64 // BM25 weight α ∈ [0, 1]; 1=full BM25, 0=full vector
}

// NewHybridSearch creates a new hybrid search engine.
// alpha is the BM25 weight (default 0.5 if outside [0,1]).
func NewHybridSearch(bm25 *index.Index, vectors VectorStore, embed Embedder, alpha float64) *HybridSearch {
	if alpha < 0 || alpha > 1 {
		alpha = 0.5
	}
	return &HybridSearch{
		bm25:    bm25,
		vectors: vectors,
		embed:   embed,
		alpha:   alpha,
	}
}

// HybridResult extends SearchResult with additional scoring info.
type HybridResult struct {
	index.SearchResult
	BM25Score  float64 `json:"bm25_score,omitempty"`
	VecScore   float64 `json:"vec_score,omitempty"`
}

// Search performs a hybrid BM25 + vector search and returns merged,
// normalized, and re-ranked results.
func (h *HybridSearch) Search(ctx context.Context, query string, opts index.SearchOpts) ([]HybridResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	// 1. BM25 search
	bm25Results, err := h.bm25.Search(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("bm25 search: %w", err)
	}

	// 2. Vector search: embed query → search vector store
	queryVec, err := h.embed.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	vectorResults, err := h.vectors.Search(ctx, queryVec, limit*2) // fetch more for merging
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}

	if len(bm25Results) == 0 && len(vectorResults) == 0 {
		return []HybridResult{}, nil
	}

	// 3. Build a merged result map keyed by path
	type entry struct {
		bm25Score float64
		vecScore  float64
		title     string
		path      string
	}
	merged := make(map[string]*entry)

	// BM25 results
	for _, r := range bm25Results {
		merged[r.Path] = &entry{
			bm25Score: r.Score,
			title:     r.Title,
			path:      r.Path,
		}
	}

	// Vector results (merge into existing entries or create new ones)
	for _, r := range vectorResults {
		if e, ok := merged[r.Path]; ok {
			e.vecScore = r.Score
		} else {
			merged[r.Path] = &entry{
				vecScore: r.Score,
				title:    r.Title,
				path:     r.Path,
			}
		}
	}

	// 4. Normalize BM25 scores
	maxBM25 := 0.0
	for _, e := range merged {
		if e.bm25Score > maxBM25 {
			maxBM25 = e.bm25Score
		}
	}

	// 5. Compute final combined scores
	type scoredEntry struct {
		path       string
		title      string
		score      float64 // combined final score
		bm25Score  float64 // original BM25
		vecScore   float64 // original vector cosine
	}
	results := make([]scoredEntry, 0, len(merged))

	for _, e := range merged {
		normBM25 := 0.0
		if maxBM25 > 0 {
			normBM25 = e.bm25Score / maxBM25
		}
		// Vector cosine similarity is already in [0, 1]
		normVec := e.vecScore
		if math.IsNaN(normVec) {
			normVec = 0
		}

		combined := h.alpha*normBM25 + (1-h.alpha)*normVec

		results = append(results, scoredEntry{
			path:      e.path,
			title:     e.title,
			score:     combined,
			bm25Score: e.bm25Score,
			vecScore:  e.vecScore,
		})
	}

	// 6. Sort by combined score descending, then take top N
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if limit > len(results) {
		limit = len(results)
	}

	out := make([]HybridResult, limit)
	for i, r := range results[:limit] {
		out[i] = HybridResult{
			SearchResult: index.SearchResult{
				Path:  r.path,
				Title: r.title,
				Score: r.score,
			},
			BM25Score: r.bm25Score,
			VecScore:  r.vecScore,
		}
	}
	return out, nil
}
