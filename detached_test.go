package sonzai

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestChat_AbortsOnCallerCancel proves the baseline contract: the
// existing (cancellation-honoring) Chat method aborts when its caller
// cancels the context mid-stream. This is the trap that fired the
// orchestrator wakeup incident; ChatDetached exists to avoid it.
func TestChat_AbortsOnCallerCancel(t *testing.T) {
	streamStarted := make(chan struct{})
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flush")
		}
		// Emit one delta, flush, then block until the client gives up.
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"},\"finish_reason\":null,\"index\":0}]}\n\n")
		flusher.Flush()
		close(streamStarted)
		<-r.Context().Done()
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := client.Agents.Chat(ctx, AgentChatParams{
			AgentID:     "agent-1",
			ChatOptions: ChatOptions{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		})
		errCh <- err
	}()

	<-streamStarted
	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected Chat to error after caller cancel, got nil")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Chat did not return after caller cancel within 3s")
	}
}

// TestChatDetached_SurvivesCallerCancel locks in the fix: the detached
// variant continues to completion even after the caller's context is
// cancelled, mirroring the production wakeup handler use case.
func TestChatDetached_SurvivesCallerCancel(t *testing.T) {
	var (
		streamStarted = make(chan struct{})
		releaseStream = make(chan struct{})
	)
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flush")
		}
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hel\"},\"finish_reason\":null,\"index\":0}]}\n\n")
		flusher.Flush()
		close(streamStarted)

		// Hold the stream open simulating a slow LLM until the test
		// signals release — then complete normally.
		<-releaseStream
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\",\"index\":0}]}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})
	defer server.Close()

	parent, cancel := context.WithCancel(context.Background())

	type result struct {
		resp *ChatResponse
		err  error
	}
	resCh := make(chan result, 1)

	// Route the detached-call warning through a custom logger and
	// assert it fires when the parent is cancelled mid-stream.
	var logBuf bytes.Buffer
	var logMu sync.Mutex
	logger := slog.New(slog.NewTextHandler(&syncWriter{w: &logBuf, mu: &logMu}, &slog.HandlerOptions{Level: slog.LevelWarn}))

	go func() {
		resp, err := client.Agents.ChatDetached(parent, AgentChatParams{
			AgentID:     "agent-1",
			ChatOptions: ChatOptions{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		}, DetachOptions{Timeout: 10 * time.Second, Logger: logger})
		resCh <- result{resp, err}
	}()

	<-streamStarted
	cancel() // Parent dies mid-stream — must NOT abort the call.

	// Brief pause to let the watchdog goroutine pick up the cancel and
	// emit the warning before we let the stream complete.
	time.Sleep(50 * time.Millisecond)
	close(releaseStream)

	select {
	case r := <-resCh:
		if r.err != nil {
			t.Fatalf("ChatDetached failed despite parent cancel: %v", r.err)
		}
		if r.resp == nil || r.resp.Content != "hello" {
			t.Fatalf("expected 'hello', got %+v", r.resp)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ChatDetached did not complete within 5s")
	}

	logMu.Lock()
	got := logBuf.String()
	logMu.Unlock()
	if !strings.Contains(got, "parent context cancelled during detached streaming call") {
		t.Fatalf("expected warning about parent cancel, got log output: %q", got)
	}
}

// TestChatDetached_OnParentCancelCallback verifies the callback hook is
// preferred over the slog warning when supplied.
func TestChatDetached_OnParentCancelCallback(t *testing.T) {
	var (
		streamStarted = make(chan struct{})
		releaseStream = make(chan struct{})
	)
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"x\"},\"finish_reason\":\"stop\",\"index\":0}]}\n\n")
		flusher.Flush()
		close(streamStarted)
		<-releaseStream
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})
	defer server.Close()

	parent, cancel := context.WithCancel(context.Background())
	var callbackHits atomic.Int32

	done := make(chan error, 1)
	go func() {
		_, err := client.Agents.ChatDetached(parent, AgentChatParams{
			AgentID:     "agent-1",
			ChatOptions: ChatOptions{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		}, DetachOptions{
			Timeout: 5 * time.Second,
			OnParentCancel: func(err error) {
				callbackHits.Add(1)
			},
		})
		done <- err
	}()

	<-streamStarted
	cancel()
	time.Sleep(50 * time.Millisecond)
	close(releaseStream)

	if err := <-done; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := callbackHits.Load(); got != 1 {
		t.Fatalf("expected OnParentCancel to fire exactly once, got %d", got)
	}
}

// TestChatDetached_NoWarningWhenParentLives ensures we don't spam logs
// for well-behaved callers whose context outlives the stream.
func TestChatDetached_NoWarningWhenParentLives(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		sseResponse(w,
			`{"choices":[{"delta":{"content":"Hi"},"finish_reason":"stop","index":0}]}`,
		)
	})
	defer server.Close()

	var logBuf bytes.Buffer
	var logMu sync.Mutex
	logger := slog.New(slog.NewTextHandler(&syncWriter{w: &logBuf, mu: &logMu}, &slog.HandlerOptions{Level: slog.LevelWarn}))

	_, err := client.Agents.ChatDetached(context.Background(), AgentChatParams{
		AgentID:     "agent-1",
		ChatOptions: ChatOptions{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
	}, DetachOptions{Logger: logger})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give the watchdog a chance to observe a non-cancelled parent and
	// exit cleanly.
	time.Sleep(50 * time.Millisecond)

	logMu.Lock()
	got := logBuf.String()
	logMu.Unlock()
	if got != "" {
		t.Fatalf("expected no warning when parent lives, got: %q", got)
	}
}

// syncWriter serialises writes from the slog handler goroutine and the
// test goroutine inspecting the buffer. Required because slog can be
// invoked from a watchdog goroutine concurrently with the test reading.
type syncWriter struct {
	w  *bytes.Buffer
	mu *sync.Mutex
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}
