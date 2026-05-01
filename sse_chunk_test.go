package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChunkPayload_SmallPayload(t *testing.T) {
	payload := json.RawMessage(`{"content":"hello"}`)
	frames := ChunkPayload(payload, 1024)

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}
	if string(frames[0]) != "data: {\"content\":\"hello\"}\n\n" {
		t.Fatalf("unexpected frame: %q", string(frames[0]))
	}
}

func TestChunkPayload_LargePayload(t *testing.T) {
	big := strings.Repeat("x", 1000)
	payload := json.RawMessage(fmt.Sprintf(`{"content":"%s"}`, big))
	frames := ChunkPayload(payload, 256)

	if len(frames) < 2 {
		t.Fatalf("expected multiple frames, got %d", len(frames))
	}

	for i, frame := range frames {
		s := string(frame)
		if !strings.HasPrefix(s, "data: ") || !strings.HasSuffix(s, "\n\n") {
			t.Fatalf("frame %d not in SSE format: %q", i, s)
		}
		var env sseChunkEnvelope
		if err := json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(s, "data: "), "\n\n")), &env); err != nil {
			t.Fatalf("frame %d: unmarshal error: %v", i, err)
		}
		if env.Chunk == nil {
			t.Fatalf("frame %d: missing __chunk", i)
		}
		if env.Chunk.Index != i {
			t.Fatalf("frame %d: expected index %d, got %d", i, i, env.Chunk.Index)
		}
		if env.Chunk.Total != len(frames) {
			t.Fatalf("frame %d: expected total %d, got %d", i, len(frames), env.Chunk.Total)
		}
	}
}

func TestChunkPayload_Reassemble(t *testing.T) {
	original := json.RawMessage(`{"key":"value","number":42,"nested":{"a":"b"}}`)
	frames := ChunkPayload(original, 10)

	if len(frames) < 2 {
		t.Fatalf("expected chunking, got %d frames", len(frames))
	}

	var parts []string
	for _, frame := range frames {
		s := strings.TrimSuffix(strings.TrimPrefix(string(frame), "data: "), "\n\n")
		var env sseChunkEnvelope
		if err := json.Unmarshal([]byte(s), &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		parts = append(parts, env.Data)
	}

	reassembled := strings.Join(parts, "")
	if reassembled != string(original) {
		t.Fatalf("reassembled mismatch:\n  got:  %s\n  want: %s", reassembled, string(original))
	}
}

func TestStreamSSE_ChunkedEvents(t *testing.T) {
	original := `{"choices":[{"delta":{"content":"Hello world this is a very long message that needs chunking"}}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)

		frames := ChunkPayload(json.RawMessage(original), 30)
		for _, f := range frames {
			w.Write(f)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))

	var received []json.RawMessage
	err := client.http.StreamSSE(context.Background(), "POST", "/test", nil, func(raw json.RawMessage) error {
		received = append(received, raw)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 reassembled event, got %d", len(received))
	}
	if string(received[0]) != original {
		t.Fatalf("reassembled mismatch:\n  got:  %s\n  want: %s", string(received[0]), original)
	}
}

func TestStreamSSE_MixedChunkedAndNormal(t *testing.T) {
	normalEvent := `{"choices":[{"delta":{"content":"Hi"}}]}`
	largeOriginal := `{"choices":[{"delta":{"content":"` + strings.Repeat("x", 200) + `"}}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)

		fmt.Fprintf(w, "data: %s\n\n", normalEvent)

		frames := ChunkPayload(json.RawMessage(largeOriginal), 50)
		for _, f := range frames {
			w.Write(f)
		}

		fmt.Fprintf(w, "data: %s\n\n", normalEvent)
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))

	var received []string
	err := client.http.StreamSSE(context.Background(), "POST", "/test", nil, func(raw json.RawMessage) error {
		received = append(received, string(raw))
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received) != 3 {
		t.Fatalf("expected 3 events (normal + chunked + normal), got %d", len(received))
	}
	if received[0] != normalEvent {
		t.Fatalf("event 0 mismatch: got %s", received[0])
	}
	if received[1] != largeOriginal {
		t.Fatalf("event 1 (chunked) mismatch:\n  got:  %s\n  want: %s", received[1], largeOriginal)
	}
	if received[2] != normalEvent {
		t.Fatalf("event 2 mismatch: got %s", received[2])
	}
}

func TestStreamSSE_NonChunkedPassthrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sseResponse(w,
			`{"choices":[{"delta":{"content":"Hello"},"finish_reason":null,"index":0}]}`,
			`{"choices":[{"delta":{"content":" world"},"finish_reason":"stop","index":0}]}`,
		)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))

	var events []ChatStreamEvent
	err := client.Agents.ChatStream(context.Background(), AgentChatParams{
		AgentID:     "agent-1",
		ChatOptions: ChatOptions{Messages: []ChatMessage{{Role: "user", Content: "Hi"}}},
	}, func(event ChatStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Content() != "Hello" {
		t.Fatalf("expected 'Hello', got '%s'", events[0].Content())
	}
}

func TestChunkPayload_DefaultMaxChunkSize(t *testing.T) {
	small := json.RawMessage(`{"ok":true}`)
	frames := ChunkPayload(small, 0)
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame with default size, got %d", len(frames))
	}
}
