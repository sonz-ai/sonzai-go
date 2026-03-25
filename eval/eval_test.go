package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestBackend(handler http.HandlerFunc) *testBackend {
	server := httptest.NewServer(handler)
	return &testBackend{server: server}
}

type testBackend struct {
	server *httptest.Server
}

func (b *testBackend) Get(ctx context.Context, path string, params map[string]string, result interface{}) error {
	resp, err := http.Get(b.server.URL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (b *testBackend) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	resp, err := http.Post(b.server.URL+path, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (b *testBackend) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return b.Post(ctx, path, body, result) // reuse for tests
}

func (b *testBackend) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return b.Post(ctx, path, body, result) // reuse for tests
}

func (b *testBackend) Delete(ctx context.Context, path string, result interface{}) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, b.server.URL+path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (b *testBackend) StreamSSE(ctx context.Context, method, path string, body interface{}, callback func(json.RawMessage) error) error {
	resp, err := http.Get(b.server.URL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return err
	}
	return callback(raw)
}

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

func TestTemplatesList(t *testing.T) {
	backend := newTestBackend(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/eval-templates" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, TemplateListResponse{
			Templates: []Template{{ID: "tpl-1", Name: "Quality Check"}},
		})
	})
	defer backend.server.Close()

	client := New(backend)
	result, err := client.Templates.List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Templates[0].Name != "Quality Check" {
		t.Fatalf("expected 'Quality Check', got '%s'", result.Templates[0].Name)
	}
}

func TestTemplateCreate(t *testing.T) {
	backend := newTestBackend(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 201, Template{ID: "tpl-2", Name: "New Template"})
	})
	defer backend.server.Close()

	client := New(backend)
	result, err := client.Templates.Create(context.Background(), TemplateCreateOptions{
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
// Runs
// ---------------------------------------------------------------------------

func TestRunsList(t *testing.T) {
	backend := newTestBackend(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/eval-runs" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, RunListResponse{
			Runs:       []Run{{ID: "run-1", Status: "completed", TotalTurns: 20}},
			TotalCount: 1,
		})
	})
	defer backend.server.Close()

	client := New(backend)
	result, err := client.Runs.List(context.Background(), "", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Runs[0].Status != "completed" {
		t.Fatalf("expected 'completed', got '%s'", result.Runs[0].Status)
	}
}

// ---------------------------------------------------------------------------
// Evaluate
// ---------------------------------------------------------------------------

func TestEvaluate(t *testing.T) {
	backend := newTestBackend(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, EvaluationResult{
			Score:    0.85,
			Feedback: "Good response",
			Categories: []Category{
				{Name: "Empathy", Score: 0.9, Feedback: "Strong empathy"},
			},
		})
	})
	defer backend.server.Close()

	client := New(backend)
	result, err := client.Evaluate(context.Background(), "agent-1", EvaluateOptions{
		Messages:   []Message{{Role: "user", Content: "I'm sad"}},
		TemplateID: "tpl-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 0.85 {
		t.Fatalf("expected 0.85, got %f", result.Score)
	}
	if len(result.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(result.Categories))
	}
}

// ---------------------------------------------------------------------------
// Simulate
// ---------------------------------------------------------------------------

func TestSimulate(t *testing.T) {
	// Simulate now uses two-step: POST returns RunRef, then StreamEvents
	callCount := 0
	backend := newTestBackend(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call: POST /simulate → returns RunRef
			jsonResponse(w, 200, RunRef{RunID: "run-sim-1", Status: "running"})
		} else {
			// Second call: GET /events → returns SSE event
			jsonResponse(w, 200, SimulationEvent{
				Type:    "session_start",
				Message: "Starting session 1",
			})
		}
	})
	defer backend.server.Close()

	client := New(backend)
	var events []SimulationEvent
	ref, err := client.Simulate(context.Background(), "agent-1", SimulateOptions{}, func(event SimulationEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.RunID != "run-sim-1" {
		t.Fatalf("expected run_id 'run-sim-1', got '%s'", ref.RunID)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "session_start" {
		t.Fatalf("expected 'session_start', got '%s'", events[0].Type)
	}
}

// Suppress unused import warning.
var _ = fmt.Sprintf
