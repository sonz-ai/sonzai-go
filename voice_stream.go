package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sonz-ai/sonzai-go/internal/ws"
)

// VoiceStreamConfig configures a voice live streaming session.
type VoiceStreamConfig struct {
	AgentID     string `json:"agent_id"`
	VoiceName   string `json:"voice_name,omitempty"`
	Language    string `json:"language,omitempty"`
	AudioFormat string `json:"audio_format,omitempty"`
}

// VoiceStream is a bidirectional WebSocket connection for real-time voice chat
// via Gemini Live. Send audio with SendAudio, send text with SendText,
// and receive events with Recv.
type VoiceStream struct {
	conn *ws.Conn
}

// VoiceUsage contains token usage statistics from a voice session.
type VoiceUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// VoiceStreamEvent represents a server event from the voice live stream.
type VoiceStreamEvent struct {
	// Type is the event type: "ready", "session_ready", "input_transcript",
	// "output_transcript", "agent_state", "turn_complete", "tool_activity",
	// "side_effects", "usage", "session_ended", "error", or "audio".
	Type string `json:"type"`

	// SessionID is set on "ready" and "session_ready" events.
	SessionID string `json:"sessionId,omitempty"`

	// Text is set on "input_transcript" and "output_transcript" events.
	Text string `json:"text,omitempty"`

	// IsFinal indicates whether the transcript is final or interim.
	IsFinal *bool `json:"isFinal,omitempty"`

	// Speaking is set on "agent_state" events.
	Speaking *bool `json:"speaking,omitempty"`

	// TurnIndex is set on "turn_complete" events.
	TurnIndex *int `json:"turnIndex,omitempty"`

	// Name is set on "tool_activity" events (tool name).
	Name string `json:"name,omitempty"`

	// Status is set on "tool_activity" events ("called" or "resolved").
	Status string `json:"status,omitempty"`

	// Facts is set on "side_effects" events.
	Facts json.RawMessage `json:"facts,omitempty"`

	// Emotions is set on "side_effects" events.
	Emotions json.RawMessage `json:"emotions,omitempty"`

	// RelationshipDelta is set on "side_effects" events.
	RelationshipDelta json.RawMessage `json:"relationshipDelta,omitempty"`

	// PromptTokens is set on "usage" events.
	PromptTokens int `json:"promptTokens,omitempty"`

	// CompletionTokens is set on "usage" events.
	CompletionTokens int `json:"completionTokens,omitempty"`

	// TotalTokens is set on "usage" events.
	TotalTokens int `json:"totalTokens,omitempty"`

	// Reason is set on "session_ended" events ("normal", "error", "timeout").
	Reason string `json:"reason,omitempty"`

	// TotalUsage is set on "session_ended" events.
	TotalUsage *VoiceUsage `json:"totalUsage,omitempty"`

	// TurnCount is set on "session_ended" events.
	TurnCount int `json:"turnCount,omitempty"`

	// VoiceName is set on "session_ready" events.
	VoiceName string `json:"voiceName,omitempty"`

	// Error is set on "error" events.
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"errorCode,omitempty"`

	// Audio is set on "audio" events (raw binary PCM audio data, 24kHz 16-bit mono).
	Audio []byte `json:"-"`
}

// VoiceStreamToken represents the token needed to establish a voice live WebSocket.
type VoiceStreamToken struct {
	WSURL     string `json:"wsUrl"`
	AuthToken string `json:"authToken"`
}

// VoiceTokenOptions configures a voice live WebSocket token request.
type VoiceTokenOptions struct {
	VoiceName            string `json:"voiceName,omitempty"`
	Language             string `json:"language,omitempty"`
	UserID               string `json:"userId,omitempty"`
	CompiledSystemPrompt string `json:"compiledSystemPrompt,omitempty"`
}

// GetToken obtains a short-lived token for voice live WebSocket streaming.
// The token expires in 60 seconds and is single-use.
func (v *VoiceResource) GetToken(ctx context.Context, agentID string, opts VoiceTokenOptions) (*VoiceStreamToken, error) {
	var result VoiceStreamToken
	err := v.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/voice/live-ws-token", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Stream opens a bidirectional WebSocket for real-time voice chat via Gemini Live.
// The token should be obtained from GetToken. After connecting, the session
// is authenticated automatically and a "ready" event is sent.
//
// Usage:
//
//	token, _ := client.Agents.Voice.GetToken(ctx, agentID, opts)
//	stream, _ := client.Agents.Voice.Stream(ctx, token)
//	defer stream.Close()
//
//	// Send PCM audio chunks (16kHz, 16-bit, mono)
//	stream.SendAudio(audioBytes)
//
//	// Or send text input instead of audio
//	stream.SendText("Hello!")
//
//	// Receive events
//	for {
//	    event, err := stream.Recv()
//	    if err == io.EOF { break }
//	    switch event.Type {
//	    case "input_transcript": fmt.Println("User:", event.Text)
//	    case "output_transcript": fmt.Println("Agent:", event.Text)
//	    case "audio": playPCMAudio(event.Audio) // 24kHz PCM
//	    case "turn_complete": // ready for next turn
//	    case "session_ended": // session finished
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

	// Binary frames are PCM audio.
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

// SendText sends a text message to the agent instead of audio.
func (s *VoiceStream) SendText(text string) error {
	data, err := json.Marshal(map[string]string{"type": "text_input", "text": text})
	if err != nil {
		return fmt.Errorf("voice stream: marshal text_input: %w", err)
	}
	return s.conn.WriteText(data)
}

// EndSession gracefully ends the voice session.
func (s *VoiceStream) EndSession() error {
	return s.conn.WriteText([]byte(`{"type":"end_session"}`))
}

// Configure sends a config message to change audio format or sample rate
// mid-session without reconnecting.
func (s *VoiceStream) Configure(audioFormat string, sampleRate int) error {
	msg := map[string]any{"type": "config"}
	if audioFormat != "" {
		msg["audioFormat"] = audioFormat
	}
	if sampleRate > 0 {
		msg["sampleRate"] = sampleRate
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("voice stream: marshal config: %w", err)
	}
	return s.conn.WriteText(data)
}

// Close closes the voice stream.
func (s *VoiceStream) Close() error {
	return s.conn.Close()
}
