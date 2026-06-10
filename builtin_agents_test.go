package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// Pin URL shapes, HTTP verbs, query params, and body payloads so the Go SDK
// stays in sync with the built-in agent handlers in services/platform/api
// and the sibling Python / TypeScript SDKs. Response parsing is smoke-tested
// by decoding a few fields; the SSE path is exercised against a synthetic
// stream.

func TestBuiltinAgents_List_URLAndDecode(t *testing.T) {
	var seen struct {
		path, method string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"agents": []map[string]any{
				{
					"slug":        "lead_research",
					"name":        "Lead Research",
					"description": "Deep-researches a company and contact.",
					"model":       "sonzai-task-1",
					"provisioned": true,
				},
				{
					"slug":        "market_intel",
					"name":        "Market Intel",
					"description": "Market and competitor intelligence.",
					"model":       "sonzai-task-1",
					"provisioned": false,
				},
			},
		})
	})
	client := newTestClient(t, h)

	agents, err := client.BuiltinAgents.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents" {
		t.Errorf("path: got %q", seen.path)
	}
	if len(agents) != 2 || agents[0].Slug != "lead_research" || !agents[0].Provisioned {
		t.Errorf("decoded list mismatch: %+v", agents)
	}
	if agents[1].Provisioned {
		t.Errorf("agents[1].Provisioned: got true, want false")
	}
}

func TestBuiltinAgents_Invoke_Blocking_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method, stream string
		body                 map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		seen.stream = r.URL.Query().Get("stream")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"findings":   map[string]any{"company": "Ayala Land", "score": 87},
			"summary":    "Strong enterprise lead.",
			"session_id": "sess-1",
			"model":      "sonzai-task-1",
			"byok":       true,
			"usage": map[string]any{
				"input_tokens":                1200,
				"output_tokens":               340,
				"cache_creation_input_tokens": 50,
				"cache_read_input_tokens":     900,
			},
			"running_seconds": 412.5,
			"cost_usd":        0.0731,
		})
	})
	client := newTestClient(t, h)

	got, err := client.BuiltinAgents.Invoke(context.Background(), BuiltinAgentLeadResearch, BuiltinAgentInvokeParams{
		Input: map[string]any{"company": "Ayala Land"},
		Title: "Ayala Land research",
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/lead_research/invoke" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.stream != "false" {
		t.Errorf("stream query: got %q, want \"false\"", seen.stream)
	}
	input, ok := seen.body["input"].(map[string]any)
	if !ok || input["company"] != "Ayala Land" {
		t.Errorf("body.input: got %v", seen.body["input"])
	}
	if seen.body["title"] != "Ayala Land research" {
		t.Errorf("body.title: got %v", seen.body["title"])
	}
	if got.SessionID != "sess-1" || got.Summary != "Strong enterprise lead." || !got.BYOK {
		t.Errorf("decoded result mismatch: %+v", got)
	}
	if got.Usage.InputTokens != 1200 || got.Usage.CacheReadInputTokens != 900 {
		t.Errorf("decoded usage mismatch: %+v", got.Usage)
	}
	if got.RunningSeconds != 412.5 || got.CostUSD != 0.0731 {
		t.Errorf("decoded cost mismatch: %+v", got)
	}
	var findings struct {
		Company string `json:"company"`
	}
	if err := json.Unmarshal(got.Findings, &findings); err != nil || findings.Company != "Ayala Land" {
		t.Errorf("findings raw JSON mismatch: %s (err=%v)", string(got.Findings), err)
	}
}

func TestBuiltinAgents_InvokeStream_SSEParse(t *testing.T) {
	var seen struct {
		path, stream, accept string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.stream = r.URL.Query().Get("stream")
		seen.accept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: update\n")
		fmt.Fprint(w, "data: {\"type\":\"tool_use\",\"tool\":\"web_search\",\"detail\":\"ayala land leadership\",\"elapsed\":3.2}\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "event: update\n")
		fmt.Fprint(w, "data: {\"type\":\"text\",\"text\":\"Compiling findings...\",\"elapsed\":41.0}\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "event: result\n")
		fmt.Fprint(w, "data: {\"findings\":{\"ok\":true},\"summary\":\"done\",\"session_id\":\"sess-2\",\"model\":\"sonzai-task-1\",\"byok\":false,\"usage\":{\"input_tokens\":10,\"output_tokens\":5,\"cache_creation_input_tokens\":0,\"cache_read_input_tokens\":0},\"running_seconds\":44.1,\"cost_usd\":0.002}\n")
		fmt.Fprint(w, "\n")
	})
	client := newTestClient(t, h)

	var updates []BuiltinAgentStreamUpdate
	got, err := client.BuiltinAgents.InvokeStream(context.Background(), BuiltinAgentMarketIntel,
		BuiltinAgentInvokeParams{Input: map[string]any{"market": "PH real estate"}},
		func(u BuiltinAgentStreamUpdate) { updates = append(updates, u) })
	if err != nil {
		t.Fatalf("InvokeStream: %v", err)
	}
	if seen.path != "/api/v1/builtin-agents/market_intel/invoke" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.stream != "true" {
		t.Errorf("stream query: got %q, want \"true\"", seen.stream)
	}
	if seen.accept != "text/event-stream" {
		t.Errorf("Accept header: got %q", seen.accept)
	}
	if len(updates) != 2 {
		t.Fatalf("updates: got %d, want 2 (%+v)", len(updates), updates)
	}
	if updates[0].Type != "tool_use" || updates[0].Tool != "web_search" || updates[0].Elapsed != 3.2 {
		t.Errorf("updates[0] mismatch: %+v", updates[0])
	}
	if updates[1].Type != "text" || updates[1].Text != "Compiling findings..." {
		t.Errorf("updates[1] mismatch: %+v", updates[1])
	}
	if got == nil || got.SessionID != "sess-2" || got.Summary != "done" || got.RunningSeconds != 44.1 {
		t.Errorf("decoded result mismatch: %+v", got)
	}
}

func TestBuiltinAgents_InvokeStream_ErrorEvent(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: update\n")
		fmt.Fprint(w, "data: {\"type\":\"text\",\"text\":\"starting\"}\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "event: error\n")
		fmt.Fprint(w, "data: {\"error\":\"agent run failed: upstream timeout\"}\n")
		fmt.Fprint(w, "\n")
	})
	client := newTestClient(t, h)

	_, err := client.BuiltinAgents.InvokeStream(context.Background(), BuiltinAgentLeadScore,
		BuiltinAgentInvokeParams{Input: map[string]any{"lead_id": "l-1"}}, nil)
	if err == nil {
		t.Fatal("InvokeStream: expected error from error event, got nil")
	}
	if want := "agent run failed: upstream timeout"; err.Error() != "builtin agent: "+want {
		t.Errorf("error: got %q, want suffix %q", err.Error(), want)
	}
}

func TestBuiltinAgents_CreateSession_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "sess-3",
			"agent":      "lead_qualifier",
			"model":      "sonzai-task-1",
			"status":     "active",
			"title":      "Qualify inbound",
			"byok":       false,
			"cost_usd":   0,
			"created_at": "2026-06-10T00:00:00Z",
		})
	})
	client := newTestClient(t, h)

	got, err := client.BuiltinAgents.CreateSession(context.Background(), BuiltinAgentSessionParams{
		Agent: BuiltinAgentLeadQualifier,
		Title: "Qualify inbound",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/sessions" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.body["agent"] != "lead_qualifier" || seen.body["title"] != "Qualify inbound" {
		t.Errorf("body mismatch: %+v", seen.body)
	}
	if got.ID != "sess-3" || got.Agent != "lead_qualifier" || got.Status != "active" {
		t.Errorf("decoded session mismatch: %+v", got)
	}
}

func TestBuiltinAgents_ListSessions_URLAndLimit(t *testing.T) {
	var seen struct {
		path, method, limit string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		seen.limit = r.URL.Query().Get("limit")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sessions": []map[string]any{
				{"id": "sess-1", "agent": "lead_research", "status": "completed"},
			},
		})
	})
	client := newTestClient(t, h)

	sessions, err := client.BuiltinAgents.ListSessions(context.Background(), 25)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/sessions" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.limit != "25" {
		t.Errorf("limit query: got %q, want 25", seen.limit)
	}
	if len(sessions) != 1 || sessions[0].ID != "sess-1" {
		t.Errorf("decoded sessions mismatch: %+v", sessions)
	}
}

func TestBuiltinAgents_GetSession_URLAndBilledFields(t *testing.T) {
	var seen struct {
		path, method string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                           "sess-4",
			"agent":                        "lead_extract",
			"model":                        "sonzai-task-1",
			"status":                       "completed",
			"title":                        "Extract leads",
			"byok":                         true,
			"cost_usd":                     0.12,
			"created_at":                   "2026-06-10T00:00:00Z",
			"billed_input_tokens":          5000,
			"billed_output_tokens":         1200,
			"billed_cache_read_tokens":     800,
			"billed_cache_creation_tokens": 300,
		})
	})
	client := newTestClient(t, h)

	got, err := client.BuiltinAgents.GetSession(context.Background(), "sess-4")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/sessions/sess-4" {
		t.Errorf("path: got %q", seen.path)
	}
	if got.BilledInputTokens != 5000 || got.BilledOutputTokens != 1200 ||
		got.BilledCacheReadTokens != 800 || got.BilledCacheCreationTokens != 300 {
		t.Errorf("billed fields mismatch: %+v", got)
	}
}

func TestBuiltinAgents_SendMessage_Blocking_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method, stream string
		body                 map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		seen.stream = r.URL.Query().Get("stream")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"reply":    "The decision maker is the VP of Leasing.",
			"findings": map[string]any{"contact": "VP of Leasing"},
			"usage": map[string]any{
				"input_tokens":                300,
				"output_tokens":               80,
				"cache_creation_input_tokens": 0,
				"cache_read_input_tokens":     250,
			},
			"turn_cost_usd":   0.004,
			"running_seconds": 9.7,
		})
	})
	client := newTestClient(t, h)

	got, err := client.BuiltinAgents.SendMessage(context.Background(), "sess-1", "Who is the decision maker?", nil)
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/sessions/sess-1/messages" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.stream != "false" {
		t.Errorf("stream query: got %q, want \"false\"", seen.stream)
	}
	if seen.body["text"] != "Who is the decision maker?" {
		t.Errorf("body.text: got %v", seen.body["text"])
	}
	if got.Reply != "The decision maker is the VP of Leasing." || got.TurnCostUSD != 0.004 {
		t.Errorf("decoded turn mismatch: %+v", got)
	}
	if got.Usage.OutputTokens != 80 || got.RunningSeconds != 9.7 {
		t.Errorf("decoded usage mismatch: %+v", got)
	}
}

func TestBuiltinAgents_SendMessage_Streaming_SSEParse(t *testing.T) {
	var seen struct {
		stream string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.stream = r.URL.Query().Get("stream")
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: update\n")
		fmt.Fprint(w, "data: {\"type\":\"thinking\",\"text\":\"checking notes\"}\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "event: result\n")
		fmt.Fprint(w, "data: {\"reply\":\"It is the CFO.\",\"usage\":{\"input_tokens\":50,\"output_tokens\":12,\"cache_creation_input_tokens\":0,\"cache_read_input_tokens\":40},\"turn_cost_usd\":0.001,\"running_seconds\":4.2}\n")
		fmt.Fprint(w, "\n")
	})
	client := newTestClient(t, h)

	var updates []BuiltinAgentStreamUpdate
	got, err := client.BuiltinAgents.SendMessage(context.Background(), "sess-1", "Who signs off?",
		func(u BuiltinAgentStreamUpdate) { updates = append(updates, u) })
	if err != nil {
		t.Fatalf("SendMessage (stream): %v", err)
	}
	if seen.stream != "true" {
		t.Errorf("stream query: got %q, want \"true\"", seen.stream)
	}
	if len(updates) != 1 || updates[0].Type != "thinking" {
		t.Errorf("updates mismatch: %+v", updates)
	}
	if got == nil || got.Reply != "It is the CFO." || got.TurnCostUSD != 0.001 {
		t.Errorf("decoded turn mismatch: %+v", got)
	}
}

func TestBuiltinAgent_SlugConstants(t *testing.T) {
	// Lock the wire values — these round-trip through the platform API
	// and the sibling sonzai-python / sonzai-typescript SDKs.
	cases := map[string]string{
		BuiltinAgentLeadResearch:  "lead_research",
		BuiltinAgentMarketIntel:   "market_intel",
		BuiltinAgentLeadExtract:   "lead_extract",
		BuiltinAgentLeadScore:     "lead_score",
		BuiltinAgentLeadQualifier: "lead_qualifier",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("slug constant: got %q, want %q", got, want)
		}
	}
}
