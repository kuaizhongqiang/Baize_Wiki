package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
	"github.com/kuaizhongqiang/baize-wiki/internal/mcp"
)

// RunServe starts the MCP Server for the given Wiki directory.
// transportType: "stdio" (default) or "tcp"
// addr: TCP listen address (only used when transportType is "tcp")
func RunServe(ctx context.Context, wikiDir, transportType, addr string) error {
	if wikiDir == "" {
		wikiDir = "./wiki"
	}

	absWiki, err := filepath.Abs(wikiDir)
	if err != nil {
		return fmt.Errorf("wiki directory: %w", err)
	}

	// Validate wiki directory exists
	store := storage.NewStore()
	if _, err := store.ReadMeta(absWiki); err != nil {
		if err == model.ErrWikiNotFound {
			return fmt.Errorf("wiki not found at %s (run 'baize-wiki build' first)", absWiki)
		}
		return fmt.Errorf("wiki: %w", err)
	}

	// Create transport
	var transport mcp.Transport

	switch transportType {
	case "", "stdio":
		transport = mcp.NewStdioTransport()
		fmt.Fprintf(os.Stderr, "MCP Server starting (stdio mode, wiki: %s)\n", absWiki)
	case "tcp":
		if addr == "" {
			addr = ":8080"
		}
		tcpTrans, err := mcp.NewTCPServerTransport(addr)
		if err != nil {
			return fmt.Errorf("tcp listen: %w", err)
		}
		fmt.Fprintf(os.Stderr, "MCP Server listening on %s (wiki: %s)\n", addr, absWiki)

		// Wait for a single connection
		if err := tcpTrans.Accept(); err != nil {
			return fmt.Errorf("tcp accept: %w", err)
		}
		fmt.Fprintf(os.Stderr, "TCP client connected\n")
		transport = tcpTrans
	default:
		return fmt.Errorf("unknown transport: %s (must be stdio or tcp)", transportType)
	}

	// Create server and register tools
	server := mcp.NewServer(transport)
	mcp.RegisterAllTools(server, absWiki, func(ctx context.Context, source, output, configPath string, level int, draft, quiet, scanAll bool) (bool, int64, int, int, []string) {
		r := RunBuild(ctx, source, output, configPath, level, draft, quiet, scanAll)
		return r.Success, r.DurationMs, r.Summary.Pages, r.Summary.Directories, r.Errors
	})

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case sig := <-sigCh:
			fmt.Fprintf(os.Stderr, "Received signal %v, shutting down...\n", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	fmt.Fprintf(os.Stderr, "MCP Server ready\n")
	return server.Run(ctx)
}
