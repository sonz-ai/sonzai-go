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

// SimulateRoundsParams is the request body for a single-call learning
// simulation. The platform runs the entire closed loop in-process — many
// rounds of (accrue outcomes → train the auto-tuned scoring model → bandit
// decide/reward/learn → off-policy evaluation) — so an integrator can trigger
// and observe the whole self-learning pipeline with one call instead of wiring
// TrainScoring + DecideNBA + LearnNBA + EvaluateOPE together by hand.
//
// It runs on a built-in synthetic Scenario (a learnable, reproducible world),
// so it is the canonical way to smoke-test, demo, or benchmark the learning
// machinery end-to-end without supplying real data. The use_case argument
// scopes an ephemeral model that never touches your production models or the
// cross-tenant global prior.
type SimulateRoundsParams struct {
	// Scenario selects the built-in synthetic world to learn (e.g.
	// "real_estate"). Leave empty for the platform default.
	Scenario string `json:"scenario,omitempty"`

	// Rounds is how many learning rounds to run (the platform clamps to a
	// sane range). Leave 0 for the platform default.
	Rounds int `json:"rounds,omitempty"`

	// Seed optionally makes the run reproducible (same seed → same curve).
	// Leave nil for a fresh cold-start each call.
	Seed *int `json:"seed,omitempty"`
}

// SimulateRoundPoint is one round of the learning curve. AUC is the scoring
// model's held-out accuracy that round; NBAValue is the bandit policy's
// off-policy value; OPEDR/CILow/CIHigh report the doubly-robust estimate and
// its confidence interval on that round's logged decisions.
type SimulateRoundPoint struct {
	Round     int     `json:"round"`
	N         int     `json:"n"`
	AUC       float64 `json:"auc"`
	NBAValue  float64 `json:"nba_value"`
	NBAReward float64 `json:"nba_reward"`
	OPEDR     float64 `json:"ope_dr"`
	CILow     float64 `json:"ci_low"`
	CIHigh    float64 `json:"ci_high"`
}

// SimulateModelSummary reports the final scoring model after the last round —
// the same shape TrainScoring returns, summarized for display.
type SimulateModelSummary struct {
	AUC               float64             `json:"auc"`
	Brier             float64             `json:"brier"`
	ECE               float64             `json:"ece"`
	N                 int                 `json:"n"`
	CalibrationMethod string              `json:"calibration_method"`
	BestParams        map[string]any      `json:"best_params"`
	Importances       []FeatureImportance `json:"importances"`
}

// SimulatePolicyActionScore is one action's learned value within a segment's
// policy, with a human-readable label.
type SimulatePolicyActionScore struct {
	ActionID string  `json:"action_id"`
	Score    float64 `json:"score"`
	Label    string  `json:"label"`
}

// SimulatePolicyEntry is the learned next-best-action policy for one lead
// segment: the recommended action plus every action's learned value.
type SimulatePolicyEntry struct {
	Segment           string                      `json:"segment"`
	RecommendedAction string                      `json:"recommended_action"`
	RecommendedLabel  string                      `json:"recommended_label"`
	Scores            []SimulatePolicyActionScore `json:"scores"`
}

// SimulateRoundsResult is the outcome of a learning simulation: the per-round
// learning curve plus the final trained model, the learned policy per segment,
// and the off-policy value estimate of the final policy.
type SimulateRoundsResult struct {
	Scenario     string                `json:"scenario"`
	ActionLabels map[string]string     `json:"action_labels"`
	Series       []SimulateRoundPoint  `json:"series"`
	Model        *SimulateModelSummary `json:"model"`
	Policy       []SimulatePolicyEntry `json:"policy"`
	OPE          struct {
		DR     float64 `json:"dr"`
		CILow  float64 `json:"ci_low"`
		CIHigh float64 `json:"ci_high"`
	} `json:"ope"`
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

// SimulateRounds runs the entire self-learning loop end-to-end in one call:
// the platform repeatedly accrues outcomes, retrains the auto-tuned scoring
// model, runs the contextual bandit (decide → reward → learn), and off-policy-
// evaluates the policy — for `Rounds` rounds on a built-in synthetic scenario —
// then returns the per-round learning curve plus the final model and learned
// policy. This is the easiest way to trigger and observe all of the platform's
// learning machinery without composing the individual primitives yourself.
//
// The run is self-contained and reproducible (pass a Seed) and uses an
// ephemeral model scoped to useCase, so it never affects production models or
// the cross-tenant global prior. The job runs for tens of seconds; pass a
// context with a generous deadline.
func (c *MLResource) SimulateRounds(ctx context.Context, useCase string, params SimulateRoundsParams) (*SimulateRoundsResult, error) {
	var result SimulateRoundsResult
	path := fmt.Sprintf("/api/v1/builtin-agents/ml/%s/simulate-rounds", useCase)
	if err := c.http.PostLongRunning(ctx, path, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
