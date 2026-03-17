package ws

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"net"
	"net/http"
	"testing"
)

// upgradeWebSocket performs the server-side WebSocket handshake.
func upgradeWebSocket(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	key := r.Header.Get("Sec-Websocket-Key")
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", 500)
		return nil, nil
	}

	conn, rw, _ := hj.Hijack()
	rw.WriteString("HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Accept: " + accept + "\r\n\r\n")
	rw.Flush()
	return conn, nil
}

func TestDialAndEcho(t *testing.T) {
	// Start a test HTTP server that upgrades to WebSocket and echoes.
	ready := make(chan string, 1)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		ready <- "ws://" + ln.Addr().String()
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgradeWebSocket(w, r)
			if err != nil || conn == nil {
				return
			}
			defer conn.Close()

			// Read one text message and echo it back as unmasked server frame.
			// Read client frame (masked).
			header := make([]byte, 2)
			conn.Read(header)
			length := int(header[1] & 0x7F)
			mask := make([]byte, 4)
			conn.Read(mask)
			payload := make([]byte, length)
			conn.Read(payload)
			for i := range payload {
				payload[i] ^= mask[i%4]
			}

			// Write back as server (unmasked).
			resp := []byte{0x81, byte(len(payload))}
			resp = append(resp, payload...)
			conn.Write(resp)

			// Send a close frame.
			conn.Write([]byte{0x88, 0})
		})
		http.Serve(ln, mux)
	}()

	wsURL := <-ready

	conn, err := Dial(context.Background(), wsURL+"/ws")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send text.
	if err := conn.WriteText([]byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read echo.
	op, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if op != OpText {
		t.Fatalf("expected text, got opcode %d", op)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
}
