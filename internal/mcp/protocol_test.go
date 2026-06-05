package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestMarshal(t *testing.T) {
	rawID := json.RawMessage(`1`)
	req := NewRequest(rawID, "ping", nil)

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed Request
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, Version, parsed.JSONRPC)
	assert.Equal(t, "ping", parsed.Method)
	assert.JSONEq(t, `1`, string(parsed.ID))
}

func TestRequestWithParams(t *testing.T) {
	rawID := json.RawMessage(`2`)
	params := map[string]any{"path": "test.md"}
	req := NewRequest(rawID, "wiki_read", params)

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed Request
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "wiki_read", parsed.Method)
	assert.Contains(t, string(parsed.Params), "test.md")
}

func TestResponseMarshal(t *testing.T) {
	rawID := json.RawMessage(`1`)
	resp := NewResponse(rawID, map[string]string{"status": "ok"})

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed Response
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, Version, parsed.JSONRPC)
	assert.Nil(t, parsed.Error)
	assert.Contains(t, string(parsed.Result), "ok")
}

func TestErrorResponse(t *testing.T) {
	rawID := json.RawMessage(`1`)
	resp := NewErrorResponse(rawID, ErrMethodNotFound, "not found")

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed Response
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	require.NotNil(t, parsed.Error)
	assert.Equal(t, ErrMethodNotFound, parsed.Error.Code)
	assert.Contains(t, parsed.Error.Message, "not found")
	assert.Nil(t, parsed.Result)
}

func TestNewParseError(t *testing.T) {
	rawID := json.RawMessage(nil) // null id for parse errors
	resp := NewParseError(rawID)

	assert.Equal(t, ErrParse, resp.Error.Code)
}

func TestNewMethodNotFoundError(t *testing.T) {
	rawID := json.RawMessage(`1`)
	resp := NewMethodNotFoundError(rawID, "unknown_method")

	assert.Equal(t, ErrMethodNotFound, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "unknown_method")
}

func TestMCPToolResult(t *testing.T) {
	result := NewMCPToolResult("hello")
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "hello", result.Content[0].Text)

	errResult := NewMCPErrorResult("error msg")
	assert.True(t, errResult.IsError)
	assert.Equal(t, "error msg", errResult.Content[0].Text)
}

func TestRequestUnmarshalInvalid(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"not json", "not json"},
		{"missing jsonrpc", `{"id":1,"method":"ping"}`},
		{"wrong version", `{"jsonrpc":"1.0","id":1,"method":"ping"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req Request
			err := json.Unmarshal([]byte(tt.data), &req)
			if err == nil {
				// If it parsed, the version should at least be correct
				assert.Equal(t, Version, req.JSONRPC)
			}
		})
	}
}

func TestMCPToolDefinition(t *testing.T) {
	def := MCPToolDefinition{
		Name:        "wiki_read",
		Description: "Read a wiki page",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"path": {
					Type:        "string",
					Description: "Page path",
				},
			},
			Required: []string{"path"},
		},
	}

	data, err := json.Marshal(def)
	require.NoError(t, err)

	var parsed MCPToolDefinition
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "wiki_read", parsed.Name)
	assert.Equal(t, "object", parsed.InputSchema.Type)
	assert.Contains(t, parsed.InputSchema.Required, "path")
}
