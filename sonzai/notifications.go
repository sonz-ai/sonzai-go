package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// NotificationsResource provides notification operations for an agent.
type NotificationsResource struct {
	http *httpClient
}

// NotificationListOptions configures a notification list request.
type NotificationListOptions struct {
	Status string
	UserID string
	Limit  int
}

// List returns notifications for an agent.
func (n *NotificationsResource) List(ctx context.Context, agentID string, opts *NotificationListOptions) (*NotificationListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}

	var result NotificationListResponse
	err := n.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/notifications", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Consume marks a notification as consumed.
func (n *NotificationsResource) Consume(ctx context.Context, agentID, messageID string) (*SessionResponse, error) {
	var result SessionResponse
	err := n.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/notifications/%s/consume", agentID, messageID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// History returns notification history.
func (n *NotificationsResource) History(ctx context.Context, agentID string, limit int) (*NotificationListResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	var result NotificationListResponse
	err := n.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/notifications/history", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
