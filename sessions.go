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
	UserID          string           `json:"user_id"`
	UserDisplayName string           `json:"user_display_name,omitempty"`
	SessionID       string           `json:"session_id"`
	InstanceID      string           `json:"instance_id,omitempty"`
	ToolDefinitions []ToolDefinition `json:"tool_definitions,omitempty"`
}

// SessionEndOptions configures a session end request.
type SessionEndOptions struct {
	UserID          string        `json:"user_id"`
	SessionID       string        `json:"session_id"`
	InstanceID      string        `json:"instance_id,omitempty"`
	TotalMessages   int           `json:"total_messages"`
	DurationSeconds int           `json:"duration_seconds"`
	Messages        []ChatMessage `json:"messages,omitempty"`
	UserDisplayName string        `json:"user_display_name,omitempty"`
	UserTimezone    string        `json:"user_timezone,omitempty"`
}

// SessionToolsOptions configures tools for a session.
type SessionToolsOptions struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolDefinition defines an external tool available during a session.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Start begins a chat session.
func (s *SessionsResource) Start(ctx context.Context, agentID string, opts SessionStartOptions) (*SessionResponse, error) {
	var result SessionResponse
	err := s.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/start", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// End concludes a chat session.
func (s *SessionsResource) End(ctx context.Context, agentID string, opts SessionEndOptions) (*SessionResponse, error) {
	var result SessionResponse
	err := s.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/end", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SetTools configures the tools available for a session.
// The Platform API expects a raw JSON array as the request body.
func (s *SessionsResource) SetTools(ctx context.Context, agentID, sessionID string, opts SessionToolsOptions) (*SessionResponse, error) {
	var result SessionResponse
	err := s.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/%s/tools", agentID, sessionID), opts.Tools, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
