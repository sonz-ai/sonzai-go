package sonzai

import (
	"context"
	"strconv"
	"time"
)

// IngestResource provides the adapter-ingestion surface: a customer-owned
// adapter (running in the customer's environment, e.g. a CRM sidecar under
// app-runtime) normalizes its own events/contacts and POSTs them here, so the
// platform's pipelines, lead-assignment ledger, and outbound webhooks can
// react without the platform ever holding CRM tables.
type IngestResource struct {
	http *httpClient
}

// DomainEvent v1 closed event-type enum (POST /ingest/events' `type` field).
const (
	IngestEventLeadCreated      = "lead.created"
	IngestEventLeadUpdated      = "lead.updated"
	IngestEventLeadStageChanged = "lead.stage_changed"
	IngestEventInventoryUpdated = "inventory.updated"
	IngestEventPriceChanged     = "price.changed"
	IngestEventMessageReceived  = "message.received"
	IngestEventOutcomeRecorded  = "outcome.recorded"
)

// IngestEventOptions configures a POST /ingest/events call: one normalized
// DomainEvent v1 row.
type IngestEventOptions struct {
	// EventID is the adapter-chosen idempotency key (UUID); replaying the
	// same event_id into the same project is a no-op.
	EventID string `json:"event_id"`
	// Type is one of the closed DomainEvent v1 types (see the
	// IngestEvent* constants).
	Type string `json:"type"`
	// OccurredAt is when the event happened in the source system.
	OccurredAt time.Time `json:"occurred_at"`
	// LeadRef is a stable external lead identifier (CRM record id, …).
	LeadRef string `json:"lead_ref,omitempty"`
	// ContactRef is a stable external contact identifier (see UpsertContact).
	ContactRef string `json:"contact_ref,omitempty"`
	// Payload is the type-specific event body, stored verbatim and fanned
	// out to subscribers.
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// IngestEventResult is the outcome of storing one domain event.
type IngestEventResult struct {
	EventID string `json:"event_id"`
	Type    string `json:"type"`
	// Duplicate is true when this event_id was already stored for the
	// project — the replay was a no-op and no fan-out happened.
	Duplicate bool `json:"duplicate"`
}

// IngestContactOptions configures a POST /ingest/contacts call: an upsert
// keyed by ContactRef within the project.
type IngestContactOptions struct {
	// ContactRef is a stable external identifier the adapter owns; the
	// upsert key within the project.
	ContactRef string `json:"contact_ref"`
	// Kind is "rep" (the tenant's own salesperson) or "lead_contact" (an
	// end customer).
	Kind        string `json:"kind"`
	DisplayName string `json:"display_name,omitempty"`
	PhoneE164   string `json:"phone_e164,omitempty"`
	Email       string `json:"email,omitempty"`
	// CRMOwnerID is the CRM-side owner/user id (e.g. Salesforce OwnerId) for
	// write-back routing.
	CRMOwnerID string `json:"crm_owner_id,omitempty"`
	// Metadata is free-form registry metadata (channel identities, desk,
	// brand assignments, …).
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IngestContact is a stored contact/rep registry entry.
type IngestContact struct {
	ID          string                 `json:"id"`
	ContactRef  string                 `json:"contact_ref"`
	Kind        string                 `json:"kind"`
	DisplayName string                 `json:"display_name,omitempty"`
	PhoneE164   string                 `json:"phone_e164,omitempty"`
	Email       string                 `json:"email,omitempty"`
	CRMOwnerID  string                 `json:"crm_owner_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// IngestedEvent is one row of the ListEvents read path — the envelope minus
// the payload body (a completeness/gap check needs the keys, not the
// verbatim body).
type IngestedEvent struct {
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	LeadRef    string    `json:"lead_ref,omitempty"`
	ContactRef string    `json:"contact_ref,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListIngestEventsOptions configures a GET /ingest/events call.
type ListIngestEventsOptions struct {
	// Since is required: return events stored at/after this time.
	Since time.Time
	// Limit caps events per page (platform default 200, max 1000).
	Limit int
	// Cursor is an opaque pagination token from a prior page's NextCursor;
	// when set it overrides Since.
	Cursor string
}

// ListIngestEventsResult is a page of stored domain events.
type ListIngestEventsResult struct {
	Events []IngestedEvent `json:"events"`
	// NextCursor is set only when a full page was returned; feed it back as
	// Cursor to fetch the next page. Empty means the scan is exhausted.
	NextCursor string `json:"next_cursor,omitempty"`
}

// SendEvent stores a normalized DomainEvent v1 (idempotent on EventID per
// project) and fans it out to the project's notification channels under the
// event's own type.
func (i *IngestResource) SendEvent(ctx context.Context, opts IngestEventOptions) (*IngestEventResult, error) {
	var result IngestEventResult
	if err := i.http.Post(ctx, "/api/v1/ingest/events", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpsertContact creates or replaces a contact/rep registry entry keyed by
// ContactRef within the project (last write wins).
func (i *IngestResource) UpsertContact(ctx context.Context, opts IngestContactOptions) (*IngestContact, error) {
	var result IngestContact
	if err := i.http.Post(ctx, "/api/v1/ingest/contacts", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListEvents returns a project's stored events in keyset (created_at, id)
// order at/after Since, paged by Cursor. A read-only completeness/gap check
// for write-back adapters: webhook delivery is best-effort, so an adapter
// reconciles by listing platform-known events and re-applying any it has not
// yet synced.
func (i *IngestResource) ListEvents(ctx context.Context, opts ListIngestEventsOptions) (*ListIngestEventsResult, error) {
	params := map[string]string{
		"since": opts.Since.Format(time.RFC3339),
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.Cursor != "" {
		params["cursor"] = opts.Cursor
	}
	var result ListIngestEventsResult
	if err := i.http.Get(ctx, "/api/v1/ingest/events", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
