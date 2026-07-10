package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestRuntimeContextBundlePreservesLicenseCode(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v1/agents/agent%2Fone/context-bundle" {
			t.Fatalf("path = %q", r.URL.EscapedPath())
		}
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail":"context bundle access denied","errors":[{"location":"license_state","message":"license_insufficient_credit"}]}`))
	}))

	_, err := client.Runtime.ContextBundle(context.Background(), "agent/one", RuntimeContextBundleParams{UserID: "u1", SessionID: "s1"})
	apiErr, ok := err.(*RuntimeAPIError)
	if !ok {
		t.Fatalf("error = %T %v, want *RuntimeAPIError", err, err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Code != "license_insufficient_credit" {
		t.Fatalf("unexpected runtime error: %+v", apiErr)
	}
}

func TestRuntimeBackendAgentArtifactsAreConfigurationOnly(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/runtime/backend-agent-artifacts" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, http.StatusOK, RuntimeBackendAgentArtifactList{Artifacts: []RuntimeBackendAgentArtifact{{Slug: "lead_score", System: "score locally", Version: "v1"}}})
	}))
	out, err := client.Runtime.BackendAgentArtifacts(context.Background())
	if err != nil || len(out.Artifacts) != 1 || out.Artifacts[0].Slug != "lead_score" {
		t.Fatalf("BackendAgentArtifacts = %+v, %v", out, err)
	}
}

func TestRuntimeReportTurnsUsesTypedContract(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent-1/turns" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		var body RuntimeTurnReport
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.UserID != "user-1" || len(body.Messages) != 2 {
			t.Fatalf("body = %+v", body)
		}
		jsonResponse(w, http.StatusAccepted, RuntimeTurnReportResult{Accepted: true, AgentID: "agent-1", MessagesStored: 2})
	}))

	text := "hello"
	result, err := client.Runtime.ReportTurns(context.Background(), "agent-1", RuntimeTurnReport{
		UserID: "user-1", SessionID: "session-1",
		Messages: []RuntimeTurnMessage{{Role: "user", Content: &text}, {Role: "assistant", Content: &text}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Accepted || result.MessagesStored != 2 {
		t.Fatalf("result = %+v", result)
	}
}

func TestRuntimeUsageSignatureStableAcrossCounterOrder(t *testing.T) {
	now := time.Date(2026, 7, 10, 8, 0, 0, 123, time.UTC)
	a := RuntimeUsageCounter{ProjectID: "p1", AgentID: "a1", Provider: "openrouter", Model: "m1", UseCase: "chat", BillingMode: RuntimeBillingModeBYOK, TokensIn: 10, TokensOut: 20, CacheReadTokens: 3, Turns: 1}
	b := RuntimeUsageCounter{ProjectID: "p1", AgentID: "a2", Provider: "gemini", Model: "m2", UseCase: "chat", BillingMode: RuntimeBillingModeStandard, TokensIn: 30, TokensOut: 40, CacheCreationTokens: 2, Turns: 1}
	report := RuntimeUsageReport{SchemaVersion: 2, ReportID: "r1", TenantID: "t1", InstanceID: "i1", PeriodStart: now, PeriodEnd: now.Add(time.Minute), HeartbeatAt: now.Add(time.Minute), Counters: []RuntimeUsageCounter{a, b}}
	sig := SignRuntimeUsageReport(report, "secret")
	report.Counters = []RuntimeUsageCounter{b, a}
	if got := SignRuntimeUsageReport(report, "secret"); got != sig {
		t.Fatalf("signature changed with counter order: %s != %s", got, sig)
	}
	const expected = "4d209106751b9768c4e8afe82c544fbdfc43b83c3bb9fb42e79bb81765301308"
	if sig != expected {
		t.Fatalf("signature = %s, want shared contract vector %s", sig, expected)
	}
}

func TestRuntimeSubmitUsageReportMirrorsSignatureHeader(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/usage/reports" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("X-Sonzai-Metering-Signature"); got != "signed" {
			t.Fatalf("signature header = %q", got)
		}
		jsonResponse(w, http.StatusAccepted, RuntimeUsageReportResult{Accepted: true, ReportID: "r1"})
	}))

	result, err := client.Runtime.SubmitUsageReport(context.Background(), RuntimeUsageReport{SchemaVersion: 2, ReportID: "r1", Signature: "signed"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Accepted || result.ReportID != "r1" {
		t.Fatalf("result = %+v", result)
	}
}
