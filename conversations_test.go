package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestConversationsList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/conversations" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("channel") != "whatsapp" || q.Get("agent_id") != "agent-1" || q.Get("user_id") != "user-1" {
			t.Fatalf("unexpected filters: %s", r.URL.RawQuery)
		}
		if q.Get("controller") != "human" || q.Get("status") != "open" || q.Get("q") != "billing" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		if q.Get("cursor") != "cur-1" || q.Get("limit") != "25" {
			t.Fatalf("unexpected pagination: %s", r.URL.RawQuery)
		}
		jsonResponse(w, 200, ConversationListResponse{
			Conversations: []ConversationListItem{{ID: "conv-1", Agent: "agent-1", Channel: "whatsapp"}},
			Items:         []ConversationListItem{{ID: "conv-1", Agent: "agent-1", Channel: "whatsapp"}},
			HasMore:       true,
			NextCursor:    "cur-2",
			Total:         1,
		})
	})
	defer server.Close()

	result, err := client.Conversations.List(context.Background(), &ConversationListOptions{
		Channel:    "whatsapp",
		AgentID:    "agent-1",
		UserID:     "user-1",
		Controller: "human",
		Status:     "open",
		Query:      "billing",
		Cursor:     "cur-1",
		Limit:      25,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NextCursor != "cur-2" || !result.HasMore {
		t.Fatalf("unexpected pagination response: %+v", result)
	}
	if len(result.Conversations) != 1 || result.Conversations[0].ID != "conv-1" {
		t.Fatalf("unexpected conversations: %+v", result.Conversations)
	}
}

func TestConversationActions(t *testing.T) {
	var sawTakeover, sawSend, sawUpdate bool
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conversations/conv-1":
			jsonResponse(w, 200, ConversationDetailResponse{
				Conversation: Conversation{ConversationID: "conv-1", Controller: "agent"},
				Source:       "omnichannel",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conversations/conv-1/messages":
			if r.URL.Query().Get("cursor") != "cur-1" || r.URL.Query().Get("limit") != "10" {
				t.Fatalf("unexpected message query: %s", r.URL.RawQuery)
			}
			jsonResponse(w, 200, ConversationMessagesResponse{
				Messages: []ConversationMessage{{MessageID: "msg-1", ConversationID: "conv-1", Content: "hi"}},
				Items:    []ConversationMessage{{MessageID: "msg-1", ConversationID: "conv-1", Content: "hi"}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/conversations/conv-1/takeover":
			sawTakeover = true
			if r.URL.Query().Get("operator_id") != "op-1" || r.URL.Query().Get("force") != "true" {
				t.Fatalf("unexpected takeover query: %s", r.URL.RawQuery)
			}
			jsonResponse(w, 200, Conversation{ConversationID: "conv-1", Controller: "human"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/conversations/conv-1/takeover":
			jsonResponse(w, 200, Conversation{ConversationID: "conv-1", Controller: "agent"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/conversations/conv-1/messages":
			sawSend = true
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["content"] != "hello" {
				t.Fatalf("unexpected send body: %+v", body)
			}
			jsonResponse(w, 200, Conversation{ConversationID: "conv-1", LastMessagePreview: "hello"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/conversations/conv-1/read":
			jsonResponse(w, 200, Conversation{ConversationID: "conv-1", UnreadCount: 0})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/conversations/conv-1":
			sawUpdate = true
			var body UpdateConversationOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Status != "closed" || body.AgentID != "agent-2" {
				t.Fatalf("unexpected update body: %+v", body)
			}
			jsonResponse(w, 200, Conversation{ConversationID: "conv-1", Status: "closed", AgentID: "agent-2"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	detail, err := client.Conversations.Get(context.Background(), "conv-1")
	if err != nil || detail.Conversation.ConversationID != "conv-1" {
		t.Fatalf("get: %+v %v", detail, err)
	}
	messages, err := client.Conversations.Messages(context.Background(), "conv-1", &ConversationMessagesOptions{Limit: 10, Cursor: "cur-1"})
	if err != nil || len(messages.Messages) != 1 {
		t.Fatalf("messages: %+v %v", messages, err)
	}
	if _, err := client.Conversations.TakeOver(context.Background(), "conv-1", &TakeOverConversationOptions{OperatorID: "op-1", Force: true}); err != nil {
		t.Fatalf("takeover: %v", err)
	}
	if _, err := client.Conversations.Release(context.Background(), "conv-1"); err != nil {
		t.Fatalf("release: %v", err)
	}
	if _, err := client.Conversations.SendAsAgent(context.Background(), "conv-1", SendConversationMessageOptions{Content: "hello"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if _, err := client.Conversations.MarkRead(context.Background(), "conv-1"); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if _, err := client.Conversations.Update(context.Background(), "conv-1", UpdateConversationOptions{Status: "closed", AgentID: "agent-2"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if !sawTakeover || !sawSend || !sawUpdate {
		t.Fatalf("missing action calls: takeover=%v send=%v update=%v", sawTakeover, sawSend, sawUpdate)
	}
}

func TestConversationsStream(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/conversations/stream" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("project_id") != "project-1" {
			t.Fatalf("unexpected project_id: %s", r.URL.RawQuery)
		}
		sseResponse(w,
			`{"type":"conversation.message","conversation":{"conversation_id":"conv-1"},"message":{"message_id":"msg-1","content":"hi"}}`,
		)
	})
	defer server.Close()

	var events []ConversationStreamEvent
	err := client.Conversations.Stream(context.Background(), &ConversationStreamOptions{ProjectID: "project-1"}, func(event ConversationStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 || events[0].Type != "conversation.message" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestConversationsPush(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/conversations/push" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body PushMessageOptions
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.AgentID != "agent-1" || body.UserID != "user-1" || body.Content != "New lead: score 82 Hot" || body.ChannelType != "whatsapp" {
			t.Fatalf("unexpected push body: %+v", body)
		}
		jsonResponse(w, 200, PushMessageResult{
			ConversationID: "conv-1",
			ChannelType:    "whatsapp",
			ExternalID:     "+639171234567",
			DeliveryStatus: "sent",
			SessionID:      "sess-1",
			UsedTemplate:   true,
		})
	})
	defer server.Close()

	result, err := client.Conversations.Push(context.Background(), PushMessageOptions{
		AgentID:     "agent-1",
		UserID:      "user-1",
		Content:     "New lead: score 82 Hot",
		ChannelType: "whatsapp",
	})
	if err != nil {
		t.Fatalf("Push: %v", err)
	}
	if result.ConversationID != "conv-1" || result.DeliveryStatus != "sent" || !result.UsedTemplate {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestWebhookConversationEventConstants(t *testing.T) {
	events := []string{
		WebhookEventConversationStarted,
		WebhookEventConversationMessage,
		WebhookEventConversationTakeoverStarted,
		WebhookEventConversationTakeoverReleased,
		WebhookEventConversationMessageFailed,
		WebhookEventConversationUnrouted,
	}
	for _, event := range events {
		if !strings.HasPrefix(event, "conversation.") {
			t.Fatalf("unexpected event constant: %s", event)
		}
	}
}
