package sonzai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
)

// Pin URL shapes, HTTP verbs, query params, and body payloads for the
// Support Tickets resource so the Go SDK stays in sync with the Platform
// API handlers. Response parsing is smoke-tested by decoding key fields.

func TestListSupportTickets_URLAndParams(t *testing.T) {
	var seen struct {
		path, method, rawQuery string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		seen.rawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(TicketListResponse{
			Tickets: []TicketSummary{{TicketID: "t-1", Title: "Login broken", Status: "open"}},
			Total:   1,
			HasMore: false,
		})
	})
	client := newTestClient(t, h)

	resp, err := client.SupportTickets.List(context.Background(), &ListTicketsOptions{
		Limit:  25,
		Offset: 10,
		Status: "open",
		Type:   "bug",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/support/tickets" {
		t.Errorf("path: got %q, want /api/v1/support/tickets", seen.path)
	}
	// Order within the encoded query is not guaranteed — inspect individually.
	q := parseQuery(t, seen.rawQuery)
	if q.Get("limit") != "25" {
		t.Errorf("limit: got %q, want 25", q.Get("limit"))
	}
	if q.Get("offset") != "10" {
		t.Errorf("offset: got %q, want 10", q.Get("offset"))
	}
	if q.Get("status") != "open" {
		t.Errorf("status: got %q, want open", q.Get("status"))
	}
	if q.Get("type") != "bug" {
		t.Errorf("type: got %q, want bug", q.Get("type"))
	}
	if resp.Total != 1 || len(resp.Tickets) != 1 || resp.Tickets[0].TicketID != "t-1" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestListSupportTickets_NoOptionsOmitsQuery(t *testing.T) {
	var seen string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(TicketListResponse{Tickets: []TicketSummary{}, Total: 0, HasMore: false})
	})
	client := newTestClient(t, h)

	if _, err := client.SupportTickets.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if seen != "" {
		t.Errorf("expected empty query string when opts is nil, got %q", seen)
	}
}

func TestCreateSupportTicket_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &seen.body)
		_ = json.NewEncoder(w).Encode(SupportTicket{
			TicketID:    "t-new",
			Title:       "Login broken",
			Description: "Cannot log in",
			Type:        "bug",
			Status:      "open",
			Priority:    "medium",
		})
	})
	client := newTestClient(t, h)

	got, err := client.SupportTickets.Create(context.Background(), CreateTicketRequest{
		Title:       "Login broken",
		Description: "Cannot log in",
		Type:        "bug",
		Priority:    "high",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/support/tickets" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.body["title"] != "Login broken" || seen.body["description"] != "Cannot log in" ||
		seen.body["type"] != "bug" || seen.body["priority"] != "high" {
		t.Errorf("body missing fields: %+v", seen.body)
	}
	if got.TicketID != "t-new" {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestGetSupportTicket_URL(t *testing.T) {
	var seen struct {
		path, method string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(TicketDetailResponse{
			Ticket: SupportTicket{
				TicketID: "t-42",
				Title:    "Refund",
				Status:   "open",
				Comments: []SupportTicketComment{
					{CommentID: "c-1", Content: "Thanks, will look into it.", IsInternal: false},
				},
			},
			History: []SupportTicketHistory{
				{HistoryID: "h-1", FieldChanged: "status", OldValue: "open", NewValue: "in_progress"},
			},
		})
	})
	client := newTestClient(t, h)

	got, err := client.SupportTickets.Get(context.Background(), "t-42")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/support/tickets/t-42" {
		t.Errorf("path: got %q", seen.path)
	}
	if got.Ticket.TicketID != "t-42" || len(got.Ticket.Comments) != 1 || len(got.History) != 1 {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestCloseSupportTicket_URL(t *testing.T) {
	var seen struct {
		path, method string
		bodyLen      int
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		b, _ := io.ReadAll(r.Body)
		seen.bodyLen = len(b)
		_ = json.NewEncoder(w).Encode(SupportTicket{TicketID: "t-42", Status: "closed"})
	})
	client := newTestClient(t, h)

	got, err := client.SupportTickets.Close(context.Background(), "t-42")
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/support/tickets/t-42/close" {
		t.Errorf("path: got %q", seen.path)
	}
	if got.Status != "closed" {
		t.Errorf("unexpected status: %+v", got)
	}
}

func TestAddSupportTicketComment_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &seen.body)
		_ = json.NewEncoder(w).Encode(SupportTicketComment{
			CommentID:  "c-new",
			TicketID:   "t-42",
			AuthorType: "user",
			Content:    "Still broken on iOS.",
			IsInternal: false,
		})
	})
	client := newTestClient(t, h)

	got, err := client.SupportTickets.AddComment(context.Background(), "t-42", AddCommentRequest{
		Content: "Still broken on iOS.",
	})
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/support/tickets/t-42/comments" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.body["content"] != "Still broken on iOS." {
		t.Errorf("body.content: got %v", seen.body["content"])
	}
	if got.CommentID != "c-new" || got.IsInternal {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestSupportTicketsResource_WiredOnClient(t *testing.T) {
	c := MustNewClient("test-key")
	if c.SupportTickets == nil {
		t.Fatal("SupportTickets is nil")
	}
}

// parseQuery parses the raw query string and fails the test on error.
func parseQuery(t *testing.T, raw string) url.Values {
	t.Helper()
	v, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return v
}
