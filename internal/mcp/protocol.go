// Package mcp implements a Model Context Protocol server.
//
// The protocol follows MCP 2025-03-26 using JSON-RPC 2.0 message format.
package mcp

import "encoding/json"

// JSON-RPC 2.0 constants.
const (
	Version        = "2.0"
	ErrParse       = -32700
	ErrInvalidReq  = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams = -32602
	ErrInternal    = -32603
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorObj       `json:"error,omitempty"`
}

// ErrorObj represents a JSON-RPC 2.0 error object.
type ErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *ErrorObj) Error() string {
	return e.Message
}

// NewRequest creates a new JSON-RPC 2.0 request.
func NewRequest(id json.RawMessage, method string, params any) *Request {
	var rawParams json.RawMessage
	if params != nil {
		rawParams, _ = json.Marshal(params)
	}
	return &Request{
		JSONRPC: Version,
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}
}

// NewResponse creates a success response.
func NewResponse(id json.RawMessage, result any) *Response {
	var rawResult json.RawMessage
	if result != nil {
		rawResult, _ = json.Marshal(result)
	}
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Result:  rawResult,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(id json.RawMessage, code int, message string) *Response {
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Error: &ErrorObj{
			Code:    code,
			Message: message,
		},
	}
}

// NewParseError creates a parse error response (for invalid JSON).
func NewParseError(id json.RawMessage) *Response {
	return NewErrorResponse(id, ErrParse, "Parse error")
}

// NewMethodNotFoundError creates a method not found error.
func NewMethodNotFoundError(id json.RawMessage, method string) *Response {
	return NewErrorResponse(id, ErrMethodNotFound, "Method not found: "+method)
}

// MCPToolDefinition describes an MCP tool for the tools/list response.
type MCPToolDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is a JSON Schema describing tool parameters.
type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema describes a single tool parameter.
type PropertySchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Items       *struct {
		Type string `json:"type"`
	} `json:"items,omitempty"`
}

// MCPToolResult wraps tool output in the format expected by MCP.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent is a single piece of content in an MCP tool result.
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewMCPToolResult creates a success tool result.
func NewMCPToolResult(text string) MCPToolResult {
	return MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: text}},
	}
}

// NewMCPErrorResult creates an error tool result.
func NewMCPErrorResult(text string) MCPToolResult {
	return MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: text}},
		IsError: true,
	}
}
