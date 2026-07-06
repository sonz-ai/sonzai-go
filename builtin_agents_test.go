package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestBuiltinAgentsRecordLeadOutcome(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/builtin-agents/lead_score/outcome":
			var body LeadOutcomeOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.LeadRef != "lead-1" || body.Outcome != "won" || body.PredictedScore != 72 || body.PredictedBand != "Hot" {
				t.Fatalf("unexpected outcome body: %+v", body)
			}
			if body.Features["budget"] != float64(500000) {
				t.Fatalf("unexpected features: %+v", body.Features)
			}
			jsonResponse(w, 200, LeadCalibration{
				Segments: []SegmentCalibration{{Segment: "hot|financed", N: 10, Conversions: 4, PHat: 0.4, Adjust: 5}},
				Bands:    []LeadBandAccuracy{{Band: "Hot", N: 10, Conversions: 4, PredictedRate: 0.5, ActualRate: 0.4, AvgScore: 70, CalibrationGap: -0.1}},
				BaseRate: 0.25,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	cal, err := client.BuiltinAgents.RecordLeadOutcome(context.Background(), LeadOutcomeOptions{
		LeadRef:        "lead-1",
		Outcome:        "won",
		PredictedScore: 72,
		PredictedBand:  "Hot",
		Features:       map[string]interface{}{"budget": 500000},
	})
	if err != nil {
		t.Fatalf("RecordLeadOutcome: %v", err)
	}
	if len(cal.Segments) != 1 || cal.Segments[0].Segment != "hot|financed" {
		t.Fatalf("unexpected calibration: %+v", cal)
	}
	if len(cal.Bands) != 1 || cal.Bands[0].Band != "Hot" {
		t.Fatalf("unexpected bands: %+v", cal)
	}
	if cal.BaseRate != 0.25 {
		t.Fatalf("unexpected base rate: %v", cal.BaseRate)
	}
}

func TestBuiltinAgentsGetLeadCalibration(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/builtin-agents/lead_score/calibration":
			jsonResponse(w, 200, LeadCalibration{BaseRate: 0.3})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	cal, err := client.BuiltinAgents.GetLeadCalibration(context.Background())
	if err != nil {
		t.Fatalf("GetLeadCalibration: %v", err)
	}
	if cal.BaseRate != 0.3 {
		t.Fatalf("unexpected base rate: %v", cal.BaseRate)
	}
}
