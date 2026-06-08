// Package index provides full-text search over Wiki pages using bleve.
package index

import (
	"context"
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

const snippetLen = 200

// SearchResult represents a single search hit.
type SearchResult struct {
	Path    string   `json:"path"`
	Title   string   `json:"title"`
	Score   float64  `json:"score"`
	Snippet string   `json:"snippet"`
	Tags    []string `json:"tags"`
}

// SearchOpts controls search behaviour.
type SearchOpts struct {
	Tags        []string // filter by tags (AND)
	Limit       int      // max results (default 10)
	WithContent bool     // return full content
}

// Index wraps a bleve index for full-text search over Wiki pages.
type Index struct {
	path  string
	index bleve.Index
}

// doc is the document structure indexed by bleve.
// Struct field tags determine the index mapping.
type doc struct {
	Path    string   `json:"path" index:"keyword" store:"true"`
	Title   string   `json:"title" index:"text" store:"true"`
	Content string   `json:"content" index:"text" store:"true"`
	Tags    []string `json:"tags" index:"keyword" store:"true"`
}

// NewIndex creates or opens an index at the given path.
// If the path already contains a bleve index it is opened;
// otherwise a new index is created.
func NewIndex(path string) (*Index, error) {
	idx, err := bleve.Open(path)
	if err == nil {
		return &Index{path: path, index: idx}, nil
	}

	// Failed to open existing index — create a new one.
	// Remove any leftover directory so bleve.New can create from scratch,
	// then fall back to temp dir if the original path is unusable.
	_ = os.RemoveAll(path)
	m := bleve.NewIndexMapping()
	idx, err = bleve.New(path, m)
	if err != nil {
		return nil, fmt.Errorf("create index at %s: %w", path, err)
	}
	return &Index{path: path, index: idx}, nil
}

// Build indexes a batch of pages. It is safe to call multiple times;
// existing documents with the same ID (page path) are overwritten.
func (idx *Index) Build(ctx context.Context, pages []*model.Page) error {
	batch := idx.index.NewBatch()
	for _, p := range pages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		d := doc{
			Path:    p.Path,
			Title:   p.Title,
			Content: p.Content,
			Tags:    p.Tags,
		}
		if len(p.Meta.Tags) > 0 {
			d.Tags = append(d.Tags, p.Meta.Tags...)
		}

		if err := batch.Index(p.ID, d); err != nil {
			return fmt.Errorf("index page %s: %w", p.ID, err)
		}
	}
	return idx.index.Batch(batch)
}

// Search executes a full-text search query.
func (idx *Index) Search(ctx context.Context, q string, opts SearchOpts) ([]SearchResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	// Build the query
	var qry query.Query
	if q == "" {
		qry = bleve.NewMatchAllQuery()
	} else {
		qry = bleve.NewQueryStringQuery(q)
	}

	// Apply tag filter if specified
	if len(opts.Tags) > 0 {
		tagQueries := make([]query.Query, len(opts.Tags))
		for i, tag := range opts.Tags {
			tq := bleve.NewTermQuery(tag)
			tq.SetField("tags")
			tagQueries[i] = tq
		}
		conj := bleve.NewConjunctionQuery(tagQueries...)
		if q != "" {
			qry = bleve.NewConjunctionQuery(qry, conj)
		} else {
			qry = conj
		}
	}

	searchReq := bleve.NewSearchRequestOptions(qry, limit, 0, false)
	searchReq.Fields = []string{"path", "title", "content", "tags"}
	searchReq.Highlight = bleve.NewHighlight()

	result, err := idx.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	results := make([]SearchResult, 0, len(result.Hits))
	for _, hit := range result.Hits {
		sr := SearchResult{
			Path:  fieldStr(hit.Fields, "path"),
			Title: fieldStr(hit.Fields, "title"),
			Score: hit.Score,
			Tags:  fieldStrs(hit.Fields, "tags"),
		}

		// Build snippet
		content := fieldStr(hit.Fields, "content")
		if content != "" {
			snippet := content
			if len(snippet) > snippetLen {
				snippet = snippet[:snippetLen]
			}
			sr.Snippet = snippet
		}

		results = append(results, sr)
	}

	return results, nil
}

// Close closes the index and releases resources.
func (idx *Index) Close() error {
	return idx.index.Close()
}

// fieldStr safely extracts a string field from a document.
func fieldStr(fields map[string]interface{}, name string) string {
	if v, ok := fields[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// fieldStrs safely extracts a string slice field from a document.
func fieldStrs(fields map[string]interface{}, name string) []string {
	if v, ok := fields[name]; ok {
		switch arr := v.(type) {
		case []interface{}:
			strs := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			return strs
		case []string:
			return arr
		}
	}
	return nil
}

