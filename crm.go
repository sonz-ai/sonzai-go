package sonzai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

const crmTenantHeader = "X-Sonzai-Tenant-ID"

// CrmResource provides adapter-token access to a deployed runtime-local CRM.
// Configure the runtime target with WithRuntimeBaseURL or SONZAI_RUNTIME_BASE_URL.
//
// The runtime's staff CRM CRUD routes are protected by browser session cookies
// and are intentionally not exposed here. This resource mirrors the adapter
// surface mounted behind Authorization: Bearer ADAPTER_TOKEN.
type CrmResource struct {
	http *httpClient
}

// CrmContact mirrors the runtime CRM contact JSON shape.
type CrmContact struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id,omitempty"`
	ProjectID   string          `json:"project_id,omitempty"`
	FirstName   string          `json:"first_name,omitempty"`
	LastName    string          `json:"last_name,omitempty"`
	Emails      json.RawMessage `json:"emails"`
	Phones      json.RawMessage `json:"phones"`
	LeadRef     string          `json:"lead_ref,omitempty"`
	OwnerUserID string          `json:"owner_user_id,omitempty"`
	Source      string          `json:"source,omitempty"`
	ExternalRef string          `json:"external_ref,omitempty"`
	Custom      json.RawMessage `json:"custom"`
	Archived    bool            `json:"archived"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

// CrmCompany mirrors the runtime CRM company JSON shape.
type CrmCompany struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenant_id,omitempty"`
	ProjectID string          `json:"project_id,omitempty"`
	Name      string          `json:"name"`
	Domain    string          `json:"domain,omitempty"`
	Custom    json.RawMessage `json:"custom"`
	Archived  bool            `json:"archived"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

// CrmPipeline mirrors the runtime CRM pipeline JSON shape.
type CrmPipeline struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenant_id,omitempty"`
	ProjectID string     `json:"project_id,omitempty"`
	Name      string     `json:"name"`
	IsDefault bool       `json:"is_default"`
	Archived  bool       `json:"archived"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CrmStage mirrors the runtime CRM stage JSON shape.
type CrmStage struct {
	ID         string     `json:"id"`
	PipelineID string     `json:"pipeline_id"`
	TenantID   string     `json:"tenant_id,omitempty"`
	Name       string     `json:"name"`
	Kind       string     `json:"kind"`
	SortOrder  int        `json:"sort_order"`
	Archived   bool       `json:"archived"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// CrmDeal mirrors the runtime CRM deal JSON shape.
type CrmDeal struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id,omitempty"`
	ProjectID     string          `json:"project_id,omitempty"`
	ContactID     string          `json:"contact_id,omitempty"`
	CompanyID     string          `json:"company_id,omitempty"`
	PipelineID    string          `json:"pipeline_id"`
	StageID       string          `json:"stage_id"`
	CatalogItemID string          `json:"catalog_item_id,omitempty"`
	ValueCents    *int64          `json:"value_cents,omitempty"`
	Currency      string          `json:"currency,omitempty"`
	OwnerUserID   string          `json:"owner_user_id,omitempty"`
	LeadRef       string          `json:"lead_ref,omitempty"`
	Custom        json.RawMessage `json:"custom"`
	Archived      bool            `json:"archived"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty"`
}

// CrmDealStageHistory mirrors the runtime CRM deal stage history JSON shape.
type CrmDealStageHistory struct {
	ID            string    `json:"id"`
	DealID        string    `json:"deal_id"`
	TenantID      string    `json:"tenant_id,omitempty"`
	FromStageID   string    `json:"from_stage_id,omitempty"`
	ToStageID     string    `json:"to_stage_id"`
	MovedByUserID string    `json:"moved_by_user_id,omitempty"`
	MovedAt       time.Time `json:"moved_at"`
}

// CrmActivity mirrors the runtime CRM activity JSON shape.
type CrmActivity struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id,omitempty"`
	ProjectID    string          `json:"project_id,omitempty"`
	Kind         string          `json:"kind"`
	ContactID    string          `json:"contact_id,omitempty"`
	DealID       string          `json:"deal_id,omitempty"`
	Body         string          `json:"body,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	DueAt        *time.Time      `json:"due_at,omitempty"`
	DoneAt       *time.Time      `json:"done_at,omitempty"`
	AuthorUserID string          `json:"author_user_id,omitempty"`
	Archived     bool            `json:"archived"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
}

// CrmCustomField mirrors the runtime CRM custom field registry JSON shape.
type CrmCustomField struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenant_id,omitempty"`
	ObjectType string          `json:"object_type"`
	FieldKey   string          `json:"field_key"`
	Label      string          `json:"label"`
	FieldType  string          `json:"field_type"`
	Options    json.RawMessage `json:"options,omitempty"`
	Required   bool            `json:"required"`
	Archived   bool            `json:"archived"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  *time.Time      `json:"deleted_at,omitempty"`
}

// CrmEvent is one ordered runtime CRM change-feed event.
type CrmEvent struct {
	Cursor     string          `json:"cursor"`
	TenantID   string          `json:"tenant_id,omitempty"`
	EventType  string          `json:"event"`
	EntityID   string          `json:"entity_id,omitempty"`
	EntityType string          `json:"entity_type,omitempty"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"at"`
}

// CrmEventEnvelope is the cursor-paginated response from GET /api/rt/crm/events.
type CrmEventEnvelope struct {
	Events     []CrmEvent `json:"events"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

// CrmImportItem is one contact to bulk upsert into the runtime CRM. ExternalRef
// is required by the runtime and makes import idempotent.
type CrmImportItem struct {
	ProjectID   string          `json:"project_id,omitempty"`
	ExternalRef string          `json:"external_ref"`
	FirstName   string          `json:"first_name,omitempty"`
	LastName    string          `json:"last_name,omitempty"`
	Emails      json.RawMessage `json:"emails,omitempty"`
	Phones      json.RawMessage `json:"phones,omitempty"`
	LeadRef     string          `json:"lead_ref,omitempty"`
	OwnerUserID string          `json:"owner_user_id,omitempty"`
	Source      string          `json:"source,omitempty"`
	Custom      json.RawMessage `json:"custom,omitempty"`
}

// CrmImportResult is returned by the runtime CRM bulk import route.
type CrmImportResult struct {
	Imported int          `json:"imported"`
	Contacts []CrmContact `json:"contacts"`
}

// CrmImportOptions configures a runtime CRM import request.
type CrmImportOptions struct {
	// TenantID is sent as X-Sonzai-Tenant-ID for managed/shared runtimes where
	// TENANT_ID is not pinned in the app-runtime process.
	TenantID string
}

// CrmEventsOptions configures a runtime CRM events request.
type CrmEventsOptions struct {
	// Cursor is the previous response's NextCursor. The runtime currently uses
	// integer event IDs as cursors.
	Cursor string

	// Limit caps the number of events returned. The runtime defaults to 100 and
	// clamps invalid or oversized values to 100.
	Limit int

	// TenantID is sent as X-Sonzai-Tenant-ID for managed/shared runtimes where
	// TENANT_ID is not pinned in the app-runtime process.
	TenantID string
}

// ErrRuntimeBaseURLRequired is returned when Crm is used without configuring a
// runtime target via WithRuntimeBaseURL or SONZAI_RUNTIME_BASE_URL.
var ErrRuntimeBaseURLRequired = errors.New("sonzai: runtime base URL is required for Crm; configure WithRuntimeBaseURL or SONZAI_RUNTIME_BASE_URL")

// Import bulk upserts contacts into the runtime CRM. The runtime is idempotent
// by external_ref and returns the resulting contacts.
func (c *CrmResource) Import(ctx context.Context, contacts []CrmImportItem, opts *CrmImportOptions) (*CrmImportResult, error) {
	if err := c.requireRuntime(); err != nil {
		return nil, err
	}
	body := struct {
		Contacts []CrmImportItem `json:"contacts"`
	}{Contacts: contacts}
	var result CrmImportResult
	if err := c.http.PostWithHeaders(ctx, "/api/rt/crm/import", body, crmHeaders(importTenantID(opts)), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportContacts is an alias for Import.
func (c *CrmResource) ImportContacts(ctx context.Context, contacts []CrmImportItem, opts *CrmImportOptions) (*CrmImportResult, error) {
	return c.Import(ctx, contacts, opts)
}

// Events returns one cursor-paginated page from the runtime CRM change feed.
// Call repeatedly with the previous response's NextCursor to continue polling.
func (c *CrmResource) Events(ctx context.Context, opts *CrmEventsOptions) (*CrmEventEnvelope, error) {
	if err := c.requireRuntime(); err != nil {
		return nil, err
	}
	params := map[string]string{}
	if opts != nil {
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}
	var result CrmEventEnvelope
	if err := c.http.GetWithHeaders(ctx, "/api/rt/crm/events", params, crmHeaders(eventsTenantID(opts)), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EventIterator returns a pull iterator over CRM events. Next returns io.EOF
// when the current poll returns no events; callers should persist Cursor()
// between polls for write-back sync loops.
func (c *CrmResource) EventIterator(opts CrmEventsOptions) *CrmEventIterator {
	return &CrmEventIterator{crm: c, opts: opts}
}

// CrmEventIterator is a pull iterator over the runtime CRM change feed.
type CrmEventIterator struct {
	crm  *CrmResource
	opts CrmEventsOptions
	page *CrmEventEnvelope
	idx  int
}

// Cursor returns the latest cursor observed by the iterator.
func (it *CrmEventIterator) Cursor() string {
	return it.opts.Cursor
}

// Next returns the next CRM event or io.EOF when the current poll has no more
// events. Calling Next again after io.EOF polls from the current cursor.
func (it *CrmEventIterator) Next(ctx context.Context) (*CrmEvent, error) {
	for it.page == nil || it.idx >= len(it.page.Events) {
		page, err := it.crm.Events(ctx, &it.opts)
		if err != nil {
			return nil, err
		}
		it.page = page
		it.idx = 0
		if page.NextCursor != "" {
			it.opts.Cursor = page.NextCursor
		}
		if len(page.Events) == 0 {
			return nil, io.EOF
		}
	}
	event := it.page.Events[it.idx]
	it.idx++
	if event.Cursor != "" {
		it.opts.Cursor = event.Cursor
	}
	return &event, nil
}

func (c *CrmResource) requireRuntime() error {
	if c == nil || c.http == nil {
		return ErrRuntimeBaseURLRequired
	}
	return nil
}

func crmHeaders(tenantID string) map[string]string {
	if tenantID == "" {
		return nil
	}
	return map[string]string{crmTenantHeader: tenantID}
}

func importTenantID(opts *CrmImportOptions) string {
	if opts == nil {
		return ""
	}
	return opts.TenantID
}

func eventsTenantID(opts *CrmEventsOptions) string {
	if opts == nil {
		return ""
	}
	return opts.TenantID
}

func (e *CrmEvent) String() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s %s %s", e.Cursor, e.EventType, e.EntityID)
}
