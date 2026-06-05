package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTransport is an in-memory transport for testing.
type testTransport struct {
	readData  []byte
	readErr   error
	writeBuf  bytes.Buffer
	closed    bool
	readCh    chan struct{}
}

func newTestTransport(input string) *testTransport {
	return &testTransport{
		readData: []byte(input),
		readCh:   make(chan struct{}, 1),
	}
}

func (t *testTransport) Read() ([]byte, error) {
	if t.readErr != nil {
		return nil, t.readErr
	}
	// Return data once, then EOF
	data := t.readData
	t.readData = nil
	t.readErr = io.EOF
	return data, nil
}

func (t *testTransport) Write(data []byte) error {
	t.writeBuf.Write(data)
	t.writeBuf.WriteString("\n")
	return nil
}

func (t *testTransport) Close() error {
	t.closed = true
	return nil
}

func TestPing(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF) // transport EOF is expected after one read

	assert.Contains(t, trans.writeBuf.String(), `"result":"pong"`)
	assert.Contains(t, trans.writeBuf.String(), `"id":1`)
}

func TestToolList(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	// Register a test tool
	srv.RegisterTool(MCPToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: InputSchema{Type: "object"},
	}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		return "ok", nil
	})

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, `"tools"`)
	assert.Contains(t, output, "test_tool")
}

func TestUnknownMethod(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"unknown_method"}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, `"code":-32601`)
	assert.Contains(t, output, "Method not found")
}

func TestInvalidJSONRPCVersion(t *testing.T) {
	req := `{"jsonrpc":"1.0","id":1,"method":"ping"}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, `"code":-32600`)
}

func TestToolCall(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"echo","params":{"msg":"hello"}}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	srv.RegisterTool(MCPToolDefinition{
		Name:        "echo",
		Description: "echoes input",
		InputSchema: InputSchema{Type: "object"},
	}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		var input struct {
			Msg string `json:"msg"`
		}
		if err := json.Unmarshal(params, &input); err != nil {
			return nil, &ErrorObj{Code: ErrInvalidParams, Message: err.Error()}
		}
		return map[string]string{"echo": input.Msg}, nil
	})

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, "hello")
	assert.NotContains(t, output, "error")
}

func TestToolError(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"failing_tool"}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	srv.RegisterTool(MCPToolDefinition{
		Name: "failing_tool", Description: "always fails",
		InputSchema: InputSchema{Type: "object"},
	}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		return nil, &ErrorObj{Code: ErrInternal, Message: "something went wrong"}
	})

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, `"code":-32603`)
	assert.Contains(t, output, "something went wrong")
}

func TestGracefulShutdown(t *testing.T) {
	// Create a transport that blocks on Read
	r, w := io.Pipe()
	trans := NewStdioTransportWith(r, io.Discard)
	srv := NewServer(trans)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run the server; it should exit when context is cancelled
	err := srv.Run(ctx)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second, "shutdown should be quick")
	w.Close()
}

func TestNotificationNoID(t *testing.T) {
	// Notifications (requests without id) should not produce responses
	req := `{"jsonrpc":"2.0","method":"ping"}`
	trans := newTestTransport(req)
	srv := NewServer(trans)

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	// No response should be written for notifications
	assert.Empty(t, trans.writeBuf.String())
}

func TestMultipleRequests(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
	}
	input := strings.Join(requests, "\n") + "\n"

	trans := &multiLineTransport{lines: strings.Split(input, "\n")}
	srv := NewServer(trans)

	srv.RegisterTool(MCPToolDefinition{
		Name: "echo", Description: "echo",
		InputSchema: InputSchema{Type: "object"},
	}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) {
		return "echo", nil
	})

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, `"id":1`)
	assert.Contains(t, output, `"id":2`)
	// Should have 2 responses
	assert.Equal(t, 2, strings.Count(output, `"jsonrpc":"2.0"`))
}

// multiLineTransport returns lines one at a time, then EOF.
type multiLineTransport struct {
	lines    []string
	index    int
	writeBuf bytes.Buffer
}

func (m *multiLineTransport) Read() ([]byte, error) {
	if m.index >= len(m.lines) {
		return nil, io.EOF
	}
	line := m.lines[m.index]
	m.index++
	return []byte(line), nil
}

func (m *multiLineTransport) Write(data []byte) error {
	m.writeBuf.Write(data)
	m.writeBuf.WriteString("\n")
	return nil
}

func (m *multiLineTransport) Close() error { return nil }

func TestBadJSON(t *testing.T) {
	trans := newTestTransport(`not json`)
	srv := NewServer(trans)

	err := srv.Run(context.Background())
	require.ErrorIs(t, err, io.EOF)

	output := trans.writeBuf.String()
	assert.Contains(t, output, `"code":-32700`) // parse error
}

func TestToolListOrder(t *testing.T) {
	// Register tools in reverse order to test sorting
	srv := NewServer(newTestTransport(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	srv.RegisterTool(MCPToolDefinition{Name: "z_tool"}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) { return nil, nil })
	srv.RegisterTool(MCPToolDefinition{Name: "a_tool"}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) { return nil, nil })

	_ = srv.Run(context.Background())

	output := srv.transport.(*testTransport).writeBuf.String()
	// a_tool should appear before z_tool
	aPos := strings.Index(output, "a_tool")
	zPos := strings.Index(output, "z_tool")
	assert.Less(t, aPos, zPos, "tools should be sorted alphabetically")
}

func TestServerRegisterTool(t *testing.T) {
	srv := NewServer(newTestTransport(""))
	srv.RegisterTool(MCPToolDefinition{Name: "test"}, func(ctx context.Context, params json.RawMessage) (any, *ErrorObj) { return "ok", nil })
	assert.Len(t, srv.tools, 1)
}
