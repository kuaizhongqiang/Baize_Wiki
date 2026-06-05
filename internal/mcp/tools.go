package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
)

// RunBuildFunc is the signature of the Wiki build function provided by the app layer.
// It must be injected to avoid an import cycle (mcp → app → mcp).
type RunBuildFunc func(ctx context.Context, source, output, configPath string, level int, draft, quiet bool) BuildResult

// BuildResult mirrors app.BuildResult, defined here to break the import cycle.
type BuildResult struct {
	Success    bool          `json:"success"`
	DurationMs int64         `json:"duration_ms"`
	Summary    BuildSummary  `json:"summary"`
	Errors     []string      `json:"errors"`
	Warnings   []string      `json:"warnings"`
}

// BuildSummary mirrors app.Summary.
type BuildSummary struct {
	TotalFiles  int `json:"total_files"`
	Parsed      int `json:"parsed"`
	Pages       int `json:"pages"`
	Directories int `json:"directories"`
}

// toolWikiBuild handles wiki_build: builds or updates a Wiki from source.
func toolWikiBuild(buildFn RunBuildFunc) ToolHandler {
	return func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		var p struct {
			Source string `json:"source"`
			Output string `json:"output"`
			Config string `json:"config"`
			Level  int    `json:"level"`
		}
		if params != nil {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid params: " + err.Error()}
			}
		}

		result := buildFn(ctx, p.Source, p.Output, p.Config, p.Level, false, false)
		if !result.Success {
			errMsg := "build failed"
			if len(result.Errors) > 0 {
				errMsg = result.Errors[0]
			}
			return NewMCPErrorResult(fmt.Sprintf(`{"code":"ERR_BUILD_FAILED","message":"%s"}`, errMsg)), nil
		}

		data, _ := json.Marshal(map[string]any{
			"success":    true,
			"duration_ms": result.DurationMs,
			"summary": map[string]int{
				"pages":       result.Summary.Pages,
				"directories": result.Summary.Directories,
			},
		})
		return NewMCPToolResult(string(data)), nil
	}
}

// toolWikiRead handles wiki_read: reads a page from the Wiki directory.
func toolWikiRead(wikiDir string) ToolHandler {
	return func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		var p struct {
			Path   string `json:"path"`
			Format string `json:"format"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid params"}
		}
		if p.Path == "" {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: "path is required"}
		}

		// Path security: prevent directory traversal
		safePath, err := secureJoin(wikiDir, p.Path)
		if err != nil {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid path: " + err.Error()}
		}

		data, err := os.ReadFile(safePath)
		if err != nil {
			if os.IsNotExist(err) {
				return NewMCPErrorResult(`{"code":"ERR_PAGE_NOT_FOUND","message":"page not found"}`), nil
			}
			return nil, &ErrorObj{Code: ErrInternal, Message: err.Error()}
		}

		return NewMCPToolResult(string(data)), nil
	}
}

// toolWikiList handles wiki_list: lists the Wiki directory tree.
func toolWikiList(wikiDir string) ToolHandler {
	return func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		var p struct {
			Dir          string `json:"dir"`
			Depth        int    `json:"depth"`
			IncludePages bool   `json:"include_pages"`
		}
		if params != nil {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid params"}
			}
		}
		if p.Depth == 0 {
			p.Depth = 1
		}
		if !p.IncludePages {
			p.IncludePages = true // default true
		}

		rootDir := wikiDir
		if p.Dir != "" {
			var err error
			rootDir, err = secureJoin(wikiDir, p.Dir)
			if err != nil {
				return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid dir: " + err.Error()}
			}
		}

		tree, err := buildDirTree(rootDir, wikiDir, p.Depth, p.IncludePages)
		if err != nil {
			return nil, &ErrorObj{Code: ErrInternal, Message: err.Error()}
		}

		data, _ := json.Marshal(tree)
		return NewMCPToolResult(string(data)), nil
	}
}

// dirEntry represents a directory tree node for wiki_list output.
type dirEntry struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"` // "directory" or "page"
	Children []dirEntry `json:"children,omitempty"`
	Title    string     `json:"title,omitempty"`
}

// buildDirTree recursively builds a directory tree structure.
func buildDirTree(root, baseDir string, depth int, includePages bool) (dirEntry, error) {
	rel, err := filepath.Rel(baseDir, root)
	if err != nil {
		return dirEntry{}, err
	}

	entry := dirEntry{
		Name: rel,
		Type: "directory",
	}

	if depth == 0 {
		return entry, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return entry, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir() // directories first
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		if e.Name() == ".baize" {
			continue
		}

		childPath := filepath.Join(root, e.Name())

		if e.IsDir() {
			child, err := buildDirTree(childPath, baseDir, depth-1, includePages)
			if err != nil {
				continue
			}
			entry.Children = append(entry.Children, child)
		} else if includePages && filepath.Ext(e.Name()) == ".md" {
			entry.Children = append(entry.Children, dirEntry{
				Name:  e.Name(),
				Type:  "page",
				Title: trimExt(e.Name()),
			})
		}
	}

	return entry, nil
}

// toolWikiAdd handles wiki_add: adds or updates a page in the Wiki.
func toolWikiAdd(wikiDir string) ToolHandler {
	return func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		var p struct {
			Path      string   `json:"path"`
			Content   string   `json:"content"`
			Tags      []string `json:"tags"`
			Overwrite bool     `json:"overwrite"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid params"}
		}
		if p.Path == "" || p.Content == "" {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: "path and content are required"}
		}

		// Path security
		safePath, err := secureJoin(wikiDir, p.Path)
		if err != nil {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: "invalid path: " + err.Error()}
		}

		// Ensure .md extension
		if filepath.Ext(safePath) == "" {
			safePath += ".md"
		}

		// Check if file exists and overwrite is not allowed
		if !p.Overwrite {
			if _, err := os.Stat(safePath); err == nil {
				return NewMCPErrorResult(`{"code":"ERR_PAGE_EXISTS","message":"page already exists (use overwrite=true to replace)"}`), nil
			}
		}

		// Add frontmatter with tags if provided
		content := p.Content
		if len(p.Tags) > 0 {
			tagStr := "["
			for i, tag := range p.Tags {
				if i > 0 {
					tagStr += ", "
				}
				tagStr += `"` + tag + `"`
			}
			tagStr += "]"
			content = fmt.Sprintf("---\ntags: %s\n---\n\n%s", tagStr, p.Content)
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(safePath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return nil, &ErrorObj{Code: ErrInternal, Message: err.Error()}
		}

		// Atomic write
		tmpPath := safePath + ".tmp"
		if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
			return nil, &ErrorObj{Code: ErrInternal, Message: err.Error()}
		}
		if err := os.Rename(tmpPath, safePath); err != nil {
			os.Remove(tmpPath)
			return nil, &ErrorObj{Code: ErrInternal, Message: err.Error()}
		}

		return NewMCPToolResult(fmt.Sprintf(`{"path":"%s","status":"created"}`, p.Path)), nil
	}
}

// toolWikiStats handles wiki_stats: returns Wiki statistics.
func toolWikiStats(wikiDir string) ToolHandler {
	return func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		store := storage.NewStore()
		wiki, err := store.ReadMeta(wikiDir)
		if err != nil {
			if err == model.ErrWikiNotFound {
				return NewMCPErrorResult(`{"code":"ERR_WIKI_NOT_FOUND","message":"wiki not found at ` + wikiDir + `"}`), nil
			}
			return nil, &ErrorObj{Code: ErrInternal, Message: err.Error()}
		}

		// Count .md files and directories
		pageFiles := 0
		dirCount := 0
		tags := make(map[string]bool)

		filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(wikiDir, path)
			if rel == ".baize" || (len(rel) > 6 && rel[:7] == ".baize"+string(filepath.Separator)) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.IsDir() {
				if path != wikiDir {
					dirCount++
				}
			} else if filepath.Ext(path) == ".md" {
				pageFiles++
			}
			return nil
		})

		stats := map[string]any{
			"name":            wiki.Name,
			"page_count":      wiki.PageCount,
			"page_files":      pageFiles,
			"directory_count": dirCount,
			"updated_at":      wiki.UpdatedAt.Format(time.RFC3339),
			"wiki_path":       wikiDir,
		}

		if len(tags) > 0 {
			tagList := make([]string, 0, len(tags))
			for t := range tags {
				tagList = append(tagList, t)
			}
			sort.Strings(tagList)
			stats["tags"] = tagList
		}

		data, _ := json.Marshal(stats)
		return NewMCPToolResult(string(data)), nil
	}
}

// RegisterAllTools registers all 5 MCP tools on the server.
func RegisterAllTools(srv *Server, wikiDir string, buildFn RunBuildFunc) {
	srv.RegisterTool(MCPToolDefinition{
		Name:        "wiki_build",
		Description: "Build or update Wiki from source directory. Scans specified path and generates structured Wiki output at configurable complexity levels.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"source": {Type: "string", Description: "Source file or directory path"},
				"output": {Type: "string", Description: "Wiki output directory"},
				"config": {Type: "string", Description: "Config file path (default ./baize.yaml)"},
				"level":  {Type: "integer", Description: "Output complexity: 1 | 2 | 3"},
			},
		},
	}, toolWikiBuild(buildFn))

	srv.RegisterTool(MCPToolDefinition{
		Name:        "wiki_read",
		Description: "Read a Wiki page's full Markdown content by its relative path.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"path":   {Type: "string", Description: "Page path relative to Wiki root (e.g. \"architecture/data-model.md\")"},
				"format": {Type: "string", Description: "Return format: markdown | html | text", Enum: []string{"markdown", "html", "text"}},
			},
			Required: []string{"path"},
		},
	}, toolWikiRead(wikiDir))

	srv.RegisterTool(MCPToolDefinition{
		Name:        "wiki_list",
		Description: "Browse the Wiki directory structure. Lists pages and subdirectories at the specified path.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"dir":           {Type: "string", Description: "Directory path relative to Wiki root (default: root)"},
				"depth":         {Type: "integer", Description: "Recursion depth (default 1, -1 for unlimited)", Default: 1},
				"include_pages": {Type: "boolean", Description: "Include pages in listing (default true)", Default: true},
			},
		},
	}, toolWikiList(wikiDir))

	srv.RegisterTool(MCPToolDefinition{
		Name:        "wiki_add",
		Description: "Add a new page or update an existing page in the Wiki. Content is in Markdown format.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"path":      {Type: "string", Description: "Page path relative to Wiki root (e.g. \"guide/debugging.md\")"},
				"content":   {Type: "string", Description: "Markdown content of the page"},
				"tags":      {Type: "array", Description: "List of tags", Items: &struct{ Type string }{Type: "string"}},
				"overwrite": {Type: "boolean", Description: "Overwrite existing page (default false)", Default: false},
			},
			Required: []string{"path", "content"},
		},
	}, toolWikiAdd(wikiDir))

	srv.RegisterTool(MCPToolDefinition{
		Name:        "wiki_stats",
		Description: "Get overall Wiki statistics including page count, directory count, tags, and last update time.",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]PropertySchema{},
		},
	}, toolWikiStats(wikiDir))
}

// secureJoin joins a base directory with a user-supplied path and ensures
// the result does not escape the base directory.
func secureJoin(baseDir, userPath string) (string, error) {
	cleanPath := filepath.Clean(userPath)
	if filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	// Check for directory traversal
	joined := filepath.Join(baseDir, cleanPath)
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	// Ensure the resolved path is within the base directory
	rel, err := filepath.Rel(absBase, absJoined)
	if err != nil {
		return "", err
	}
	if rel == ".." || (len(rel) > 2 && rel[:3] == ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes wiki directory")
	}

	return absJoined, nil
}

// trimExt removes the extension from a filename.
func trimExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}
