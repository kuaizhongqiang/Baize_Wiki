package parser

import (
	"regexp"
	"strings"
)

// LinkRef represents a parsed [[wiki-link]] reference.
type LinkRef struct {
	Target string // the link target (e.g. "数据模型", "目录/页面名")
	Text   string // optional display text after |
}

var wikiLinkRe = regexp.MustCompile(`\[\[([^\[\]]+?)(?:\|([^\[\]]+?))?\]\]`)

// ExtractWikiLinks extracts [[wiki-link]] references from Markdown content.
// It skips links inside code blocks (triple-backtick fenced blocks).
func ExtractWikiLinks(content string) []LinkRef {
	// Remove fenced code blocks first to avoid matching [[links]] inside them
	cleaned := removeFencedCodeBlocks(content)
	return extractLinks(cleaned)
}

// removeFencedCodeBlocks strips fenced code blocks (triple backticks) from content.
// [[links]] inside code blocks are replaced so they won't be extracted.
func removeFencedCodeBlocks(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			result.WriteString(line)
			result.WriteByte('\n')
			continue
		}
		if inCodeBlock {
			// Replace [[ and ]] inside code blocks with placeholder so links aren't extracted
			replaced := strings.ReplaceAll(line, "[[", "﹝﹝")
			replaced = strings.ReplaceAll(replaced, "]]", "﹞﹞")
			result.WriteString(replaced)
			result.WriteByte('\n')
		} else {
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}

// extractLinks finds [[link]] patterns in content.
func extractLinks(content string) []LinkRef {
	matches := wikiLinkRe.FindAllStringSubmatch(content, -1)
	if matches == nil {
		return nil
	}

	refs := make([]LinkRef, 0, len(matches))
	seen := make(map[string]bool)

	for _, m := range matches {
		target := strings.TrimSpace(m[1])
		if target == "" {
			continue
		}

		// Deduplicate by target
		if seen[target] {
			continue
		}
		seen[target] = true

		ref := LinkRef{Target: target}
		if len(m) > 2 && m[2] != "" {
			ref.Text = strings.TrimSpace(m[2])
		}
		refs = append(refs, ref)
	}

	return refs
}
