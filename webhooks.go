package sonzai

import "context"

// WebhooksResource provides webhook management operations.
type WebhooksResource struct {
	http *httpClient
}

// WebhookEndpoint represents a registered webhook.
type WebhookEndpoint struct {
	EventType  string `json:"event_type"`
	WebhookURL string `json:"webhook_url"`
	AuthHeader string `json:"auth_header,omitempty"`
}

// WebhookRegisterOptions configures a webhook registration request.
type WebhookRegisterOptions struct {
	WebhookURL string `json:"webhook_url"`
	AuthHeader string `json:"auth_header,omitempty"`
}

// WebhookListResponse is the response from listing webhooks.
type WebhookListResponse struct {
	Webhooks []WebhookEndpoint `json:"webhooks"`
}

// Register registers (or updates) a webhook URL for an event type.
// Event types include: "on_wakeup_ready", "on_diary_generated",
// "on_personality_updated", "on_recurring_event_due", etc.
func (w *WebhooksResource) Register(ctx context.Context, eventType string, opts WebhookRegisterOptions) error {
	return w.http.Put(ctx, "/api/v1/webhooks/"+eventType, opts, nil)
}

// List returns all registered webhooks.
func (w *WebhooksResource) List(ctx context.Context) (*WebhookListResponse, error) {
	var result WebhookListResponse
	if err := w.http.Get(ctx, "/api/v1/webhooks", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a webhook for an event type.
func (w *WebhooksResource) Delete(ctx context.Context, eventType string) error {
	return w.http.Delete(ctx, "/api/v1/webhooks/"+eventType, nil)
}
