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
// Uses goldmark AST to extract heading-based structure.
func parseMarkdown(content string) []model.Section {
	reader := text.NewReader([]byte(content))
	doc := goldmark.New().Parser().Parse(reader)

	var sections []model.Section
	// Stack stores indices into sections slice for building the tree.
	// When sections grows, its backing array may be reallocated, so we
	// use indices rather than pointers to avoid stale references.
	var stack []int

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

		sec := model.Section{
			ID:    generateSectionID(title),
			Level: heading.Level,
			Title: title,
		}

		// Pop stack until we find a parent with a lower heading level.
		for len(stack) > 0 && sections[stack[len(stack)-1]].Level >= heading.Level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) > 0 {
			// Add as child of the nearest ancestor.
			parentIdx := stack[len(stack)-1]
			sections[parentIdx].Children = append(sections[parentIdx].Children, sec)
		} else {
			// Add as a root-level section.
			sections = append(sections, sec)
		}

		// Push this section's index onto the stack.
		stack = append(stack, len(sections)-1)

		return ast.WalkContinue, nil
	})

	return sections
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
