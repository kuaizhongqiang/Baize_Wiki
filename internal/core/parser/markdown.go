package parser

import (
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const sectionContentMaxLen = 200

// parseMarkdown parses markdown content into a tree of sections.
// Uses goldmark AST to extract heading-based structure,
// then builds a tree via recursive partitioning.
func parseMarkdown(content string) []model.Section {
	reader := text.NewReader([]byte(content))
	doc := goldmark.New().Parser().Parse(reader)

	// Collect all headings into a flat list.
	var headings []model.Section
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		title := extractHeadingText(heading, reader)
		if title == "" {
			return ast.WalkContinue, nil
		}
		headings = append(headings, model.Section{
			ID:    generateSectionID(title),
			Level: heading.Level,
			Title: title,
		})
		return ast.WalkContinue, nil
	})

	if len(headings) == 0 {
		return nil
	}

	// Build tree via recursive partitioning.
	// Each call processes a slice of headings at a given level:
	//   - headings at `level` become root nodes
	//   - headings between two `level` nodes are children of the first
	return buildTree(headings, 1)
}

// buildTree recursively partitions headings into a section tree.
// headings at `level` become immediate children of the result;
// entries between them (with level > minLevel) are grouped as children.
func buildTree(headings []model.Section, level int) []model.Section {
	var result []model.Section
	i := 0
	for i < len(headings) {
		if headings[i].Level != level {
			i++
			continue
		}

		sec := headings[i]
		i++

		// Collect child headings (those between this heading and the
		// next heading at the same or shallower level).
		childStart := i
		for i < len(headings) && headings[i].Level > level {
			i++
		}
		if childStart < i {
			sec.Children = buildTree(headings[childStart:i], level+1)
		}

		result = append(result, sec)
	}
	return result
}

// extractHeadingText retrieves the text content of a heading node.
// Heading children are inline nodes (Text, Emphasis, etc.) — use Segment directly.
func extractHeadingText(n *ast.Heading, reader text.Reader) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			segment := t.Segment
			b.Write(reader.Value(segment))
		}
	}
	return strings.TrimSpace(b.String())
}

// generateSectionID creates a simple anchor ID from a heading title.
func generateSectionID(title string) string {
	id := strings.ToLower(title)
	id = strings.NewReplacer(
		" ", "-",
		".", "-",
		"_", "-",
		"/", "-",
		"\\", "-",
	).Replace(id)
	return id
}

// collectSectionContent gathers text content under a heading.
// Uses a simple line-based approach to extract section content summary.
func collectSectionContent(content string, startLine int) string {
	lines := strings.Split(content, "\n")
	var b strings.Builder
	for i := startLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			break
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(line)
		if b.Len() >= sectionContentMaxLen {
			break
		}
	}

	result := b.String()
	if len(result) > sectionContentMaxLen {
		result = result[:sectionContentMaxLen]
	}
	return result
}
