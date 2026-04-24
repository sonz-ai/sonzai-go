package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestSessionEnd_LegacyPassthrough guarantees the pre-async 200 path is
// unchanged: no processing_id, no polling, one request.
func TestSessionEnd_LegacyPassthrough(t *testing.T) {
	var hits int
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		hits++
		jsonResponse(w, 200, SessionResponse{Success: true})
	})
	defer server.Close()

	result, err := client.Agents.Sessions.End(context.Background(), "agent-1", SessionEndOptions{
		UserID: "user-1", SessionID: "sess-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if hits != 1 {
		t.Fatalf("expected 1 request, got %d", hits)
	}
}

// TestSessionEnd_AsyncPollsUntilDone covers the 202 → poll loop path.
// First /status hit returns "processing", second returns "done"; End must
// return only after observing "done" and the caller sees Success=true.
func TestSessionEnd_AsyncPollsUntilDone(t *testing.T) {
	var statusHits int
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/sessions/end"):
			jsonResponse(w, 202, asyncSessionEndAccept{
				Success:      true,
				Async:        true,
				ProcessingID: "11111111-1111-1111-1111-111111111111",
				StatusURL:    "/api/v1/sessions/end/status/11111111-1111-1111-1111-111111111111",
				SessionID:    "sess-1",
				AgentID:      "agent-1",
				EnqueuedAt:   "2026-04-24T10:00:00Z",
			})
		case strings.Contains(r.URL.Path, "/sessions/end/status/"):
			statusHits++
			if statusHits == 1 {
				jsonResponse(w, 200, sessionEndStatus{
					State:      "processing",
					EnqueuedAt: "2026-04-24T10:00:00Z",
					SessionID:  "sess-1",
					AgentID:    "agent-1",
				})
				return
			}
			jsonResponse(w, 200, sessionEndStatus{
				State:      "done",
				EnqueuedAt: "2026-04-24T10:00:00Z",
				StartedAt:  "2026-04-24T10:00:01Z",
				FinishedAt: "2026-04-24T10:00:30Z",
				SessionID:  "sess-1",
				AgentID:    "agent-1",
			})
		default:
			http.Error(w, "unexpected path "+r.URL.Path, http.StatusNotFound)
		}
	})
	defer server.Close()

	result, err := client.Agents.Sessions.End(context.Background(), "agent-1", SessionEndOptions{
		UserID: "user-1", SessionID: "sess-1", TotalMessages: 2, DurationSeconds: 30,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success after polling")
	}
	if statusHits != 2 {
		t.Fatalf("expected 2 status hits, got %d", statusHits)
	}
}

// TestSessionEnd_AsyncFailedBubblesError covers the terminal-failure
// case: state=failed must surface as an error with the reason.
func TestSessionEnd_AsyncFailedBubblesError(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/sessions/end"):
			jsonResponse(w, 202, asyncSessionEndAccept{
				Success:      true,
				Async:        true,
				ProcessingID: "22222222-2222-2222-2222-222222222222",
				StatusURL:    "/api/v1/sessions/end/status/22222222-2222-2222-2222-222222222222",
				SessionID:    "sess-1",
				AgentID:      "agent-1",
			})
		default:
			jsonResponse(w, 200, sessionEndStatus{
				State:      "failed",
				EnqueuedAt: "2026-04-24T10:00:00Z",
				SessionID:  "sess-1",
				AgentID:    "agent-1",
				Error:      "LLM upstream timeout",
			})
		}
	})
	defer server.Close()

	_, err := client.Agents.Sessions.End(context.Background(), "agent-1", SessionEndOptions{
		UserID: "user-1", SessionID: "sess-1",
	})
	if err == nil {
		t.Fatal("expected error on failed state")
	}
	if !strings.Contains(err.Error(), "LLM upstream timeout") {
		t.Fatalf("expected error to mention LLM upstream timeout, got: %v", err)
	}
}

// TestSessionEnd_AsyncTimeoutBounded verifies PollTimeout is respected so
// a wedged server can't hang the caller forever.
func TestSessionEnd_AsyncTimeoutBounded(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/sessions/end"):
			jsonResponse(w, 202, asyncSessionEndAccept{
				Success:      true,
				Async:        true,
				ProcessingID: "33333333-3333-3333-3333-333333333333",
				SessionID:    "sess-1",
				AgentID:      "agent-1",
			})
		default:
			// Always return pending — simulates a stuck worker.
			jsonResponse(w, 200, sessionEndStatus{State: "pending", SessionID: "sess-1", AgentID: "agent-1"})
		}
	})
	defer server.Close()

	start := time.Now()
	_, err := client.Agents.Sessions.End(context.Background(), "agent-1", SessionEndOptions{
		UserID: "user-1", SessionID: "sess-1", PollTimeout: 300 * time.Millisecond,
	})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected 'timed out' error, got: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("PollTimeout=300ms but End took %s", elapsed)
	}
}

// jsonResponse already exists in client_test.go; adding a small helper
// for readability in case we need to inspect request bodies later.
func dumpJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return fmt.Sprintf("%s", b)
}
