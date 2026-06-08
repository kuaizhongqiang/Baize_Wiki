// Package graph provides the Level 3 knowledge graph engine.
// It discovers relationships between pages by analyzing entities,
// assigns architectural layers, and persists the graph as graph.json.
package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

// GraphNode represents a page in the knowledge graph.
type GraphNode struct {
	Path      string   `json:"path"`
	Title     string   `json:"title"`
	Layer     string   `json:"layer,omitempty"`
	Abstract  string   `json:"abstract,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
}

// GraphEdge represents a relationship between two pages.
type GraphEdge struct {
	Source      string `json:"source"`      // source page path
	Target      string `json:"target"`      // target page path
	Relation    string `json:"relation"`    // depends_on | uses | implements | extends | related_to
	Description string `json:"description,omitempty"`
}

// GraphLayer represents an architectural layer grouping.
type GraphLayer struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	MemberPaths []string `json:"member_paths"`
}

// KnowledgeGraph is the complete graph structure for a Wiki.
type KnowledgeGraph struct {
	Version int          `json:"version"`
	Nodes   []GraphNode  `json:"nodes"`
	Edges   []GraphEdge  `json:"edges"`
	Layers  []GraphLayer `json:"layers,omitempty"`
}

// BuildGraph constructs the knowledge graph from parsed pages.
func BuildGraph(pages []*model.Page) *KnowledgeGraph {
	g := &KnowledgeGraph{
		Version: 1,
		Nodes:   make([]GraphNode, 0, len(pages)),
		Edges:   nil, // computed below
	}

	// Build nodes
	nodeMap := make(map[string]int) // path -> index
	for i, p := range pages {
		node := GraphNode{
			Path:     p.Path,
			Title:    p.Title,
			Abstract: p.Abstract,
			Keywords: p.Keywords,
		}
		g.Nodes = append(g.Nodes, node)
		nodeMap[p.Path] = i
	}

	// Build edges from entity cross-references
	// If page A defines entity X and page B uses/imports entity X → edge A→B
	definedBy := make(map[string]string) // entity name -> page path (that defines it)
	for _, p := range pages {
		for _, e := range p.Entities {
			if e.Role == "defined" {
				definedBy[e.Name] = p.Path
			}
		}
	}

	edgeSet := make(map[string]bool) // dedup "source:target"
	for _, p := range pages {
		for _, e := range p.Entities {
			if e.Role == "defined" {
				continue // skip self
			}
			defPath, found := definedBy[e.Name]
			if !found || defPath == p.Path {
				continue
			}
			key := defPath + ":" + p.Path
			if edgeSet[key] {
				continue
			}
			edgeSet[key] = true
			g.Edges = append(g.Edges, GraphEdge{
				Source:      defPath,
				Target:      p.Path,
				Relation:    e.Role,
				Description: fmt.Sprintf("references %s", e.Name),
			})
		}
	}

	// Discover layers from path patterns and common entities
	g.Layers = discoverLayers(pages, g.Nodes)

	return g
}

// discoverLayers groups pages into architectural layers based on path patterns and entity analysis.
func discoverLayers(pages []*model.Page, nodes []GraphNode) []GraphLayer {
	layerPatterns := []struct {
		Name    string
		Pattern []string // path segments to match
	}{
		{Name: "model", Pattern: []string{"model"}},
		{Name: "data", Pattern: []string{"data"}},
		{Name: "storage", Pattern: []string{"storage"}},
		{Name: "parser", Pattern: []string{"parser"}},
		{Name: "scanner", Pattern: []string{"scanner"}},
		{Name: "generator", Pattern: []string{"generator"}},
		{Name: "index", Pattern: []string{"index"}},
		{Name: "vector", Pattern: []string{"vector"}},
		{Name: "linker", Pattern: []string{"linker"}},
		{Name: "catalog", Pattern: []string{"catalog"}},
		{Name: "app", Pattern: []string{"app"}},
		{Name: "mcp", Pattern: []string{"mcp"}},
		{Name: "config", Pattern: []string{"config"}},
	}

	var layers []GraphLayer
	used := make(map[string]bool)

	for _, lp := range layerPatterns {
		var members []string
		for _, node := range nodes {
			pathLower := strings.ToLower(node.Path)
			for _, seg := range lp.Pattern {
				if strings.Contains(pathLower, seg) && !used[node.Path] {
					members = append(members, node.Path)
					used[node.Path] = true
					break
				}
			}
		}
		if len(members) > 0 {
			sort.Strings(members)
			layers = append(layers, GraphLayer{
				Name:        lp.Name,
				MemberPaths: members,
			})
		}
	}

	// Collect unassigned pages into a "other" layer
	var unassigned []string
	for _, node := range nodes {
		if !used[node.Path] {
			unassigned = append(unassigned, node.Path)
		}
	}
	if len(unassigned) > 0 {
		sort.Strings(unassigned)
		layers = append(layers, GraphLayer{
			Name:        "other",
			MemberPaths: unassigned,
		})
	}

	return layers
}

// GraphPath returns the path to the graph.json file within a Wiki directory.
func GraphPath(wikiDir string) string {
	return filepath.Join(wikiDir, ".baize", "graph.json")
}

// SaveGraph persists the knowledge graph to graph.json.
func SaveGraph(graph *KnowledgeGraph, wikiDir string) error {
	path := GraphPath(wikiDir)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir graph dir: %w", err)
	}

	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write graph: %w", err)
	}
	return nil
}

// LoadGraph reads and returns the knowledge graph from graph.json.
func LoadGraph(wikiDir string) (*KnowledgeGraph, error) {
	data, err := os.ReadFile(GraphPath(wikiDir))
	if err != nil {
		return nil, fmt.Errorf("read graph: %w", err)
	}

	var graph KnowledgeGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("unmarshal graph: %w", err)
	}
	return &graph, nil
}
