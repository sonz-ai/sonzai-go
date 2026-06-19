package sonzai

import (
	"context"
	"fmt"
)

// The Sonzai ML resource exposes the platform's generalized, multi-tenant /
// multi-vertical ML & RL primitives: supervised scoring (train + calibrated
// predict), contextual-bandit next-best-action (decide + learn), and
// off-policy evaluation. Every endpoint is keyed by a free-form use_case
// string (e.g. "lead_score", "claim_triage", "churn") so a single tenant can
// run many independent models. All calls are tenant-scoped server-side by the
// caller's API key — no tenant argument is needed.
//
// Endpoints live under /api/v1/builtin-agents/ml/{use_case}/...

// ScoringTrainRow is one labeled training example for supervised scoring.
type ScoringTrainRow struct {
	// Features is the example's feature map. Keys and value types are
	// use-case-defined; the platform infers the schema across the batch.
	Features map[string]any `json:"features"`

	// Label is the binary target (0 or 1).
	Label int `json:"label"`
}

// TrainScoringParams is the request body for training a scoring model.
type TrainScoringParams struct {
	// Rows are the labeled training examples. Required.
	Rows []ScoringTrainRow `json:"rows"`

	// OptimizeBudget optionally caps the hyperparameter-search trial budget.
	// Leave nil for the platform default.
	OptimizeBudget *int `json:"optimize_budget,omitempty"`
}

// FeatureImportance reports a single feature's contribution to the trained
// model, by total gain.
type FeatureImportance struct {
	Name string  `json:"name"`
	Gain float64 `json:"gain"`
}

// TrainScoringResult reports the trained model's held-out metrics, the chosen
// hyperparameters, and the new model version. Brier is post-calibration;
// BrierUncalibrated and BrierBaseline contextualize the calibration lift.
type TrainScoringResult struct {
	AUC               float64             `json:"auc"`
	Logloss           float64             `json:"logloss"`
	Brier             float64             `json:"brier"`
	BrierUncalibrated float64             `json:"brier_uncalibrated"`
	BrierBaseline     float64             `json:"brier_baseline"`
	ECE               float64             `json:"ece"`
	N                 int                 `json:"n"`
	Importances       []FeatureImportance `json:"importances"`
	BestParams        map[string]any      `json:"best_params"`
	CalibrationMethod string              `json:"calibration_method"`
	Trials            int                 `json:"trials"`
	ModelVersion      int                 `json:"model_version"`
}

// PredictScoreParams is the request body for scoring a single example.
type PredictScoreParams struct {
	// Features is the example's feature map. Required.
	Features map[string]any `json:"features"`
}

// PredictScoreResult is the calibrated prediction for one example. Score is
// the calibrated probability; Raw is the model's pre-calibration output.
// ServedFrom reports which model tier answered (e.g. cache vs. live).
type PredictScoreResult struct {
	Score             float64 `json:"score"`
	Raw               float64 `json:"raw"`
	ModelVersion      int     `json:"model_version"`
	ServedFrom        string  `json:"served_from"`
	CalibrationMethod string  `json:"calibration_method"`
}

// NBAAction is one candidate action for a next-best-action decision.
type NBAAction struct {
	// ID identifies the action. Required.
	ID string `json:"id"`

	// Features is the action's feature map (combined with the request
	// context to score the action).
	Features map[string]any `json:"features"`
}

// DecideNBAParams is the request body for a next-best-action decision.
type DecideNBAParams struct {
	// Context is the decision context shared across all candidate actions.
	// Required.
	Context map[string]any `json:"context"`

	// Actions are the candidate actions to choose among. Required.
	Actions []NBAAction `json:"actions"`

	// Explore optionally forces exploration on (true) or off (false). Leave
	// nil to let the policy decide.
	Explore *bool `json:"explore,omitempty"`
}

// NBAActionScore is the scored result for one candidate action.
type NBAActionScore struct {
	ActionID   string  `json:"action_id"`
	Score      float64 `json:"score"`
	Propensity float64 `json:"propensity"`
}

// DecideNBAResult is the policy's chosen action plus the full scored slate.
// Propensity is the probability the policy assigned to the chosen action
// (record it and pass it back to LearnNBA for unbiased learning). Explore
// reports whether the choice was an exploration step.
type DecideNBAResult struct {
	ActionID   string           `json:"action_id"`
	Propensity float64          `json:"propensity"`
	Scores     []NBAActionScore `json:"scores"`
	Explore    bool             `json:"explore"`
	ModelN     int              `json:"model_n"`
}

// LearnNBAParams is the request body for recording the realized reward of a
// next-best-action decision.
type LearnNBAParams struct {
	// Context is the decision context that was used. Required.
	Context map[string]any `json:"context"`

	// ActionID is the action that was taken. Required.
	ActionID string `json:"action_id"`

	// ActionFeatures is the taken action's feature map.
	ActionFeatures map[string]any `json:"action_features"`

	// Propensity is the probability the policy assigned to the taken action
	// at decision time (from DecideNBAResult.Propensity). Pass it for
	// unbiased off-policy learning; leave nil if unknown.
	Propensity *float64 `json:"propensity,omitempty"`

	// Reward is the realized reward for the taken action. Required.
	Reward float64 `json:"reward"`
}

// LearnNBAResult acknowledges a recorded reward. N is the running count of
// learning examples the policy has ingested.
type LearnNBAResult struct {
	OK bool `json:"ok"`
	N  int  `json:"n"`
}

// OPELoggedRecord is one logged decision used for off-policy evaluation:
// the context, the action taken, the logging policy's propensity for that
// action, and the realized reward.
type OPELoggedRecord struct {
	Context    map[string]any `json:"context"`
	ActionID   string         `json:"action_id"`
	Propensity float64        `json:"propensity"`
	Reward     float64        `json:"reward"`
}

// EvaluateOPEParams is the request body for off-policy evaluation.
type EvaluateOPEParams struct {
	// Logged are the logged decisions to evaluate the current policy against.
	// Required.
	Logged []OPELoggedRecord `json:"logged"`
}

// EvaluateOPEResult reports off-policy value estimates for the current policy
// against the logged data: inverse-propensity (IPS), self-normalized IPS
// (SNIPS), and doubly-robust (DR), with a confidence interval and the
// effective sample size (ESS). EstimatorCI names which estimator the CI
// bounds describe.
type EvaluateOPEResult struct {
	IPS         float64 `json:"ips"`
	SNIPS       float64 `json:"snips"`
	DR          float64 `json:"dr"`
	CILow       float64 `json:"ci_low"`
	CIHigh      float64 `json:"ci_high"`
	N           int     `json:"n"`
	ESS         float64 `json:"ess"`
	EstimatorCI string  `json:"estimator_ci"`
}

// MLResource provides the platform's generalized ML & RL primitives keyed by
// a use_case string: supervised scoring, contextual-bandit next-best-action,
// and off-policy evaluation.
type MLResource struct {
	http *httpClient
}

// TrainScoring trains (or retrains) the calibrated scoring model for the given
// use case from the labeled rows and returns its held-out metrics, chosen
// hyperparameters, and new model version. Training can run for a while on
// large batches; pass a context with a generous deadline.
func (c *MLResource) TrainScoring(ctx context.Context, useCase string, params TrainScoringParams) (*TrainScoringResult, error) {
	var result TrainScoringResult
	path := fmt.Sprintf("/api/v1/builtin-agents/ml/%s/scoring/train", useCase)
	if err := c.http.PostLongRunning(ctx, path, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PredictScore returns the calibrated score for a single example using the
// use case's current scoring model.
func (c *MLResource) PredictScore(ctx context.Context, useCase string, params PredictScoreParams) (*PredictScoreResult, error) {
	var result PredictScoreResult
	path := fmt.Sprintf("/api/v1/builtin-agents/ml/%s/scoring/predict", useCase)
	if err := c.http.Post(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DecideNBA chooses the next best action among the candidate slate for the
// given use case and context, returning the chosen action and the full scored
// slate. Record the returned Propensity and pass it back to LearnNBA when the
// reward is realized.
func (c *MLResource) DecideNBA(ctx context.Context, useCase string, params DecideNBAParams) (*DecideNBAResult, error) {
	var result DecideNBAResult
	path := fmt.Sprintf("/api/v1/builtin-agents/ml/%s/nba/decide", useCase)
	if err := c.http.Post(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LearnNBA records the realized reward for a previously taken action, updating
// the use case's bandit policy.
func (c *MLResource) LearnNBA(ctx context.Context, useCase string, params LearnNBAParams) (*LearnNBAResult, error) {
	var result LearnNBAResult
	path := fmt.Sprintf("/api/v1/builtin-agents/ml/%s/nba/learn", useCase)
	if err := c.http.Post(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EvaluateOPE runs off-policy evaluation of the use case's current policy
// against a batch of logged decisions, returning IPS/SNIPS/DR value estimates
// with a confidence interval and effective sample size.
func (c *MLResource) EvaluateOPE(ctx context.Context, useCase string, params EvaluateOPEParams) (*EvaluateOPEResult, error) {
	var result EvaluateOPEResult
	path := fmt.Sprintf("/api/v1/builtin-agents/ml/%s/ope/evaluate", useCase)
	if err := c.http.Post(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
