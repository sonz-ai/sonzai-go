package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// Sonzai Built-in Agents are platform-hosted vertical task agents. They run
// fully managed on the Sonzai platform — no provisioning, prompting, or tool
// wiring required — and bill per invocation (with BYOK pass-through where
// configured). Catalog slugs are listed below; call List for live metadata.
const (
	BuiltinAgentLeadResearch  = "lead_research"
	BuiltinAgentMarketIntel   = "market_intel"
	BuiltinAgentLeadExtract   = "lead_extract"
	BuiltinAgentLeadScore     = "lead_score"
	BuiltinAgentLeadQualifier = "lead_qualifier"
)

// BuiltinAgent describes one entry in the built-in agent catalog.
type BuiltinAgent struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	Provisioned bool   `json:"provisioned"`
}

// BuiltinAgentInvokeParams is the request body for invoking a built-in agent.
type BuiltinAgentInvokeParams struct {
	// Input is the agent-specific task payload (e.g. {"company": "..."} for
	// lead_research). Required.
	Input map[string]any `json:"input"`

	// Title optionally names the session created for this invocation.
	Title string `json:"title,omitempty"`
}

// BuiltinAgentUsage carries token accounting for an invocation or chat turn.
type BuiltinAgentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// BuiltinAgentInvokeResult is the terminal result of a built-in agent
// invocation (blocking or streaming).
type BuiltinAgentInvokeResult struct {
	// Findings is the raw structured output of the agent. Its shape is
	// agent-specific; decode it into your own struct.
	Findings json.RawMessage `json:"findings"`

	// Summary is a human-readable digest of the findings.
	Summary string `json:"summary"`

	// SessionID identifies the session created for this invocation. Use it
	// with SendMessage to ask follow-up questions.
	SessionID string `json:"session_id"`

	// Model is the model the agent ran on.
	Model string `json:"model"`

	// BYOK is true when the invocation billed through a customer-provided key.
	BYOK bool `json:"byok"`

	Usage          BuiltinAgentUsage `json:"usage"`
	RunningSeconds float64           `json:"running_seconds"`
	CostUSD        float64           `json:"cost_usd"`
}

// BuiltinAgentStreamUpdate is one progress frame from a streaming invocation
// or chat turn (SSE `event: update`).
type BuiltinAgentStreamUpdate struct {
	// Type is the update kind (e.g. "thinking", "tool_use", "text").
	Type string `json:"type"`

	// Tool is set on tool-activity updates.
	Tool string `json:"tool,omitempty"`

	// Text is set on text/progress updates.
	Text string `json:"text,omitempty"`

	// Detail carries additional context for the update, when present.
	Detail string `json:"detail,omitempty"`

	// Elapsed is seconds since the invocation started, when present.
	Elapsed float64 `json:"elapsed,omitempty"`
}

// BuiltinAgentSessionParams is the request body for creating a session.
type BuiltinAgentSessionParams struct {
	// Agent is the built-in agent slug (e.g. BuiltinAgentLeadResearch). Required.
	Agent string `json:"agent"`

	// Title optionally names the session.
	Title string `json:"title,omitempty"`
}

// BuiltinAgentSession is a conversation session with a built-in agent. The
// Billed* fields are populated on GetSession only.
type BuiltinAgentSession struct {
	ID        string  `json:"id"`
	Agent     string  `json:"agent"`
	Model     string  `json:"model"`
	Status    string  `json:"status"`
	Title     string  `json:"title"`
	BYOK      bool    `json:"byok"`
	CostUSD   float64 `json:"cost_usd"`
	CreatedAt string  `json:"created_at"`

	BilledInputTokens         int `json:"billed_input_tokens,omitempty"`
	BilledOutputTokens        int `json:"billed_output_tokens,omitempty"`
	BilledCacheReadTokens     int `json:"billed_cache_read_tokens,omitempty"`
	BilledCacheCreationTokens int `json:"billed_cache_creation_tokens,omitempty"`
}

// BuiltinAgentChatTurnResult is the terminal result of one chat turn in a
// built-in agent session.
type BuiltinAgentChatTurnResult struct {
	// Reply is the agent's textual answer for this turn.
	Reply string `json:"reply"`

	// Findings is present when the turn produced new structured output.
	Findings json.RawMessage `json:"findings,omitempty"`

	Usage          BuiltinAgentUsage `json:"usage"`
	TurnCostUSD    float64           `json:"turn_cost_usd"`
	RunningSeconds float64           `json:"running_seconds"`
}

// BuiltinAgentsResource provides access to Sonzai Built-in Agents —
// platform-hosted vertical task agents for lead research, market intel,
// lead extraction, scoring, and qualification.
type BuiltinAgentsResource struct {
	http *httpClient
}

// List returns the built-in agent catalog with per-project provisioning state.
func (c *BuiltinAgentsResource) List(ctx context.Context) ([]BuiltinAgent, error) {
	var result struct {
		Agents []BuiltinAgent `json:"agents"`
	}
	if err := c.http.Get(ctx, "/api/v1/builtin-agents", nil, &result); err != nil {
		return nil, err
	}
	return result.Agents, nil
}

// Invoke runs a built-in agent to completion and returns its result. Deep
// research invocations can run for 15+ minutes; the call blocks until the
// agent finishes and is capped only by the ctx deadline — pass a context
// with a generous timeout (or none) rather than relying on short defaults.
func (c *BuiltinAgentsResource) Invoke(ctx context.Context, slug string, params BuiltinAgentInvokeParams) (*BuiltinAgentInvokeResult, error) {
	var result BuiltinAgentInvokeResult
	path := fmt.Sprintf("/api/v1/builtin-agents/%s/invoke", slug)
	if err := c.http.PostLongRunning(ctx, path, map[string]string{"stream": "false"}, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InvokeStream runs a built-in agent and streams progress updates while it
// works. onUpdate is called for every `update` frame (pass nil to ignore
// them); the terminal `result` frame is decoded and returned. A terminal
// `error` frame is returned as an error. Like Invoke, the call is capped
// only by the ctx deadline.
func (c *BuiltinAgentsResource) InvokeStream(ctx context.Context, slug string, params BuiltinAgentInvokeParams, onUpdate func(BuiltinAgentStreamUpdate)) (*BuiltinAgentInvokeResult, error) {
	path := fmt.Sprintf("/api/v1/builtin-agents/%s/invoke", slug)
	var result *BuiltinAgentInvokeResult
	err := c.http.StreamSSENamed(ctx, "POST", path, map[string]string{"stream": "true"}, params, func(event string, data json.RawMessage) error {
		switch event {
		case "update":
			if onUpdate == nil {
				return nil
			}
			var u BuiltinAgentStreamUpdate
			if err := json.Unmarshal(data, &u); err != nil {
				return fmt.Errorf("builtin agent: decode update: %w", err)
			}
			onUpdate(u)
		case "result":
			var r BuiltinAgentInvokeResult
			if err := json.Unmarshal(data, &r); err != nil {
				return fmt.Errorf("builtin agent: decode result: %w", err)
			}
			result = &r
		case "error":
			return builtinAgentStreamError(data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("builtin agent: stream ended without a result event")
	}
	return result, nil
}

// CreateSession opens a conversation session with a built-in agent without
// invoking it. Use SendMessage to drive the conversation.
func (c *BuiltinAgentsResource) CreateSession(ctx context.Context, params BuiltinAgentSessionParams) (*BuiltinAgentSession, error) {
	var result BuiltinAgentSession
	if err := c.http.Post(ctx, "/api/v1/builtin-agents/sessions", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListSessions returns recent built-in agent sessions, newest first. Pass
// limit <= 0 for the server default.
func (c *BuiltinAgentsResource) ListSessions(ctx context.Context, limit int) ([]BuiltinAgentSession, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result struct {
		Sessions []BuiltinAgentSession `json:"sessions"`
	}
	if err := c.http.Get(ctx, "/api/v1/builtin-agents/sessions", params, &result); err != nil {
		return nil, err
	}
	return result.Sessions, nil
}

// GetSession returns one session including billed token totals.
func (c *BuiltinAgentsResource) GetSession(ctx context.Context, sessionID string) (*BuiltinAgentSession, error) {
	var result BuiltinAgentSession
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/builtin-agents/sessions/%s", sessionID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SendMessage sends one chat turn into a built-in agent session. With a nil
// onUpdate the call blocks (stream=false) and returns the turn result
// directly; with a non-nil onUpdate the turn streams (stream=true) and
// onUpdate receives every `update` frame before the terminal result. Turns
// can run for minutes when the agent re-researches; only the ctx deadline
// caps the call.
func (c *BuiltinAgentsResource) SendMessage(ctx context.Context, sessionID, text string, onUpdate func(BuiltinAgentStreamUpdate)) (*BuiltinAgentChatTurnResult, error) {
	path := fmt.Sprintf("/api/v1/builtin-agents/sessions/%s/messages", sessionID)
	body := map[string]string{"text": text}

	if onUpdate == nil {
		var result BuiltinAgentChatTurnResult
		if err := c.http.PostLongRunning(ctx, path, map[string]string{"stream": "false"}, body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	var result *BuiltinAgentChatTurnResult
	err := c.http.StreamSSENamed(ctx, "POST", path, map[string]string{"stream": "true"}, body, func(event string, data json.RawMessage) error {
		switch event {
		case "update":
			var u BuiltinAgentStreamUpdate
			if err := json.Unmarshal(data, &u); err != nil {
				return fmt.Errorf("builtin agent: decode update: %w", err)
			}
			onUpdate(u)
		case "result":
			var r BuiltinAgentChatTurnResult
			if err := json.Unmarshal(data, &r); err != nil {
				return fmt.Errorf("builtin agent: decode result: %w", err)
			}
			result = &r
		case "error":
			return builtinAgentStreamError(data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("builtin agent: stream ended without a result event")
	}
	return result, nil
}

// builtinAgentStreamError decodes a terminal SSE `error` frame.
func builtinAgentStreamError(data json.RawMessage) error {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(data, &e) == nil && e.Error != "" {
		return fmt.Errorf("builtin agent: %s", e.Error)
	}
	return fmt.Errorf("builtin agent: stream error: %s", string(data))
}
