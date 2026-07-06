package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMLDecideNBA(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/builtin-agents/ml/lead_score/nba/decide":
			var body DecideNBAOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if len(body.Actions) != 2 || body.Actions[0].ID != "call" || body.Explore == nil || *body.Explore != true {
				t.Fatalf("unexpected decide body: %+v", body)
			}
			jsonResponse(w, 200, DecideNBAResult{
				ActionID:   "call",
				Propensity: 0.6,
				Scores: []NBAActionScore{
					{ActionID: "call", Score: 0.9, Propensity: 0.6},
					{ActionID: "sms", Score: 0.4, Propensity: 0.4},
				},
				Explore: true,
				ModelN:  42,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	explore := true
	result, err := client.ML.DecideNBA(context.Background(), "lead_score", DecideNBAOptions{
		Context: map[string]interface{}{"score": 80},
		Actions: []NBAAction{
			{ID: "call", Features: map[string]interface{}{"cost": 5}},
			{ID: "sms", Features: map[string]interface{}{"cost": 1}},
		},
		Explore: &explore,
	})
	if err != nil {
		t.Fatalf("DecideNBA: %v", err)
	}
	if result.ActionID != "call" || result.Propensity != 0.6 || result.ModelN != 42 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.Scores) != 2 {
		t.Fatalf("unexpected scores: %+v", result.Scores)
	}
}

func TestMLLearnNBA(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/builtin-agents/ml/lead_score/nba/learn":
			var body LearnNBAOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.ActionID != "call" || body.Reward != 1 || body.Propensity == nil || *body.Propensity != 0.6 {
				t.Fatalf("unexpected learn body: %+v", body)
			}
			jsonResponse(w, 200, LearnNBAResult{OK: true, N: 43})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	propensity := 0.6
	result, err := client.ML.LearnNBA(context.Background(), "lead_score", LearnNBAOptions{
		ActionID:   "call",
		Reward:     1,
		Propensity: &propensity,
	})
	if err != nil {
		t.Fatalf("LearnNBA: %v", err)
	}
	if !result.OK || result.N != 43 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestMLRecordFeedback(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/builtin-agents/ml/lead_score/feedback":
			var body MLFeedbackOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.SubjectID != "lead-42" || !body.Converted || body.ActionID != "call" {
				t.Fatalf("unexpected feedback body: %+v", body)
			}
			jsonResponse(w, 200, MLFeedbackResult{
				OK:              true,
				UseCase:         "lead_score",
				Converted:       true,
				OutcomeRecorded: true,
				BanditUpdated:   true,
				BanditN:         44,
				Message:         "outcome recorded; the scoring model retrains automatically on the platform schedule",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.ML.RecordFeedback(context.Background(), "lead_score", MLFeedbackOptions{
		SubjectID: "lead-42",
		Features:  map[string]interface{}{"budget": 3000000, "intent": 0.7},
		Converted: true,
		ActionID:  "call",
	})
	if err != nil {
		t.Fatalf("RecordFeedback: %v", err)
	}
	if !result.OK || !result.OutcomeRecorded || !result.BanditUpdated || result.BanditN != 44 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
