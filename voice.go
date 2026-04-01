package sonzai

import (
	"context"
	"fmt"
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

// TTSOptions configures a text-to-speech request.
type TTSOptions struct {
	Text         string `json:"text"`
	VoiceName    string `json:"voiceName,omitempty"`
	Language     string `json:"language,omitempty"`
	OutputFormat string `json:"outputFormat,omitempty"` // "wav" or "opus"
}

// TTSResponse is the response from text-to-speech synthesis.
type TTSResponse struct {
	Audio       string `json:"audio"`
	ContentType string `json:"contentType"`
	DurationMs  int64  `json:"durationMs,omitempty"`
	Usage       *struct {
		PromptTokens     int    `json:"promptTokens"`
		CompletionTokens int    `json:"completionTokens"`
		TotalTokens      int    `json:"totalTokens"`
		Model            string `json:"model"`
	} `json:"usage,omitempty"`
}

// STTOptions configures a speech-to-text request.
type STTOptions struct {
	Audio       string `json:"audio"`
	AudioFormat string `json:"audioFormat"`
	Language    string `json:"language,omitempty"`
}

// STTResponse is the response from speech-to-text transcription.
type STTResponse struct {
	Transcript   string  `json:"transcript"`
	Confidence   float64 `json:"confidence"`
	LanguageCode string  `json:"languageCode,omitempty"`
}

// TTS converts text to speech audio using Gemini TTS.
func (v *VoiceResource) TTS(ctx context.Context, agentID string, opts TTSOptions) (*TTSResponse, error) {
	var result TTSResponse
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/tts", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// STT transcribes audio to text using Gemini STT.
func (v *VoiceResource) STT(ctx context.Context, agentID string, opts STTOptions) (*STTResponse, error) {
	var result STTResponse
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/stt", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
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
