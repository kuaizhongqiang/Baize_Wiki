package model

import (
	"fmt"
	"hash/fnv"
	"time"
)

// PageID generates a deterministic ID from a file path.
func PageID(path string) string {
	h := fnv.New64a()
	h.Write([]byte(path))
	return fmt.Sprintf("page_%x", h.Sum(nil))
}

// Page represents a single Wiki page.
type Page struct {
	ID         string      `json:"id"`
	WikiID     string      `json:"wiki_id,omitempty"`
	Path       string      `json:"path"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	Meta       Frontmatter `json:"meta,omitempty"`
	Sections   []Section   `json:"sections,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	Depth      int         `json:"depth"`
	Weight     int         `json:"weight"`
	SourceFile string      `json:"source_file,omitempty"`
	UpdatedAt  time.Time   `json:"updated_at,omitempty"`
}

// Section represents a heading-level section in a document, forming a tree.
type Section struct {
	ID       string    `json:"id"`
	Level    int       `json:"level"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Children []Section `json:"children,omitempty"`
}

// FileInfo contains metadata about a scanned file.
type FileInfo struct {
	Path      string `json:"path"`
	AbsPath   string `json:"abs_path"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
}

// Frontmatter represents YAML frontmatter metadata in a Markdown file.
type Frontmatter struct {
	Title       string         `yaml:"title" json:"title"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	Aliases     []string       `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	Weight      *int           `yaml:"weight,omitempty" json:"weight,omitempty"`
	Draft       bool           `yaml:"draft,omitempty" json:"draft,omitempty"`
	Category    string         `yaml:"category,omitempty" json:"category,omitempty"`
	Custom      map[string]any `yaml:",inline" json:"custom,omitempty"`
}
