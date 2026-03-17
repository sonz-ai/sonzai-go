// Package sonzai provides a Go SDK for the Sonzai Character Engine API.
//
// Usage:
//
//	client := sonzai.NewClient("your-api-key")
//
//	// Chat with an agent
//	resp, err := client.Agents.Chat(ctx, "agent-id", sonzai.ChatOptions{
//	    Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
//	})
//
//	// Stream chat
//	err := client.Agents.ChatStream(ctx, "agent-id", opts, func(event sonzai.ChatStreamEvent) error {
//	    fmt.Print(event.Content())
//	    return nil
//	})
package sonzai

import (
	"os"
	"time"
)

const (
	defaultBaseURL = "https://api.sonz.ai"
	defaultTimeout = 30 * time.Second
)

// ClientOption configures the Sonzai client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL string
	timeout time.Duration
}

// WithBaseURL sets the API base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) { c.baseURL = url }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) { c.timeout = d }
}

// Client is the Sonzai Character Engine API client.
type Client struct {
	Agents        *AgentsResource
	Voice         *VoiceResource
	EvalTemplates *EvalTemplatesResource
	EvalRuns      *EvalRunsResource
}

// NewClient creates a new Sonzai client with the given API key.
// If apiKey is empty, it falls back to the SONZAI_API_KEY environment variable.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("SONZAI_API_KEY")
	}
	if apiKey == "" {
		panic("sonzai: apiKey must be provided or set via SONZAI_API_KEY environment variable")
	}

	cfg := &clientConfig{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}

	if envURL := os.Getenv("SONZAI_BASE_URL"); envURL != "" {
		cfg.baseURL = envURL
	}

	for _, opt := range opts {
		opt(cfg)
	}

	http := newHTTPClient(cfg.baseURL, apiKey, cfg.timeout)

	return &Client{
		Agents:        newAgentsResource(http),
		Voice:         &VoiceResource{http: http},
		EvalTemplates: newEvalTemplatesResource(http),
		EvalRuns:      newEvalRunsResource(http),
	}
}
