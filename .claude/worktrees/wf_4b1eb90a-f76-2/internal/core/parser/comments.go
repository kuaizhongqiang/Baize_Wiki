package parser

import (
	"path/filepath"
	"strings"
)

// commentPrefixes maps file extensions to their single-line comment prefix.
var commentPrefixes = map[string]string{
	".go": "//",
	".js": "//",
	".jsx": "//",
	".ts": "//",
	".tsx": "//",
	".rs": "//",
	".c":  "//",
	".cpp": "//",
	".h":  "//",
	".hpp": "//",
	".java": "//",
	".swift": "//",
	".kt": "//",
	".py": "#",
	".rb": "#",
	".sh": "#",
	".pl": "#",
	".php": "#",
	".yaml": "#",
	".yml": "#",
	".toml": "#",
	".ini": "#",
	".cfg": "#",
}

// multiLineOpen maps file extensions to their multi-line comment opening delimiter.
var multiLineOpen = map[string]string{
	".go": "/*",
	".js": "/*",
	".jsx": "/*",
	".ts": "/*",
	".tsx": "/*",
	".rs": "/*",
	".c":  "/*",
	".cpp": "/*",
	".h":  "/*",
	".hpp": "/*",
	".java": "/*",
	".swift": "/*",
	".kt": "/*",
}

// multiLineClose maps file extensions to their multi-line comment closing delimiter.
var multiLineClose = map[string]string{
	".go": "*/",
	".js": "*/",
	".jsx": "*/",
	".ts": "*/",
	".tsx": "*/",
	".rs": "*/",
	".c":  "*/",
	".cpp": "*/",
	".h":  "*/",
	".hpp": "*/",
	".java": "*/",
	".swift": "*/",
	".kt": "*/",
}

// ExtractComments extracts documentation comments from source code content.
// It returns the file-level comment block (top-of-file comments before any code).
// Supports Go, Python, JavaScript, TypeScript, Rust, and other common languages.
func ExtractComments(filename string, content string) string {
	ext := filepath.Ext(filename)
	linePrefix, hasLine := commentPrefixes[ext]
	if !hasLine && ext == "" {
		return ""
	}

	lines := strings.Split(content, "\n")

	// Collect top-of-file comment block
	var comments []string
	inBlock := false
	seenCode := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle multi-line block comments
		openDelim := multiLineOpen[ext]
		closeDelim := multiLineClose[ext]

		if inBlock {
			// End of block
			if closeDelim != "" && strings.Contains(trimmed, closeDelim) {
				// Extract comment text before the closing delimiter
				beforeClose := trimmed[:strings.Index(trimmed, closeDelim)]
				if cleaned := cleanComment(beforeClose, "", ext); cleaned != "" {
					comments = append(comments, cleaned)
				}
				inBlock = false
				continue
			}
			if cleaned := cleanComment(trimmed, "", ext); cleaned != "" {
				comments = append(comments, cleaned)
			}
			continue
		}

		if seenCode {
			break
		}

		// Check for block comment start
		if openDelim != "" && strings.HasPrefix(trimmed, openDelim) {
			// Check if it starts and ends on the same line
			if closeDelim != "" && strings.Contains(trimmed, closeDelim) {
				inner := trimmed[len(openDelim) : len(trimmed)-len(closeDelim)]
				if cleaned := cleanComment(inner, "", ext); cleaned != "" {
					comments = append(comments, cleaned)
				}
				continue
			}
			// Multi-line block comment
			afterOpen := trimmed[len(openDelim):]
			if cleaned := cleanComment(afterOpen, "", ext); cleaned != "" {
				comments = append(comments, cleaned)
			}
			inBlock = true
			continue
		}

		// Check for line comment
		if hasLine && strings.HasPrefix(trimmed, linePrefix) {
			comment := strings.TrimSpace(trimmed[len(linePrefix):])
			if comment != "" {
				comments = append(comments, comment)
			}
			continue
		}

		// Empty line — continue if we haven't seen code yet
		if trimmed == "" {
			continue
		}

		// Non-comment code line: mark seenCode
		seenCode = true
		if len(comments) > 0 {
			break
		}
	}

	// If we're still in a block comment at EOF, that's fine — include what we collected
	return strings.Join(comments, "\n")
}

// cleanComment removes common leading comment markers and whitespace.
func cleanComment(line, prefix, ext string) string {
	trimmed := strings.TrimSpace(line)

	// Remove leading *
	for strings.HasPrefix(trimmed, "*") {
		trimmed = strings.TrimSpace(trimmed[1:])
	}

	return strings.TrimSpace(trimmed)
}
