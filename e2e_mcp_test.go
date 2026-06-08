package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/app"
	"github.com/kuaizhongqiang/baize-wiki/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// e2eMockBuild is a stub build function for E2E MCP tests.
var e2eMockBuild mcp.RunBuildFunc = func(ctx context.Context, source, output, configPath string, level int, catalogLevel int, draft, quiet, scanAll bool) (bool, int64, int, int, []string) {
	return true, 100, 1, 1, nil
}

// e2eMCPTransport implements mcp.Transport using two separate pipes:
// one for requests (client→server) and one for responses (server→client).
// This prevents the server from reading its own response.
type e2eMCPTransport struct {
	serverReadPipe  *io.PipeReader // server reads requests from here
	clientWritePipe *io.PipeWriter // client writes requests here
	serverWritePipe *io.PipeWriter // server writes responses here
	clientReadPipe  *io.PipeReader // client reads responses from here
}

func newE2eMCPTransport() *e2eMCPTransport {
	reqR, reqW := io.Pipe()   // request pipe: client→server
	respR, respW := io.Pipe() // response pipe: server→client
	return &e2eMCPTransport{
		serverReadPipe:  reqR,
		clientWritePipe: reqW,
		serverWritePipe: respW,
		clientReadPipe:  respR,
	}
}

// Read is called by the server to receive a request from the client.
func (t *e2eMCPTransport) Read() ([]byte, error) {
	return readLine(t.serverReadPipe)
}

// Write is called by the server to send a response to the client.
func (t *e2eMCPTransport) Write(data []byte) error {
	_, err := t.serverWritePipe.Write(data)
	if err == nil {
		t.serverWritePipe.Write([]byte("\n"))
	}
	return err
}

// ClientWrite sends a request to the server (used by test helpers).
func (t *e2eMCPTransport) ClientWrite(data []byte) error {
	_, err := t.clientWritePipe.Write(append(data, '\n'))
	return err
}

// ClientRead reads a response from the server (used by test helpers).
func (t *e2eMCPTransport) ClientRead() ([]byte, error) {
	return readLine(t.clientReadPipe)
}

func (t *e2eMCPTransport) Close() error {
	t.clientWritePipe.Close()
	t.serverReadPipe.Close()
	t.serverWritePipe.Close()
	return t.clientReadPipe.Close()
}

// readLine reads a single newline-delimited line from a reader.
func readLine(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	tmp := make([]byte, 1)
	for {
		n, err := r.Read(tmp)
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

// setupE2EWiki creates a temporary Wiki with test data for E2E MCP testing.
func setupE2EWiki(t *testing.T) (string, string) {
	t.Helper()

	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "doc.md"), []byte("---\ntitle: Test Doc\n---\n\n# Test Doc\n\nHello world"), 0644))

	outDir := t.TempDir()

	// Build the wiki at Level 2 so file paths are preserved
	result := app.RunBuild(context.Background(), srcDir, outDir, "", 2, false, true, false)
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

	// Write request via the client→server pipe
	err = trans.ClientWrite(data)
	require.NoError(t, err)

	// Read response from the server→client pipe
	respData, err := trans.ClientRead()
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
	mcp.RegisterAllTools(server, wikiDir, e2eMockBuild)

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
		"path": "doc.md",
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
	assert.Contains(t, string(result), "doc.md")
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
		"path":    "doc.md",
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
	_, wikiDir := setupE2EWiki(t)
	trans := newE2eMCPTransport()
	defer trans.Close()

	cancel := runMCPServer(t, wikiDir, trans)
	defer cancel()

	_, result, errObj := sendMCPRequest(t, trans, "wiki_read", map[string]string{
		"path": "doc.md",
	})
	assert.Nil(t, errObj)
	assert.Contains(t, string(result), "Hello world")
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
	// Test the raw stdio-like transport behavior using two pipes
	reqR, reqW := io.Pipe()   // client writes requests, server reads
	respR, respW := io.Pipe() // server writes responses, client reads
	trans := mcp.NewStdioTransportWith(reqR, respW)

	_, wikiDir := setupE2EWiki(t)
	server := mcp.NewServer(trans)
	mcp.RegisterAllTools(server, wikiDir, e2eMockBuild)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		server.Run(ctx)
	}()

	// Send a ping request via the request pipe
	_, err := reqW.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"))
	require.NoError(t, err)

	// Read response from the response pipe
	respData, err := readLine(respR)
	require.NoError(t, err)

	var resp mcp.Response
	err = json.Unmarshal(respData, &resp)
	require.NoError(t, err)

	cancel()
	reqW.Close()
	respW.Close()

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
	_, _, errObj := sendMCPRequest(t, trans, "wiki_add", map[string]any{
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
	err := trans.ClientWrite([]byte("not json"))
	require.NoError(t, err)
}
