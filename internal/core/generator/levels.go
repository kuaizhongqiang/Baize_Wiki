package generator

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

const mergeSplitSize = 50 * 1024 // 50KB

// PageRef is a lightweight reference to a page within a directory node.
type PageRef struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

// DirNode represents a directory in the Wiki output tree.
type DirNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Pages    []PageRef  `json:"pages"`
	Children []*DirNode `json:"children,omitempty"`
	Depth    int        `json:"depth"`
}

// LevelBuilder builds a DirNode tree from pages at a given level.
type LevelBuilder struct{}

// NewLevelBuilder creates a new LevelBuilder.
func NewLevelBuilder() *LevelBuilder {
	return &LevelBuilder{}
}

// Build constructs the directory tree for the given level.
func (lb *LevelBuilder) Build(pages []*model.Page, level int) *DirNode {
	switch level {
	case 1:
		return lb.Flat(pages)
	case 2:
		return lb.Structured(pages)
	case 3:
		return lb.Deep(pages)
	default:
		return lb.Structured(pages)
	}
}

// Flat builds a single-level directory with merged content.
func (lb *LevelBuilder) Flat(pages []*model.Page) *DirNode {
	root := &DirNode{Name: "wiki", Path: ".", Depth: 0}

	// Group by category
	type categoryPages struct {
		category string
		pages    []*model.Page
	}
	groups := make(map[string][]*model.Page)
	var categories []string

	for _, p := range pages {
		cat := p.Meta.Category
		if cat == "" {
			// Use first path component as category
			parts := strings.SplitN(strings.ReplaceAll(p.Path, "\\", "/"), "/", 2)
			if len(parts) > 1 {
				cat = parts[0]
			} else {
				cat = "uncategorized"
			}
		}
		if _, ok := groups[cat]; !ok {
			categories = append(categories, cat)
		}
		groups[cat] = append(groups[cat], p)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		group := groups[cat]

		// Sort by weight then filename
		sort.Slice(group, func(i, j int) bool {
			if group[i].Weight != group[j].Weight {
				return group[i].Weight < group[j].Weight
			}
			return group[i].Title < group[j].Title
		})

		// Merge content
		if len(group) == 0 {
			continue
		}

		// Handle single-page group
		if len(group) == 1 {
			p := group[0]
			fileName := cat + ".md"
			if !strings.HasSuffix(p.Path, ".md") {
				fileName = cat + ".md"
			}
			root.Pages = append(root.Pages, PageRef{
				Title: cat,
				Path:  fileName,
			})
			p.Path = fileName
			if p.Title == "" {
				p.Title = cat
			}
			continue
		}

		// Multi-page group: merge with 50KB split
		mergedGroups := mergePages(group, cat)
		for _, mg := range mergedGroups {
			root.Pages = append(root.Pages, PageRef{
				Title: mg.Title,
				Path:  mg.Path,
			})
		}
	}

	return root
}

// mergePages concatenates page content with `---` separators and splits at 50KB.
func mergePages(pages []*model.Page, category string) []*model.Page {
	var result []*model.Page
	var current *model.Page

	for _, p := range pages {
		if current == nil {
			current = &model.Page{
				Path:    category + ".md",
				Title:   category,
				Content: p.Content,
			}
			continue
		}

		// Check if adding this page would exceed 50KB
		separator := "\n\n---\n\n"
		projected := current.Content + separator + p.Content
		if len(projected) > mergeSplitSize {
			result = append(result, current)
			// Start new file with suffix
			suffix := len(result)
			current = &model.Page{
				Path:    fmt.Sprintf("%s-%d.md", category, suffix),
				Title:   category,
				Content: p.Content,
			}
		} else {
			current.Content = projected
		}
	}

	if current != nil {
		result = append(result, current)
	}

	return result
}

// Structured builds a two-level directory tree.
func (lb *LevelBuilder) Structured(pages []*model.Page) *DirNode {
	root := &DirNode{Name: "wiki", Path: ".", Depth: 0}

	// Group by first directory component
	dirs := make(map[string][]*model.Page)
	var dirNames []string

	for _, p := range pages {
		dirName := firstDir(p.Path)
		if dirName == "" {
			dirName = "_root"
		}
		if _, ok := dirs[dirName]; !ok {
			dirNames = append(dirNames, dirName)
		}
		dirs[dirName] = append(dirs[dirName], p)
	}
	sort.Strings(dirNames)

	for _, name := range dirNames {
		dir := &DirNode{
			Name:  name,
			Path:  name,
			Depth: 1,
		}

		for _, p := range dirs[name] {
			// For Level 2, strip the first directory component from path
			relativePath := stripFirstDir(p.Path)
			if relativePath == "" {
				relativePath = p.Title + ".md"
			}
			dir.Pages = append(dir.Pages, PageRef{
				Title: p.Title,
				Path:  name + "/" + relativePath,
			})
		}

		// Sort pages by title
		sort.Slice(dir.Pages, func(i, j int) bool {
			return dir.Pages[i].Title < dir.Pages[j].Title
		})

		root.Children = append(root.Children, dir)
	}

	return root
}

// Deep builds a full directory tree (max 3 levels).
func (lb *LevelBuilder) Deep(pages []*model.Page) *DirNode {
	root := &DirNode{Name: "wiki", Path: ".", Depth: 0}

	for _, p := range pages {
		parts := splitPath(p.Path)
		depth := len(parts)
		if depth > 3 {
			parts = parts[:3]
			depth = 3
		}

		current := root
		for i := 0; i < depth-1; i++ {
			found := false
			for _, child := range current.Children {
				if child.Name == parts[i] {
					current = child
					found = true
					break
				}
			}
			if !found {
				dirPath := path.Join(current.Path, parts[i])
				if dirPath == "." {
					dirPath = parts[i]
				}
				newDir := &DirNode{
					Name:  parts[i],
					Path:  dirPath,
					Depth: i + 1,
				}
				current.Children = append(current.Children, newDir)
				current = newDir
			}
		}

		// Last part is the filename
		fileName := parts[len(parts)-1]
		pagePath := path.Join(current.Path, fileName)
		if current.Path == "." {
			pagePath = fileName
		}

		current.Pages = append(current.Pages, PageRef{
			Title: p.Title,
			Path:  pagePath,
		})
	}

	return root
}

// firstDir returns the first directory component of a path.
func firstDir(filePath string) string {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	parts := strings.SplitN(normalized, "/", 2)
	if len(parts) <= 1 {
		return ""
	}
	return parts[0]
}

// stripFirstDir removes the first directory component from a path.
func stripFirstDir(filePath string) string {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	parts := strings.SplitN(normalized, "/", 2)
	if len(parts) <= 1 {
		return ""
	}
	return parts[1]
}

// splitPath splits a path into components.
func splitPath(filePath string) []string {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	return strings.Split(normalized, "/")
}
