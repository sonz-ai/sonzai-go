package sonzai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// Pin URL shapes, HTTP verbs, and body payloads for the generalized ML & RL
// endpoints so the Go SDK stays in sync with the builtin-agents/ml handlers
// in services/platform/api and the sibling Python / TypeScript SDKs. Response
// parsing is smoke-tested by decoding the load-bearing fields.

func TestML_TrainScoring_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"auc":                0.91,
			"logloss":            0.32,
			"brier":              0.11,
			"brier_uncalibrated": 0.14,
			"brier_baseline":     0.25,
			"ece":                0.03,
			"n":                  1200,
			"importances": []map[string]any{
				{"name": "net_worth", "gain": 42.5},
				{"name": "intent", "gain": 18.2},
			},
			"best_params":        map[string]any{"max_depth": 6.0, "eta": 0.1},
			"calibration_method": "isotonic",
			"trials":             40,
			"model_version":      7,
		})
	})
	client := newTestClient(t, h)

	budget := 40
	res, err := client.ML.TrainScoring(context.Background(), "lead_score", TrainScoringParams{
		Rows: []ScoringTrainRow{
			{Features: map[string]any{"net_worth": 5_000_000, "intent": "high"}, Label: 1},
			{Features: map[string]any{"net_worth": 50_000, "intent": "low"}, Label: 0},
		},
		OptimizeBudget: &budget,
	})
	if err != nil {
		t.Fatalf("TrainScoring: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/ml/lead_score/scoring/train" {
		t.Errorf("path: got %q", seen.path)
	}
	rows, ok := seen.body["rows"].([]any)
	if !ok || len(rows) != 2 {
		t.Fatalf("body rows: got %v", seen.body["rows"])
	}
	if got := seen.body["optimize_budget"]; got != float64(40) {
		t.Errorf("optimize_budget: got %v, want 40", got)
	}
	if res.AUC != 0.91 || res.ModelVersion != 7 || res.CalibrationMethod != "isotonic" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
	if len(res.Importances) != 2 || res.Importances[0].Name != "net_worth" || res.Importances[0].Gain != 42.5 {
		t.Errorf("decoded importances mismatch: %+v", res.Importances)
	}
}

func TestML_TrainScoring_OmitsBudgetWhenNil(t *testing.T) {
	var seen map[string]any
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen)
		_ = json.NewEncoder(w).Encode(map[string]any{"model_version": 1})
	})
	client := newTestClient(t, h)

	_, err := client.ML.TrainScoring(context.Background(), "churn", TrainScoringParams{
		Rows: []ScoringTrainRow{{Features: map[string]any{"x": 1}, Label: 0}},
	})
	if err != nil {
		t.Fatalf("TrainScoring: %v", err)
	}
	if _, present := seen["optimize_budget"]; present {
		t.Errorf("optimize_budget should be omitted when nil, body: %v", seen)
	}
}

func TestML_PredictScore_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"score":              0.73,
			"raw":                0.81,
			"model_version":      7,
			"served_from":        "live",
			"calibration_method": "isotonic",
		})
	})
	client := newTestClient(t, h)

	res, err := client.ML.PredictScore(context.Background(), "lead_score", PredictScoreParams{
		Features: map[string]any{"net_worth": 2_000_000},
	})
	if err != nil {
		t.Fatalf("PredictScore: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/ml/lead_score/scoring/predict" {
		t.Errorf("path: got %q", seen.path)
	}
	if res.Score != 0.73 || res.Raw != 0.81 || res.ServedFrom != "live" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
}

func TestML_DecideNBA_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"action_id":  "sms",
			"propensity": 0.62,
			"scores": []map[string]any{
				{"action_id": "sms", "score": 0.71, "propensity": 0.62},
				{"action_id": "call", "score": 0.55, "propensity": 0.38},
			},
			"explore": false,
			"model_n": 5000,
		})
	})
	client := newTestClient(t, h)

	explore := true
	res, err := client.ML.DecideNBA(context.Background(), "outreach", DecideNBAParams{
		Context: map[string]any{"hour": 14, "band": "hot"},
		Actions: []NBAAction{
			{ID: "sms", Features: map[string]any{"cost": 0.01}},
			{ID: "call", Features: map[string]any{"cost": 1.0}},
		},
		Explore: &explore,
	})
	if err != nil {
		t.Fatalf("DecideNBA: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/ml/outreach/nba/decide" {
		t.Errorf("path: got %q", seen.path)
	}
	if got := seen.body["explore"]; got != true {
		t.Errorf("explore body: got %v, want true", got)
	}
	actions, ok := seen.body["actions"].([]any)
	if !ok || len(actions) != 2 {
		t.Fatalf("body actions: got %v", seen.body["actions"])
	}
	if res.ActionID != "sms" || res.Propensity != 0.62 || res.Explore {
		t.Errorf("decoded result mismatch: %+v", res)
	}
	if len(res.Scores) != 2 || res.Scores[1].ActionID != "call" || res.Scores[1].Score != 0.55 {
		t.Errorf("decoded scores mismatch: %+v", res.Scores)
	}
	if res.ModelN != 5000 {
		t.Errorf("model_n: got %d, want 5000", res.ModelN)
	}
}

func TestML_LearnNBA_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "n": 5001})
	})
	client := newTestClient(t, h)

	prop := 0.62
	res, err := client.ML.LearnNBA(context.Background(), "outreach", LearnNBAParams{
		Context:        map[string]any{"hour": 14, "band": "hot"},
		ActionID:       "sms",
		ActionFeatures: map[string]any{"cost": 0.01},
		Propensity:     &prop,
		Reward:         1.0,
	})
	if err != nil {
		t.Fatalf("LearnNBA: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/ml/outreach/nba/learn" {
		t.Errorf("path: got %q", seen.path)
	}
	if got := seen.body["propensity"]; got != 0.62 {
		t.Errorf("propensity body: got %v, want 0.62", got)
	}
	if !res.OK || res.N != 5001 {
		t.Errorf("decoded result mismatch: %+v", res)
	}
}

func TestML_EvaluateOPE_URLBodyAndDecode(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ips":          0.42,
			"snips":        0.44,
			"dr":           0.43,
			"ci_low":       0.38,
			"ci_high":      0.48,
			"n":            3000,
			"ess":          1820.5,
			"estimator_ci": "dr",
		})
	})
	client := newTestClient(t, h)

	res, err := client.ML.EvaluateOPE(context.Background(), "outreach", EvaluateOPEParams{
		Logged: []OPELoggedRecord{
			{Context: map[string]any{"band": "hot"}, ActionID: "sms", Propensity: 0.6, Reward: 1.0},
			{Context: map[string]any{"band": "cold"}, ActionID: "call", Propensity: 0.4, Reward: 0.0},
		},
	})
	if err != nil {
		t.Fatalf("EvaluateOPE: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/builtin-agents/ml/outreach/ope/evaluate" {
		t.Errorf("path: got %q", seen.path)
	}
	logged, ok := seen.body["logged"].([]any)
	if !ok || len(logged) != 2 {
		t.Fatalf("body logged: got %v", seen.body["logged"])
	}
	if res.IPS != 0.42 || res.SNIPS != 0.44 || res.DR != 0.43 {
		t.Errorf("decoded estimates mismatch: %+v", res)
	}
	if res.CILow != 0.38 || res.CIHigh != 0.48 || res.EstimatorCI != "dr" {
		t.Errorf("decoded CI mismatch: %+v", res)
	}
	if res.N != 3000 || res.ESS != 1820.5 {
		t.Errorf("decoded n/ess mismatch: %+v", res)
	}
}
