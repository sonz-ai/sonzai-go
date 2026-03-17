package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// VoiceResource provides per-agent voice operations.
type VoiceResource struct {
	http *httpClient
}

// VoiceMatchOptions configures a voice matching request.
type VoiceMatchOptions struct {
	Big5            *Big5Scores `json:"big5,omitempty"`
	PreferredGender string      `json:"preferred_gender,omitempty"`
}

// VoiceMatchResponse is the response from voice matching.
type VoiceMatchResponse struct {
	VoiceID    string  `json:"voice_id"`
	VoiceName  string  `json:"voice_name"`
	MatchScore float64 `json:"match_score"`
	Reasoning  string  `json:"reasoning,omitempty"`
}

// EmotionalContext provides emotional hints for TTS generation.
type EmotionalContext struct {
	Themes []string `json:"themes,omitempty"`
	Tone   string   `json:"tone,omitempty"`
}

// TTSOptions configures a text-to-speech request.
type TTSOptions struct {
	Text             string           `json:"text"`
	VoiceName        string           `json:"voice_name,omitempty"`
	Language         string           `json:"language,omitempty"`
	EmotionalContext *EmotionalContext `json:"emotional_context,omitempty"`
}

// TTSResponse is the response from text-to-speech.
type TTSResponse struct {
	Audio       string `json:"audio"`        // base64-encoded audio
	ContentType string `json:"content_type"`
	VoiceName   string `json:"voice_name,omitempty"`
	DurationMs  int    `json:"duration_ms,omitempty"`
}

// Match finds the best matching voice for an agent based on personality and preferences.
func (v *VoiceResource) Match(ctx context.Context, agentID string, opts VoiceMatchOptions) (*VoiceMatchResponse, error) {
	var result VoiceMatchResponse
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/match", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VoiceChatOptions configures a single-turn voice chat request.
type VoiceChatOptions struct {
	UserID            string `json:"user_id,omitempty"`
	Audio             string `json:"audio"`                  // base64-encoded audio
	AudioFormat       string `json:"audio_format,omitempty"` // "webm", "wav", etc.
	VoiceName         string `json:"voice_name,omitempty"`
	ContinuationToken string `json:"continuation_token,omitempty"`
	Language          string `json:"language,omitempty"`
}

// VoiceChatResponse is the response from single-turn voice chat.
type VoiceChatResponse struct {
	Transcript        string `json:"transcript"`
	Response          string `json:"response"`
	Audio             string `json:"audio"` // base64-encoded
	ContentType       string `json:"content_type"`
	ContinuationToken string `json:"continuation_token,omitempty"`
}

// Chat performs a single-turn voice chat: send audio, receive text + audio response.
func (v *VoiceResource) Chat(ctx context.Context, agentID string, opts VoiceChatOptions) (*VoiceChatResponse, error) {
	var result VoiceChatResponse
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/chat", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TTS converts text to speech using the agent's voice.
func (v *VoiceResource) TTS(ctx context.Context, agentID string, opts TTSOptions) (*TTSResponse, error) {
	var result TTSResponse
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/tts", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
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
