package parser

import (
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"gopkg.in/yaml.v3"
)

const frontmatterDelim = "---"

// extractFrontmatter parses YAML frontmatter from markdown content.
// Returns the parsed Frontmatter and the content body (without frontmatter).
// If no frontmatter is found, returns an empty Frontmatter and the original content.
// On parse errors, it logs a warning and returns the content as-is.
func extractFrontmatter(content string) (model.Frontmatter, string, string) {
	content = strings.TrimLeft(content, "\n\r\t ")
	fm, body, warning := splitFrontmatter(content)
	if fm == "" {
		return model.Frontmatter{}, content, ""
	}

	var frontmatter model.Frontmatter
	if err := yaml.Unmarshal([]byte(fm), &frontmatter); err != nil {
		return model.Frontmatter{}, content, "invalid frontmatter: " + err.Error()
	}

	return frontmatter, body, warning
}

// splitFrontmatter splits content into frontmatter YAML and body.
// Returns the frontmatter block (without delimiters), the body, and an
// optional warning string.
func splitFrontmatter(content string) (fm, body, warning string) {
	if !strings.HasPrefix(content, frontmatterDelim) {
		return "", content, ""
	}

	// Find closing delimiter
	rest := content[len(frontmatterDelim):]
	endIdx := strings.Index(rest, "\n"+frontmatterDelim)
	if endIdx == -1 {
		return "", content, ""
	}

	// Check if there's more content after the closing ---
	bodyStart := endIdx + 1 + len(frontmatterDelim)
	if bodyStart < len(rest) && rest[bodyStart] == '\n' {
		bodyStart++
	}

	fm = strings.TrimSpace(rest[:endIdx])
	body = rest[bodyStart:]
	return fm, body, ""
}
