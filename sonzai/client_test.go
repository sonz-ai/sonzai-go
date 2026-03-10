package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient("test-key", WithBaseURL(server.URL))
	return server, client
}

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func sseResponse(w http.ResponseWriter, events ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)
	for _, e := range events {
		fmt.Fprintf(w, "data: %s\n\n", e)
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
}

// ---------------------------------------------------------------------------
// Client Init
// ---------------------------------------------------------------------------

func TestNewClientPanicsWithoutKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewClient("")
}

func TestNewClientCreatesResources(t *testing.T) {
	c := NewClient("test-key")
	if c.Agents == nil {
		t.Fatal("Agents is nil")
	}
	if c.EvalTemplates == nil {
		t.Fatal("EvalTemplates is nil")
	}
	if c.EvalRuns == nil {
		t.Fatal("EvalRuns is nil")
	}
}

// ---------------------------------------------------------------------------
// Chat
// ---------------------------------------------------------------------------

func TestChatAggregated(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		sseResponse(w,
			`{"choices":[{"delta":{"content":"Hello"},"finish_reason":null,"index":0}]}`,
			`{"choices":[{"delta":{"content":" world"},"finish_reason":"stop","index":0}],"usage":{"promptTokens":10,"completionTokens":5,"totalTokens":15}}`,
		)
	})
	defer server.Close()

	resp, err := client.Agents.Chat(context.Background(), "agent-1", ChatOptions{
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello world" {
		t.Fatalf("expected 'Hello world', got '%s'", resp.Content)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 15 {
		t.Fatalf("expected 15 total tokens, got %+v", resp.Usage)
	}
}

func TestChatStream(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		sseResponse(w,
			`{"choices":[{"delta":{"content":"Hi"},"finish_reason":null,"index":0}]}`,
			`{"choices":[{"delta":{"content":"!"},"finish_reason":"stop","index":0}]}`,
		)
	})
	defer server.Close()

	var events []ChatStreamEvent
	err := client.Agents.ChatStream(context.Background(), "agent-1", ChatOptions{
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
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
	if events[0].Content() != "Hi" {
		t.Fatalf("expected 'Hi', got '%s'", events[0].Content())
	}
	if !events[1].IsFinished() {
		t.Fatal("expected last event to be finished")
	}
}

// ---------------------------------------------------------------------------
// Memory
// ---------------------------------------------------------------------------

func TestMemoryList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/memory" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, MemoryResponse{
			Nodes: []MemoryNode{{NodeID: "n1", Title: "Favorites", Importance: 0.8}},
		})
	})
	defer server.Close()

	result, err := client.Agents.Memory.List(context.Background(), "agent-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Nodes))
	}
	if result.Nodes[0].NodeID != "n1" {
		t.Fatalf("expected 'n1', got '%s'", result.Nodes[0].NodeID)
	}
}

func TestMemorySearch(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "food" {
			t.Fatalf("expected q=food, got %s", r.URL.Query().Get("q"))
		}
		jsonResponse(w, 200, MemorySearchResponse{
			Results: []MemorySearchResult{{FactID: "f1", Content: "Likes pizza", Score: 0.95}},
		})
	})
	defer server.Close()

	result, err := client.Agents.Memory.Search(context.Background(), "agent-1", MemorySearchOptions{Query: "food"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Results[0].Content != "Likes pizza" {
		t.Fatalf("expected 'Likes pizza', got '%s'", result.Results[0].Content)
	}
}

// ---------------------------------------------------------------------------
// Personality
// ---------------------------------------------------------------------------

func TestPersonalityGet(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, PersonalityResponse{
			Profile: PersonalityProfile{
				Name: "Luna",
				Big5: Big5{Openness: Big5Trait{Score: 0.8, Percentile: 85}},
			},
		})
	})
	defer server.Close()

	result, err := client.Agents.Personality.Get(context.Background(), "agent-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile.Name != "Luna" {
		t.Fatalf("expected 'Luna', got '%s'", result.Profile.Name)
	}
	if result.Profile.Big5.Openness.Score != 0.8 {
		t.Fatalf("expected 0.8, got %f", result.Profile.Big5.Openness.Score)
	}
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func TestSessionStart(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, SessionResponse{Success: true})
	})
	defer server.Close()

	result, err := client.Agents.Sessions.Start(context.Background(), "agent-1", SessionStartOptions{
		UserID: "user-1", SessionID: "sess-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestSessionEnd(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, SessionResponse{Success: true})
	})
	defer server.Close()

	result, err := client.Agents.Sessions.End(context.Background(), "agent-1", SessionEndOptions{
		UserID: "user-1", SessionID: "sess-1", TotalMessages: 10, DurationSeconds: 300,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

// ---------------------------------------------------------------------------
// Instances
// ---------------------------------------------------------------------------

func TestInstancesList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, InstanceListResponse{
			Instances: []AgentInstance{{InstanceID: "inst-1", Name: "Default", Status: "active"}},
		})
	})
	defer server.Close()

	result, err := client.Agents.Instances.List(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(result.Instances))
	}
	if result.Instances[0].Name != "Default" {
		t.Fatalf("expected 'Default', got '%s'", result.Instances[0].Name)
	}
}

func TestInstanceCreate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 201, AgentInstance{InstanceID: "inst-2", Name: "Test", Status: "active"})
	})
	defer server.Close()

	result, err := client.Agents.Instances.Create(context.Background(), "agent-1", "Test", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InstanceID != "inst-2" {
		t.Fatalf("expected 'inst-2', got '%s'", result.InstanceID)
	}
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

func TestNotificationsList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, NotificationListResponse{
			Notifications: []Notification{{MessageID: "msg-1", GeneratedMessage: "Hey there!", Status: "pending"}},
		})
	})
	defer server.Close()

	result, err := client.Agents.Notifications.List(context.Background(), "agent-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Notifications[0].GeneratedMessage != "Hey there!" {
		t.Fatalf("expected 'Hey there!', got '%s'", result.Notifications[0].GeneratedMessage)
	}
}

func TestNotificationConsume(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, SessionResponse{Success: true})
	})
	defer server.Close()

	result, err := client.Agents.Notifications.Consume(context.Background(), "agent-1", "msg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

// ---------------------------------------------------------------------------
// Eval Templates
// ---------------------------------------------------------------------------

func TestEvalTemplatesList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, EvalTemplateListResponse{
			Templates: []EvalTemplate{{ID: "tpl-1", Name: "Quality Check"}},
		})
	})
	defer server.Close()

	result, err := client.EvalTemplates.List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Templates[0].Name != "Quality Check" {
		t.Fatalf("expected 'Quality Check', got '%s'", result.Templates[0].Name)
	}
}

func TestEvalTemplateCreate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 201, EvalTemplate{ID: "tpl-2", Name: "New Template"})
	})
	defer server.Close()

	result, err := client.EvalTemplates.Create(context.Background(), EvalTemplateCreateOptions{
		Name: "New Template", ScoringRubric: "Score well",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "tpl-2" {
		t.Fatalf("expected 'tpl-2', got '%s'", result.ID)
	}
}

// ---------------------------------------------------------------------------
// Eval Runs
// ---------------------------------------------------------------------------

func TestEvalRunsList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, EvalRunListResponse{
			Runs:       []EvalRun{{ID: "run-1", Status: "completed", TotalTurns: 20}},
			TotalCount: 1,
		})
	})
	defer server.Close()

	result, err := client.EvalRuns.List(context.Background(), "", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Runs[0].Status != "completed" {
		t.Fatalf("expected 'completed', got '%s'", result.Runs[0].Status)
	}
}

// ---------------------------------------------------------------------------
// Error Handling
// ---------------------------------------------------------------------------

func TestError401(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 401, map[string]string{"error": "Invalid API key"})
	})
	defer server.Close()

	_, err := client.Agents.Memory.List(context.Background(), "agent-1", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestError404(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 404, map[string]string{"error": "Not found"})
	})
	defer server.Close()

	_, err := client.Agents.Personality.Get(context.Background(), "bad-id", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestError400(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 400, map[string]string{"error": "invalid limit"})
	})
	defer server.Close()

	_, err := client.Agents.Memory.List(context.Background(), "agent-1", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*BadRequestError); !ok {
		t.Fatalf("expected BadRequestError, got %T: %v", err, err)
	}
}
