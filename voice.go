package sonzai

import (
	"context"
	"strconv"
)

// VoiceResource provides per-agent voice live operations.
type VoiceResource struct {
	http *httpClient
}

// VoicesResource provides global voice catalog operations.
type VoicesResource struct {
	http *httpClient
}

// Voice represents an available voice in the catalog.
type Voice struct {
	VoiceID        string `json:"voice_id"`
	VoiceName      string `json:"voice_name"`
	Gender         string `json:"gender"`
	Tier           int    `json:"tier"`
	Provider       string `json:"provider"`
	Language       string `json:"language"`
	Accent         string `json:"accent,omitempty"`
	AgeProfile     string `json:"age_profile,omitempty"`
	Description    string `json:"description,omitempty"`
	SampleAudioURL string `json:"sample_audio_url,omitempty"`
	Availability   string `json:"availability"`
}

// VoiceListResponse is the response from listing voices.
type VoiceListResponse struct {
	Voices     []Voice `json:"voices"`
	TotalCount int     `json:"total_count"`
	HasMore    bool    `json:"has_more"`
}

// VoiceListOptions configures a voice listing request.
type VoiceListOptions struct {
	Tier     int
	Gender   string
	Language string
	Limit    int
	Offset   int
}

// List returns available voices from the catalog.
func (v *VoicesResource) List(ctx context.Context, opts *VoiceListOptions) (*VoiceListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Tier > 0 {
			params["tier"] = strconv.Itoa(opts.Tier)
		}
		if opts.Gender != "" {
			params["gender"] = opts.Gender
		}
		if opts.Language != "" {
			params["language"] = opts.Language
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = strconv.Itoa(opts.Offset)
		}
	}

	var result VoiceListResponse
	err := v.http.Get(ctx, "/api/v1/voices", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
