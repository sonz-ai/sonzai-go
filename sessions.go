package sonzai

import (
	"context"
	"fmt"
	"time"
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
	// Wait runs the CE pipeline synchronously before responding.
	// Useful for benchmarks or tests that query memory immediately after session end.
	Wait bool `json:"wait,omitempty"`
	// PollTimeout caps how long End blocks polling the async status
	// endpoint before returning a timeout error. Zero falls back to the
	// default (15 minutes) — ample for a stalled LLM pipeline while
	// still surfacing a clear error when the worker is truly wedged.
	// Only consulted when the server returns a processing_id.
	PollTimeout time.Duration `json:"-"`
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

// asyncSessionEndAccept mirrors the 202 response when the server is in
// ENABLE_ASYNC_SESSION_END=true mode. Deliberately separate from
// SessionResponse so the public shape stays stable.
type asyncSessionEndAccept struct {
	Success      bool   `json:"success"`
	Async        bool   `json:"async"`
	ProcessingID string `json:"processing_id"`
	StatusURL    string `json:"status_url"`
	SessionID    string `json:"session_id"`
	AgentID      string `json:"agent_id"`
	EnqueuedAt   string `json:"enqueued_at"`
}

// sessionEndStatus mirrors redisinfra.SessionEndStatus on the wire —
// what GET /sessions/end/status/{pid} returns.
type sessionEndStatus struct {
	State      string `json:"state"`
	EnqueuedAt string `json:"enqueued_at"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
	SessionID  string `json:"session_id"`
	AgentID    string `json:"agent_id"`
	Error      string `json:"error,omitempty"`
	Attempt    int    `json:"attempt,omitempty"`
}

// sessionEndPoll* knobs pair with the Python SDK's polling defaults so
// the two SDKs behave identically under an async deploy. Overall timeout
// is overridable via SessionEndOptions.PollTimeout.
const (
	sessionEndPollInitialInterval = 500 * time.Millisecond
	sessionEndPollMaxInterval     = 5 * time.Second
	sessionEndPollOverallTimeout  = 15 * time.Minute
)

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
//
// Behaviour matrix vs. the server:
//   - Legacy 200 {success:true,async:*} — return immediately.
//   - 202 with processing_id (ENABLE_ASYNC_SESSION_END=true) — block on
//     /sessions/end/status/{pid} until state is done or failed so the
//     caller sees the same synchronous shape it always had. Use Wait:true
//     in the request to force the inline path on an async server.
func (s *SessionsResource) End(ctx context.Context, agentID string, opts SessionEndOptions) (*SessionResponse, error) {
	var maybeAsync asyncSessionEndAccept
	if err := s.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/end", agentID), opts, &maybeAsync); err != nil {
		return nil, err
	}
	if maybeAsync.ProcessingID == "" {
		return &SessionResponse{Success: maybeAsync.Success}, nil
	}
	timeout := opts.PollTimeout
	if timeout <= 0 {
		timeout = sessionEndPollOverallTimeout
	}
	if err := s.pollSessionEndStatus(ctx, maybeAsync.ProcessingID, timeout); err != nil {
		return nil, err
	}
	return &SessionResponse{Success: true}, nil
}

// pollSessionEndStatus polls /sessions/end/status/{pid} with exponential
// backoff until state is done or failed. Returns nil on done; an error
// describing the failure on failed; or a timeout error if we exceed the
// overall deadline.
func (s *SessionsResource) pollSessionEndStatus(ctx context.Context, processingID string, overallTimeout time.Duration) error {
	deadline := time.Now().Add(overallTimeout)
	interval := sessionEndPollInitialInterval
	var lastState string
	for {
		var status sessionEndStatus
		if err := s.http.Get(ctx, fmt.Sprintf("/api/v1/sessions/end/status/%s", processingID), nil, &status); err != nil {
			return fmt.Errorf("session end status poll: %w", err)
		}
		lastState = status.State
		switch status.State {
		case "done":
			return nil
		case "failed":
			if status.Error != "" {
				return fmt.Errorf("session end failed: %s", status.Error)
			}
			return fmt.Errorf("session end failed")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("session end poll timed out after %s (last state: %s)", overallTimeout, lastState)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
		interval = time.Duration(float64(interval) * 1.5)
		if interval > sessionEndPollMaxInterval {
			interval = sessionEndPollMaxInterval
		}
	}
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
