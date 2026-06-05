package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// RunInfo displays information about a Wiki directory.
func RunInfo(wikiDir string, showTree, showStats, jsonOutput bool) (interface{}, error) {
	if wikiDir == "" {
		wikiDir = "./wiki"
	}

	metaPath := filepath.Join(wikiDir, ".baize", "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("wiki not found at %s (run 'baize-wiki build' first)", wikiDir)
		}
		return nil, err
	}

	if jsonOutput {
		return map[string]interface{}{
			"success": true,
			"path":    wikiDir,
			"meta":    string(metaData),
		}, nil
	}

	fmt.Printf("Wiki 目录: %s\n", wikiDir)
	fmt.Printf("元数据: %s\n", string(metaData))

	if showStats {
		// Stats are shown via meta.json content above
	}

	if showTree {
		fmt.Printf("%s/\n", filepath.Base(wikiDir))
		printTree(wikiDir, "")
	}

	return nil, nil
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
