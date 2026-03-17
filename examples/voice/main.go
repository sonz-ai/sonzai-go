// Example: match a voice to an agent and generate speech.
//
// Usage:
//
//	export SONZAI_API_KEY=sk-...
//	go run ./examples/voice -agent <agent-id>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
	agentID := flag.String("agent", "", "Agent ID")
	text := flag.String("text", "Hello! It's great to meet you.", "Text to speak")
	flag.Parse()

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/voice -agent <agent-id>")
		os.Exit(1)
	}

	client := sonzai.NewClient("")
	ctx := context.Background()

	// Find the best voice match
	match, err := client.Agents.Voice.Match(ctx, *agentID, sonzai.VoiceMatchOptions{
		PersonalityTraits: []string{"warm", "calm"},
		VoiceTier:         2,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "match error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Matched voice: %s (confidence: %.2f)\n", match.VoiceName, match.Confidence)

	// Generate speech
	tts, err := client.Agents.Voice.TTS(ctx, *agentID, sonzai.TTSOptions{
		Text:    *text,
		VoiceID: match.VoiceID,
		Speed:   1.0,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "tts error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Audio: %s (%dms duration)\n", tts.AudioURL, tts.DurationMs)
}
