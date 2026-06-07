package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/parser"
	"github.com/spf13/cobra"
)

// NewInfoCmd creates the `info` subcommand.
func NewInfoCmd() *cobra.Command {
	var showTree, showStats, jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info [path]",
		Short: "View Wiki or page information",
		Long:  "Display information about a Wiki directory or a specific page.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wikiDir := "./wiki"
			if len(args) > 0 {
				wikiDir = args[0]
			}
			result, err := RunInfo(wikiDir, showTree, showStats, jsonOutput)
			if err != nil {
				return err
			}
			if jsonOutput && result != nil {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showTree, "tree", "t", false, "Show Wiki directory as a tree")
	cmd.Flags().BoolVarP(&showStats, "stats", "s", false, "Show Wiki statistics")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "JSON format output")

	return cmd
}

// PageInfo holds detailed information about a single page.
type PageInfo struct {
	Title     string `json:"title"`
	Path      string `json:"path"`
	Filesize  int64  `json:"filesize"`
	LinkCount int    `json:"link_count"`
}

// RunInfo displays information about a Wiki directory or a specific page.
func RunInfo(wikiDir string, showTree, showStats, jsonOutput bool) (interface{}, error) {
	if wikiDir == "" {
		wikiDir = "./wiki"
	}

	// Check if argument is a specific page file
	info, err := os.Stat(wikiDir)
	if err == nil && !info.IsDir() && (filepath.Ext(wikiDir) == ".md" || filepath.Ext(wikiDir) == ".mdx") {
		return showPageInfo(wikiDir, jsonOutput)
	}

	// Wiki-level info
	metaPath := filepath.Join(wikiDir, ".baize", "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("wiki not found at %s (run 'baize-wiki build' first)", wikiDir)
		}
		return nil, err
	}

	if showStats {
		printStats(wikiDir, metaData)
		return nil, nil
	}

	if jsonOutput {
		return buildJSONOutput(wikiDir, metaData, showTree), nil
	}

	fmt.Printf("Wiki 目录: %s\n", wikiDir)
	fmt.Printf("元数据: %s\n", string(metaData))

	if showTree {
		fmt.Printf("%s/\n", filepath.Base(wikiDir))
		printTree(wikiDir, "")
	}

	return nil, nil
}

// showPageInfo displays information about a single page.
func showPageInfo(path string, jsonOutput bool) (*PageInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	refs := parser.ExtractWikiLinks(string(content))

	info := &PageInfo{
		Title:     title,
		Path:      path,
		Filesize:  int64(len(content)),
		LinkCount: len(refs),
	}

	if jsonOutput {
		return info, nil
	}

	fmt.Printf("页面: %s\n", path)
	fmt.Printf("  标题: %s\n", title)
	fmt.Printf("  大小: %d 字节\n", info.Filesize)
	fmt.Printf("  链接数: %d\n", info.LinkCount)

	return info, nil
}

// buildJSONOutput builds the JSON response for wiki info.
func buildJSONOutput(wikiDir string, metaData []byte, showTree bool) map[string]interface{} {
	out := map[string]interface{}{
		"success": true,
		"path":    wikiDir,
		"meta":    string(metaData),
	}
	if showTree {
		out["tree"] = buildTreeJSON(wikiDir, "")
	}
	return out
}

// buildTreeJSON recursively builds a JSON tree structure.
func buildTreeJSON(dir, prefix string) []map[string]interface{} {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var nodes []map[string]interface{}
	for _, entry := range entries {
		if entry.Name() == ".baize" {
			continue
		}
		node := map[string]interface{}{
			"name": entry.Name(),
		}
		if entry.IsDir() {
			node["type"] = "directory"
			node["children"] = buildTreeJSON(filepath.Join(dir, entry.Name()), prefix+"  ")
		} else {
			node["type"] = "page"
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// printStats prints Wiki statistics from meta.json.
func printStats(wikiDir string, metaData []byte) {
	var meta map[string]interface{}
	if err := json.Unmarshal(metaData, &meta); err != nil {
		fmt.Printf("无法解析 meta.json: %v\n", err)
		return
	}

	pageCount := 0
	if pc, ok := meta["page_count"].(float64); ok {
		pageCount = int(pc)
	}

	// Count files/directories and aggregate link data
	fileCount := 0
	pageFiles := 0
	dirCount := 0
	totalLinks := 0

	filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if path == wikiDir {
			return nil
		}
		rel, _ := filepath.Rel(wikiDir, path)
		if rel == ".baize" || strings.HasPrefix(rel, ".baize"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
			if filepath.Ext(path) == ".md" {
				pageFiles++
				// Count [[wiki-links]] in the page
				content, err := os.ReadFile(path)
				if err == nil {
					refs := parser.ExtractWikiLinks(string(content))
					totalLinks += len(refs)
				}
			}
		}
		return nil
	})

	fmt.Printf("  页面数: %d\n", pageCount)
	fmt.Printf("  .md 文件: %d\n", pageFiles)
	fmt.Printf("  目录数: %d\n", dirCount)
	fmt.Printf("  总文件数: %d\n", fileCount)
	fmt.Printf("  链接总数: %d\n", totalLinks)

	if name, ok := meta["name"].(string); ok && name != "" {
		fmt.Printf("  Wiki 名称: %s\n", name)
	}
	if ver, ok := meta["version"].(float64); ok && ver > 0 {
		fmt.Printf("  版本: %.0f\n", ver)
	}
}

// printTree recursively prints a directory tree of wiki files.
func printTree(dir, prefix string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for i, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == ".baize" {
				continue
			}
			fmt.Printf("%s├── %s/\n", prefix, entry.Name())
			printTree(filepath.Join(dir, entry.Name()), prefix+"│   ")
		} else if filepath.Ext(entry.Name()) == ".md" {
			branch := "├──"
			if i == len(entries)-1 {
				branch = "└──"
			}
			fmt.Printf("%s%s %s\n", prefix, branch, entry.Name())
		}
	}
}
