package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kuaizhongqiang/baize-wiki/internal/config"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/generator"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/index"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/parser"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/scanner"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
	"github.com/spf13/cobra"
)

// NewBuildCmd creates the `build` subcommand.
func NewBuildCmd() *cobra.Command {
	var output string
	var level int
	var draft, quiet bool

	cmd := &cobra.Command{
		Use:   "build [source]",
		Short: "Build/update Wiki from source directory",
		Long:  "Scan source documents, parse them, and generate a structured Wiki at the configured output directory.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := ""
			if len(args) > 0 {
				source = args[0]
			}

			// Validate level flag if explicitly set
			if cmd.Flags().Changed("level") {
				if level < 1 || level > 3 {
					return fmt.Errorf("invalid level: %d, must be 1, 2, or 3", level)
				}
			}

			configPath, _ := cmd.Flags().GetString("config")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			result := RunBuild(context.Background(), source, output, configPath, level, draft, quiet)

			if jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				if !result.Success {
					os.Exit(1)
				}
				return nil
			}

			if !result.Success {
				for _, err := range result.Errors {
					fmt.Fprintf(os.Stderr, "✗ %s\n", err)
				}
				return fmt.Errorf("build failed")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory (overrides config)")
	cmd.Flags().IntVarP(&level, "level", "l", 0, "Output complexity: 1=flat, 2=structured, 3=deep")
	cmd.Flags().BoolVar(&draft, "draft", false, "Include pages marked as draft")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (errors and summary only)")

	return cmd
}

// BuildResult holds the outcome of a Wiki build.
type BuildResult struct {
	Success    bool     `json:"success"`
	DurationMs int64    `json:"duration_ms"`
	Summary    Summary  `json:"summary"`
	Errors     []string `json:"errors"`
	Warnings   []string `json:"warnings"`
}

// Summary contains build statistics.
type Summary struct {
	TotalFiles  int `json:"total_files"`
	Parsed      int `json:"parsed"`
	Pages       int `json:"pages"`
	Directories int `json:"directories"`
}

// RunBuild executes the full build pipeline: config → scanner → parser → generator → storage.
func RunBuild(ctx context.Context, source, output, configPath string, level int, draft, quiet bool) BuildResult {
	start := time.Now()
	result := BuildResult{}

	// 1. Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		result.Errors = append(result.Errors, "config: "+err.Error())
		return result
	}
	cfg = cfg.Merge(level, output)

	// Determine source path
	sourceDir := source
	if sourceDir == "" {
		if len(cfg.Scan.Paths) > 0 {
			sourceDir = cfg.Scan.Paths[0]
		} else {
			sourceDir = "."
		}
	}
	absSource, err := filepath.Abs(sourceDir)
	if err != nil {
		result.Errors = append(result.Errors, "source: "+err.Error())
		return result
	}

	// Validate source
	if _, err := os.Stat(absSource); os.IsNotExist(err) {
		result.Errors = append(result.Errors, "source directory does not exist: "+absSource)
		return result
	}

	// Resolve output directory early for source/output overlap check
	outputDir := cfg.Output.Dir
	if output != "" {
		outputDir = output
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		result.Errors = append(result.Errors, "output: "+err.Error())
		return result
	}
	if absSource == absOutput {
		result.Errors = append(result.Errors, "source and output directories must be different: "+absSource)
		return result
	}

	// 2. Scan
	scanCfg := scanner.ScanConfig{
		MaxSize: cfg.Scan.MaxSize,
		Exclude: cfg.Scan.Exclude,
	}
	files, err := scanner.Scan(ctx, absSource, scanCfg)
	if err != nil {
		result.Errors = append(result.Errors, "scan: "+err.Error())
		return result
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "✓ 扫描完成: 找到 %d 个文件 (跳过 %d 个二进制, 忽略 %d 个)\n",
			len(files), 0, 0)
	}

	if len(files) == 0 {
		result.Errors = append(result.Errors, "no valid files found in source directory")
		return result
	}

	// 3. Parse
	pages, warnings := parser.ParseBatch(ctx, files)
	result.Warnings = warnings
	if !quiet {
		fmt.Fprintf(os.Stderr, "✓ 解析完成: %d/%d 成功\n", len(pages), len(files))
	}

	if len(pages) == 0 {
		result.Errors = append(result.Errors, "no pages could be parsed")
		return result
	}

	// 4. Generate
	wiki := model.NewWiki(cfg.Name, absSource, absOutput, cfg)

	store := storage.NewStore()
	gen := generator.NewGenerator(store)

	if err := gen.Generate(ctx, wiki, pages); err != nil {
		result.Errors = append(result.Errors, "generate: "+err.Error())
		return result
	}

	// Count directories
	dirCount := countDirs(pages, cfg.Output.Level)

	if !quiet {
		fmt.Fprintf(os.Stderr, "✓ 生成完成: 输出到 %s (%d 页面, %d 目录, Level %d)\n",
			absOutput, len(pages), dirCount, cfg.Output.Level)
	}

	// 5. Build full-text index (non-blocking on failure)
	indexPath := filepath.Join(absOutput, ".baize", "index.bleve")
	if idx, err := index.NewIndex(indexPath); err == nil {
		if err := idx.Build(ctx, pages); err != nil {
			result.Warnings = append(result.Warnings, "index build warning: "+err.Error())
		}
		idx.Close()
	} else {
		result.Warnings = append(result.Warnings, "index create warning: "+err.Error())
	}

	result.Success = true
	result.DurationMs = time.Since(start).Milliseconds()
	result.Summary = Summary{
		TotalFiles:  len(files),
		Parsed:      len(pages),
		Pages:       len(pages),
		Directories: dirCount,
	}

	return result
}

// countDirs estimates the number of output directories based on level and pages.
func countDirs(pages []*model.Page, level int) int {
	if level == 1 {
		return 1
	}
	dirs := make(map[string]bool)
	for _, p := range pages {
		dir := filepath.Dir(p.Path)
		if dir != "." {
			parts := splitPath(dir)
			if level == 2 && len(parts) > 0 {
				dirs[parts[0]] = true
			} else {
				dirs[dir] = true
			}
		}
	}
	count := len(dirs)
	if count == 0 {
		count = 1
	}
	return count + 1 // +1 for root
}

func splitPath(p string) []string {
	dirs := make([]string, 0)
	dir := filepath.Dir(p)
	for dir != "." {
		dirs = append([]string{filepath.Base(dir)}, dirs...)
		dir = filepath.Dir(dir)
	}
	return dirs
}
