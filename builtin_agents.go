package sonzai

import (
	"context"
	"time"
)

// BuiltinAgentsResource provides operations for the platform's built-in agents
// (e.g. lead_score): recording realized outcomes and reading the learned
// calibration snapshot.
type BuiltinAgentsResource struct {
	http *httpClient
}

// LeadOutcomeOptions configures a realized lead-scoring outcome to record.
type LeadOutcomeOptions struct {
	LeadRef        string                 `json:"lead_ref"`
	Outcome        string                 `json:"outcome"`
	PredictedScore int                    `json:"predicted_score,omitempty"`
	PredictedBand  string                 `json:"predicted_band,omitempty"`
	Features       map[string]interface{} `json:"features,omitempty"`
	ScoreSignal    string                 `json:"score_signal,omitempty"`
	Note           string                 `json:"note,omitempty"`
}

// SegmentCalibration is the calibration for one lead-scoring context segment.
type SegmentCalibration struct {
	Segment     string  `json:"segment"`
	N           int     `json:"n"`
	Conversions int     `json:"conversions"`
	PHat        float64 `json:"p_hat"`
	Adjust      int     `json:"adjust"`
}

// LeadBandAccuracy is the predicted-vs-actual panel row for one lead-scoring band.
type LeadBandAccuracy struct {
	Band           string  `json:"band"`
	N              int     `json:"n"`
	Conversions    int     `json:"conversions"`
	PredictedRate  float64 `json:"predicted_rate"`
	ActualRate     float64 `json:"actual_rate"`
	AvgScore       float64 `json:"avg_score"`
	CalibrationGap float64 `json:"calibration_gap"`
}

// LeadCalibration is the lead_score agent's learned calibration snapshot:
// per-segment score adjustments and per-band predicted-vs-actual accuracy.
type LeadCalibration struct {
	Segments  []SegmentCalibration `json:"segments"`
	Bands     []LeadBandAccuracy   `json:"bands"`
	BaseRate  float64              `json:"base_rate"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// RecordLeadOutcome records a realized lead-scoring outcome (won/lost/…) and
// returns the recomputed calibration the lead_score agent applies to future
// leads in the same segment.
func (b *BuiltinAgentsResource) RecordLeadOutcome(ctx context.Context, opts LeadOutcomeOptions) (*LeadCalibration, error) {
	var result LeadCalibration
	if err := b.http.Post(ctx, "/api/v1/builtin-agents/lead_score/outcome", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLeadCalibration returns the current lead-scoring calibration: per-segment
// score adjustments and per-band predicted-vs-actual accuracy.
func (b *BuiltinAgentsResource) GetLeadCalibration(ctx context.Context) (*LeadCalibration, error) {
	var result LeadCalibration
	if err := b.http.Get(ctx, "/api/v1/builtin-agents/lead_score/calibration", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
