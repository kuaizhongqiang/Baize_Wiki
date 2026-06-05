package main

import (
	"fmt"
	"os"

	"github.com/kuaizhongqiang/baize-wiki/internal/app"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	cfgFile string
	verbose bool
	jsonOut bool
)

var rootCmd = &cobra.Command{
	Use:   "baize-wiki",
	Short: "Baize Wiki — AI Agent oriented Wiki generation and consumption tool",
	Long: `Baize Wiki (白泽维基) scans a source directory of documents,
parses them, and generates a structured Wiki output at configurable
complexity Levels 1/2/3.

Documentation: https://github.com/kuaizhongqiang/baize-wiki`,
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "baize.yaml", "Config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&jsonOut, "json", "j", false, "JSON format output")

	rootCmd.AddCommand(
		app.NewInitCmd(),
		app.NewBuildCmd(),
		app.NewInfoCmd(),
		app.NewMCPCmd(),
	)

	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	rootCmd.SetVersionTemplate("baize-wiki version {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
