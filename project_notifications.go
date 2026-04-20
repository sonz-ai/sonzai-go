package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// ProjectNotificationsResource provides project-scoped notification polling.
type ProjectNotificationsResource struct {
	http *httpClient
}

// ProjectNotificationListOptions configures a project notification list request.
type ProjectNotificationListOptions struct {
	AgentID   string
	EventType string
	Limit     int
}

// ProjectNotificationListResponse is the response from listing project notifications.
type ProjectNotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Count         int            `json:"count"`
}

// AcknowledgeOptions configures which notifications to acknowledge.
type AcknowledgeOptions struct {
	NotificationIDs []string `json:"notification_ids"`
}

// AcknowledgeResponse is the response from acknowledging notifications.
type AcknowledgeResponse struct {
	Acknowledged int `json:"acknowledged"`
}

// AcknowledgeAllOptions configures an acknowledge-all request.
type AcknowledgeAllOptions struct {
	AgentID   string
	EventType string
}

// List returns pending notifications for a project.
func (n *ProjectNotificationsResource) List(ctx context.Context, projectID string, opts *ProjectNotificationListOptions) (*ProjectNotificationListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.AgentID != "" {
			params["agent_id"] = opts.AgentID
		}
		if opts.EventType != "" {
			params["event_type"] = opts.EventType
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}

	var result ProjectNotificationListResponse
	if err := n.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/notifications", projectID), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Acknowledge marks specific notifications as acknowledged.
func (n *ProjectNotificationsResource) Acknowledge(ctx context.Context, projectID string, opts AcknowledgeOptions) (*AcknowledgeResponse, error) {
	var result AcknowledgeResponse
	if err := n.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/notifications/acknowledge", projectID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AcknowledgeAll marks all pending notifications for a project as acknowledged.
func (n *ProjectNotificationsResource) AcknowledgeAll(ctx context.Context, projectID string, opts *AcknowledgeAllOptions) (*AcknowledgeResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.AgentID != "" {
			params["agent_id"] = opts.AgentID
		}
		if opts.EventType != "" {
			params["event_type"] = opts.EventType
		}
	}

	// POST with query params — encode into URL
	path := fmt.Sprintf("/api/v1/projects/%s/notifications/acknowledge-all", projectID)
	var result AcknowledgeResponse
	if err := n.http.Post(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
