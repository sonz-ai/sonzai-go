package sonzai

import (
	"context"
	"net/http"
	"testing"
)

func TestMLResource_EV_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path, seen.method = r.URL.Path, r.Method
		_, _ = w.Write([]byte(`{"probability":0.4,"value":100,"expected_value":40,"prob_served_from":"model","value_served_from":"prior"}`))
	})
	client := newTestClient(t, h)

	result, err := client.ML.EV(context.Background(), EVParams{Features: map[string]float64{"x": 1}})
	if err != nil {
		t.Fatalf("EV: %v", err)
	}
	if seen.method != http.MethodPost || seen.path != "/api/v1/ml/ev" {
		t.Errorf("method/path = %s %s", seen.method, seen.path)
	}
	if result.ExpectedValue != 40 {
		t.Errorf("ExpectedValue = %v", result.ExpectedValue)
	}
}

func TestMLResource_Forecast_WindowDaysParam(t *testing.T) {
	var gotQuery string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"expected_revenue":1,"p10":1,"p50":2,"p90":3,"open_leads":5,"assumptions":["a"]}`))
	})
	client := newTestClient(t, h)

	result, err := client.ML.Forecast(context.Background(), ForecastOptions{WindowDays: 90})
	if err != nil {
		t.Fatalf("Forecast: %v", err)
	}
	if gotQuery != "window_days=90" {
		t.Errorf("query = %q", gotQuery)
	}
	if result.OpenLeads != 5 {
		t.Errorf("OpenLeads = %v", result.OpenLeads)
	}
}

func TestMLResource_TimingSuggest_URLAndBody(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ml/contact_timing/suggest" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"send_bucket":{"daypart":"morning","day_type":"weekday"},"channel":"whatsapp","action_id":"a1","propensity":0.5}`))
	})
	client := newTestClient(t, h)

	result, err := client.ML.TimingSuggest(context.Background(), ContactTimingSuggestParams{ContactID: "c1", Channels: []string{"whatsapp"}})
	if err != nil {
		t.Fatalf("TimingSuggest: %v", err)
	}
	if result.Channel != "whatsapp" || result.SendBucket.Daypart != "morning" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestMLResource_RevivalQueue_Decode(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ml/revival-queue" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"computed_at":"2026-01-01T00:00:00Z","items":[{"contact_id":"c1","agent_id":"a1","revival_score":0.9,"dormancy_days":30,"reason":"went quiet"}]}`))
	})
	client := newTestClient(t, h)

	result, err := client.ML.RevivalQueue(context.Background())
	if err != nil {
		t.Fatalf("RevivalQueue: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ContactID != "c1" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestModelsResource_Export_FormatParam(t *testing.T) {
	var gotQuery, gotPath string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"use_case":"lead_score","version":"v1","format":"onnx-v1","artifact_b64":"abc"}`))
	})
	client := newTestClient(t, h)

	result, err := client.Models.Export(context.Background(), "lead_score", ModelExportOptions{Format: "onnx"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if gotPath != "/api/v1/models/lead_score/export" || gotQuery != "format=onnx" {
		t.Errorf("path/query = %s %s", gotPath, gotQuery)
	}
	if result.ArtifactB64 != "abc" {
		t.Errorf("ArtifactB64 = %q", result.ArtifactB64)
	}
}

func TestModelsResource_Export_404_NotFoundError(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"no model"}`, http.StatusNotFound)
	})
	client := newTestClient(t, h)

	_, err := client.Models.Export(context.Background(), "lead_score", ModelExportOptions{})
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func TestAgentsResource_ListUserConversations(t *testing.T) {
	var gotPath, gotQuery string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"messages":[{"role":"summary","content":"hi","timestamp":"2026-01-01T00:00:00Z","session_id":"s1"}],"source":"memory_timeline"}`))
	})
	client := newTestClient(t, h)

	result, err := client.Agents.ListUserConversations(context.Background(), "agent-1", "user-1", &ListUserConversationsOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListUserConversations: %v", err)
	}
	if gotPath != "/api/v1/agents/agent-1/users/user-1/conversations" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "limit=10" {
		t.Errorf("query = %q", gotQuery)
	}
	if result.Source != "memory_timeline" || len(result.Messages) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
}
