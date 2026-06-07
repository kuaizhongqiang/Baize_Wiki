// Package linker computes cross-references between Wiki pages.
package linker

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/parser"
)

// Linker resolves [[wiki-link]] references between pages and computes backlinks.
type Linker struct{}

// New creates a new Linker.
func New() *Linker {
	return &Linker{}
}

// Link resolves all [[wiki-link]] references across pages, populating
// each page's Links and Backlinks fields.
func (l *Linker) Link(ctx context.Context, pages []*model.Page) error {
	if len(pages) == 0 {
		return nil
	}

	// Build path → page and title → page indexes
	pathIndex := make(map[string]*model.Page, len(pages))
	titleIndex := make(map[string]*model.Page, len(pages))
	fileNameIndex := make(map[string]*model.Page, len(pages))

	for _, p := range pages {
		pathIndex[p.Path] = p
		titleIndex[p.Title] = p
		base := filepath.Base(p.Path)
		// Strip .md/.mdx extension for filename matching
		base = strings.TrimSuffix(base, filepath.Ext(base))
		fileNameIndex[base] = p
	}

	// Resolve each page's links
	resolvedLinks := make(map[string][]model.Link) // pageID → resolved links

	for _, src := range pages {
		// If Links is already populated (from parser), resolve them;
		// otherwise extract from content.
		var refs []parser.LinkRef
		if len(src.Links) == 0 && src.Content != "" {
			refs = parser.ExtractWikiLinks(src.Content)
		} else {
			// Convert existing links back to refs for re-resolution
			for _, link := range src.Links {
				refs = append(refs, parser.LinkRef{Target: link.TargetPath, Text: link.Text})
			}
		}

		if len(refs) == 0 {
			continue
		}

		links := make([]model.Link, 0, len(refs))
		for _, ref := range refs {
			link := resolveLink(src.ID, ref, pathIndex, titleIndex, fileNameIndex)
			links = append(links, link)
		}
		resolvedLinks[src.ID] = links
	}

	// Write back resolved Links and build Backlinks index
	backlinksIndex := make(map[string][]model.Link) // pageID → backlinks

	for _, src := range pages {
		links := resolvedLinks[src.ID]
		src.Links = links

		for _, link := range links {
			if link.TargetID != "" {
				backlinksIndex[link.TargetID] = append(backlinksIndex[link.TargetID], model.Link{
					SourceID:   src.ID,
					TargetID:   link.TargetID,
					TargetPath: link.TargetPath,
					Text:       link.Text,
					Type:       link.Type,
				})
			}
		}
	}

	// Write back Backlinks
	for _, p := range pages {
		p.Backlinks = backlinksIndex[p.ID]
		if p.Backlinks == nil {
			p.Backlinks = []model.Link{}
		}
	}

	return nil
}

// resolveLink matches a LinkRef to a target page or classifies it as external/resource.
func resolveLink(sourceID string, ref parser.LinkRef, pathIndex, titleIndex, fileNameIndex map[string]*model.Page) model.Link {
	target := ref.Target

	// External link
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return model.Link{
			SourceID:   sourceID,
			TargetPath: target,
			Text:       ref.Text,
			Type:       model.LinkExternal,
		}
	}

	// Anchor link (same page)
	if strings.HasPrefix(target, "#") {
		return model.Link{
			SourceID:   sourceID,
			TargetPath: target,
			Text:       ref.Text,
			Type:       model.LinkInternal,
		}
	}

	// Resource link (has file extension, not .md)
	if strings.Contains(target, ".") && !strings.HasSuffix(target, ".md") && !strings.HasSuffix(target, ".mdx") {
		return model.Link{
			SourceID:   sourceID,
			TargetPath: target,
			Text:       ref.Text,
			Type:       model.LinkResource,
		}
	}

	// Internal wiki link — try to match
	targetPage := matchPage(target, pathIndex, titleIndex, fileNameIndex)
	if targetPage != nil {
		text := ref.Text
		if text == "" {
			text = targetPage.Title
		}
		return model.Link{
			SourceID:   sourceID,
			TargetID:   targetPage.ID,
			TargetPath: targetPage.Path,
			Text:       text,
			Type:       model.LinkInternal,
		}
	}

	// Dangling link (no match)
	return model.Link{
		SourceID:   sourceID,
		TargetPath: target,
		Text:       ref.Text,
		Type:       model.LinkInternal,
	}
}

// matchPage tries to find a page matching the link target using
// the configured resolution strategy.
func matchPage(target string, pathIndex, titleIndex, fileNameIndex map[string]*model.Page) *model.Page {
	// 1. Exact path match
	if p, ok := pathIndex[target]; ok {
		return p
	}

	// 2. Title match
	if p, ok := titleIndex[target]; ok {
		return p
	}

	// 3. File name match (basename without extension)
	if p, ok := fileNameIndex[target]; ok {
		return p
	}

	return nil
}
