package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewMCPCmd creates the `mcp` subcommand.
func NewMCPCmd() *cobra.Command {
	var transportType, addr string

	cmd := &cobra.Command{
		Use:   "mcp [wiki-dir]",
		Short: "Start MCP Server mode",
		Long: `Start the MCP (Model Context Protocol) Server for AI Agent integration.

In stdio mode (default), the server reads JSON-RPC 2.0 requests from stdin
and writes responses to stdout. Log output goes to stderr.

This mode is designed for MCP clients like Claude Code or Cline that
automatically launch the process and communicate over stdio.

In TCP mode, the server listens on a TCP port for a single connection.
Useful for Docker deployments or when you need to connect remotely.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wikiDir := "./wiki"
			if len(args) > 0 {
				wikiDir = args[0]
			}

			if err := RunServe(context.Background(), wikiDir, transportType, addr); err != nil {
				fmt.Fprintf(os.Stderr, "MCP Server error: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&transportType, "transport", "t", "stdio", "Transport: stdio | tcp")
	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "TCP listen address")
	return cmd
}
