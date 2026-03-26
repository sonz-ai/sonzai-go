// Example: search and browse an agent's memory tree.
//
// Usage:
//
//	export SONZAI_API_KEY=sk-...
//	go run ./examples/memory -agent <agent-id> -query "favorite food"
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
	query := flag.String("query", "", "Search query (omit to list full tree)")
	userID := flag.String("user", "", "Filter by user ID")
	flag.Parse()

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/memory -agent <agent-id> [-query <search>]")
		os.Exit(1)
	}

	client := sonzai.MustNewClient("")
	ctx := context.Background()

	if *query != "" {
		// Semantic search
		results, err := client.Agents.Memory.Search(ctx, *agentID, sonzai.MemorySearchOptions{
			Query: *query,
			Limit: 10,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Search results for %q:\n\n", *query)
		for _, r := range results.Results {
			fmt.Printf("  [%.2f] %s\n", r.Score, r.Content)
		}
		return
	}

	// Full memory tree
	opts := &sonzai.MemoryListOptions{IncludeContents: true}
	if *userID != "" {
		opts.UserID = *userID
	}

	tree, err := client.Agents.Memory.List(ctx, *agentID, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Memory tree (%d nodes):\n\n", len(tree.Nodes))
	for _, node := range tree.Nodes {
		fmt.Printf("  %s (importance: %.2f)\n", node.Title, node.Importance)
	}
}
