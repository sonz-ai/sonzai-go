package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// SupportTicketsResource provides support ticket operations for the
// authenticated user within their active tenant. All methods are scoped to
// the caller's Clerk session — callers can only see, create, comment on,
// and close tickets within their own tenant.
type SupportTicketsResource struct {
	http *httpClient
}

// SupportTicket represents a support ticket with its full detail.
type SupportTicket struct {
	TicketID        string                 `json:"ticket_id"`
	TenantID        string                 `json:"tenant_id"`
	CreatedBy       string                 `json:"created_by"`
	CreatedByEmail  string                 `json:"created_by_email"`
	AssignedTo      string                 `json:"assigned_to,omitempty"`
	AssignedToEmail string                 `json:"assigned_to_email,omitempty"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status"`
	Priority        string                 `json:"priority"`
	CommentCount    int64                  `json:"comment_count,omitempty"`
	Comments        []SupportTicketComment `json:"comments,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	ResolvedAt      string                 `json:"resolved_at,omitempty"`
}

// SupportTicketComment represents a single comment on a ticket thread.
// User-authored comments are always external (IsInternal=false); internal
// comments can only be created by staff via the admin portal.
type SupportTicketComment struct {
	CommentID   string `json:"comment_id"`
	TicketID    string `json:"ticket_id"`
	AuthorID    string `json:"author_id"`
	AuthorEmail string `json:"author_email"`
	AuthorType  string `json:"author_type"`
	Content     string `json:"content"`
	IsInternal  bool   `json:"is_internal"`
	CreatedAt   string `json:"created_at"`
}

// SupportTicketHistory represents a single audit entry on a ticket — one
// field change per record.
type SupportTicketHistory struct {
	HistoryID      string `json:"history_id"`
	TicketID       string `json:"ticket_id"`
	ChangedBy      string `json:"changed_by"`
	ChangedByEmail string `json:"changed_by_email"`
	FieldChanged   string `json:"field_changed"`
	OldValue       string `json:"old_value,omitempty"`
	NewValue       string `json:"new_value,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// TicketSummary is the trimmed ticket shape returned by list endpoints.
type TicketSummary struct {
	TicketID        string `json:"ticket_id"`
	Title           string `json:"title"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	Priority        string `json:"priority"`
	CreatedByEmail  string `json:"created_by_email"`
	AssignedToEmail string `json:"assigned_to_email,omitempty"`
	CommentCount    int64  `json:"comment_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// TicketListResponse is the response from listing the caller's tickets.
type TicketListResponse struct {
	Tickets []TicketSummary `json:"tickets"`
	Total   int64           `json:"total"`
	HasMore bool            `json:"has_more"`
}

// TicketDetailResponse is the response from fetching a single ticket — the
// ticket and its change history. Comments live on the embedded ticket.
type TicketDetailResponse struct {
	Ticket  SupportTicket          `json:"ticket"`
	History []SupportTicketHistory `json:"history,omitempty"`
}

// ListTicketsOptions configures a ticket list request. All fields are
// optional; zero values mean "server default" (limit=20, offset=0, no
// status/type filter).
type ListTicketsOptions struct {
	Limit  int    // items per page; server caps at 100
	Offset int    // pagination offset
	Status string // filter by status (open, in_progress, resolved, closed)
	Type   string // filter by type (support, bug, feature_request, billing, ...)
}

// CreateTicketRequest is the body for creating a support ticket. `Type`
// defaults to "support" and `Priority` defaults to "medium" when omitted.
type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Priority    string `json:"priority,omitempty"`
}

// AddCommentRequest is the body for appending a comment to a ticket.
// `IsInternal` is ignored for non-staff callers (always stored as false).
type AddCommentRequest struct {
	Content    string `json:"content"`
	IsInternal bool   `json:"is_internal,omitempty"`
}

// List returns tickets created by the authenticated user within their
// active tenant. Filter by status or type; paginate with limit/offset.
func (s *SupportTicketsResource) List(ctx context.Context, opts *ListTicketsOptions) (*TicketListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = strconv.Itoa(opts.Offset)
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.Type != "" {
			params["type"] = opts.Type
		}
	}
	var result TicketListResponse
	if err := s.http.Get(ctx, "/api/v1/support/tickets", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a support ticket in the caller's tenant.
func (s *SupportTicketsResource) Create(ctx context.Context, opts CreateTicketRequest) (*SupportTicket, error) {
	var result SupportTicket
	if err := s.http.Post(ctx, "/api/v1/support/tickets", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a ticket and its comment thread. Returns NotFoundError when
// the ticket belongs to a different tenant.
func (s *SupportTicketsResource) Get(ctx context.Context, ticketID string) (*TicketDetailResponse, error) {
	var result TicketDetailResponse
	if err := s.http.Get(ctx, fmt.Sprintf("/api/v1/support/tickets/%s", ticketID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Close closes the ticket if the caller is its original creator. Returns
// PermissionDeniedError when the caller is in the tenant but did not
// create the ticket.
func (s *SupportTicketsResource) Close(ctx context.Context, ticketID string) (*SupportTicket, error) {
	var result SupportTicket
	if err := s.http.Post(ctx, fmt.Sprintf("/api/v1/support/tickets/%s/close", ticketID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddComment appends a user comment to the ticket thread. User comments
// are always external — the server ignores IsInternal for non-staff
// callers.
func (s *SupportTicketsResource) AddComment(ctx context.Context, ticketID string, opts AddCommentRequest) (*SupportTicketComment, error) {
	var result SupportTicketComment
	if err := s.http.Post(ctx, fmt.Sprintf("/api/v1/support/tickets/%s/comments", ticketID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
