package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
)

// ChannelConnectionsResource provides project-scoped Meta channel connection operations.
type ChannelConnectionsResource struct {
	http *httpClient
}

// ChannelType is a supported omnichannel provider surface.
type ChannelType string

const (
	// ChannelTypeWhatsApp is the Meta WhatsApp channel.
	ChannelTypeWhatsApp ChannelType = "whatsapp"
	// ChannelTypeMessenger is the Meta Messenger channel.
	ChannelTypeMessenger ChannelType = "messenger"
	// ChannelTypeInstagram is the Meta Instagram channel.
	ChannelTypeInstagram ChannelType = "instagram"
)

// ChannelProviderMode selects how Meta channel credentials are provisioned.
type ChannelProviderMode string

const (
	// ChannelProviderModeBYOApp uses customer-provided Meta app credentials.
	ChannelProviderModeBYOApp ChannelProviderMode = "byo_app"
	// ChannelProviderModeEmbeddedSignup uses Meta Embedded Signup.
	ChannelProviderModeEmbeddedSignup ChannelProviderMode = "embedded_signup"
)

// ChannelConnection represents a configured Meta channel connection.
type ChannelConnection struct {
	ConnectionID       string          `json:"connection_id"`
	ProjectID          string          `json:"project_id"`
	ChannelType        string          `json:"channel_type"`
	ProviderMode       string          `json:"provider_mode"`
	DisplayName        string          `json:"display_name"`
	Status             string          `json:"status"`
	StatusDetail       string          `json:"status_detail,omitempty"`
	DefaultAgentID     string          `json:"default_agent_id,omitempty"`
	AppID              string          `json:"app_id,omitempty"`
	PhoneNumberID      string          `json:"phone_number_id,omitempty"`
	WABAID             string          `json:"waba_id,omitempty"`
	PageID             string          `json:"page_id,omitempty"`
	IGAccountID        string          `json:"ig_account_id,omitempty"`
	VerifyToken        string          `json:"verify_token,omitempty"`
	WebhookCallbackURL string          `json:"webhook_callback_url,omitempty"`
	TemplatesJSON      json.RawMessage `json:"templates,omitempty"`
	TestSendSucceeded  bool            `json:"test_send_succeeded,omitempty"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}

// String returns a redacted string representation of the channel connection.
func (c ChannelConnection) String() string {
	type redacted ChannelConnection
	cp := redacted(c)
	if cp.VerifyToken != "" {
		cp.VerifyToken = redactedSecret
	}
	return fmt.Sprintf("%+v", cp)
}

// GoString returns a redacted Go-syntax representation of the channel connection.
func (c ChannelConnection) GoString() string {
	return c.String()
}

// ChannelConnectionListResponse is the response from listing channel connections.
type ChannelConnectionListResponse struct {
	Connections []ChannelConnection `json:"connections"`
	Items       []ChannelConnection `json:"items"`
}

// CreateChannelConnectionOptions configures a channel connection creation request.
type CreateChannelConnectionOptions struct {
	ChannelType    ChannelType         `json:"channel_type"`
	ProviderMode   ChannelProviderMode `json:"provider_mode,omitempty"`
	DisplayName    string              `json:"display_name,omitempty"`
	DefaultAgentID string              `json:"default_agent_id,omitempty"`
	AppID          string              `json:"app_id,omitempty"`
	AppSecret      string              `json:"app_secret,omitempty"`
	PhoneNumberID  string              `json:"phone_number_id,omitempty"`
	WABAID         string              `json:"waba_id,omitempty"`
	PageID         string              `json:"page_id,omitempty"`
	IGAccountID    string              `json:"ig_account_id,omitempty"`
	AccessToken    string              `json:"access_token,omitempty"`
	VerifyToken    string              `json:"verify_token,omitempty"`
	Code           string              `json:"code,omitempty"`
	TestTo         string              `json:"test_to,omitempty"`
	TestMessage    string              `json:"test_message,omitempty"`
	Templates      any                 `json:"templates,omitempty"`
}

// String returns a redacted string representation of the creation request.
func (o CreateChannelConnectionOptions) String() string {
	type redacted CreateChannelConnectionOptions
	cp := redacted(o)
	if cp.AppSecret != "" {
		cp.AppSecret = redactedSecret
	}
	if cp.AccessToken != "" {
		cp.AccessToken = redactedSecret
	}
	if cp.VerifyToken != "" {
		cp.VerifyToken = redactedSecret
	}
	return fmt.Sprintf("%+v", cp)
}

// GoString returns a redacted Go-syntax representation of the creation request.
func (o CreateChannelConnectionOptions) GoString() string {
	return o.String()
}

// UpdateChannelConnectionOptions configures a channel connection update request.
type UpdateChannelConnectionOptions struct {
	DefaultAgentID string `json:"default_agent_id,omitempty"`
	Status         string `json:"status,omitempty"`
	Templates      any    `json:"templates,omitempty"`
}

// TestChannelConnectionOptions configures a channel connection test-send request.
type TestChannelConnectionOptions struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

const redactedSecret = "[REDACTED]"

// List returns all channel connections for a project.
func (c *ChannelConnectionsResource) List(ctx context.Context, projectID string) (*ChannelConnectionListResponse, error) {
	var result ChannelConnectionListResponse
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/channel-connections", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a channel connection for a project.
func (c *ChannelConnectionsResource) Create(ctx context.Context, projectID string, opts CreateChannelConnectionOptions) (*ChannelConnection, error) {
	var result ChannelConnection
	if err := c.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/channel-connections", projectID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single channel connection.
func (c *ChannelConnectionsResource) Get(ctx context.Context, projectID, connectionID string) (*ChannelConnection, error) {
	var result ChannelConnection
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/channel-connections/%s", projectID, connectionID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates channel connection status, default agent, or templates.
func (c *ChannelConnectionsResource) Update(ctx context.Context, projectID, connectionID string, opts UpdateChannelConnectionOptions) (*ChannelConnection, error) {
	var result ChannelConnection
	if err := c.http.Patch(ctx, fmt.Sprintf("/api/v1/projects/%s/channel-connections/%s", projectID, connectionID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a channel connection.
func (c *ChannelConnectionsResource) Delete(ctx context.Context, projectID, connectionID string) error {
	return c.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/channel-connections/%s", projectID, connectionID), nil)
}

// Test sends a test message through a channel connection and updates health metadata.
func (c *ChannelConnectionsResource) Test(ctx context.Context, projectID, connectionID string, opts TestChannelConnectionOptions) (*ChannelConnection, error) {
	var result ChannelConnection
	if err := c.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/channel-connections/%s/test", projectID, connectionID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
