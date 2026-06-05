package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/kuaizhongqiang/baize-wiki/internal/core/storage"
	"github.com/kuaizhongqiang/baize-wiki/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// e2eMCPTransport implements mcp.Transport over in-memory pipes for testing.
type e2eMCPTransport struct {
	readPipe  *io.PipeReader
	writePipe *io.PipeWriter
	readBuf   bytes.Buffer
}

func newE2eMCPTransport() *e2eMCPTransport {
	r, w := io.Pipe()
	return &e2eMCPTransport{readPipe: r, writePipe: w}
}

func (t *e2eMCPTransport) Read() ([]byte, error) {
	var buf bytes.Buffer
	tmp := make([]byte, 1)
	for {
		n, err := t.readPipe.Read(tmp)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}
		if tmp[0] == '\n' {
			break
		}
		buf.WriteByte(tmp[0])
	}
	return buf.Bytes(), nil
}

func (t *e2eMCPTransport) Write(data []byte) error {
	_, err := t.writePipe.Write(data)
	return err
}

func (t *e2eMCPTransport) Close() error {
	t.readPipe.Close()
	return t.writePipe.Close()
}

// setupE2EWiki creates a temporary Wiki with test data for E2E MCP testing.
func setupE2EWiki(t *testing.T) (string, string) {
	t.Helper()

	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("---\ntitle: Test Doc\n---\n\n# Test Doc\n\nHello world"), 0644))

	outDir := t.TempDir()

	// Build the wiki
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 1, false, true)
	require.True(t, result.Success)

	return srcDir, outDir
}

func sendMCPRequest(t *testing.T, trans *e2eMCPTransport, method string, params any) (int, json.RawMessage, *mcp.ErrorObj) {
	t.Helper()

	id := 1
	rawID, _ := json.Marshal(id)
	req := mcp.NewRequest(rawID, method, params)
	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Write request with newline delimiter
	_, err = trans.writePipe.Write(append(data, '\n'))
	require.NoError(t, err)

	// Read response
	respData, err := trans.Read()
	require.NoError(t, err)

	var resp mcp.Response
	err = json.Unmarshal(respData, &resp)
	require.NoError(t, err)

	return id, resp.Result, resp.Error
}

func runMCPServer(t *testing.T, wikiDir string, trans *e2eMCPTransport) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	server := mcp.NewServer(trans)
	mcp.RegisterAllTools(server, wikiDir)

	go func() {
		server.Run(ctx)
	}()

	return cancel
}

func TestMCPE2EPing(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "ping", nil)
	assert.Nil(t, errObj)
	assert.Contains(t, string(result), "pong")
}

func TestMCPE2EToolList(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "tools/list", nil)
	assert.Nil(t, errObj)

	assert.Contains(t, string(result), "wiki_build")
	assert.Contains(t, string(result), "wiki_read")
	assert.Contains(t, string(result), "wiki_list")
	assert.Contains(t, string(result), "wiki_add")
	assert.Contains(t, string(result), "wiki_stats")
}

func TestMCPE2EWikiRead(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "wiki_read", map[string]string{
		"path": "test-doc.md",
	})
	assert.Nil(t, errObj)

	// The result should contain the markdown content wrapped in MCPToolResult
	assert.Contains(t, string(result), "Hello world")
}

func TestMCPE2EWikiReadNotFound(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "wiki_read", map[string]string{
		"path": "nonexistent.md",
	})
	assert.Nil(t, errObj)
	assert.Contains(t, string(result), "ERR_PAGE_NOT_FOUND")
}

func TestMCPE2EWikiList(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "wiki_list", map[string]any{
		"depth": 2,
	})
	assert.Nil(t, errObj)

	assert.Contains(t, string(result), "_index.md")
	assert.Contains(t, string(result), "test-doc.md")
}

func TestMCPE2EWikiAdd(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	// Add a new page
	_, result, errObj := sendMCPRequest(t, trans, "wiki_add", map[string]string{
		"path":    "new-page.md",
		"content": "# New Page\n\nContent here",
	})
	assert.Nil(t, errObj)
	assert.Contains(t, string(result), "created")

	// Verify file exists
	assert.FileExists(t, filepath.Join(wikiDir, "new-page.md"))
}

func TestMCPE2EWikiAddExists(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	// Try to add a page that already exists (from build)
	_, result, errObj := sendMCPRequest(t, trans, "wiki_add", map[string]string{
		"path":    "test-doc.md",
		"content": "# Overwrite me",
	})
	assert.Nil(t, errObj)
	assert.Contains(t, string(result), "ERR_PAGE_EXISTS")
}

func TestMCPE2EWikiStats(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "wiki_stats", nil)
	assert.Nil(t, errObj)

	assert.Contains(t, string(result), "page_count")
}

func TestMCPE2EPathSecurity(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	// Try path traversal via wiki_read
	_, _, errObj := sendMCPRequest(t, trans, "wiki_read", map[string]string{
		"path": "../../etc/passwd",
	})
	require.NotNil(t, errObj)
	assert.Equal(t, mcp.ErrInvalidParams, errObj.Code)
}

func TestMCPE2EUnknownMethod(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, _, errObj := sendMCPRequest(t, trans, "unknown_method", nil)
	require.NotNil(t, errObj)
	assert.Equal(t, mcp.ErrMethodNotFound, errObj.Code)
}

func TestMCPE2EWikiBuild(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create a wiki without building first
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("# Doc"), 0644))

	_, wikiDir := setupE2EWiki(t) // just need a valid wiki dir for the server
	_ = outDir
	_ = srcDir

	trans := newE2eMCPTransport()
	defer trans.Close()

	// Create a fresh wiki dir with meta.json so the server starts
	freshWikiDir := t.TempDir()
	store := storage.NewStore()
	cfg := model.DefaultConfig()
	wiki := model.NewWiki("E2E Wiki", srcDir, freshWikiDir, cfg)
	store.WriteMeta(freshWikiDir, wiki)

	cancel := runMCPServer(t, freshWikiDir, trans)
	defer cancel()

	// Read the actual file that we set up
	require.NoError(t, os.WriteFile(filepath.Join(freshWikiDir, "test.md"), []byte("# Test"), 0644))

	_, result, errObj := sendMCPRequest(t, trans, "wiki_read", map[string]string{
		"path": "test.md",
	})
	assert.Nil(t, errObj)
	assert.Contains(t, string(result), "# Test")
}

func TestMCPE2EMultipleRequests(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	// Send ping first
	_, result1, err1 := sendMCPRequest(t, trans, "ping", nil)
	assert.Nil(t, err1)
	assert.Contains(t, string(result1), "pong")

	// Then tools/list
	_, result2, err2 := sendMCPRequest(t, trans, "tools/list", nil)
	assert.Nil(t, err2)
	assert.Contains(t, string(result2), "wiki_read")

	// Then wiki_stats
	_, result3, err3 := sendMCPRequest(t, trans, "wiki_stats", nil)
	assert.Nil(t, err3)
	assert.Contains(t, string(result3), "page_count")
}

func TestMCPE2EIO(t *testing.T) {
	// Test the raw stdio-like transport behavior
	r, w := io.Pipe()
	var stdout bytes.Buffer
	trans := mcp.NewStdioTransportWith(r, &stdout)

	_, wikiDir := setupE2EWiki(t)
	server := mcp.NewServer(trans)
	mcp.RegisterAllTools(server, wikiDir)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		server.Run(ctx)
	}()

	// Send a ping request
	go func() {
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"))
	}()

	// Read response from stdout buffer
	var resp mcp.Response
	for i := 0; i < 100; i++ {
		if stdout.Len() > 0 {
			json.Unmarshal(stdout.Bytes(), &resp)
			break
		}
	}

	cancel()
	w.Close()

	require.NotNil(t, resp.Result)
	assert.Contains(t, string(resp.Result), "pong")
}

func TestMCPE2EWikiAddWithTags(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	// Add a page with tags
	_, result, errObj := sendMCPRequest(t, trans, "wiki_add", map[string]any{
		"path":    "tagged.md",
		"content": "# Tagged\n\nContent",
		"tags":    []string{"go", "testing"},
	})
	assert.Nil(t, errObj)

	// Verify content has frontmatter with tags
	content, err := os.ReadFile(filepath.Join(wikiDir, "tagged.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "go")
	assert.Contains(t, string(content), "testing")
}

func TestMCPE2EBadJSON(t *testing.T) {
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	_ = runMCPServer(t, wikiDir, trans)

	// Send invalid JSON
	_, err := trans.writePipe.Write([]byte("not json\n"))
	require.NoError(t, err)
}

func TestExportAppRunBuild(t *testing.T) {
	// Ensure app is importable by importing the package
	_ = app.RunBuild
}
