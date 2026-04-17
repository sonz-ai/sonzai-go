package sonzai

import (
	"encoding/json"
	"testing"
)

func TestPostProcessingModelMap_RoundTripJSON(t *testing.T) {
	want := PostProcessingModelMap{
		"gemini-3.1-pro-preview":  {Provider: "gemini", Model: "gemini-3.1-flash-lite-preview"},
		"claude-opus-4.6":         {Provider: "openrouter", Model: "anthropic/claude-haiku-4.5"},
		PostProcessingWildcardKey: {Provider: "gemini", Model: "gemini-3.1-flash-lite-preview"},
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Decode via the helper the SDK uses on the GET path so a round-trip
	// through an untyped interface{} (how the server returns config values)
	// yields the same typed map.
	var untyped interface{}
	if err := json.Unmarshal(data, &untyped); err != nil {
		t.Fatalf("unmarshal untyped: %v", err)
	}
	got, err := unmarshalPostProcessingMap(untyped)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d, want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("entry %q: got %+v, want %+v", k, got[k], v)
		}
	}
}

func TestPostProcessingModelMap_NilValue(t *testing.T) {
	got, err := unmarshalPostProcessingMap(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("nil input should decode to nil map, got %+v", got)
	}
}

func TestPostProcessingModelMap_Wildcard(t *testing.T) {
	// A map with only a wildcard entry must be valid — that's how operators
	// express "use one model for every chat model" in a single line.
	m := PostProcessingModelMap{
		PostProcessingWildcardKey: {Provider: "gemini", Model: "gemini-3.1-flash-lite-preview"},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var untyped interface{}
	_ = json.Unmarshal(data, &untyped)
	got, err := unmarshalPostProcessingMap(untyped)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got[PostProcessingWildcardKey].Model != "gemini-3.1-flash-lite-preview" {
		t.Fatalf("wildcard entry lost in round-trip: %+v", got)
	}
}
