package sonzai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestSessions_Turn_PostsBodyAndDecodesResult verifies the happy path:
// the SDK POSTs to /turn with the right body shape and decodes the
// mood + extraction fields the server returns.
func TestSessions_Turn_PostsBodyAndDecodesResult(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody TurnOptions

	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		jsonResponse(w, 200, TurnResult{
			Success:          true,
			ExtractionID:     "ext-123",
			ExtractionStatus: "queued",
			Mood: &TurnMood{
				Valence:     0.4,
				Arousal:     0.2,
				Tension:     -0.1,
				Affiliation: 0.3,
				Reason:      "user shared good news",
				TriggerType: "emotional_response",
			},
		})
	})
	defer server.Close()

	res, err := client.Agents.Sessions.Turn(context.Background(), "agent-1", "sess-1", TurnOptions{
		UserID:   "user-1",
		Provider: "anthropic",
		Model:    "claude-3-5-haiku",
		Messages: []TurnMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/api/v1/agents/agent-1/sessions/sess-1/turn" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody.UserID != "user-1" || gotBody.Provider != "anthropic" || gotBody.Model != "claude-3-5-haiku" {
		t.Fatalf("body did not round-trip: %+v", gotBody)
	}
	if len(gotBody.Messages) != 1 || gotBody.Messages[0].Content != "hi" {
		t.Fatalf("messages did not round-trip: %+v", gotBody.Messages)
	}
	if !res.Success || res.ExtractionID != "ext-123" || res.ExtractionStatus != "queued" {
		t.Fatalf("response did not decode: %+v", res)
	}
	if res.Mood == nil || res.Mood.Valence != 0.4 || res.Mood.Reason != "user shared good news" {
		t.Fatalf("mood did not decode: %+v", res.Mood)
	}
}

// TestSessions_TurnStatus_GetsAndDecodes covers the GET status path —
// the typical "poll until done" loop calls this with backoff.
func TestSessions_TurnStatus_GetsAndDecodes(t *testing.T) {
	var gotPath string
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, TurnStatusResult{
			ExtractionID: "ext-123",
			State:        "done",
		})
	})
	defer server.Close()

	res, err := client.Agents.Sessions.TurnStatus(context.Background(), "agent-1", "ext-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/agents/agent-1/turns/ext-123/status" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if res.State != "done" || !res.IsTerminal() {
		t.Fatalf("unexpected status: %+v", res)
	}
}

// TestSession_StartSession_ReturnsHandle verifies StartSession produces a
// *Session with the IDs and defaults pre-populated.
func TestSession_StartSession_ReturnsHandle(t *testing.T) {
	var gotStartBody SessionStartOptions
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotStartBody)
		jsonResponse(w, 200, SessionResponse{Success: true})
	})
	defer server.Close()

	sess, err := client.Agents.Sessions.StartSession(context.Background(), "agent-1", SessionStartOptions{
		UserID:    "user-1",
		SessionID: "sess-1",
		Provider:  "gemini",
		Model:     "gemini-3.1-flash-lite",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotStartBody.Provider != "gemini" || gotStartBody.Model != "gemini-3.1-flash-lite" {
		t.Fatalf("provider/model not forwarded on /start: %+v", gotStartBody)
	}
	if sess.AgentID != "agent-1" || sess.UserID != "user-1" || sess.SessionID != "sess-1" {
		t.Fatalf("Session handle not populated: %+v", sess)
	}
	if sess.Provider != "gemini" || sess.Model != "gemini-3.1-flash-lite" {
		t.Fatalf("Session defaults not stored: %+v", sess)
	}
}

// TestSession_Turn_AppliesDefaultsAndOverride covers the per-call vs
// session-default precedence: when the caller passes provider+model on
// the Turn options, they override the session defaults; otherwise the
// session defaults flow through to the server.
func TestSession_Turn_AppliesDefaultsAndOverride(t *testing.T) {
	var bodies []TurnOptions
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sessions/start") {
			jsonResponse(w, 200, SessionResponse{Success: true})
			return
		}
		var body TurnOptions
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		bodies = append(bodies, body)
		jsonResponse(w, 200, TurnResult{
			Success:          true,
			ExtractionID:     "ext",
			ExtractionStatus: "queued",
		})
	})
	defer server.Close()

	sess, err := client.Agents.Sessions.StartSession(context.Background(), "agent-1", SessionStartOptions{
		UserID:    "user-1",
		SessionID: "sess-1",
		Provider:  "gemini",
		Model:     "gemini-3.1-flash-lite",
	})
	if err != nil {
		t.Fatalf("StartSession err: %v", err)
	}

	// First Turn — no override, should inherit session defaults.
	if _, err := sess.Turn(context.Background(), TurnOptions{
		Messages: []TurnMessage{{Role: "user", Content: "hello"}},
	}); err != nil {
		t.Fatalf("Turn 1 err: %v", err)
	}
	// Second Turn — override provider+model.
	if _, err := sess.Turn(context.Background(), TurnOptions{
		Provider: "anthropic",
		Model:    "claude-3-5-haiku",
		Messages: []TurnMessage{{Role: "user", Content: "switch model"}},
	}); err != nil {
		t.Fatalf("Turn 2 err: %v", err)
	}

	if len(bodies) != 2 {
		t.Fatalf("expected 2 turn bodies, got %d", len(bodies))
	}
	if bodies[0].Provider != "gemini" || bodies[0].Model != "gemini-3.1-flash-lite" {
		t.Fatalf("turn 1 should have inherited session defaults: %+v", bodies[0])
	}
	if bodies[0].UserID != "user-1" {
		t.Fatalf("turn 1 should have inherited UserID: %+v", bodies[0])
	}
	if bodies[1].Provider != "anthropic" || bodies[1].Model != "claude-3-5-haiku" {
		t.Fatalf("turn 2 should have used per-call override: %+v", bodies[1])
	}
}

// TestSession_End_FillsIDsFromHandle verifies End uses the session's
// agent/user/session IDs without requiring the caller to repeat them.
func TestSession_End_FillsIDsFromHandle(t *testing.T) {
	var endHits int
	var gotEndBody SessionEndOptions
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sessions/start") {
			jsonResponse(w, 200, SessionResponse{Success: true})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/sessions/end") {
			endHits++
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &gotEndBody)
			jsonResponse(w, 200, SessionResponse{Success: true})
			return
		}
		http.Error(w, "unexpected path "+r.URL.Path, http.StatusNotFound)
	})
	defer server.Close()

	sess, err := client.Agents.Sessions.StartSession(context.Background(), "agent-1", SessionStartOptions{
		UserID:    "user-1",
		SessionID: "sess-1",
	})
	if err != nil {
		t.Fatalf("StartSession err: %v", err)
	}
	if err := sess.End(context.Background()); err != nil {
		t.Fatalf("End err: %v", err)
	}
	if endHits != 1 {
		t.Fatalf("expected 1 end call, got %d", endHits)
	}
	if gotEndBody.UserID != "user-1" || gotEndBody.SessionID != "sess-1" {
		t.Fatalf("End did not pre-fill IDs from Session: %+v", gotEndBody)
	}
}
