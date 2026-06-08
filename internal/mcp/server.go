package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ToolHandler processes an MCP tool call and returns a result or error.
type ToolHandler func(ctx context.Context, params json.RawMessage) (any, *ErrorObj)

// toolEntry stores a registered tool's handler and definition.
type toolEntry struct {
	def     MCPToolDefinition
	handler ToolHandler
}

// ResourceDefinition describes an MCP resource (wiki page).
type ResourceDefinition struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// PromptDefinition describes an MCP prompt template.
type PromptDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Arguments   []PromptArgument   `json:"arguments,omitempty"`
}

// PromptArgument describes a prompt argument.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Server is an MCP JSON-RPC 2.0 server.
// It reads requests from a Transport, dispatches them to registered tools,
// and writes responses back.
type Server struct {
	transport Transport
	tools     map[string]toolEntry
	mu        sync.RWMutex
}

// NewServer creates an MCP server with the given transport and registers
// built-in methods (ping, tools/list).
func NewServer(transport Transport) *Server {
	s := &Server{
		transport: transport,
		tools:     make(map[string]toolEntry),
	}
	return s
}

// RegisterTool registers an MCP tool with its definition and handler.
func (s *Server) RegisterTool(def MCPToolDefinition, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[def.Name] = toolEntry{def: def, handler: handler}
}

// Run starts the main MCP server loop. It reads requests from the
// transport, dispatches them, and returns responses. The loop exits
// when the transport returns an error (e.g. EOF) or the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := s.transport.Read()
		if err != nil {
			// If the context was cancelled while blocked on Read, prefer the context error
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("transport read: %w", err)
		}

		// Skip empty lines
		if len(bytes.TrimSpace(msg)) == 0 {
			continue
		}

		resp := s.handleMessage(ctx, msg)
		if resp == nil {
			continue // notification (no id)
		}

		data, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal response: %w", err)
		}

		if err := s.transport.Write(data); err != nil {
			return fmt.Errorf("transport write: %w", err)
		}
	}
}

// handleMessage parses a raw JSON message and dispatches it.
// Returns nil for notifications (requests without an id).
func (s *Server) handleMessage(ctx context.Context, msg []byte) *Response {
	var raw struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return NewParseError(extractID(msg))
	}

	// JSON-RPC 2.0 notification: no "id" member (field absent)
	if raw.ID == nil {
		return nil
	}

	if raw.JSONRPC != Version {
		return NewErrorResponse(raw.ID, ErrInvalidReq, "invalid jsonrpc version")
	}

	// Handle built-in methods
	switch raw.Method {
	case "ping":
		return NewResponse(raw.ID, "pong")
	case "tools/list":
		return s.handleToolList(raw.ID)
	case "resources/list":
		return s.handleResourceList(raw.ID, raw.Params)
	case "resources/read":
		return s.handleResourceRead(raw.ID, raw.Params)
	case "prompts/list":
		return s.handlePromptList(raw.ID)
	case "prompts/get":
		return s.handlePromptGet(raw.ID, raw.Params)
	}

	// Look up registered tool
	s.mu.RLock()
	entry, ok := s.tools[raw.Method]
	s.mu.RUnlock()

	if !ok {
		return NewMethodNotFoundError(raw.ID, raw.Method)
	}

	result, errObj := entry.handler(ctx, raw.Params)
	if errObj != nil {
		errResp := &Response{
			JSONRPC: Version,
			ID:      raw.ID,
			Error:   errObj,
		}
		return errResp
	}

	return NewResponse(raw.ID, result)
}

// handleToolList returns the list of all registered tools.
func (s *Server) handleToolList(id json.RawMessage) *Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	defs := make([]MCPToolDefinition, 0, len(s.tools))
	for _, entry := range s.tools {
		defs = append(defs, entry.def)
	}

	// Sort by name for deterministic output
	for i := 0; i < len(defs); i++ {
		for j := i + 1; j < len(defs); j++ {
			if defs[i].Name > defs[j].Name {
				defs[i], defs[j] = defs[j], defs[i]
			}
		}
	}

	return NewResponse(id, map[string]any{"tools": defs})
}

// handleResourceList returns available wiki pages as MCP resources.
func (s *Server) handleResourceList(id json.RawMessage, params json.RawMessage) *Response {
	return NewResponse(id, map[string]any{
		"resources": []ResourceDefinition{
			{
				URI:         "wiki:///",
				Name:        "Wiki Root",
				Description: "Root Wiki directory listing",
				MimeType:    "text/markdown",
			},
		},
	})
}

// handleResourceRead reads a wiki page resource.
func (s *Server) handleResourceRead(id json.RawMessage, params json.RawMessage) *Response {
	return NewResponse(id, map[string]any{
		"contents": []ResourceDefinition{
			{
				URI:      "wiki:///",
				Name:     "Wiki Root",
				MimeType: "text/markdown",
			},
		},
	})
}

// handlePromptList returns available prompt templates.
func (s *Server) handlePromptList(id json.RawMessage) *Response {
	return NewResponse(id, map[string]any{
		"prompts": []PromptDefinition{
			{
				Name:        "summarize-page",
				Description: "Generate a summary of a wiki page",
				Arguments: []PromptArgument{
					{Name: "path", Description: "Page path", Required: true},
				},
			},
			{
				Name:        "explain-architecture",
				Description: "Explain the architecture of a section",
				Arguments: []PromptArgument{
					{Name: "section", Description: "Section path", Required: true},
				},
			},
		},
	})
}

// handlePromptGet returns a specific prompt template.
func (s *Server) handlePromptGet(id json.RawMessage, params json.RawMessage) *Response {
	var p struct {
		Name string `json:"name"`
		Args map[string]string `json:"arguments"`
	}
	if params != nil {
		_ = json.Unmarshal(params, &p)
	}

	messages := []map[string]any{}

	switch p.Name {
	case "summarize-page":
		content := "Please summarize the wiki page."
		if path, ok := p.Args["path"]; ok {
			content = "Please summarize the wiki page at: " + path
		}
		messages = append(messages, map[string]any{
			"role":    "user",
			"content": content,
		})
	case "explain-architecture":
		content := "Please explain the architecture."
		if section, ok := p.Args["section"]; ok {
			content = "Please explain the architecture of: " + section
		}
		messages = append(messages, map[string]any{
			"role":    "user",
			"content": content,
		})
	}

	return NewResponse(id, map[string]any{
		"description": "Prompt template for " + p.Name,
		"messages":    messages,
	})
}

// extractID attempts to extract the "id" field from raw JSON for error responses.
func extractID(msg []byte) json.RawMessage {
	var withID struct {
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(msg, &withID); err == nil && len(withID.ID) > 0 {
		return withID.ID
	}
	return json.RawMessage(`null`)
}

