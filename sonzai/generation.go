package sonzai

import (
	"context"
	"fmt"
)

// GenerationResource provides content generation operations for an agent.
type GenerationResource struct {
	http *httpClient
}

// GenerateBio generates an AI biography for an agent.
func (g *GenerationResource) GenerateBio(ctx context.Context, agentID string, params GenerateBioParams) (*GenerateBioResult, error) {
	var result GenerateBioResult
	err := g.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/generate-bio", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateImage generates an image from a text prompt.
func (g *GenerationResource) GenerateImage(ctx context.Context, agentID string, params GenerateImageParams) (*GenerateImageResult, error) {
	var result GenerateImageResult
	err := g.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/image/generate", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateCharacter generates a full character profile from a name, gender, and description.
func (g *GenerationResource) GenerateCharacter(ctx context.Context, agentID string, params GenerateCharacterParams) (*GenerateCharacterResult, error) {
	var result GenerateCharacterResult
	err := g.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/generate-character", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateSeedMemories generates seed memories via LLM for an agent.
func (g *GenerationResource) GenerateSeedMemories(ctx context.Context, agentID string, params GenerateSeedMemoriesParams) (*GenerateSeedMemoriesResult, error) {
	var result GenerateSeedMemoriesResult
	err := g.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/generate-seed-memories", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
