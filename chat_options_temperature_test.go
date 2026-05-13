package sonzai

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestChatOptions_Temperature_OmittedWhenNil verifies that the JSON encoder
// drops the temperature field entirely when ChatOptions.Temperature is nil.
// This is the contract that lets the AI service apply its per-model default
// and lets the Platform adapt the value for providers that need it.
func TestChatOptions_Temperature_OmittedWhenNil(t *testing.T) {
	opts := ChatOptions{
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	}
	b, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "temperature") {
		t.Fatalf("expected temperature key to be omitted when nil, got %s", string(b))
	}
}

// TestChatOptions_Temperature_PresentWhenSet verifies that an explicit
// temperature value is serialised through to the wire.
func TestChatOptions_Temperature_PresentWhenSet(t *testing.T) {
	v := 0.7
	opts := ChatOptions{
		Messages:    []ChatMessage{{Role: "user", Content: "hello"}},
		Temperature: &v,
	}
	b, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Round-trip into a generic map to assert the exact JSON shape.
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	raw, ok := decoded["temperature"]
	if !ok {
		t.Fatalf("expected temperature key in JSON, got %s", string(b))
	}
	if got := strings.TrimSpace(string(raw)); got != "0.7" {
		t.Fatalf("expected temperature=0.7, got %s", got)
	}
}

// TestChatOptions_Temperature_ZeroValuePreserved verifies that a temperature
// of 0 is still emitted (i.e. the pointer-vs-value distinction works — a
// caller asking for deterministic output is not silently dropped).
func TestChatOptions_Temperature_ZeroValuePreserved(t *testing.T) {
	v := 0.0
	opts := ChatOptions{
		Messages:    []ChatMessage{{Role: "user", Content: "hello"}},
		Temperature: &v,
	}
	b, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"temperature":0`) {
		t.Fatalf("expected temperature=0 to be emitted, got %s", string(b))
	}
}
