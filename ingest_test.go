package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestIngestSendEvent(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/ingest/events":
			var body IngestEventOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.EventID != "11111111-1111-1111-1111-111111111111" || body.Type != IngestEventLeadCreated || body.LeadRef != "lead-1" {
				t.Fatalf("unexpected event body: %+v", body)
			}
			if body.Payload["source"] != "salesforce" {
				t.Fatalf("unexpected payload: %+v", body.Payload)
			}
			jsonResponse(w, 200, IngestEventResult{EventID: body.EventID, Type: body.Type, Duplicate: false})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.Ingest.SendEvent(context.Background(), IngestEventOptions{
		EventID:    "11111111-1111-1111-1111-111111111111",
		Type:       IngestEventLeadCreated,
		OccurredAt: time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC),
		LeadRef:    "lead-1",
		Payload:    map[string]interface{}{"source": "salesforce"},
	})
	if err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	if result.Duplicate || result.Type != IngestEventLeadCreated {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestIngestUpsertContact(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/ingest/contacts":
			var body IngestContactOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.ContactRef != "rep-1" || body.Kind != "rep" || body.CRMOwnerID != "sf-owner-1" {
				t.Fatalf("unexpected contact body: %+v", body)
			}
			jsonResponse(w, 200, IngestContact{
				ID:         "contact-1",
				ContactRef: "rep-1",
				Kind:       "rep",
				CRMOwnerID: "sf-owner-1",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.Ingest.UpsertContact(context.Background(), IngestContactOptions{
		ContactRef: "rep-1",
		Kind:       "rep",
		CRMOwnerID: "sf-owner-1",
	})
	if err != nil {
		t.Fatalf("UpsertContact: %v", err)
	}
	if result.ID != "contact-1" || result.ContactRef != "rep-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestIngestListEvents(t *testing.T) {
	since := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/ingest/events":
			q := r.URL.Query()
			if q.Get("since") != since.Format(time.RFC3339) || q.Get("limit") != "50" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			jsonResponse(w, 200, ListIngestEventsResult{
				Events:     []IngestedEvent{{EventID: "evt-1", Type: IngestEventOutcomeRecorded, LeadRef: "lead-1"}},
				NextCursor: "cursor-2",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.Ingest.ListEvents(context.Background(), ListIngestEventsOptions{Since: since, Limit: 50})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(result.Events) != 1 || result.Events[0].EventID != "evt-1" || result.NextCursor != "cursor-2" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
