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
// Agent CRUD
// ---------------------------------------------------------------------------

func TestAgentCreate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 201, CreateAgentResult{AgentID: "agent-new", Status: "created"})
	})
	defer server.Close()

	result, err := client.Agents.Create(context.Background(), CreateAgentParams{
		UserID: "user-1", AgentName: "Luna", Gender: "female",
		Big5: Big5Scores{Openness: 0.8, Conscientiousness: 0.6},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "agent-new" {
		t.Fatalf("expected 'agent-new', got '%s'", result.AgentID)
	}
}

func TestAgentGet(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/agents/agent-1" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, AgentProfile{AgentID: "agent-1", Name: "Luna", Gender: "female"})
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
		if r.Method != "PATCH" || r.URL.Path != "/api/v1/agents/agent-1" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, UpdateAgentResult{Success: true})
	})
	defer server.Close()

	result, err := client.Agents.Update(context.Background(), "agent-1", UpdateAgentParams{Name: "Luna v2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestAgentDelete(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/v1/agents/agent-1" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.Agents.Delete(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Memory Seed / Facts / Reset
// ---------------------------------------------------------------------------

func TestMemorySeed(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/memory/seed" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, SeedMemoriesResult{MemoriesCreated: 3})
	})
	defer server.Close()

	result, err := client.Agents.Memory.Seed(context.Background(), "agent-1", SeedMemoriesParams{
		UserID: "user-1",
		Memories: []MemoryCandidate{
			{Content: "Likes coffee", FactType: "preference", Importance: 0.7},
			{Content: "Lives in Singapore", FactType: "location", Importance: 0.9},
			{Content: "Works as engineer", FactType: "occupation", Importance: 0.8},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MemoriesCreated != 3 {
		t.Fatalf("expected 3, got %d", result.MemoriesCreated)
	}
}

func TestMemoryListFacts(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/memory/facts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("fact_type") != "preference" {
			t.Fatalf("expected fact_type=preference, got %s", r.URL.Query().Get("fact_type"))
		}
		jsonResponse(w, 200, ListFactsResult{
			Facts:      []StoredFact{{FactID: "f1", Content: "Likes coffee", FactType: "preference"}},
			TotalCount: 1,
		})
	})
	defer server.Close()

	result, err := client.Agents.Memory.ListFacts(context.Background(), "agent-1", &ListFactsOptions{
		UserID: "user-1", FactType: "preference",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("expected 1, got %d", result.TotalCount)
	}
	if result.Facts[0].Content != "Likes coffee" {
		t.Fatalf("expected 'Likes coffee', got '%s'", result.Facts[0].Content)
	}
}

func TestMemoryReset(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/memory/reset" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, ResetMemoryResult{Success: true, FactsDeleted: 5, NodesDeleted: 2})
	})
	defer server.Close()

	result, err := client.Agents.Memory.Reset(context.Background(), "agent-1", ResetMemoryParams{UserID: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success || result.FactsDeleted != 5 {
		t.Fatalf("expected success with 5 facts deleted, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// Personality Update
// ---------------------------------------------------------------------------

func TestPersonalityUpdate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/v1/agents/agent-1/personality" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, UpdatePersonalityResult{Success: true})
	})
	defer server.Close()

	result, err := client.Agents.Personality.Update(context.Background(), "agent-1", UpdatePersonalityParams{
		Big5: Big5Scores{Openness: 0.9, Conscientiousness: 0.7},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

// ---------------------------------------------------------------------------
// Voice
// ---------------------------------------------------------------------------

func TestVoiceTTS(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/voice/tts" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, TextToSpeechResult{ContentType: "audio/mp3", VoiceName: "alloy", DurationMs: 1500})
	})
	defer server.Close()

	result, err := client.Voice.TextToSpeech(context.Background(), TextToSpeechParams{
		AgentID: "agent-1", Text: "Hello world",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VoiceName != "alloy" {
		t.Fatalf("expected 'alloy', got '%s'", result.VoiceName)
	}
}

func TestVoiceMatch(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/voice/match" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, VoiceMatchResult{VoiceID: "v1", VoiceName: "nova", MatchScore: 0.92})
	})
	defer server.Close()

	result, err := client.Voice.VoiceMatch(context.Background(), VoiceMatchParams{
		Big5: Big5Scores{Openness: 0.8, Extraversion: 0.7},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VoiceName != "nova" {
		t.Fatalf("expected 'nova', got '%s'", result.VoiceName)
	}
}

func TestVoiceListVoices(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/voice/voices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, ListVoicesResult{
			Voices: []VoiceInfo{{Name: "alloy", Gender: "neutral"}, {Name: "nova", Gender: "female"}},
		})
	})
	defer server.Close()

	result, err := client.Voice.ListVoices(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Voices) != 2 {
		t.Fatalf("expected 2 voices, got %d", len(result.Voices))
	}
}

// ---------------------------------------------------------------------------
// Generation
// ---------------------------------------------------------------------------

func TestGenerateBio(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/generate-bio" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, GenerateBioResult{Bio: "A curious soul...", Tone: "warm", Confidence: 0.9})
	})
	defer server.Close()

	result, err := client.Agents.Generation.GenerateBio(context.Background(), "agent-1", GenerateBioParams{
		UserID: "user-1", Style: "poetic",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bio != "A curious soul..." {
		t.Fatalf("expected 'A curious soul...', got '%s'", result.Bio)
	}
}

func TestGenerateImage(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/image/generate" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, GenerateImageResult{Success: true, ImageID: "img-1", PublicURL: "https://cdn.example.com/img.png"})
	})
	defer server.Close()

	result, err := client.Agents.Generation.GenerateImage(context.Background(), "agent-1", GenerateImageParams{
		Prompt: "A cute cat",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success || result.ImageID != "img-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGenerateCharacter(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/generate-character" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, GenerateCharacterResult{
			Bio:               "A warm and curious companion",
			PersonalityPrompt: "You are warm...",
			Big5:              &Big5Scores{Openness: 0.85},
		})
	})
	defer server.Close()

	result, err := client.Agents.Generation.GenerateCharacter(context.Background(), "agent-1", GenerateCharacterParams{
		Name: "Luna", Gender: "female", Description: "A warm companion",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Big5 == nil || result.Big5.Openness != 0.85 {
		t.Fatalf("unexpected Big5: %+v", result.Big5)
	}
}

// ---------------------------------------------------------------------------
// Dialogue
// ---------------------------------------------------------------------------

func TestDialogue(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/dialogue" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, AgentDialogueResult{Response: "Hello there!"})
	})
	defer server.Close()

	result, err := client.Agents.Dialogue(context.Background(), "agent-1", AgentDialogueParams{
		UserID: "user-1", SceneGuidance: "casual greeting",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "Hello there!" {
		t.Fatalf("expected 'Hello there!', got '%s'", result.Response)
	}
}

// ---------------------------------------------------------------------------
// Game Events
// ---------------------------------------------------------------------------

func TestTriggerGameEvent(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/events" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, TriggerGameEventResult{Accepted: true, EventID: "evt-1"})
	})
	defer server.Close()

	result, err := client.Agents.TriggerGameEvent(context.Background(), "agent-1", TriggerGameEventParams{
		UserID:    "user-1",
		EventType: "achievement",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accepted {
		t.Fatal("expected accepted")
	}
}

// ---------------------------------------------------------------------------
// Custom States
// ---------------------------------------------------------------------------

func TestCustomStatesList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/custom-states" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, CustomStateListResponse{
			States: []CustomState{{StateID: "s1", Key: "level", AgentID: "agent-1"}},
		})
	})
	defer server.Close()

	result, err := client.Agents.CustomStates.List(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.States) != 1 || result.States[0].Key != "level" {
		t.Fatalf("unexpected states: %+v", result.States)
	}
}

func TestCustomStatesCreate(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/agents/agent-1/custom-states" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 201, CustomState{StateID: "s2", Key: "score"})
	})
	defer server.Close()

	result, err := client.Agents.CustomStates.Create(context.Background(), "agent-1", CustomStateCreateParams{
		Key: "score", Value: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Key != "score" {
		t.Fatalf("expected 'score', got '%s'", result.Key)
	}
}

// ---------------------------------------------------------------------------
// Context Engine Extended
// ---------------------------------------------------------------------------

func TestGetConstellation(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/constellation" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}})
	})
	defer server.Close()

	result, err := client.Agents.GetConstellation(context.Background(), "agent-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestGetBreakthroughs(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/breakthroughs" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]interface{}{"breakthroughs": []interface{}{}})
	})
	defer server.Close()

	result, err := client.Agents.GetBreakthroughs(context.Background(), "agent-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ---------------------------------------------------------------------------
// New Resource Wiring
// ---------------------------------------------------------------------------

func TestNewClientCreatesAllResources(t *testing.T) {
	c := NewClient("test-key")
	if c.Voice == nil {
		t.Fatal("Voice is nil")
	}
	if c.Agents.Generation == nil {
		t.Fatal("Generation is nil")
	}
	if c.Agents.CustomStates == nil {
		t.Fatal("CustomStates is nil")
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
