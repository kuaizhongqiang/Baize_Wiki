package parser

import (
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const sectionContentMaxLen = 200

// parseMarkdown parses markdown content into sections.
// It uses goldmark to parse the AST and extracts heading-based sections.
func parseMarkdown(content string) []model.Section {
	reader := text.NewReader([]byte(content))
	doc := goldmark.New().Parser().Parse(reader)

	var sections []model.Section
	var stack []*model.Section
	lastLevel := 0

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		level := heading.Lines().Len()
		_ = level

		title := extractHeadingText(heading, reader)
		if title == "" {
			return ast.WalkContinue, nil
		}

		sec := model.Section{
			ID:    generateSectionID(title),
			Level: heading.Level,
			Title: title,
		}

		// Build the section tree
		if heading.Level > lastLevel {
			// Child section: add to parent
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, sec)
			} else {
				sections = append(sections, sec)
			}
			stack = append(stack, &sections[len(sections)-1])
		} else if heading.Level == lastLevel {
			// Sibling: replace top of stack
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			sections = append(sections, sec)
			stack = append(stack, &sections[len(sections)-1])
		} else {
			// Parent level: pop stack and add
			for len(stack) > 0 && stack[len(stack)-1].Level >= heading.Level {
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, sec)
			} else {
				sections = append(sections, sec)
			}
			stack = append(stack, &sections[len(sections)-1])
		}

		lastLevel = heading.Level

		return ast.WalkContinue, nil
	})

	return sections
}

// extractHeadingText retrieves the text content of a heading node.
func extractHeadingText(n *ast.Heading, reader text.Reader) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		lines := c.Lines()
		if lines.Len() == 0 {
			if t, ok := c.(*ast.Text); ok {
				segment := t.Segment
				b.Write(reader.Value(segment))
			}
			continue
		}
		for i := 0; i < lines.Len(); i++ {
			segment := lines.At(i)
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
