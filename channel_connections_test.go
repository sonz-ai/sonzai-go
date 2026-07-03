package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestChannelConnectionsCRUD(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/channel-connections":
			jsonResponse(w, 200, ChannelConnectionListResponse{
				Connections: []ChannelConnection{{ConnectionID: "conn-1", ChannelType: "whatsapp"}},
				Items:       []ChannelConnection{{ConnectionID: "conn-1", ChannelType: "whatsapp"}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/channel-connections":
			var body CreateChannelConnectionOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.ProviderMode != ChannelProviderModeBYOApp || body.AppID != "app-1" || body.AppSecret != "secret" {
				t.Fatalf("unexpected create body: %+v", body)
			}
			if body.PhoneNumberID != "phone-1" || body.WABAID != "waba-1" || body.AccessToken != "token" {
				t.Fatalf("unexpected BYO fields: %+v", body)
			}
			jsonResponse(w, 200, ChannelConnection{ConnectionID: "conn-1", ProjectID: "project-1", ChannelType: "whatsapp"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/channel-connections/conn-1":
			jsonResponse(w, 200, ChannelConnection{ConnectionID: "conn-1", ProjectID: "project-1"})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/projects/project-1/channel-connections/conn-1":
			var body UpdateChannelConnectionOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.DefaultAgentID != "agent-1" || body.Status != "active" {
				t.Fatalf("unexpected update body: %+v", body)
			}
			jsonResponse(w, 200, ChannelConnection{ConnectionID: "conn-1", DefaultAgentID: "agent-1", Status: "active"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/channel-connections/conn-1/test":
			var body TestChannelConnectionOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.To != "user-1" || body.Message != "ping" {
				t.Fatalf("unexpected test body: %+v", body)
			}
			jsonResponse(w, 200, ChannelConnection{ConnectionID: "conn-1", TestSendSucceeded: true})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/project-1/channel-connections/conn-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	list, err := client.ChannelConnections.List(context.Background(), "project-1")
	if err != nil || len(list.Connections) != 1 {
		t.Fatalf("list: %+v %v", list, err)
	}
	created, err := client.ChannelConnections.Create(context.Background(), "project-1", CreateChannelConnectionOptions{
		ChannelType:   ChannelTypeWhatsApp,
		ProviderMode:  ChannelProviderModeBYOApp,
		DisplayName:   "Support",
		AppID:         "app-1",
		AppSecret:     "secret",
		PhoneNumberID: "phone-1",
		WABAID:        "waba-1",
		AccessToken:   "token",
		VerifyToken:   "verify",
	})
	if err != nil || created.ConnectionID != "conn-1" {
		t.Fatalf("create: %+v %v", created, err)
	}
	if _, err := client.ChannelConnections.Get(context.Background(), "project-1", "conn-1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if _, err := client.ChannelConnections.Update(context.Background(), "project-1", "conn-1", UpdateChannelConnectionOptions{DefaultAgentID: "agent-1", Status: "active"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := client.ChannelConnections.Test(context.Background(), "project-1", "conn-1", TestChannelConnectionOptions{To: "user-1", Message: "ping"}); err != nil {
		t.Fatalf("test: %v", err)
	}
	if err := client.ChannelConnections.Delete(context.Background(), "project-1", "conn-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestChannelConnectionStringRedactsSecrets(t *testing.T) {
	opts := CreateChannelConnectionOptions{
		ChannelType: ChannelTypeWhatsApp,
		AppSecret:   "app-secret",
		AccessToken: "access-token",
		VerifyToken: "verify-token",
	}
	got := fmt.Sprintf("%v %#v", opts, opts)
	for _, secret := range []string{"app-secret", "access-token", "verify-token"} {
		if strings.Contains(got, secret) {
			t.Fatalf("secret leaked in string output: %s", got)
		}
	}
	if !strings.Contains(got, redactedSecret) {
		t.Fatalf("expected redaction marker in string output: %s", got)
	}

	conn := ChannelConnection{ConnectionID: "conn-1", VerifyToken: "verify-token"}
	if strings.Contains(fmt.Sprintf("%v %#v", conn, conn), "verify-token") {
		t.Fatalf("verify token leaked in connection string output")
	}
}
