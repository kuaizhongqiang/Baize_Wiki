package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the `init` subcommand.
func NewInitCmd() *cobra.Command {
	var output, name string
	var force bool

	cmd := &cobra.Command{
		Use:   "init [source]",
		Short: "Initialize a Wiki project (generate baize.yaml)",
		Long: `Initialize a Wiki project by generating a baize.yaml configuration template.
Run 'baize-wiki build' after initialization to generate the Wiki.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := "."
			if len(args) > 0 {
				source = args[0]
			}
			return RunInit(source, output, name, force)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "./wiki", "Output directory")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Wiki name (default: source directory basename)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")

	return cmd
}

// RunInit initializes a Wiki project by generating a baize.yaml template.
func RunInit(source, output, name string, force bool) error {
	sourceDir := source
	if sourceDir == "" {
		sourceDir = "."
	}

	if name == "" {
		name = filepath.Base(sourceDir)
	}

	// Check if config already exists
	configPath := "baize.yaml"
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("baize.yaml already exists (use --force to overwrite)")
	}

	template := fmt.Sprintf(`# 由 Baize Wiki 自动生成
name: "%s"
scan:
  paths:
    - %s
  exclude: []
  max_size: 10485760
output:
  dir: %s
  level: 2
  clean: false
organize:
  by: directory
features:
  draft: false
`, name, sourceDir, output)

	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write baize.yaml: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ 已生成 baize.yaml\n")
	fmt.Fprintf(os.Stderr, "运行 'baize-wiki build' 生成 Wiki\n")
	return nil
}
