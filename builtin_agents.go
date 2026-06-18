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

// ---- Lead-scoring feedback loop (bandit) ----

// SegmentCal is a per-segment calibration adjustment derived from realized
// lead outcomes. Adjust is a multiplicative correction the lead_score agent
// applies to future leads matching this segment.
type SegmentCal struct {
	Segment     string  `json:"segment"`
	N           int     `json:"n"`
	Conversions int     `json:"conversions"`
	PHat        float64 `json:"p_hat"`
	Adjust      float64 `json:"adjust"`
}

// BandAccuracy compares the predicted conversion rate of a score band against
// the realized rate, exposing the calibration gap per band.
type BandAccuracy struct {
	Band           string  `json:"band"`
	N              int     `json:"n"`
	Conversions    int     `json:"conversions"`
	PredictedRate  float64 `json:"predicted_rate"`
	ActualRate     float64 `json:"actual_rate"`
	AvgScore       float64 `json:"avg_score"`
	CalibrationGap float64 `json:"calibration_gap"`
}

// Calibration is the project's current lead-scoring calibration: predicted-vs-
// actual accuracy by band plus multiplicative segment adjustments. The
// lead_score agent applies it to future leads.
type Calibration struct {
	Segments  []SegmentCal   `json:"segments"`
	Bands     []BandAccuracy `json:"bands"`
	BaseRate  float64        `json:"base_rate"`
	UpdatedAt string         `json:"updated_at"`
}

// LeadOutcomeParams is the request body for recording a realized lead-scoring
// outcome. LeadRef and Outcome are required.
type LeadOutcomeParams struct {
	// LeadRef identifies the lead this outcome belongs to. Required.
	LeadRef string `json:"lead_ref"`

	// PredictedScore is the score the agent assigned, if known.
	PredictedScore int `json:"predicted_score,omitempty"`

	// PredictedBand is the score band the agent assigned, if known.
	PredictedBand string `json:"predicted_band,omitempty"`

	// Features captures the lead's scored features for segment calibration.
	Features map[string]any `json:"features,omitempty"`

	// Outcome is the realized result (e.g. "won", "lost"). Required.
	Outcome string `json:"outcome"`

	// ScoreSignal optionally overrides how the outcome maps to a conversion.
	ScoreSignal string `json:"score_signal,omitempty"`

	// Note is an optional free-text annotation.
	Note string `json:"note,omitempty"`
}

// RecordLeadOutcome records a realized lead-scoring outcome (won/lost/…) and
// returns the recomputed project calibration the lead_score agent applies to
// future leads.
func (c *BuiltinAgentsResource) RecordLeadOutcome(ctx context.Context, params LeadOutcomeParams) (*Calibration, error) {
	var result Calibration
	if err := c.http.Post(ctx, "/api/v1/builtin-agents/lead_score/outcome", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLeadCalibration returns the current lead-scoring calibration (predicted-
// vs-actual by band plus per-segment adjustments).
func (c *BuiltinAgentsResource) GetLeadCalibration(ctx context.Context) (*Calibration, error) {
	var result Calibration
	if err := c.http.Get(ctx, "/api/v1/builtin-agents/lead_score/calibration", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---- Async lead enrichment ----

// EnrichLeadInput is the raw lead payload submitted for enrichment. All
// fields are optional; richer input yields a richer enriched profile.
type EnrichLeadInput struct {
	Name     string `json:"name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Company  string `json:"company,omitempty"`
	Brand    string `json:"brand,omitempty"`
	Vertical string `json:"vertical,omitempty"`
	Raw      string `json:"raw,omitempty"`
}

// EnrichLeadParams is the request body for enqueueing a lead-enrichment job.
type EnrichLeadParams struct {
	// Lead is the raw lead payload to enrich. Required.
	Lead EnrichLeadInput `json:"lead"`

	// WebhookURL optionally receives a callback when the job completes.
	WebhookURL string `json:"webhook_url,omitempty"`
}

// EnrichJob is the state of an async lead-enrichment job. Status moves
// "queued" → "processing" → "done" (or "error"). When done, Result carries
// the rich, evolving enrichment object (identity, affiliations,
// current_location, net_worth, score, band, intent, value, recommended_brand,
// next_best_action, recommended_message, …) — decode it into your own struct
// as needed.
type EnrichJob struct {
	JobID  string         `json:"job_id"`
	Status string         `json:"status"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// EnrichLead enqueues an async lead-enrichment job and returns the job
// handle (job_id + status "queued"). Poll GetEnrichment with the returned
// JobID until Status is "done" or "error".
func (c *BuiltinAgentsResource) EnrichLead(ctx context.Context, params EnrichLeadParams) (*EnrichJob, error) {
	var result EnrichJob
	if err := c.http.Post(ctx, "/api/v1/builtin-agents/lead/enrich", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEnrichment returns the current state of an async lead-enrichment job.
// When Status is "done", Result carries the enriched lead profile; when
// "error", Error carries the failure reason.
func (c *BuiltinAgentsResource) GetEnrichment(ctx context.Context, jobID string) (*EnrichJob, error) {
	var result EnrichJob
	path := fmt.Sprintf("/api/v1/builtin-agents/lead/enrich/%s", jobID)
	if err := c.http.Get(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---- Closed-loop agent self-improvement (learned guidance) ----

// LearnResult is the outcome of one agent distillation cycle. Changed reports
// whether a new guidance version was applied; Guidance carries the applied
// guidance when it did, and Violations lists any bounds the candidate breached.
type LearnResult struct {
	Changed    bool           `json:"changed"`
	Reason     string         `json:"reason,omitempty"`
	Guidance   map[string]any `json:"guidance,omitempty"`
	Violations []string       `json:"violations,omitempty"`
}

// AgentGuidance is an agent's learned guidance: the active version plus recent
// version history. Both fields may be nil when no guidance has been distilled.
type AgentGuidance struct {
	Active  map[string]any   `json:"active"`
	History []map[string]any `json:"history"`
}

// LearnAgent runs one distillation cycle for an agent, turning accumulated
// critiques/outcomes into a new, bounded, auto-applied guidance version. It
// respects the project kill switch (see SetAgentLearning). Pass nil evidence
// to distill from accumulated signals only.
func (c *BuiltinAgentsResource) LearnAgent(ctx context.Context, slug string, evidence any) (*LearnResult, error) {
	var result LearnResult
	path := fmt.Sprintf("/api/v1/builtin-agents/%s/learn", slug)
	if err := c.http.Post(ctx, path, map[string]any{"evidence": evidence}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAgentGuidance returns an agent's active learned guidance plus recent
// version history.
func (c *BuiltinAgentsResource) GetAgentGuidance(ctx context.Context, slug string) (*AgentGuidance, error) {
	var result AgentGuidance
	path := fmt.Sprintf("/api/v1/builtin-agents/%s/guidance", slug)
	if err := c.http.Get(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RollbackAgentGuidance rolls an agent's active guidance back to the prior
// version and returns the now-active guidance.
func (c *BuiltinAgentsResource) RollbackAgentGuidance(ctx context.Context, slug string) (*AgentGuidance, error) {
	var result AgentGuidance
	path := fmt.Sprintf("/api/v1/builtin-agents/%s/guidance/rollback", slug)
	if err := c.http.Post(ctx, path, map[string]any{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetAgentLearning toggles closed-loop auto-apply for the project (the kill
// switch). When disabled, LearnAgent distills but does not auto-apply guidance.
func (c *BuiltinAgentsResource) SetAgentLearning(ctx context.Context, enabled bool) error {
	var result struct {
		Enabled bool `json:"enabled"`
	}
	return c.http.Put(ctx, "/api/v1/builtin-agents/learning", map[string]bool{"enabled": enabled}, &result)
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
