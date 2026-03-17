package sonzai

import "context"

// VoiceResource provides voice operations.
type VoiceResource struct {
	http *httpClient
}

// TextToSpeech generates speech from text.
func (v *VoiceResource) TextToSpeech(ctx context.Context, params TextToSpeechParams) (*TextToSpeechResult, error) {
	var result TextToSpeechResult
	err := v.http.post(ctx, "/api/v1/voice/tts", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VoiceMatch matches personality to an appropriate voice.
func (v *VoiceResource) VoiceMatch(ctx context.Context, params VoiceMatchParams) (*VoiceMatchResult, error) {
	var result VoiceMatchResult
	err := v.http.post(ctx, "/api/v1/voice/match", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListVoices returns available voices, optionally filtered by gender.
func (v *VoiceResource) ListVoices(ctx context.Context, gender string) (*ListVoicesResult, error) {
	params := map[string]string{}
	if gender != "" {
		params["gender"] = gender
	}
	var result ListVoicesResult
	err := v.http.get(ctx, "/api/v1/voice/voices", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VoiceChat performs a single-turn voice chat (STT → LLM → TTS).
func (v *VoiceResource) VoiceChat(ctx context.Context, params VoiceChatParams) (*VoiceChatResult, error) {
	var result VoiceChatResult
	err := v.http.post(ctx, "/api/v1/voice/chat", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
