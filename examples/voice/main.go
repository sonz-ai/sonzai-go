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
		Big5:            &sonzai.Big5Scores{Openness: 80, Agreeableness: 70, Extraversion: 60},
		PreferredGender: "female",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "match error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Matched voice: %s (score: %.2f)\n", match.VoiceName, match.MatchScore)

	// Generate speech
	tts, err := client.Agents.Voice.TTS(ctx, *agentID, sonzai.TTSOptions{
		Text:      *text,
		VoiceName: match.VoiceName,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "tts error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Audio: %s (%dms duration)\n", tts.ContentType, tts.DurationMs)
}
