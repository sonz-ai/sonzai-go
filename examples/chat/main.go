// Example: streaming chat with a Sonzai agent.
//
// Usage:
//
//	export SONZAI_API_KEY=sk-...
//	go run ./examples/chat -agent <agent-id>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
	agentID := flag.String("agent", "", "Agent ID to chat with")
	userID := flag.String("user", "example-user", "User ID for the session")
	message := flag.String("message", "Hello! Tell me about yourself.", "Message to send")
	flag.Parse()

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/chat -agent <agent-id>")
		os.Exit(1)
	}

	client := sonzai.NewClient("")

	fmt.Printf("You: %s\n\nAgent: ", *message)

	err := client.Agents.ChatStream(
		context.Background(),
		sonzai.AgentChatParams{
			AgentID: *agentID,
			ChatOptions: sonzai.ChatOptions{
				Messages: []sonzai.ChatMessage{{Role: "user", Content: *message}},
				UserID:   *userID,
			},
		},
		func(event sonzai.ChatStreamEvent) error {
			fmt.Print(event.Content())
			return nil
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
}
