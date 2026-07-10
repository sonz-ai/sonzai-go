package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRoutingResourceConfigAndPermanentRoute(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/projects/project-1/routing-config":
			var body RoutingConfig
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.GuideAgent.AgentID != "guide-1" {
				t.Fatalf("body = %+v", body)
			}
			jsonResponse(w, http.StatusOK, body)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/permanent-routes/user-1/override":
			jsonResponse(w, http.StatusOK, PermanentRoute{UserID: "user-1", OverrideAgentID: "agent-2", Overridden: true})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))

	cfg, err := client.Routing.PutConfig(context.Background(), "project-1", RoutingConfig{GuideAgent: RoutingGuideAgent{AgentID: "guide-1"}})
	if err != nil || cfg.GuideAgent.AgentID != "guide-1" {
		t.Fatalf("PutConfig = %+v, %v", cfg, err)
	}
	route, err := client.Routing.OverridePermanentRoute(context.Background(), "project-1", "user-1", "agent-2")
	if err != nil || !route.Overridden || route.OverrideAgentID != "agent-2" {
		t.Fatalf("OverridePermanentRoute = %+v, %v", route, err)
	}
}
