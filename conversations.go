package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"sync/atomic"
)

// ConversationsResource provides omnichannel conversation operations.
type ConversationsResource struct {
	http *httpClient
}

// Conversation represents a durable omnichannel conversation.
type Conversation struct {
	ConversationID       string          `json:"conversation_id"`
	ProjectID            string          `json:"project_id"`
	AgentID              string          `json:"agent_id"`
	UserID               string          `json:"user_id"`
	ChannelType          string          `json:"channel_type"`
	ConnectionID         string          `json:"connection_id,omitempty"`
	SessionID            string          `json:"session_id,omitempty"`
	Controller           string          `json:"controller"`
	ControllerOperatorID string          `json:"controller_operator_id,omitempty"`
	Status               string          `json:"status"`
	LastMessageAt        string          `json:"last_message_at"`
	LastMessagePreview   string          `json:"last_message_preview,omitempty"`
	LastDirection        string          `json:"last_direction,omitempty"`
	UnreadCount          int             `json:"unread_count"`
	TakeoverStartedAt    string          `json:"takeover_started_at,omitempty"`
	HandoffsJSON         json.RawMessage `json:"handoffs,omitempty"`
	MetaJSON             json.RawMessage `json:"meta,omitempty"`
	CreatedAt            string          `json:"created_at"`
	UpdatedAt            string          `json:"updated_at"`
}

// ConversationListItem represents a conversation row from the list endpoint.
// The Platform may return durable omnichannel rows or legacy memory-timeline
// rows; this struct mirrors the list response shape.
type ConversationListItem struct {
	ID            string                 `json:"id"`
	Agent         string                 `json:"agent"`
	AgentName     string                 `json:"agent_name"`
	Model         string                 `json:"model"`
	Status        string                 `json:"status"`
	Title         string                 `json:"title"`
	Channel       string                 `json:"channel"`
	Tier          string                 `json:"tier"`
	LastMessage   string                 `json:"last_message"`
	LastActivity  string                 `json:"last_activity"`
	CreatedAt     string                 `json:"created_at"`
	CostUSD       float64                `json:"cost_usd"`
	Tags          []string               `json:"tags"`
	Handoffs      []ConversationHandoff  `json:"handoffs"`
	ExtraMetadata map[string]interface{} `json:"-"`
}

// ConversationHandoff records a controller handoff in a legacy list row.
type ConversationHandoff struct {
	From string `json:"from"`
	To   string `json:"to"`
	When string `json:"when"`
}

// ConversationMessage represents a message in an omnichannel conversation.
type ConversationMessage struct {
	MessageID        string          `json:"message_id"`
	ConversationID   string          `json:"conversation_id"`
	Direction        string          `json:"direction"`
	AuthorType       string          `json:"author_type"`
	AuthorID         string          `json:"author_id,omitempty"`
	Role             string          `json:"role"`
	Content          string          `json:"content"`
	AttachmentsJSON  json.RawMessage `json:"attachments,omitempty"`
	ChannelMessageID string          `json:"channel_message_id,omitempty"`
	SessionID        string          `json:"session_id,omitempty"`
	DeliveryStatus   string          `json:"delivery_status,omitempty"`
	DeliveryDetail   string          `json:"delivery_detail,omitempty"`
	CreatedAt        string          `json:"created_at"`
}

// ConversationListOptions configures a conversation list request.
type ConversationListOptions struct {
	ProjectID  string
	Channel    string
	AgentID    string
	UserID     string
	Controller string
	Status     string
	Query      string
	Cursor     string
	Limit      int
	Search     string // legacy alias for Query
	Agent      string // legacy alias for AgentID
	Tier       string // legacy memory-timeline filter
}

// ConversationListResponse is the cursor-paginated response from listing conversations.
type ConversationListResponse struct {
	Conversations []ConversationListItem `json:"conversations"`
	Items         []ConversationListItem `json:"items"`
	NextCursor    string                 `json:"next_cursor,omitempty"`
	HasMore       bool                   `json:"has_more"`
	Total         int                    `json:"total"`
}

// ConversationDetailResponse is the response from getting a single conversation.
type ConversationDetailResponse struct {
	Conversation Conversation `json:"conversation"`
	Source       string       `json:"source"`
}

// ConversationMessagesOptions configures a message list request.
type ConversationMessagesOptions struct {
	Limit  int
	Cursor string
}

// ConversationMessagesResponse is the cursor-paginated message list response.
type ConversationMessagesResponse struct {
	Messages   []ConversationMessage `json:"messages"`
	Items      []ConversationMessage `json:"items"`
	NextCursor string                `json:"next_cursor,omitempty"`
	HasMore    bool                  `json:"has_more"`
}

// ConversationStreamOptions configures a conversation SSE stream.
type ConversationStreamOptions struct {
	ProjectID string
}

// ConversationStreamEvent represents a single SSE event from the conversation stream.
type ConversationStreamEvent struct {
	Type         string               `json:"type,omitempty"`
	EventType    string               `json:"event_type,omitempty"`
	Conversation *Conversation        `json:"conversation,omitempty"`
	Message      *ConversationMessage `json:"message,omitempty"`
	Data         json.RawMessage      `json:"data,omitempty"`
	Error        *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// TakeOverConversationOptions configures a conversation takeover request.
type TakeOverConversationOptions struct {
	OperatorID string
	Force      bool
}

// SendConversationMessageOptions configures an operator/agent message send.
type SendConversationMessageOptions struct {
	Content     string `json:"content"`
	Attachments any    `json:"attachments,omitempty"`
}

// UpdateConversationOptions configures a conversation update request.
type UpdateConversationOptions struct {
	Status  string `json:"status,omitempty"`
	AgentID string `json:"agent_id,omitempty"`
}

// List returns a cursor-paginated page of omnichannel conversations.
func (c *ConversationsResource) List(ctx context.Context, opts *ConversationListOptions) (*ConversationListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.ProjectID != "" {
			params["project_id"] = opts.ProjectID
		}
		if opts.Channel != "" {
			params["channel"] = opts.Channel
		}
		if opts.AgentID != "" {
			params["agent_id"] = opts.AgentID
		}
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.Controller != "" {
			params["controller"] = opts.Controller
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.Query != "" {
			params["q"] = opts.Query
		}
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.Search != "" {
			params["search"] = opts.Search
		}
		if opts.Agent != "" {
			params["agent"] = opts.Agent
		}
		if opts.Tier != "" {
			params["tier"] = opts.Tier
		}
	}

	var result ConversationListResponse
	if err := c.http.Get(ctx, "/api/v1/conversations", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single omnichannel conversation.
func (c *ConversationsResource) Get(ctx context.Context, conversationID string) (*ConversationDetailResponse, error) {
	var result ConversationDetailResponse
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/conversations/%s", conversationID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Messages returns a cursor-paginated page of messages for a conversation.
func (c *ConversationsResource) Messages(ctx context.Context, conversationID string, opts *ConversationMessagesOptions) (*ConversationMessagesResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
	}
	var result ConversationMessagesResponse
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/conversations/%s/messages", conversationID), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Stream connects to the conversation SSE stream and calls callback for each event.
func (c *ConversationsResource) Stream(ctx context.Context, opts *ConversationStreamOptions, callback func(ConversationStreamEvent) error) error {
	path := "/api/v1/conversations/stream"
	if opts != nil && opts.ProjectID != "" {
		q := url.Values{}
		q.Set("project_id", opts.ProjectID)
		path += "?" + q.Encode()
	}

	var totalEvents, malformedEvents atomic.Int64
	return c.http.StreamSSE(ctx, "GET", path, nil, func(raw json.RawMessage) error {
		total := totalEvents.Add(1)
		var event ConversationStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			malformed := malformedEvents.Add(1)
			slog.Warn("skipping malformed SSE event in Conversations.Stream",
				"error", err,
				"malformed_count", malformed,
				"total_count", total,
			)
			if total >= 4 && malformed*2 > total {
				return fmt.Errorf("too many malformed SSE events: %d/%d", malformed, total)
			}
			return nil
		}
		return callback(event)
	})
}

// StreamChannel connects to the conversation SSE stream and returns events on a channel.
func (c *ConversationsResource) StreamChannel(ctx context.Context, opts *ConversationStreamOptions) (<-chan ConversationStreamEvent, <-chan error) {
	ch := make(chan ConversationStreamEvent, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		err := c.Stream(ctx, opts, func(event ConversationStreamEvent) error {
			select {
			case ch <- event:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil {
			errCh <- err
		}
	}()

	return ch, errCh
}

// TakeOver assigns human control for a conversation.
func (c *ConversationsResource) TakeOver(ctx context.Context, conversationID string, opts *TakeOverConversationOptions) (*Conversation, error) {
	path := fmt.Sprintf("/api/v1/conversations/%s/takeover", conversationID)
	if opts != nil {
		q := url.Values{}
		if opts.OperatorID != "" {
			q.Set("operator_id", opts.OperatorID)
		}
		if opts.Force {
			q.Set("force", "true")
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var result Conversation
	if err := c.http.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Release releases human control and returns the conversation to the agent.
func (c *ConversationsResource) Release(ctx context.Context, conversationID string) (*Conversation, error) {
	var result Conversation
	if err := c.http.Delete(ctx, fmt.Sprintf("/api/v1/conversations/%s/takeover", conversationID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SendAsAgent sends an outbound message in a conversation.
func (c *ConversationsResource) SendAsAgent(ctx context.Context, conversationID string, opts SendConversationMessageOptions) (*Conversation, error) {
	var result Conversation
	if err := c.http.Post(ctx, fmt.Sprintf("/api/v1/conversations/%s/messages", conversationID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MarkRead marks all messages in a conversation as read.
func (c *ConversationsResource) MarkRead(ctx context.Context, conversationID string) (*Conversation, error) {
	var result Conversation
	if err := c.http.Post(ctx, fmt.Sprintf("/api/v1/conversations/%s/read", conversationID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates conversation status or reassigns it to another agent.
func (c *ConversationsResource) Update(ctx context.Context, conversationID string, opts UpdateConversationOptions) (*Conversation, error) {
	var result Conversation
	if err := c.http.Patch(ctx, fmt.Sprintf("/api/v1/conversations/%s", conversationID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PushMessageOptions configures a proactive agent→channel push (POST
// /conversations/push): an agent-authored message delivered to a user's
// connected messaging channel (WhatsApp/Messenger/Instagram) outside the
// reply-to-inbound flow.
type PushMessageOptions struct {
	// ProjectID is optional; it defaults to the authenticated project/default
	// project when omitted.
	ProjectID string `json:"project_id,omitempty"`
	// AgentID is the agent UUID or name authoring the message.
	AgentID string `json:"agent_id"`
	// UserID is the platform user id to deliver to (channel identity owner).
	UserID string `json:"user_id"`
	// Content is the message text.
	Content string `json:"content"`
	// ChannelType restricts delivery to one channel (whatsapp, messenger,
	// instagram); when omitted, the first identity found is used.
	ChannelType string `json:"channel_type,omitempty"`
	// ConnectionID pins the outbound channel connection UUID; when omitted,
	// the conversation's or the project's connection for the channel is used.
	ConnectionID string `json:"connection_id,omitempty"`
}

// PushMessageResult is the outcome of a proactive channel push.
type PushMessageResult struct {
	ConversationID   string `json:"conversation_id,omitempty"`
	ChannelType      string `json:"channel_type"`
	ExternalID       string `json:"external_id"`
	ChannelMessageID string `json:"channel_message_id,omitempty"`
	// DeliveryStatus is the provider delivery status (sent|delivered|read|failed).
	DeliveryStatus string `json:"delivery_status"`
	SessionID      string `json:"session_id,omitempty"`
	// UsedTemplate is true when the 24h window was closed and the send used
	// the connection's approved re-engagement template.
	UsedTemplate bool `json:"used_template"`
}

// Push delivers an agent-authored message to a user's connected messaging
// channel (WhatsApp/Messenger/Instagram) without waiting for an inbound
// message — the proactive agent→channel delivery surface used by wakeups,
// lead-offer notifications, and research/outcome pings.
func (c *ConversationsResource) Push(ctx context.Context, opts PushMessageOptions) (*PushMessageResult, error) {
	var result PushMessageResult
	if err := c.http.Post(ctx, "/api/v1/conversations/push", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
