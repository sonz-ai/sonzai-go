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
//
// Provider/Model set session-level defaults that apply to caller-overridable
// post-processing tasks (fact extraction, mood analysis, per-turn extraction).
// Both must be set together to take effect; per-call /turn or /sessions/end
// requests can override them. Otherwise the server-side resolver picks a
// default (currently gemini-3.1-flash-lite).
type SessionStartOptions struct {
	UserID          string           `json:"user_id"`
	UserDisplayName string           `json:"user_display_name,omitempty"`
	SessionID       string           `json:"session_id"`
	InstanceID      string           `json:"instance_id,omitempty"`
	Provider        string           `json:"provider,omitempty"`
	Model           string           `json:"model,omitempty"`
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

// Start begins a chat session and returns an ergonomic *Session handle
// that owns the agent/user/session triple plus optional provider/model
// defaults. Subsequent Context / Turn / End calls on the handle thread
// these through automatically.
//
// Backward compat: *Session embeds SessionResponse, so existing callers
// that read `result.Success` on the return value keep working after
// this signature change (the field is promoted from the embedded
// SessionResponse).
func (s *SessionsResource) Start(ctx context.Context, agentID string, opts SessionStartOptions) (*Session, error) {
	var resp SessionResponse
	if err := s.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/start", agentID), opts, &resp); err != nil {
		return nil, err
	}
	return &Session{
		sessions:        s,
		SessionResponse: resp,
		AgentID:         agentID,
		UserID:          opts.UserID,
		UserDisplayName: opts.UserDisplayName,
		SessionID:       opts.SessionID,
		InstanceID:      opts.InstanceID,
		Provider:        opts.Provider,
		Model:           opts.Model,
	}, nil
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

// Turn submits a conversation turn to the realtime per-turn API.
//
// Runs sync mood-only extraction inline and publishes the rest of the
// post-processing (facts, personality, habits, etc.) as a deferred work
// item; poll the status with TurnStatus using result.ExtractionID.
//
// /turn does not auto-create sessions — call Sessions.Start first when you
// need session-level tool definitions or provider/model defaults.
func (s *SessionsResource) Turn(ctx context.Context, agentID, sessionID string, opts TurnOptions) (*TurnResult, error) {
	var result TurnResult
	err := s.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/sessions/%s/turn", agentID, sessionID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TurnStatus polls the lifecycle state of a deferred-turn job. State is
// one of queued, running, done, failed. Callers should poll with
// exponential backoff until TurnStatusResult.IsTerminal returns true.
func (s *SessionsResource) TurnStatus(ctx context.Context, agentID, extractionID string) (*TurnStatusResult, error) {
	var result TurnStatusResult
	err := s.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/turns/%s/status", agentID, extractionID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// StartSession is a deprecated alias for Start kept for callers that
// adopted the prior name. Use Start; it now returns *Session directly.
//
// Deprecated: use Start.
func (s *SessionsResource) StartSession(ctx context.Context, agentID string, opts SessionStartOptions) (*Session, error) {
	return s.Start(ctx, agentID, opts)
}

// Session is an ergonomic wrapper around the agent/user/session triple
// returned by Start. It threads provider/model defaults through
// Context/Turn/End calls and lets the caller skip repeating identifiers.
//
// SessionResponse is embedded so legacy callers that read `Success` on
// the Start return value still work via Go field promotion
// (`session.Success` reads through to the embedded SessionResponse).
//
// Per-call provider/model on TurnOptions override the session defaults;
// otherwise the server-side resolver picks a default model.
type Session struct {
	sessions *SessionsResource

	// SessionResponse mirrors the /sessions/start response body.
	// Embedded (not a named field) so `session.Success` keeps working
	// for callers written against the prior return type.
	SessionResponse

	AgentID         string
	UserID          string
	UserDisplayName string
	SessionID       string
	InstanceID      string
	Provider        string
	Model           string
}

// ContextOptions configures a Session.Context call. Mirrors the relevant
// subset of GetContextOptions; user/session/instance come from the Session
// handle.
type ContextOptions struct {
	Query    string `json:"query,omitempty"`
	Language string `json:"language,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// Context fetches the enriched agent context for this session — equivalent
// to AgentsResource.GetContext but pre-populated from the Session handle.
func (s *Session) Context(ctx context.Context, opts ContextOptions) (*EnrichedContextResponse, error) {
	return (&AgentsResource{http: s.sessions.http}).GetContext(ctx, s.AgentID, GetContextOptions{
		UserID:     s.UserID,
		SessionID:  s.SessionID,
		InstanceID: s.InstanceID,
		Query:      opts.Query,
		Language:   opts.Language,
		Timezone:   opts.Timezone,
	})
}

// Turn submits a conversation turn for this session. Provider/Model on
// opts override the session defaults; otherwise the session defaults
// flow through. UserID/InstanceID are pre-filled from the handle.
func (s *Session) Turn(ctx context.Context, opts TurnOptions) (*TurnResult, error) {
	if opts.UserID == "" {
		opts.UserID = s.UserID
	}
	if opts.UserDisplayName == "" {
		opts.UserDisplayName = s.UserDisplayName
	}
	if opts.InstanceID == "" {
		opts.InstanceID = s.InstanceID
	}
	// Per-call provider/model wins; only fall back to session defaults
	// when both sides on the call are empty.
	if opts.Provider == "" && opts.Model == "" {
		opts.Provider = s.Provider
		opts.Model = s.Model
	}
	return s.sessions.Turn(ctx, s.AgentID, s.SessionID, opts)
}

// TurnStatus polls the deferred-turn pipeline state for a previously
// submitted Turn on this session.
func (s *Session) TurnStatus(ctx context.Context, extractionID string) (*TurnStatusResult, error) {
	return s.sessions.TurnStatus(ctx, s.AgentID, extractionID)
}

// End concludes the session. The handle is unusable afterwards; create
// a new one via StartSession for the next session.
func (s *Session) End(ctx context.Context, opts ...SessionEndOptions) error {
	var endOpts SessionEndOptions
	if len(opts) > 0 {
		endOpts = opts[0]
	}
	if endOpts.UserID == "" {
		endOpts.UserID = s.UserID
	}
	if endOpts.SessionID == "" {
		endOpts.SessionID = s.SessionID
	}
	if endOpts.InstanceID == "" {
		endOpts.InstanceID = s.InstanceID
	}
	if endOpts.UserDisplayName == "" {
		endOpts.UserDisplayName = s.UserDisplayName
	}
	_, err := s.sessions.End(ctx, s.AgentID, endOpts)
	return err
}

// ---------------------------------------------------------------------------
// Per-user proxy methods — auto-scope by (AgentID, UserID, InstanceID)
// ---------------------------------------------------------------------------
//
// The platform scopes per-user rows under inst:<instanceID>:<userID> when
// a session was started with an InstanceID. Calling agent-level helpers
// (e.g. client.Agents.GetMood(ctx, agentID, userID, "")) without
// InstanceID silently targets the UNSCOPED row, which Session.Context
// then never reads — leading to the classic "I overrode mood but the
// next reply uses the old one" footgun. These wrappers eliminate it by
// always forwarding InstanceID from the Session handle.

// UpdateMood hard-sets this session-user's mood (0-100 per dimension).
func (s *Session) UpdateMood(ctx context.Context, valence, arousal, tension, affiliation float64) (*MoodResponse, error) {
	return (&AgentsResource{http: s.sessions.http}).UpdateMood(ctx, s.AgentID, UpdateMoodOptions{
		Valence: valence, Arousal: arousal, Tension: tension, Affiliation: affiliation,
		UserID: s.UserID, InstanceID: s.InstanceID,
	})
}

// GetMood returns this session-user's current mood (0-100 per dimension).
func (s *Session) GetMood(ctx context.Context) (*MoodResponse, error) {
	return (&AgentsResource{http: s.sessions.http}).GetMood(ctx, s.AgentID, s.UserID, s.InstanceID)
}

// ListFacts returns active facts about this session's user.
func (s *Session) ListFacts(ctx context.Context, limit int) (*ListAllFactsResponse, error) {
	return (&InventoryResource{http: s.sessions.http}).ListAllFacts(ctx, s.AgentID, s.UserID, ListAllFactsOptions{
		Limit: limit, InstanceID: s.InstanceID,
	})
}

// ScheduleWakeup queues a proactive wakeup for this session's user.
// Forwards UserID from the Session handle; the platform wakeup endpoint
// does not currently scope by InstanceID.
func (s *Session) ScheduleWakeup(ctx context.Context, opts ScheduleWakeupOptions) (*ScheduledWakeup, error) {
	if opts.UserID == "" {
		opts.UserID = s.UserID
	}
	return (&WakeupResource{http: s.sessions.http}).Schedule(ctx, s.AgentID, opts)
}

// Constellation returns the concept-graph constellation for this session's user.
func (s *Session) Constellation(ctx context.Context) (*ConstellationResponse, error) {
	return (&AgentsResource{http: s.sessions.http}).GetConstellation(ctx, s.AgentID, s.UserID, s.InstanceID)
}
