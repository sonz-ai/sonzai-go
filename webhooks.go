package sonzai

import (
	"context"
	"fmt"
)

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

// WebhookRegisterResponse is the response from registering a webhook.
// SigningSecret is only populated on first registration (creation), not updates.
type WebhookRegisterResponse struct {
	Success       bool   `json:"success"`
	SigningSecret string `json:"signing_secret,omitempty"`
}

// WebhookListResponse is the response from listing webhooks.
type WebhookListResponse struct {
	Webhooks []WebhookEndpoint `json:"webhooks"`
}

// WebhookDeliveryAttempt represents a single webhook delivery attempt.
type WebhookDeliveryAttempt struct {
	AttemptID     string `json:"attempt_id"`
	EventType     string `json:"event_type"`
	WebhookURL    string `json:"webhook_url"`
	ResponseCode  int    `json:"response_code"`
	ResponseBody  string `json:"response_body,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	DurationMs    int    `json:"duration_ms"`
	AttemptNumber int    `json:"attempt_number"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// DeliveryAttemptsResponse is the response from listing delivery attempts.
type DeliveryAttemptsResponse struct {
	Attempts []WebhookDeliveryAttempt `json:"attempts"`
}

// Register registers (or updates) a webhook URL for an event type.
// Returns a WebhookRegisterResponse which includes the signing secret on first
// registration. The signing secret is only returned once at creation time.
//
// Event types include: "on_wakeup_ready", "on_diary_generated",
// "on_personality_updated", "on_recurring_event_due", etc.
func (w *WebhooksResource) Register(ctx context.Context, eventType string, opts WebhookRegisterOptions) (*WebhookRegisterResponse, error) {
	var result WebhookRegisterResponse
	if err := w.http.Put(ctx, "/api/v1/webhooks/"+eventType, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
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

// ListDeliveryAttempts returns recent delivery attempts for a specific event type.
func (w *WebhooksResource) ListDeliveryAttempts(ctx context.Context, eventType string) (*DeliveryAttemptsResponse, error) {
	var result DeliveryAttemptsResponse
	if err := w.http.Get(ctx, "/api/v1/webhooks/"+eventType+"/attempts", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RotateSecret generates a new signing secret for a webhook event type.
// The new secret is returned in the response and is only shown once.
func (w *WebhooksResource) RotateSecret(ctx context.Context, eventType string) (*WebhookRegisterResponse, error) {
	var result WebhookRegisterResponse
	if err := w.http.Post(ctx, "/api/v1/webhooks/"+eventType+"/rotate-secret", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// -- Project-scoped webhooks --

// RegisterForProject registers (or updates) a webhook for a project and event type.
func (w *WebhooksResource) RegisterForProject(ctx context.Context, projectID, eventType string, opts WebhookRegisterOptions) (*WebhookRegisterResponse, error) {
	var result WebhookRegisterResponse
	if err := w.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s/webhooks/%s", projectID, eventType), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListForProject returns all registered webhooks for a project.
func (w *WebhooksResource) ListForProject(ctx context.Context, projectID string) (*WebhookListResponse, error) {
	var result WebhookListResponse
	if err := w.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/webhooks", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteForProject removes a webhook for a project event type.
func (w *WebhooksResource) DeleteForProject(ctx context.Context, projectID, eventType string) error {
	return w.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/webhooks/%s", projectID, eventType), nil)
}

// ListDeliveryAttemptsForProject returns delivery attempts for a project webhook event type.
func (w *WebhooksResource) ListDeliveryAttemptsForProject(ctx context.Context, projectID, eventType string) (*DeliveryAttemptsResponse, error) {
	var result DeliveryAttemptsResponse
	if err := w.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/webhooks/%s/attempts", projectID, eventType), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RotateSecretForProject generates a new signing secret for a project webhook event type.
func (w *WebhooksResource) RotateSecretForProject(ctx context.Context, projectID, eventType string) (*WebhookRegisterResponse, error) {
	var result WebhookRegisterResponse
	if err := w.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/webhooks/%s/rotate-secret", projectID, eventType), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
