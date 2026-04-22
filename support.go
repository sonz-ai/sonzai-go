package sonzai

import (
	"context"
	"fmt"
)

// SupportResource provides support ticket operations.
type SupportResource struct {
	http *httpClient
}

// SupportTicket represents a support ticket.
type SupportTicket struct {
	TicketID        string             `json:"ticket_id"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Type            string             `json:"type,omitempty"`
	Priority        string             `json:"priority,omitempty"`
	Status          string             `json:"status,omitempty"`
	TenantID        string             `json:"tenant_id,omitempty"`
	CreatedBy       string             `json:"created_by,omitempty"`
	CreatedByEmail  string             `json:"created_by_email,omitempty"`
	AssignedTo      string             `json:"assigned_to,omitempty"`
	AssignedToEmail string             `json:"assigned_to_email,omitempty"`
	CommentCount    int                `json:"comment_count,omitempty"`
	Comments        []SupportTicketComment `json:"comments,omitempty"`
	CreatedAt       string             `json:"created_at,omitempty"`
	UpdatedAt       string             `json:"updated_at,omitempty"`
	ResolvedAt      string             `json:"resolved_at,omitempty"`
}

// SupportTicketComment represents a comment on a support ticket.
type SupportTicketComment struct {
	CommentID  string `json:"comment_id"`
	TicketID   string `json:"ticket_id"`
	Content    string `json:"content"`
	AuthorID   string `json:"author_id,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	AuthorType string `json:"author_type,omitempty"`
	IsInternal bool   `json:"is_internal,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

// TicketListResponse is the response from listing support tickets.
type TicketListResponse struct {
	Tickets []SupportTicket `json:"tickets"`
	Total   int             `json:"total"`
	HasMore bool            `json:"has_more"`
}

// TicketDetailResponse is the response from getting a single ticket.
type TicketDetailResponse struct {
	Ticket  *SupportTicket         `json:"ticket"`
	History []map[string]any       `json:"history,omitempty"`
}

// CreateTicketOptions configures a support ticket creation request.
type CreateTicketOptions struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type,omitempty"`
	Priority    string `json:"priority,omitempty"`
}

// AddTicketCommentOptions configures adding a comment to a ticket.
type AddTicketCommentOptions struct {
	Content    string `json:"content"`
	IsInternal bool   `json:"is_internal,omitempty"`
}

// SupportListOptions configures a list support tickets request.
type SupportListOptions struct {
	Limit  int    `url:"limit,omitempty"`
	Offset int    `url:"offset,omitempty"`
	Status string `url:"status,omitempty"`
	Type   string `url:"type,omitempty"`
}

// List returns all support tickets for the authenticated user.
func (s *SupportResource) List(ctx context.Context, opts *SupportListOptions) (*TicketListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
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

// Create opens a new support ticket.
func (s *SupportResource) Create(ctx context.Context, opts CreateTicketOptions) (*SupportTicket, error) {
	var result SupportTicket
	if err := s.http.Post(ctx, "/api/v1/support/tickets", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single support ticket with its full history.
func (s *SupportResource) Get(ctx context.Context, ticketID string) (*TicketDetailResponse, error) {
	var result TicketDetailResponse
	if err := s.http.Get(ctx, fmt.Sprintf("/api/v1/support/tickets/%s", ticketID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Close marks a support ticket as resolved.
func (s *SupportResource) Close(ctx context.Context, ticketID string) error {
	return s.http.Post(ctx, fmt.Sprintf("/api/v1/support/tickets/%s/close", ticketID), nil, nil)
}

// AddComment adds a comment to a support ticket.
func (s *SupportResource) AddComment(ctx context.Context, ticketID string, opts AddTicketCommentOptions) (*SupportTicketComment, error) {
	var result SupportTicketComment
	if err := s.http.Post(ctx, fmt.Sprintf("/api/v1/support/tickets/%s/comments", ticketID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
