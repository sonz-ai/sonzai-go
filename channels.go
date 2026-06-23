package sonzai

import "context"

// Event types delivered to notification channels (and webhooks). A channel
// subscribes to one or more of these; an empty subscription means "all events".
const (
	// EventBuiltinAgentCompleted fires after a built-in agent finishes a
	// project-scoped run (lead_research, lead_score, lead_qualifier, ...). The
	// payload carries the agent slug, the original input, and the findings.
	EventBuiltinAgentCompleted = "builtin_agent.completed"

	// EventLeadEnriched fires when an async lead-enrichment job completes.
	EventLeadEnriched = "lead.enriched"
)

// ChannelsResource manages notification channels — tenant-configured delivery
// destinations (signed webhook, Resend email, or a Composio Gmail/Slack action)
// that the platform delivers events to. This is the managed alternative to
// standing up your own webhook receiver.
type ChannelsResource struct {
	http *httpClient
}

// NotificationChannel is a configured delivery destination. Secret Config
// fields (api_key, signing_secret, auth_header) are returned masked.
type NotificationChannel struct {
	ChannelID string                 `json:"channel_id"`
	ProjectID string                 `json:"project_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // "webhook" | "email" | "composio"
	Config    map[string]interface{} `json:"config,omitempty"`
	Events    []string               `json:"events"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
	Active    bool                   `json:"active"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// ChannelWriteOptions is the create/update payload for a notification channel.
//
// Config is type-specific:
//
//	webhook:  {"url": "...", "auth_header"?: "...", "signing_secret"?: "..."}
//	email:    {"api_key": "re_...", "from": "...", "to"?: "{{lead.email}}", "subject"?: "..."}
//	composio: {"agent_id": "...", "action": "gmail_send_email"|"slack_send_message", "to"?/"channel"?: "...", "subject"?: "..."}
//
// Events lists the subscribed event types (empty = all events). Filter is an
// optional predicate, e.g. {"agents": ["lead_research"], "min_score": 80}.
type ChannelWriteOptions struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
	Events []string               `json:"events,omitempty"`
	Filter map[string]interface{} `json:"filter,omitempty"`
	Active *bool                  `json:"active,omitempty"`
}

// ChannelListResponse is the response from listing channels.
type ChannelListResponse struct {
	Channels []NotificationChannel `json:"channels"`
}

// List returns all notification channels for the authenticated tenant's project.
func (c *ChannelsResource) List(ctx context.Context) (*ChannelListResponse, error) {
	var result ChannelListResponse
	if err := c.http.Get(ctx, "/api/v1/channels", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create registers a new notification channel.
func (c *ChannelsResource) Create(ctx context.Context, opts ChannelWriteOptions) (*NotificationChannel, error) {
	var result NotificationChannel
	if err := c.http.Post(ctx, "/api/v1/channels", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single notification channel by ID.
func (c *ChannelsResource) Get(ctx context.Context, channelID string) (*NotificationChannel, error) {
	var result NotificationChannel
	if err := c.http.Get(ctx, "/api/v1/channels/"+channelID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a notification channel's mutable fields.
func (c *ChannelsResource) Update(ctx context.Context, channelID string, opts ChannelWriteOptions) (*NotificationChannel, error) {
	var result NotificationChannel
	if err := c.http.Put(ctx, "/api/v1/channels/"+channelID, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a notification channel.
func (c *ChannelsResource) Delete(ctx context.Context, channelID string) error {
	return c.http.Delete(ctx, "/api/v1/channels/"+channelID, nil)
}
