// Package ws provides a minimal WebSocket client using only the Go standard
// library. It implements RFC 6455 (client-side only) with support for text,
// binary, ping/pong, and close frames.
package ws

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Frame opcodes per RFC 6455.
const (
	OpText   = 1
	OpBinary = 2
	OpClose  = 8
	OpPing   = 9
	OpPong   = 10
)

// wsGUID is the magic GUID from RFC 6455 §4.2.2.
const wsGUID = "258EAFA5-E914-47DA-95CA-5AB5DC525E41"

// Conn is a minimal WebSocket client connection.
type Conn struct {
	conn    net.Conn
	br      *bufio.Reader
	writeMu sync.Mutex
	closed  bool
}

// Dial opens a WebSocket connection to the given URL (ws:// or wss://).
func Dial(ctx context.Context, rawURL string) (*Conn, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("ws: invalid URL: %w", err)
	}

	host := u.Host
	if !hasPort(host) {
		switch u.Scheme {
		case "wss":
			host += ":443"
		default:
			host += ":80"
		}
	}

	// Dial with context timeout.
	dialer := &net.Dialer{}
	var conn net.Conn
	if u.Scheme == "wss" {
		tlsDialer := &tls.Dialer{NetDialer: dialer}
		conn, err = tlsDialer.DialContext(ctx, "tcp", host)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", host)
	}
	if err != nil {
		return nil, fmt.Errorf("ws: dial: %w", err)
	}

	// Apply deadline from context if present.
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}

	// Generate random WebSocket key.
	keyBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, keyBytes); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ws: random key: %w", err)
	}
	wsKey := base64.StdEncoding.EncodeToString(keyBytes)

	path := u.RequestURI()
	if path == "" {
		path = "/"
	}

	// Write HTTP upgrade request.
	reqStr := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + u.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + wsKey + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"\r\n"

	if _, err := conn.Write([]byte(reqStr)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ws: write upgrade: %w", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ws: read upgrade response: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 101 {
		conn.Close()
		return nil, fmt.Errorf("ws: unexpected status %d", resp.StatusCode)
	}

	// Validate Sec-WebSocket-Accept.
	h := sha1.New()
	h.Write([]byte(wsKey + wsGUID))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if resp.Header.Get("Sec-Websocket-Accept") != expected {
		conn.Close()
		return nil, fmt.Errorf("ws: invalid accept key")
	}

	// Clear deadline after handshake.
	conn.SetDeadline(time.Time{})

	return &Conn{conn: conn, br: br}, nil
}

// ReadMessage reads the next WebSocket message, automatically handling
// fragmentation, ping, and close control frames.
func (c *Conn) ReadMessage() (opcode int, payload []byte, err error) {
	for {
		fin, op, data, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}

		switch op {
		case OpPing:
			c.writeFrame(OpPong, data)
			continue
		case OpPong:
			continue
		case OpClose:
			c.writeFrame(OpClose, nil)
			return OpClose, data, io.EOF
		}

		if fin {
			return op, data, nil
		}

		// Fragmented message: collect continuation frames.
		var buf []byte
		buf = append(buf, data...)
		for {
			fin2, _, data2, err := c.readFrame()
			if err != nil {
				return 0, nil, err
			}
			buf = append(buf, data2...)
			if fin2 {
				return op, buf, nil
			}
		}
	}
}

// WriteText sends a text message.
func (c *Conn) WriteText(data []byte) error {
	return c.writeFrame(OpText, data)
}

// WriteBinary sends a binary message.
func (c *Conn) WriteBinary(data []byte) error {
	return c.writeFrame(OpBinary, data)
}

// SetReadDeadline sets the read deadline on the underlying connection.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// Close sends a close frame and closes the underlying connection.
func (c *Conn) Close() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	c.writeFrameLocked(OpClose, nil)
	return c.conn.Close()
}

// readFrame reads a single WebSocket frame (server→client, unmasked).
func (c *Conn) readFrame() (fin bool, opcode int, payload []byte, err error) {
	// Byte 0: FIN(1) + RSV(3) + opcode(4)
	var header [2]byte
	if _, err := io.ReadFull(c.br, header[:]); err != nil {
		return false, 0, nil, err
	}

	fin = header[0]&0x80 != 0
	opcode = int(header[0] & 0x0F)
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)

	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(c.br, ext[:]); err != nil {
			return false, 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(c.br, ext[:]); err != nil {
			return false, 0, nil, err
		}
		length = binary.BigEndian.Uint64(ext[:])
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(c.br, maskKey[:]); err != nil {
			return false, 0, nil, err
		}
	}

	payload = make([]byte, length)
	if _, err := io.ReadFull(c.br, payload); err != nil {
		return false, 0, nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return fin, opcode, payload, nil
}

// writeFrame writes a masked client frame (thread-safe).
func (c *Conn) writeFrame(opcode int, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.writeFrameLocked(opcode, payload)
}

// writeFrameLocked writes a masked frame (caller must hold writeMu).
func (c *Conn) writeFrameLocked(opcode int, payload []byte) error {
	if c.closed {
		return fmt.Errorf("ws: connection closed")
	}

	length := len(payload)

	// Build header: FIN + opcode, MASK + length, mask key.
	var header []byte
	header = append(header, byte(0x80|opcode)) // FIN=1, opcode

	switch {
	case length <= 125:
		header = append(header, byte(0x80|length)) // MASK=1
	case length <= 65535:
		header = append(header, byte(0x80|126))
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	default:
		header = append(header, byte(0x80|127))
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}

	// Generate 4-byte mask key — TD-SDK-002: check error.
	var maskKey [4]byte
	if _, err := io.ReadFull(rand.Reader, maskKey[:]); err != nil {
		return fmt.Errorf("ws: generate mask key: %w", err)
	}
	header = append(header, maskKey[:]...)

	// Mask the payload.
	masked := make([]byte, length)
	for i := range payload {
		masked[i] = payload[i] ^ maskKey[i%4]
	}

	// Write header + masked payload atomically.
	buf := make([]byte, 0, len(header)+len(masked))
	buf = append(buf, header...)
	buf = append(buf, masked...)
	_, err := c.conn.Write(buf)
	return err
}

func hasPort(host string) bool {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return true
		}
		if host[i] == ']' {
			return false
		}
	}
	return false
}
