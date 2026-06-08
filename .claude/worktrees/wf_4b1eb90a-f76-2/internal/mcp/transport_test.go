package mcp

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioTransportReadWrite(t *testing.T) {
	reader, writer := io.Pipe()
	var buf bytes.Buffer
	outWriter := io.MultiWriter(writer, &buf)

	trans := NewStdioTransportWith(reader, outWriter)

	// Write test data to stdin
	go func() {
		writer.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"))
		writer.Close()
	}()

	msg, err := trans.Read()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "ping")
}

func TestNewStdioTransport(t *testing.T) {
	trans := NewStdioTransport()
	assert.NotNil(t, trans)
	assert.NotNil(t, trans.reader)
}

func TestStdioTransportWrite(t *testing.T) {
	var buf bytes.Buffer
	trans := NewStdioTransportWith(bytes.NewReader(nil), &buf)

	err := trans.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"ok"}`))
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ok")
}

func TestStdioTransportClose(t *testing.T) {
	trans := NewStdioTransport()
	err := trans.Close()
	assert.NoError(t, err)
}

// pipedTransport provides an in-memory Transport implementation for testing.
type pipedTransport struct {
	reader *io.PipeReader
	writer *io.PipeWriter
	buf    bytes.Buffer
}

func newPipedTransport() *pipedTransport {
	r, w := io.Pipe()
	return &pipedTransport{reader: r, writer: w}
}

func (p *pipedTransport) Read() ([]byte, error) {
	return readLine(p.reader)
}

func (p *pipedTransport) Write(data []byte) error {
	_, err := p.buf.Write(data)
	return err
}

func (p *pipedTransport) Close() error {
	p.writer.Close()
	return p.reader.Close()
}

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

func TestTransportRoundTrip(t *testing.T) {
	p := newPipedTransport()
	defer p.Close()

	// Simulate sending a request
	go func() {
		p.writer.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"))
		p.writer.Close()
	}()

	msg, err := p.Read()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "ping")

	// Simulate writing a response
	err = p.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"pong"}`))
	require.NoError(t, err)
	assert.Contains(t, p.buf.String(), "pong")
}

func TestTransportEOF(t *testing.T) {
	p := newPipedTransport()
	p.writer.Close()

	_, err := p.Read()
	assert.ErrorIs(t, err, io.EOF)
}
