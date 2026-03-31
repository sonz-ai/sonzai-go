// Example: connect to a voice live session with an agent.
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
	"io"
	"os"

	sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
	agentID := flag.String("agent", "", "Agent ID")
	voice := flag.String("voice", "Kore", "Voice name (e.g., Kore, Puck, Aoede)")
	flag.Parse()

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/voice -agent <agent-id>")
		os.Exit(1)
	}

	client := sonzai.NewClient("")
	ctx := context.Background()

	// Get a voice live WebSocket token
	token, err := client.Agents.Voice.GetToken(ctx, *agentID, sonzai.VoiceTokenOptions{
		VoiceName: *voice,
		Language:  "en-US",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "token error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got voice live token, connecting to %s\n", token.WSURL)

	// Open the WebSocket stream
	stream, err := client.Agents.Voice.Stream(ctx, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stream error: %v\n", err)
		os.Exit(1)
	}
	defer stream.Close()

	// Send a text message instead of audio
	if err := stream.SendText("Hello! How are you today?"); err != nil {
		fmt.Fprintf(os.Stderr, "send error: %v\n", err)
		os.Exit(1)
	}

	// Receive events
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "recv error: %v\n", err)
			break
		}

		switch event.Type {
		case "ready":
			fmt.Println("Connected to proxy")
		case "session_ready":
			fmt.Printf("Gemini Live session ready (voice: %s)\n", event.VoiceName)
		case "input_transcript":
			fmt.Printf("User: %s\n", event.Text)
		case "output_transcript":
			fmt.Printf("Agent: %s\n", event.Text)
		case "audio":
			fmt.Printf("Received %d bytes of PCM audio\n", len(event.Audio))
		case "turn_complete":
			fmt.Println("Turn complete")
		case "session_ended":
			fmt.Printf("Session ended: %s\n", event.Reason)
			return
		case "error":
			fmt.Printf("Error: %s (%s)\n", event.Error, event.ErrorCode)
			return
		}
	}
}
