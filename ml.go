package sonzai

import (
	"context"
	"fmt"
)

// MLResource provides the platform's generalized, multi-tenant / multi-vertical
// ML & RL surface: contextual-bandit next-best-action (decide + learn) and the
// unified feedback call. Every method is keyed by a free-form useCase string
// (e.g. "lead_score", "claim_triage", "churn") so a single tenant can run many
// independent models. Calls are tenant-scoped server-side by the API key.
type MLResource struct {
	http *httpClient
}

// NBAAction is one candidate action for a next-best-action decision.
type NBAAction struct {
	ID       string                 `json:"id"`
	Features map[string]interface{} `json:"features,omitempty"`
}

// NBAActionScore is the scored result for one candidate action.
type NBAActionScore struct {
	ActionID   string  `json:"action_id"`
	Score      float64 `json:"score"`
	Propensity float64 `json:"propensity"`
}

// DecideNBAOptions configures a next-best-action decision request.
type DecideNBAOptions struct {
	Context map[string]interface{} `json:"context,omitempty"`
	Actions []NBAAction            `json:"actions"`
	Explore *bool                  `json:"explore,omitempty"`
}

// DecideNBAResult is the policy's chosen action plus the full scored slate.
// Propensity is the probability the policy assigned the chosen action —
// record it and pass it back to LearnNBA (or RecordFeedback) once the reward
// is realized.
type DecideNBAResult struct {
	ActionID   string           `json:"action_id"`
	Propensity float64          `json:"propensity"`
	Scores     []NBAActionScore `json:"scores"`
	Explore    bool             `json:"explore"`
	ModelN     int              `json:"model_n"`
}

// LearnNBAOptions configures recording the realized reward of a previously
// taken action.
type LearnNBAOptions struct {
	Context        map[string]interface{} `json:"context,omitempty"`
	ActionID       string                 `json:"action_id"`
	ActionFeatures map[string]interface{} `json:"action_features,omitempty"`
	Propensity     *float64               `json:"propensity,omitempty"`
	Reward         float64                `json:"reward"`
}

// LearnNBAResult acknowledges a recorded reward. N is the running count of
// learning examples the policy has ingested.
type LearnNBAResult struct {
	OK bool `json:"ok"`
	N  int  `json:"n"`
}

// MLFeedbackOptions configures the unified feedback call: the single
// operator-facing way to teach the platform from a realized outcome. Only
// Converted is required; everything else is optional.
type MLFeedbackOptions struct {
	SubjectID      string                 `json:"subject_id,omitempty"`
	Features       map[string]interface{} `json:"features,omitempty"`
	Converted      bool                   `json:"converted"`
	PredictedScore *int                   `json:"predicted_score,omitempty"`
	Note           string                 `json:"note,omitempty"`
	ActionID       string                 `json:"action_id,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
	ActionFeatures map[string]interface{} `json:"action_features,omitempty"`
	Propensity     *float64               `json:"propensity,omitempty"`
	Reward         *float64               `json:"reward,omitempty"`
}

// MLFeedbackResult acknowledges a unified feedback call. OutcomeRecorded
// reports whether the labeled scoring outcome was persisted; BanditUpdated
// reports whether the bandit was taught (only when ActionID was given), with
// BanditN as the policy's running learning-example count and BanditError
// carrying any non-fatal bandit-update error.
type MLFeedbackResult struct {
	OK              bool   `json:"ok"`
	UseCase         string `json:"use_case"`
	Converted       bool   `json:"converted"`
	OutcomeRecorded bool   `json:"outcome_recorded"`
	BanditUpdated   bool   `json:"bandit_updated"`
	BanditN         int    `json:"bandit_n,omitempty"`
	BanditError     string `json:"bandit_error,omitempty"`
	Message         string `json:"message"`
}

// DecideNBA chooses the next best action among the candidate slate for
// useCase, returning the chosen action and the full scored slate.
func (m *MLResource) DecideNBA(ctx context.Context, useCase string, opts DecideNBAOptions) (*DecideNBAResult, error) {
	var result DecideNBAResult
	if err := m.http.Post(ctx, fmt.Sprintf("/api/v1/builtin-agents/ml/%s/nba/decide", useCase), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LearnNBA records the realized reward for a previously taken action,
// updating useCase's bandit policy.
func (m *MLResource) LearnNBA(ctx context.Context, useCase string, opts LearnNBAOptions) (*LearnNBAResult, error) {
	var result LearnNBAResult
	if err := m.http.Post(ctx, fmt.Sprintf("/api/v1/builtin-agents/ml/%s/nba/learn", useCase), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RecordFeedback is the single unified operator call for teaching the
// platform from a realized outcome. It persists the labeled outcome for
// useCase's scoring model (which retrains on the platform schedule) and, when
// ActionID is set, immediately teaches the bandit the realized reward. Reward
// defaults to 1/0 from Converted when omitted.
func (m *MLResource) RecordFeedback(ctx context.Context, useCase string, opts MLFeedbackOptions) (*MLFeedbackResult, error) {
	var result MLFeedbackResult
	if err := m.http.Post(ctx, fmt.Sprintf("/api/v1/builtin-agents/ml/%s/feedback", useCase), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
