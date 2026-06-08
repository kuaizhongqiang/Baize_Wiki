package catalog

import (
	"context"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

// LocalSummarizer uses first-paragraph extraction and simple heuristics.
// No external dependencies, suitable for free tier and offline use.
type LocalSummarizer struct {
	lang      string
	maxTokens int
}

// Summarize implements Summarizer for local backend.
func (s *LocalSummarizer) Summarize(ctx context.Context, page *model.Page, lang string) (*CatalogResult, error) {
	// Use frontmatter description if available
	abstract := page.Meta.Description
	if abstract == "" {
		abstract = extractFirstParagraph(page.Content)
	}

	// Extract basic keywords from content
	keywords := extractKeywords(page.Content, page.Title)

	// Basic entity detection (class names from paths, sections)
	entities := extractEntities(page)

	return &CatalogResult{
		Abstract: abstract,
		Keywords: keywords,
		Entities: entities,
	}, nil
}

// extractFirstParagraph returns the first meaningful paragraph of content.
func extractFirstParagraph(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	var b strings.Builder
	inParagraph := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inParagraph {
				break // end of first paragraph
			}
			continue
		}

		// Skip headings, code blocks, separators
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "===") {
			if inParagraph {
				break
			}
			continue
		}

		if !inParagraph {
			inParagraph = true
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(trimmed)
	}

	result := b.String()
	if len(result) > 400 {
		result = result[:400]
	}
	return result
}

// extractKeywords extracts simple keywords from title and content.
func extractKeywords(content, title string) []string {
	seen := make(map[string]bool)
	var keywords []string

	// Add title words if meaningful
	if title != "" {
		title = strings.TrimSuffix(title, ".md")
		title = strings.NewReplacer("-", " ", "_", " ", "/", " ").Replace(title)
		words := strings.Fields(title)
		for _, w := range words {
			w = strings.TrimSpace(w)
			if len(w) > 1 && !seen[w] {
				seen[w] = true
				keywords = append(keywords, w)
			}
		}
	}

	// Scan for code identifiers (camelCase, PascalCase, snake_case)
	words := strings.Fields(content)
	for _, w := range words {
		w = strings.TrimSpace(w)
		// Skip short words and common punctuation
		if len(w) < 3 || isStopWord(w) {
			continue
		}
		// Detect PascalCase or camelCase as code identifiers
		if containsUpperCase(w) && !seen[w] {
			seen[w] = true
			keywords = append(keywords, w)
		}
		if len(keywords) >= 10 {
			break
		}
	}

	return keywords
}

// containsUpperCase checks if a string contains uppercase letters.
func containsUpperCase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

// isStopWord filters common non-keyword tokens.
func isStopWord(s string) bool {
	stops := map[string]bool{
		"the": true, "this": true, "that": true, "with": true, "from": true,
		"into": true, "over": true, "void": true, "bool": true, "true": true,
		"false": true, "null": true, "nil": true, "int": true, "var": true,
		"for": true, "while": true, "else": true, "new": true, "get": true,
		"set": true, "has": true, "not": true, "are": true, "was": true,
		"but": true, "you": true, "all": true, "can": true,
	}
	return stops[s]
}

// extractEntities extracts basic entities from the page structure.
func extractEntities(page *model.Page) []model.Entity {
	var entities []model.Entity

	// Add the page itself as a defined entity
	if page.Title != "" {
		ext := detectLanguage(page)
		eType := "module"
		if ext == "markdown" {
			eType = "concept"
		}
		entities = append(entities, model.Entity{
			Name: page.Title,
			Type: eType,
			Role: "defined",
		})
	}

	// Add section headings as concepts
	for _, sec := range page.Sections {
		entities = append(entities, model.Entity{
			Name: sec.Title,
			Type: "concept",
			Role: "defined",
		})
		if len(entities) >= 10 {
			break
		}
	}

	return entities
}
