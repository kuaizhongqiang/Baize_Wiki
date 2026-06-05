package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"golang.org/x/sync/errgroup"
)

// maxConcurrentParses limits concurrent file parsing to avoid memory spikes.
const maxConcurrentParses = 10

// Parse reads a single file and returns a Page.
// It handles .md files with frontmatter and section parsing, and
// plain-text reading for all other file types.
func Parse(file model.FileInfo) (*model.Page, string) {
	content, err := os.ReadFile(file.AbsPath)
	if err != nil {
		return nil, "failed to read file: " + err.Error()
	}

	body := string(content)
	page := &model.Page{
		ID:         model.PageID(file.Path),
		Path:       file.Path,
		SourceFile: file.AbsPath,
	}

	ext := filepath.Ext(file.AbsPath)
	warning := ""

	if ext == ".md" || ext == ".mdx" {
		// Extract frontmatter
		fm, contentBody, warn := extractFrontmatter(body)
		page.Meta = fm
		page.Title = fm.Title
		page.Tags = fm.Tags
		if fm.Weight != nil {
			page.Weight = *fm.Weight
		}
		warning = warn

		// Parse markdown sections
		page.Sections = parseMarkdown(contentBody)
		page.Content = contentBody
	} else {
		// Non-markdown: use filename as title, full content as body
		page.Title = strings.TrimSuffix(filepath.Base(file.Path), file.Ext)
		page.Content = body
	}

	// Fallback title
	if page.Title == "" {
		page.Title = strings.TrimSuffix(filepath.Base(file.Path), file.Ext)
	}

	// Set Depth from path
	page.Depth = countPathDepth(file.Path)

	return page, warning
}

// ParseBatch concurrently parses a list of files.
// Individual file failures do not stop the overall process.
// Returns parsed pages and a list of warnings.
func ParseBatch(ctx context.Context, files []model.FileInfo) ([]*model.Page, []string) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentParses)

	pages := make([]*model.Page, len(files))
	warnings := make([]string, 0)

	for i, file := range files {
		i, file := i, file
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			page, warn := Parse(file)
			pages[i] = page
			if warn != "" {
				warnings = append(warnings, file.Path+": "+warn)
			}
			return nil
		})
	}

	_ = g.Wait() // ignore error; individual file failures are non-fatal

	// Filter out nil pages (failed parses)
	result := make([]*model.Page, 0, len(pages))
	for _, p := range pages {
		if p != nil {
			result = append(result, p)
		}
	}

	return result, warnings
}

// countPathDepth counts the number of path separators in a relative path.
func countPathDepth(path string) int {
	return strings.Count(path, string(filepath.Separator)) + strings.Count(path, "/")
}
