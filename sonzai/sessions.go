package sonzai

import (
	"context"
	"fmt"
)

// SessionsResource provides session lifecycle operations.
type SessionsResource struct {
	http *httpClient
}

// SessionStartOptions configures a session start request.
type SessionStartOptions struct {
	UserID     string `json:"user_id"`
	SessionID  string `json:"session_id"`
	InstanceID string `json:"instance_id,omitempty"`
}

// SessionEndOptions configures a session end request.
type SessionEndOptions struct {
	UserID          string        `json:"user_id"`
	SessionID       string        `json:"session_id"`
	InstanceID      string        `json:"instance_id,omitempty"`
	TotalMessages   int           `json:"total_messages,omitempty"`
	DurationSeconds int           `json:"duration_seconds,omitempty"`
	Messages        []ChatMessage `json:"messages,omitempty"`
}

// Start begins a chat session.
func (s *SessionsResource) Start(ctx context.Context, agentID string, opts SessionStartOptions) (*SessionResponse, error) {
	var result SessionResponse
	err := s.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/start", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// End concludes a chat session.
func (s *SessionsResource) End(ctx context.Context, agentID string, opts SessionEndOptions) (*SessionResponse, error) {
	var result SessionResponse
	err := s.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/end", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
