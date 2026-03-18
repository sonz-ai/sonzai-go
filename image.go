package sonzai

import (
	"context"
	"fmt"
)

// ImageResource provides image generation operations for an agent.
type ImageResource struct {
	http *httpClient
}

// ImageGenerateOptions configures an image generation request.
type ImageGenerateOptions struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Model          string `json:"model,omitempty"`
	Provider       string `json:"provider,omitempty"`
}

// ImageGenerateResponse is the response from image generation.
type ImageGenerateResponse struct {
	ImageID          string `json:"image_id"`
	URL              string `json:"public_url"`
	MimeType         string `json:"mime_type"`
	GenerationTimeMs int    `json:"generation_time_ms"`
}

// Generate creates an image using the agent's context.
func (i *ImageResource) Generate(ctx context.Context, agentID string, opts ImageGenerateOptions) (*ImageGenerateResponse, error) {
	var result ImageGenerateResponse
	err := i.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/image/generate", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
