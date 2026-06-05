package generator

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
)

// Generator orchestrates Wiki generation from parsed pages.
type Generator struct {
	storage *storage.Store
}

// NewGenerator creates a new Generator with the given storage.
func NewGenerator(s *storage.Store) *Generator {
	return &Generator{storage: s}
}

// Generate creates the Wiki output directory tree and writes all files.
// It builds the DirNode tree, generates _index.md files, writes page files,
// and persists metadata.
func (g *Generator) Generate(ctx context.Context, wiki *model.Wiki, pages []*model.Page) error {
	builder := NewLevelBuilder()
	root := builder.Build(pages, wiki.Config.Output.Level)

	outputDir := wiki.OutputPath

	// Write page files
	for _, p := range pages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := g.storage.WritePage(outputDir, p.Path, p.Content); err != nil {
			return fmt.Errorf("write page %s: %w", p.Path, err)
		}
	}

	// Generate _index.md files
	if err := g.generateIndexFiles(ctx, root, outputDir); err != nil {
		return err
	}

	// Write meta.json
	wiki.PageCount = len(pages)
	if err := g.storage.WriteMeta(outputDir, wiki); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}

	return nil
}

// generateIndexFiles recursively generates _index.md files for all directories.
func (g *Generator) generateIndexFiles(ctx context.Context, node *DirNode, outputDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Generate _index.md for this directory
	content := renderIndex(node)
	indexPath := filepath.Join(node.Path, "_index.md")
	if node.Path == "." {
		indexPath = "_index.md"
	}
	if err := g.storage.WritePage(outputDir, indexPath, content); err != nil {
		return fmt.Errorf("write index %s: %w", indexPath, err)
	}

	// Recurse into children
	for _, child := range node.Children {
		if err := g.generateIndexFiles(ctx, child, outputDir); err != nil {
			return err
		}
	}

	return nil
}

// renderIndex generates the _index.md content for a directory node.
func renderIndex(node *DirNode) string {
	var b strings.Builder

	// Title
	if node.Name == "wiki" {
		b.WriteString("# Wiki Overview\n\n")
	} else {
		fmt.Fprintf(&b, "# %s\n\n", node.Name)
	}

	// Page count
	totalPages := len(node.Pages)
	for _, child := range node.Children {
		totalPages += len(child.Pages)
	}
	fmt.Fprintf(&b, "> 共 %d 个页面\n\n", totalPages)

	// Pages list
	if len(node.Pages) > 0 {
		b.WriteString("## 页面\n\n")
		sort.Slice(node.Pages, func(i, j int) bool {
			return node.Pages[i].Title < node.Pages[j].Title
		})
		for _, p := range node.Pages {
			fmt.Fprintf(&b, "- [%s](%s)\n", p.Title, p.Path)
		}
		b.WriteString("\n")
	}

	// Children (subdirectories)
	if len(node.Children) > 0 {
		b.WriteString("## 子目录\n\n")
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
		for _, child := range node.Children {
			fmt.Fprintf(&b, "- [%s](%s/)\n", child.Name, child.Path)
		}
		b.WriteString("\n")
	}

	return b.String()
}
