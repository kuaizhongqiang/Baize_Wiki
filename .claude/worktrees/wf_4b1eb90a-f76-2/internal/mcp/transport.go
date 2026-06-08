package mcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

// Transport defines the MCP transport interface.
// Implementations handle reading JSON-RPC messages from a source and
// writing responses back.
type Transport interface {
	// Read returns a single JSON-RPC message (a complete JSON object).
	Read() ([]byte, error)
	// Write sends a JSON-RPC response.
	Write([]byte) error
	// Close cleans up transport resources.
	Close() error
}

// StdioTransport reads from stdin and writes to stdout.
// Log output goes to stderr to avoid corrupting the data stream.
type StdioTransport struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewStdioTransport creates a transport that uses stdin/stdout.
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// NewStdioTransportWith creates a transport with custom reader/writer for testing.
func NewStdioTransportWith(r io.Reader, w io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

// Read reads a line-delimited JSON message from stdin.
func (t *StdioTransport) Read() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line, nil
}

// Write writes a JSON message to stdout with a newline delimiter.
func (t *StdioTransport) Write(data []byte) error {
	_, err := fmt.Fprintln(t.writer, string(data))
	return err
}

// Close is a no-op for stdio transport; stdin/stdout are managed by the process.
func (t *StdioTransport) Close() error {
	return nil
}

// TCPTransport reads from and writes to a TCP connection.
type TCPTransport struct {
	conn   net.Conn
	reader *bufio.Reader
}

// NewTCPTransport creates a transport from an established TCP connection.
func NewTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// Read reads a line-delimited JSON message from the TCP connection.
func (t *TCPTransport) Read() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line, nil
}

// Write writes a JSON message to the TCP connection with a newline delimiter.
func (t *TCPTransport) Write(data []byte) error {
	_, err := fmt.Fprintln(t.conn, string(data))
	return err
}

// Close closes the underlying TCP connection.
func (t *TCPTransport) Close() error {
	return t.conn.Close()
}

// TCPServerTransport listens on a TCP address and accepts a single connection.
// This is a convenience wrapper for the server-side listener pattern.
type TCPServerTransport struct {
	listener net.Listener
	transport *TCPTransport
}

// NewTCPServerTransport creates a TCP server transport and waits for one connection.
func NewTCPServerTransport(addr string) (*TCPServerTransport, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}
	return &TCPServerTransport{listener: listener}, nil
}

// Accept waits for a single client connection.
func (t *TCPServerTransport) Accept() error {
	conn, err := t.listener.Accept()
	if err != nil {
		return err
	}
	t.transport = NewTCPTransport(conn)
	return nil
}

// Read delegates to the underlying TCP transport.
func (t *TCPServerTransport) Read() ([]byte, error) {
	if t.transport == nil {
		return nil, io.ErrClosedPipe
	}
	return t.transport.Read()
}

// Write delegates to the underlying TCP transport.
func (t *TCPServerTransport) Write(data []byte) error {
	if t.transport == nil {
		return io.ErrClosedPipe
	}
	return t.transport.Write(data)
}

// Close closes both the listener and the connection.
func (t *TCPServerTransport) Close() error {
	if t.transport != nil {
		_ = t.transport.Close()
	}
	return t.listener.Close()
}
