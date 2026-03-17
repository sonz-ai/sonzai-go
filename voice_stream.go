package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sonz-ai/sonzai-go/internal/ws"
)

// VoiceStreamConfig configures a voice streaming session.
type VoiceStreamConfig struct {
	AgentID     string `json:"agent_id"`
	VoiceName   string `json:"voice_name,omitempty"`
	Language    string `json:"language,omitempty"`
	AudioFormat string `json:"audio_format,omitempty"` // "webm", "ogg", etc. Default: "audio/webm;codecs=opus"
}

// VoiceStream is a bidirectional WebSocket connection for real-time voice chat.
// Send audio with SendAudio, signal end-of-speech with EndOfSpeech,
// and receive events with Recv.
type VoiceStream struct {
	conn *ws.Conn
}

// VoiceStreamEvent represents a server event from the voice stream.
type VoiceStreamEvent struct {
	// Type is the event type: "ready", "vad", "transcript", "response_delta",
	// "turn_complete", "error", or "audio" (binary audio data).
	Type string `json:"type"`

	// SessionID is set on "ready" events.
	SessionID string `json:"session_id,omitempty"`

	// Speaking is set on "vad" events.
	Speaking *bool `json:"speaking,omitempty"`

	// Text is set on "transcript" and "response_delta" events.
	Text string `json:"text,omitempty"`

	// ContinuationToken is set on "turn_complete" events.
	ContinuationToken string `json:"continuation_token,omitempty"`

	// ContentType is set on "turn_complete" events (audio MIME type).
	ContentType string `json:"content_type,omitempty"`

	// Error is set on "error" events.
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`

	// Audio is set on "audio" events (raw binary TTS audio data).
	Audio []byte `json:"-"`
}

// VoiceStreamToken represents the token needed to establish a voice WebSocket.
type VoiceStreamToken struct {
	WSURL     string `json:"wsUrl"`
	AuthToken string `json:"authToken"`
}

// VoiceTokenOptions configures a voice WebSocket token request.
type VoiceTokenOptions struct {
	VoiceName     string         `json:"voiceName,omitempty"`
	Language      string         `json:"language,omitempty"`
	EntityContext *EntityContext `json:"entityContext,omitempty"`
}

// EntityContext provides agent identity for voice sessions.
type EntityContext struct {
	Name        string `json:"name,omitempty"`
	Personality string `json:"personality,omitempty"`
}

// GetVoiceToken obtains a short-lived token for WebSocket voice streaming.
// The token expires in 60 seconds and is single-use.
func (v *VoiceResource) GetToken(ctx context.Context, agentID string, opts VoiceTokenOptions) (*VoiceStreamToken, error) {
	var result VoiceStreamToken
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/ws-token", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Stream opens a bidirectional WebSocket for real-time voice chat.
// The token should be obtained from GetToken. After connecting, the session
// is authenticated automatically and a "ready" event is sent.
//
// Usage:
//
//	token, _ := client.Agents.Voice.GetToken(ctx, agentID, opts)
//	stream, _ := client.Agents.Voice.Stream(ctx, token)
//	defer stream.Close()
//
//	// Send audio chunks
//	stream.SendAudio(audioBytes)
//
//	// Signal end of speech (or let server detect silence)
//	stream.EndOfSpeech()
//
//	// Receive events
//	for {
//	    event, err := stream.Recv()
//	    if err == io.EOF { break }
//	    switch event.Type {
//	    case "transcript": fmt.Println("User:", event.Text)
//	    case "response_delta": fmt.Print(event.Text)
//	    case "audio": playAudio(event.Audio)
//	    case "turn_complete": // ready for next turn
//	    }
//	}
func (v *VoiceResource) Stream(ctx context.Context, token *VoiceStreamToken) (*VoiceStream, error) {
	conn, err := ws.Dial(ctx, token.WSURL)
	if err != nil {
		return nil, fmt.Errorf("voice stream: %w", err)
	}

	// Send auth token as first text message.
	if err := conn.WriteText([]byte(token.AuthToken)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("voice stream: send auth: %w", err)
	}

	return &VoiceStream{conn: conn}, nil
}

// Recv reads the next event from the voice stream.
// Returns io.EOF when the connection is closed.
func (s *VoiceStream) Recv() (*VoiceStreamEvent, error) {
	opcode, data, err := s.conn.ReadMessage()
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("voice stream recv: %w", err)
	}

	// Binary frames are TTS audio.
	if opcode == ws.OpBinary {
		return &VoiceStreamEvent{
			Type:  "audio",
			Audio: data,
		}, nil
	}

	// Text frames are JSON events.
	var event VoiceStreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("voice stream: unmarshal event: %w", err)
	}
	return &event, nil
}

// SendAudio sends a binary audio chunk to the server.
func (s *VoiceStream) SendAudio(audio []byte) error {
	return s.conn.WriteBinary(audio)
}

// EndOfSpeech signals the server that the user has finished speaking,
// triggering immediate processing without waiting for silence detection.
func (s *VoiceStream) EndOfSpeech() error {
	return s.conn.WriteText([]byte(`{"type":"end_of_speech"}`))
}

// Configure sends a config message to change audio format, voice, or language
// mid-session without reconnecting.
func (s *VoiceStream) Configure(audioFormat, voiceName, language string) error {
	msg := map[string]string{"type": "config"}
	if audioFormat != "" {
		msg["audio_format"] = audioFormat
	}
	if voiceName != "" {
		msg["voice_name"] = voiceName
	}
	if language != "" {
		msg["language"] = language
	}
	data, _ := json.Marshal(msg)
	return s.conn.WriteText(data)
}

// Close closes the voice stream.
func (s *VoiceStream) Close() error {
	return s.conn.Close()
}
