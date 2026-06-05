package app

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/index"
	"github.com/spf13/cobra"
)

// SearchResultJSON is the JSON output structure for search results.
type SearchResultJSON struct {
	Success bool                `json:"success"`
	Query   string              `json:"query"`
	Total   int                 `json:"total"`
	Results []index.SearchResult `json:"results"`
}

// NewSearchCmd creates the `search` subcommand.
func NewSearchCmd() *cobra.Command {
	var wikiDir string
	var limit int
	var tags []string
	var withContent, jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Wiki content",
		Long:  "Search Wiki pages by keywords. Requires a built Wiki with search index.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			result := RunSearch(context.Background(), wikiDir, query, index.SearchOpts{
				Tags:        tags,
				Limit:       limit,
				WithContent: withContent,
			})

			if jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if result.Total == 0 {
				fmt.Println("未找到匹配内容")
				return nil
			}

			fmt.Printf("找到 %d 个结果:\n\n", result.Total)
			for i, r := range result.Results {
				fmt.Printf("%d. %s (score: %.2f)\n", i+1, r.Path, r.Score)
				if r.Snippet != "" {
					fmt.Printf("   ...%s...\n", r.Snippet)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&wikiDir, "wiki", "w", "./wiki", "Wiki directory")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Max results")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", nil, "Filter by tags")
	cmd.Flags().BoolVarP(&withContent, "with-content", "c", false, "Include full content")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "JSON format output")

	return cmd
}

// RunSearch executes a search against the Wiki's bleve index.
func RunSearch(ctx context.Context, wikiDir, queryStr string, opts index.SearchOpts) SearchResultJSON {
	result := SearchResultJSON{
		Success: false,
		Query:   queryStr,
	}

	indexPath := filepath.Join(wikiDir, ".baize", "index.bleve")
	idx, err := index.NewIndex(indexPath)
	if err != nil {
		result.Results = []index.SearchResult{}
		return result
	}
	defer idx.Close()

	results, err := idx.Search(ctx, queryStr, opts)
	if err != nil {
		result.Results = []index.SearchResult{}
		return result
	}

	result.Success = true
	result.Total = len(results)
	result.Results = results
	return result
}
