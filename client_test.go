package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := MustNewClient("test-key", WithBaseURL(server.URL))
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

func TestNewClientErrorWithoutKey(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestMustNewClientPanicsWithoutKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	MustNewClient("")
}

func TestNewClientCreatesResources(t *testing.T) {
	c := MustNewClient("test-key")
	if c.Agents == nil {
		t.Fatal("Agents is nil")
	}
	if c.Eval == nil {
		t.Fatal("Eval is nil")
	}
	if c.Eval.Templates == nil {
		t.Fatal("Eval.Templates is nil")
	}
	if c.Eval.Runs == nil {
		t.Fatal("Eval.Runs is nil")
	}
	if c.Agents.CustomState == nil {
		t.Fatal("Agents.CustomState is nil")
	}
	if c.Agents.Image == nil {
		t.Fatal("Agents.Image is nil")
	}
	if c.Tenants == nil {
		t.Fatal("Tenants is nil")
	}
	if c.APIKeys == nil {
		t.Fatal("APIKeys is nil")
	}
	if c.Analytics == nil {
		t.Fatal("Analytics is nil")
	}
	if c.UserPersonas == nil {
		t.Fatal("UserPersonas is nil")
	}
	if c.Storefront == nil {
		t.Fatal("Storefront is nil")
	}
	if c.Org == nil {
		t.Fatal("Org is nil")
	}
	if c.Workbench == nil {
		t.Fatal("Workbench is nil")
	}
	if c.Support == nil {
		t.Fatal("Support is nil")
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

	resp, err := client.Agents.Chat(context.Background(), AgentChatParams{
		AgentID:     "agent-1",
		ChatOptions: ChatOptions{Messages: []ChatMessage{{Role: "user", Content: "Hi"}}},
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
// Custom State
// ---------------------------------------------------------------------------

func TestCustomStateList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/custom-states" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, CustomStateListResponse{
			States: []CustomState{{StateID: "s1", Key: "level", Value: 42, Scope: "global"}},
		})
	})
	defer server.Close()

	result, err := client.Agents.CustomState.List(context.Background(), "agent-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(result.States))
	}
	if result.States[0].Key != "level" {
		t.Fatalf("expected 'level', got '%s'", result.States[0].Key)
	}
}

func TestCustomStateCreate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 201, CustomState{StateID: "s2", Key: "score", Value: 100, Scope: "user"})
	})
	defer server.Close()

	result, err := client.Agents.CustomState.Create(context.Background(), "agent-1", CustomStateCreateOptions{
		Key: "score", Value: 100, Scope: "user", UserID: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StateID != "s2" {
		t.Fatalf("expected 's2', got '%s'", result.StateID)
	}
}

// ---------------------------------------------------------------------------
// Image
// ---------------------------------------------------------------------------

func TestImageGenerate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/image/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, ImageGenerateResponse{
			ImageID:  "img-1",
			URL:      "https://storage.example.com/img-1.png",
			MimeType: "image/png",
		})
	})
	defer server.Close()

	result, err := client.Agents.Image.Generate(context.Background(), "agent-1", ImageGenerateOptions{
		Prompt: "a sunset over mountains",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ImageID != "img-1" {
		t.Fatalf("expected 'img-1', got '%s'", result.ImageID)
	}
}

// ---------------------------------------------------------------------------
// Memory Facts & Reset
// ---------------------------------------------------------------------------

func TestMemoryListFacts(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/memory/facts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, FactListResponse{
			Facts:      []Fact{{FactID: "f1", Content: "Likes pizza", Category: "preference", Confidence: 0.95}},
			TotalCount: 1,
		})
	})
	defer server.Close()

	result, err := client.Agents.Memory.ListFacts(context.Background(), "agent-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Facts[0].Content != "Likes pizza" {
		t.Fatalf("expected 'Likes pizza', got '%s'", result.Facts[0].Content)
	}
}

func TestMemoryReset(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		jsonResponse(w, 200, MemoryResetResponse{
			AgentID: "agent-1", Status: "reset", FactsDeleted: 42,
		})
	})
	defer server.Close()

	result, err := client.Agents.Memory.Reset(context.Background(), "agent-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FactsDeleted != 42 {
		t.Fatalf("expected 42 facts deleted, got %d", result.FactsDeleted)
	}
}

// ---------------------------------------------------------------------------
// Personality Update
// ---------------------------------------------------------------------------

func TestPersonalityUpdate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		jsonResponse(w, 200, PersonalityUpdateResponse{AgentID: "agent-1", Status: "updated"})
	})
	defer server.Close()

	result, err := client.Agents.Personality.Update(context.Background(), "agent-1", PersonalityUpdateOptions{
		Big5: &Big5Scores{Openness: 75, Conscientiousness: 60, Extraversion: 80, Agreeableness: 70, Neuroticism: 30},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "updated" {
		t.Fatalf("expected 'updated', got '%s'", result.Status)
	}
}

// ---------------------------------------------------------------------------
// Voice
// ---------------------------------------------------------------------------

func TestVoiceGetToken(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/voice/live-ws-token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, VoiceStreamToken{
			WSURL: "wss://api.sonz.ai/ws/voice/live", AuthToken: "tok-123",
		})
	})
	defer server.Close()

	result, err := client.Agents.Voice.GetToken(context.Background(), "agent-1", VoiceTokenOptions{
		VoiceName: "Kore", Language: "en-US",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AuthToken != "tok-123" {
		t.Fatalf("expected 'tok-123', got '%s'", result.AuthToken)
	}
}

func TestVoicesList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/voices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, VoiceListResponse{
			Voices:     []Voice{{VoiceID: "NATM0", VoiceName: "Sophia", Gender: "female", Tier: 2}},
			TotalCount: 1,
		})
	})
	defer server.Close()

	result, err := client.Voices.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Voices[0].VoiceName != "Sophia" {
		t.Fatalf("expected 'Sophia', got '%s'", result.Voices[0].VoiceName)
	}
}

// ---------------------------------------------------------------------------
// Wakeups
// ---------------------------------------------------------------------------

func TestScheduleWakeup(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/wakeups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 201, ScheduledWakeup{
			WakeupID: "w1", AgentID: "agent-1", CheckType: "birthday", Status: "scheduled",
		})
	})
	defer server.Close()

	result, err := client.Agents.Wakeups.Schedule(context.Background(), "agent-1", ScheduleWakeupOptions{
		UserID: "user-1", CheckType: "birthday", Intent: "check_in", DelayHours: 24,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "scheduled" {
		t.Fatalf("expected 'scheduled', got '%s'", result.Status)
	}
}

// ---------------------------------------------------------------------------
// Events
// ---------------------------------------------------------------------------

func TestTriggerEvent(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/events" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, TriggerEventResponse{
			Accepted: true, EventID: "evt-1",
		})
	})
	defer server.Close()

	result, err := client.Agents.TriggerEvent(context.Background(), "agent-1", TriggerEventOptions{
		UserID: "user-1", EventType: "situation",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventID != "evt-1" {
		t.Fatalf("expected 'evt-1', got '%s'", result.EventID)
	}
}

// TestTriggerEvent_IncludesMessages verifies raw chat messages on
// TriggerEventOptions are marshaled into the outbound request body under the
// "messages" key. Without this, the server cannot tell a session-originated
// event apart from a bare metadata-only event, and falls back to lossy
// consolidation summaries (see TD-PLAT-056 / TD-ORC-102 upstream).
func TestTriggerEvent_IncludesMessages(t *testing.T) {
	var capturedBody map[string]any
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &capturedBody); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		jsonResponse(w, 200, TriggerEventResponse{Accepted: true, EventID: "evt-msg"})
	})
	defer server.Close()

	msgs := []ChatMessage{
		{Role: "user", Content: "I quit my consulting job today."},
		{Role: "assistant", Content: "Big call. What drove it?"},
		{Role: "user", Content: "Lee announced Indonesia practice cuts."},
	}

	_, err := client.Agents.TriggerEvent(context.Background(), "agent-1", TriggerEventOptions{
		UserID:    "user-1",
		EventType: "daily_summary",
		Messages:  msgs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := capturedBody["messages"]
	if !ok {
		t.Fatalf("request body missing 'messages' key; got keys %v", mapKeys(capturedBody))
	}
	rawSlice, ok := raw.([]any)
	if !ok {
		t.Fatalf("'messages' not an array; got %T", raw)
	}
	if len(rawSlice) != len(msgs) {
		t.Fatalf("messages length: got %d, want %d", len(rawSlice), len(msgs))
	}
	for i, want := range msgs {
		got, ok := rawSlice[i].(map[string]any)
		if !ok {
			t.Fatalf("messages[%d] not an object; got %T", i, rawSlice[i])
		}
		if got["role"] != want.Role {
			t.Errorf("messages[%d].role: got %v, want %q", i, got["role"], want.Role)
		}
		if got["content"] != want.Content {
			t.Errorf("messages[%d].content: got %v, want %q", i, got["content"], want.Content)
		}
	}
}

// TestTriggerEvent_OmitsMessagesWhenEmpty guarantees backwards compatibility:
// callers that don't set Messages must not see the field appear in the JSON
// body at all (omitempty), so older servers that reject unknown fields keep
// working.
func TestTriggerEvent_OmitsMessagesWhenEmpty(t *testing.T) {
	var capturedBody map[string]any
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		jsonResponse(w, 200, TriggerEventResponse{Accepted: true, EventID: "evt-empty"})
	})
	defer server.Close()

	_, err := client.Agents.TriggerEvent(context.Background(), "agent-1", TriggerEventOptions{
		UserID: "user-1", EventType: "achievement",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := capturedBody["messages"]; present {
		t.Errorf("expected 'messages' absent when caller didn't set it; got body %v", capturedBody)
	}
}

func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// ---------------------------------------------------------------------------
// Dialogue
// ---------------------------------------------------------------------------

func TestDialogue(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/dialogue" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, DialogueResponse{
			Response: "Hello! Nice to meet you.",
		})
	})
	defer server.Close()

	result, err := client.Agents.Dialogue(context.Background(), "agent-1", DialogueOptions{
		Messages: []ChatMessage{{Role: "user", Content: "Talk to each other"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "Hello! Nice to meet you." {
		t.Fatalf("expected response, got '%s'", result.Response)
	}
}

// ---------------------------------------------------------------------------
// Agent CRUD
// ---------------------------------------------------------------------------

func TestAgentCreate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 201, Agent{AgentID: "new-1", Name: "Luna"})
	})
	defer server.Close()

	result, err := client.Agents.Create(context.Background(), CreateAgentOptions{Name: "Luna"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "new-1" {
		t.Fatalf("expected 'new-1', got '%s'", result.AgentID)
	}
}

func TestAgentGet(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, Agent{AgentID: "agent-1", Name: "Luna", Bio: "A friendly soul"})
	})
	defer server.Close()

	result, err := client.Agents.Get(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Luna" {
		t.Fatalf("expected 'Luna', got '%s'", result.Name)
	}
}

func TestAgentUpdate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		jsonResponse(w, 200, Agent{AgentID: "agent-1", Name: "Luna v2"})
	})
	defer server.Close()

	result, err := client.Agents.Update(context.Background(), "agent-1", UpdateAgentOptions{Name: "Luna v2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Luna v2" {
		t.Fatalf("expected 'Luna v2', got '%s'", result.Name)
	}
}

func TestAgentDelete(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		jsonResponse(w, 200, map[string]bool{"success": true})
	})
	defer server.Close()

	err := client.Agents.Delete(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Generation
// ---------------------------------------------------------------------------

func TestGenerateBio(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/bio/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, GenerateBioResponse{Bio: "A warm soul", Tone: "friendly"})
	})
	defer server.Close()

	result, err := client.Agents.Generation.GenerateBio(context.Background(), "agent-1", GenerateBioOptions{
		Name: "Luna", Gender: "female",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bio != "A warm soul" {
		t.Fatalf("expected 'A warm soul', got '%s'", result.Bio)
	}
}

func TestGenerateCharacter(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/generate-character" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, GenerateCharacterResponse{Bio: "AI bio", PersonalityPrompt: "You are warm"})
	})
	defer server.Close()

	result, err := client.Agents.Generation.GenerateCharacter(context.Background(), GenerateCharacterOptions{
		Name: "Luna", Gender: "female", Description: "A warm soul",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PersonalityPrompt != "You are warm" {
		t.Fatalf("expected 'You are warm', got '%s'", result.PersonalityPrompt)
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

func TestChatSSEChunkErrorType(t *testing.T) {
	e := ChatSSEChunkError{Message: "stream error"}
	if e.Message != "stream error" {
		t.Errorf("unexpected message: %s", e.Message)
	}
}

// ---------------------------------------------------------------------------
// Me
// ---------------------------------------------------------------------------

func TestMe(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/me" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, MeResponse{
			UserID: "user-1",
			Email:  "user@example.com",
			Orgs:   []OrgMembership{{OrgID: "org-1", Role: "admin"}},
		})
	})
	defer srv.Close()

	result, err := client.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.UserID != "user-1" {
		t.Errorf("got UserID %q, want %q", result.UserID, "user-1")
	}
	if result.Email != "user@example.com" {
		t.Errorf("got Email %q, want %q", result.Email, "user@example.com")
	}
	if len(result.Orgs) != 1 {
		t.Errorf("got %d orgs, want 1", len(result.Orgs))
	}
	if result.Orgs[0].OrgID != "org-1" {
		t.Errorf("got OrgID %q, want %q", result.Orgs[0].OrgID, "org-1")
	}
	if result.Orgs[0].Role != "admin" {
		t.Errorf("got Role %q, want %q", result.Orgs[0].Role, "admin")
	}
}

// ---------------------------------------------------------------------------
// Tenants
// ---------------------------------------------------------------------------

func TestTenantsList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/tenants" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"tenants": []Tenant{{TenantID: "t-1", Name: "Acme"}}})
	})
	defer srv.Close()
	result, err := client.Tenants.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tenants) != 1 || result.Tenants[0].TenantID != "t-1" {
		t.Error("unexpected result")
	}
}

func TestTenantsGet(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tenants/t-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, Tenant{TenantID: "t-1", Name: "Acme", IsActive: true})
	})
	defer srv.Close()
	result, err := client.Tenants.Get(context.Background(), "t-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.TenantID != "t-1" {
		t.Error("unexpected result")
	}
}

func TestKnowledgeListOrgNodes(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tenants/t-1/knowledge/org-nodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"nodes": []KBNode{{NodeID: "n-1"}}, "total": 1})
	})
	defer srv.Close()
	result, err := client.Knowledge.ListOrgNodes(context.Background(), "t-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Errorf("got total %d, want 1", result.Total)
	}
}

// ---------------------------------------------------------------------------
// API Keys
// ---------------------------------------------------------------------------

func TestAPIKeysCreate(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj-1/keys" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, APIKey{
			KeyID:     "key-1",
			Key:       "sk-live-abc123",
			ProjectID: "proj-1",
			IsActive:  true,
		})
	})
	defer srv.Close()
	result, err := client.APIKeys.Create(context.Background(), "proj-1", CreateAPIKeyOptions{Name: "my-key"})
	if err != nil {
		t.Fatal(err)
	}
	if result.KeyID != "key-1" {
		t.Errorf("got KeyID %q, want key-1", result.KeyID)
	}
}

func TestAPIKeysList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj-1/keys" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"keys": []APIKey{{KeyID: "key-1"}}})
	})
	defer srv.Close()
	result, err := client.APIKeys.List(context.Background(), "proj-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Keys) != 1 {
		t.Errorf("got %d keys, want 1", len(result.Keys))
	}
}

func TestAPIKeysRevoke(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/projects/proj-1/keys/key-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"success": true})
	})
	defer srv.Close()
	if err := client.APIKeys.Revoke(context.Background(), "proj-1", "key-1"); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Analytics
// ---------------------------------------------------------------------------

func TestAnalyticsOverview(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/analytics/overview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"total_agents": 5})
	})
	defer srv.Close()
	result, err := client.Analytics.Overview(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ---------------------------------------------------------------------------
// UserPersonas
// ---------------------------------------------------------------------------

func TestUserPersonasCreate(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/user-personas" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, UserPersona{PersonaID: "p-1", Name: "Friendly"})
	})
	defer srv.Close()
	result, err := client.UserPersonas.Create(context.Background(), CreateUserPersonaOptions{Name: "Friendly"})
	if err != nil {
		t.Fatal(err)
	}
	if result.PersonaID != "p-1" {
		t.Errorf("got PersonaID %q, want p-1", result.PersonaID)
	}
}

func TestUserPersonasList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user-personas" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"personas": []UserPersona{{PersonaID: "p-1"}}})
	})
	defer srv.Close()
	result, err := client.UserPersonas.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Personas) != 1 {
		t.Errorf("got %d personas, want 1", len(result.Personas))
	}
}

func TestUserPersonasDelete(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/user-personas/p-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"success": true})
	})
	defer srv.Close()
	if err := client.UserPersonas.Delete(context.Background(), "p-1"); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Storefront
// ---------------------------------------------------------------------------

func TestStorefrontGet(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/storefront" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"slug": "my-store"})
	})
	defer srv.Close()
	result, err := client.Storefront.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result["slug"] != "my-store" {
		t.Error("unexpected result")
	}
}

func TestStorefrontUpsertAgent(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/storefront/agents/agent-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{})
	})
	defer srv.Close()
	if err := client.Storefront.UpsertAgent(context.Background(), "agent-1", StorefrontAgentOptions{DisplayName: "My Agent"}); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Org
// ---------------------------------------------------------------------------

func TestOrgGetBilling(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/org/billing" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"status": "active"})
	})
	defer srv.Close()
	result, err := client.Org.GetBilling(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "active" {
		t.Error("unexpected result")
	}
}

func TestOrgRedeemVoucher(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/org/vouchers/redeem" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"redeemed": true})
	})
	defer srv.Close()
	result, err := client.Org.RedeemVoucher(context.Background(), "VOUCHER123")
	if err != nil {
		t.Fatal(err)
	}
	if result["redeemed"] != true {
		t.Errorf("expected redeemed=true, got %v", result["redeemed"])
	}
}

// ---------------------------------------------------------------------------
// Workbench
// ---------------------------------------------------------------------------

func TestWorkbenchChat(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/workbench/chat" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"response": "hello"})
	})
	defer srv.Close()
	result, err := client.Workbench.Chat(context.Background(), map[string]any{"message": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if result["response"] != "hello" {
		t.Error("unexpected result")
	}
}

func TestAnalyticsCostBreakdown(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/analytics/cost/breakdown" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("days") != "7" {
			t.Errorf("expected days=7, got %q", r.URL.Query().Get("days"))
		}
		jsonResponse(w, 200, map[string]any{"total_usd": 1.23})
	})
	defer srv.Close()
	result, err := client.Analytics.CostBreakdown(context.Background(), &AnalyticsOptions{Days: 7})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ---------------------------------------------------------------------------
// Support
// ---------------------------------------------------------------------------

func TestSupportList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/support/tickets" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, TicketListResponse{
			Tickets: []SupportTicket{{TicketID: "t-1", Title: "Help"}},
			Total:   1,
		})
	})
	defer srv.Close()
	result, err := client.Support.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tickets) != 1 || result.Tickets[0].TicketID != "t-1" {
		t.Error("unexpected result")
	}
}

func TestSupportCreate(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/support/tickets" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, SupportTicket{TicketID: "t-2", Title: "New issue", Status: "open"})
	})
	defer srv.Close()
	result, err := client.Support.Create(context.Background(), CreateTicketOptions{
		Title:       "New issue",
		Description: "Something broke",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TicketID != "t-2" {
		t.Errorf("got TicketID %q, want t-2", result.TicketID)
	}
}

func TestSupportAddComment(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/support/tickets/t-1/comments" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, SupportTicketComment{CommentID: "c-1", TicketID: "t-1", Content: "Thanks"})
	})
	defer srv.Close()
	result, err := client.Support.AddComment(context.Background(), "t-1", AddTicketCommentOptions{Content: "Thanks"})
	if err != nil {
		t.Fatal(err)
	}
	if result.CommentID != "c-1" {
		t.Errorf("got CommentID %q, want c-1", result.CommentID)
	}
}
