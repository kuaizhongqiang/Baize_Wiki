package graph

import (
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraph(t *testing.T) {
	pages := []*model.Page{
		{
			Path:   "core/model/page.go",
			Title:  "Page Model",
			Abstract: "Core data structures",
			Keywords: []string{"Page", "Section"},
			Entities: []model.Entity{
				{Name: "Page", Type: "struct", Role: "defined"},
				{Name: "Section", Type: "struct", Role: "defined"},
			},
		},
		{
			Path:   "core/parser/parser.go",
			Title:  "Parser",
			Abstract: "Document parser",
			Entities: []model.Entity{
				{Name: "Parse", Type: "function", Role: "defined"},
				{Name: "Page", Type: "struct", Role: "uses"},
			},
		},
		{
			Path:   "core/scanner/scanner.go",
			Title:  "Scanner",
			Entities: []model.Entity{
				{Name: "Scan", Type: "function", Role: "defined"},
				{Name: "Page", Type: "struct", Role: "uses"},
				{Name: "Section", Type: "struct", Role: "uses"},
			},
		},
	}

	graph := BuildGraph(pages)

	// Should have 3 nodes
	assert.Len(t, graph.Nodes, 3)
	assert.Equal(t, "Page Model", graph.Nodes[0].Title)

	// Should have edges: parser uses Page, scanner uses Page, scanner uses Section
	assert.GreaterOrEqual(t, len(graph.Edges), 2)

	// Verify edges
	foundParser := false
	foundScanner := false
	for _, e := range graph.Edges {
		if e.Source == "core/model/page.go" && e.Target == "core/parser/parser.go" {
			foundParser = true
		}
		if e.Source == "core/model/page.go" && e.Target == "core/scanner/scanner.go" {
			foundScanner = true
		}
	}
	assert.True(t, foundParser, "parser should depend on page model")
	assert.True(t, foundScanner, "scanner should depend on page model")

	// Should have layers
	assert.NotEmpty(t, graph.Layers)
}

func TestBuildGraphNoEntities(t *testing.T) {
	pages := []*model.Page{
		{Path: "page1.md", Title: "Page 1"},
		{Path: "page2.md", Title: "Page 2"},
	}

	graph := BuildGraph(pages)
	assert.Len(t, graph.Nodes, 2)
	assert.Empty(t, graph.Edges) // no entities → no edges
	assert.NotEmpty(t, graph.Layers) // should have an "other" layer
}

func TestSaveLoadGraph(t *testing.T) {
	graph := &KnowledgeGraph{
		Version: 1,
		Nodes: []GraphNode{
			{Path: "test.md", Title: "Test", Abstract: "A test page"},
		},
		Edges: []GraphEdge{
			{Source: "a.md", Target: "b.md", Relation: "uses"},
		},
		Layers: []GraphLayer{
			{Name: "core", MemberPaths: []string{"test.md"}},
		},
	}

	dir := t.TempDir()
	err := SaveGraph(graph, dir)
	require.NoError(t, err)

	loaded, err := LoadGraph(dir)
	require.NoError(t, err)

	assert.Equal(t, graph.Version, loaded.Version)
	assert.Len(t, loaded.Nodes, 1)
	assert.Equal(t, "Test", loaded.Nodes[0].Title)
	assert.Len(t, loaded.Edges, 1)
	assert.Len(t, loaded.Layers, 1)
}

func TestGraphPath(t *testing.T) {
	path := GraphPath("/tmp/wiki")
	assert.Contains(t, path, ".baize")
	assert.Contains(t, path, "graph.json")
}

func TestDiscoverLayers(t *testing.T) {
	nodes := []GraphNode{
		{Path: "core/model/page.go", Title: "Page"},
		{Path: "core/parser/parser.go", Title: "Parser"},
		{Path: "internal/app/build.go", Title: "Build"},
		{Path: "docs/readme.md", Title: "Readme"},
	}

	layers := discoverLayers(nil, nodes)
	assert.NotEmpty(t, layers)

	// model and parser should be in their own layers
	foundModel := false
	foundApp := false
	for _, layer := range layers {
		for _, m := range layer.MemberPaths {
			if m == "core/model/page.go" && layer.Name == "model" {
				foundModel = true
			}
			if m == "internal/app/build.go" && layer.Name == "app" {
				foundApp = true
			}
		}
	}
	assert.True(t, foundModel, "model should be in model layer")
	assert.True(t, foundApp, "app should be in app layer")
}
