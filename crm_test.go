package sonzai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRuntimeTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	runtimeServer := httptest.NewServer(handler)
	t.Cleanup(runtimeServer.Close)

	platformServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected platform request: %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(platformServer.Close)

	client, err := NewClient("adapter-token", WithBaseURL(platformServer.URL), WithRuntimeBaseURL(runtimeServer.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func TestCrmImport_UsesRuntimeBaseURLAndTenantHeader(t *testing.T) {
	var seen struct {
		path, method, auth, tenant string
		body                       struct {
			Contacts []CrmImportItem `json:"contacts"`
		}
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		seen.auth = r.Header.Get("Authorization")
		seen.tenant = r.Header.Get(crmTenantHeader)
		if err := json.NewDecoder(r.Body).Decode(&seen.body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"imported":1,"contacts":[{"id":"c1","external_ref":"sf-1","emails":["grace@example.com"],"phones":[],"custom":{},"archived":false,"created_at":"2026-07-10T04:00:00Z","updated_at":"2026-07-10T04:00:00Z"}]}`))
	})
	client := newRuntimeTestClient(t, h)

	result, err := client.Crm.Import(context.Background(), []CrmImportItem{{
		ExternalRef: "sf-1",
		FirstName:   "Grace",
		Emails:      json.RawMessage(`["grace@example.com"]`),
	}}, &CrmImportOptions{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if seen.method != http.MethodPost || seen.path != "/api/rt/crm/import" {
		t.Fatalf("method/path = %s %s", seen.method, seen.path)
	}
	if seen.auth != "Bearer adapter-token" {
		t.Fatalf("auth = %q", seen.auth)
	}
	if seen.tenant != "tenant-a" {
		t.Fatalf("tenant header = %q", seen.tenant)
	}
	if len(seen.body.Contacts) != 1 || seen.body.Contacts[0].ExternalRef != "sf-1" {
		t.Fatalf("body = %+v", seen.body)
	}
	if result.Imported != 1 || len(result.Contacts) != 1 || result.Contacts[0].ID != "c1" {
		t.Fatalf("result = %+v", result)
	}
}

func TestCrmEvents_QueryTenantAndDecode(t *testing.T) {
	var seen struct {
		path, query, tenant string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.query = r.URL.RawQuery
		seen.tenant = r.Header.Get(crmTenantHeader)
		_, _ = w.Write([]byte(`{"events":[{"cursor":"42","tenant_id":"tenant-a","event":"contact.updated","entity_type":"contact","entity_id":"c1","payload":{"id":"c1"},"at":"2026-07-10T04:00:00Z"}],"next_cursor":"42"}`))
	})
	client := newRuntimeTestClient(t, h)

	result, err := client.Crm.Events(context.Background(), &CrmEventsOptions{
		Cursor:   "41",
		Limit:    25,
		TenantID: "tenant-a",
	})
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if seen.path != "/api/rt/crm/events" || seen.query != "cursor=41&limit=25" {
		t.Fatalf("path/query = %s %s", seen.path, seen.query)
	}
	if seen.tenant != "tenant-a" {
		t.Fatalf("tenant header = %q", seen.tenant)
	}
	if result.NextCursor != "42" || len(result.Events) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.Events[0].EventType != "contact.updated" || string(result.Events[0].Payload) != `{"id":"c1"}` {
		t.Fatalf("event = %+v", result.Events[0])
	}
}

func TestCrmEventIterator_PaginatesAndReturnsEOFOnEmptyPoll(t *testing.T) {
	var calls int
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			if got := r.URL.RawQuery; got != "limit=2" {
				t.Fatalf("first query = %q", got)
			}
			_, _ = w.Write([]byte(`{"events":[{"cursor":"1","event":"contact.created","payload":{},"at":"2026-07-10T04:00:00Z"},{"cursor":"2","event":"deal.updated","payload":{},"at":"2026-07-10T04:01:00Z"}],"next_cursor":"2"}`))
		case 2:
			if got := r.URL.RawQuery; got != "cursor=2&limit=2" {
				t.Fatalf("second query = %q", got)
			}
			_, _ = w.Write([]byte(`{"events":[],"next_cursor":"2"}`))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
	})
	client := newRuntimeTestClient(t, h)
	iter := client.Crm.EventIterator(CrmEventsOptions{Limit: 2})

	first, err := iter.Next(context.Background())
	if err != nil {
		t.Fatalf("Next first: %v", err)
	}
	second, err := iter.Next(context.Background())
	if err != nil {
		t.Fatalf("Next second: %v", err)
	}
	if first.EventType != "contact.created" || second.EventType != "deal.updated" || iter.Cursor() != "2" {
		t.Fatalf("events/cursor = %+v %+v %q", first, second, iter.Cursor())
	}
	_, err = iter.Next(context.Background())
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestCrmRequiresRuntimeBaseURL(t *testing.T) {
	client := MustNewClient("adapter-token")
	_, err := client.Crm.Events(context.Background(), nil)
	if !errors.Is(err, ErrRuntimeBaseURLRequired) {
		t.Fatalf("expected ErrRuntimeBaseURLRequired, got %v", err)
	}
}
