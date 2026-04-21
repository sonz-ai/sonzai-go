package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
)

// PostProcessingModelMapKey is the config key under which the
// chat-model → post-processing-model map is stored at both the project and
// account scope. Every layer of the server's resolver cascade reads the
// same key — see
// `sonzai-ai-monolith-ts/docs/post-processing-model-mapping.md`.
const PostProcessingModelMapKey = "post_processing_model_map"

// PostProcessingWildcardKey is the per-layer wildcard used inside the map.
// Entries keyed on "*" apply to any chat model that has no explicit entry
// at the same layer.
const PostProcessingWildcardKey = "*"

// PostProcessingModelEntry is one map value — the cheaper model post-
// processing work routes to when the agent's chat turn uses a particular
// model. Sampling (temperature, maxTokens) is intentionally omitted; the
// server inherits it from the chat ModelConfig so operators only have to
// override what actually differs.
type PostProcessingModelEntry struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// PostProcessingModelMap is the full chat-model → post-processing-entry
// mapping stored under PostProcessingModelMapKey.
//
// Example:
//
//	sonzai.PostProcessingModelMap{
//	    "gemini-3.1-pro-preview":  {Provider: "gemini",     Model: "gemini-3.1-flash-lite-preview"},
//	    "claude-opus-4.6":         {Provider: "openrouter", Model: "anthropic/claude-haiku-4.5"},
//	    sonzai.PostProcessingWildcardKey: {Provider: "gemini", Model: "gemini-3.1-flash-lite-preview"},
//	}
type PostProcessingModelMap map[string]PostProcessingModelEntry

// unmarshalPostProcessingMap decodes the server's generic config value shape
// ({"key": ..., "value": {...}}) into the typed map. When the server returns
// the raw JSONB directly the "value" path still matches the untyped
// interface{} field on *ConfigEntry since Set writes the value as-is.
func unmarshalPostProcessingMap(value interface{}) (PostProcessingModelMap, error) {
	if value == nil {
		return nil, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("re-encode config value: %w", err)
	}
	var out PostProcessingModelMap
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode post-processing map: %w", err)
	}
	return out, nil
}

// GetPostProcessingModelMap reads the project-level post-processing map.
// Returns (nil, nil) when no map is configured for the project — callers
// can then rely on the account or system-default layer.
func (c *ProjectConfigResource) GetPostProcessingModelMap(ctx context.Context, projectID string) (PostProcessingModelMap, error) {
	entry, err := c.Get(ctx, projectID, PostProcessingModelMapKey)
	if err != nil {
		return nil, err
	}
	return unmarshalPostProcessingMap(entry.Value)
}

// SetPostProcessingModelMap writes the project-level post-processing map,
// replacing whatever was stored before.
func (c *ProjectConfigResource) SetPostProcessingModelMap(ctx context.Context, projectID string, m PostProcessingModelMap) error {
	return c.Set(ctx, projectID, PostProcessingModelMapKey, m)
}

// DeletePostProcessingModelMap removes the project-level map so the
// resolver cascade falls through to the account/system layers.
func (c *ProjectConfigResource) DeletePostProcessingModelMap(ctx context.Context, projectID string) error {
	return c.Delete(ctx, projectID, PostProcessingModelMapKey)
}

// GetPostProcessingModelMap reads the tenant/account-level post-processing
// map. Returns (nil, nil) when no map is configured for the tenant.
func (c *AccountConfigResource) GetPostProcessingModelMap(ctx context.Context) (PostProcessingModelMap, error) {
	entry, err := c.Get(ctx, PostProcessingModelMapKey)
	if err != nil {
		return nil, err
	}
	return unmarshalPostProcessingMap(entry.Value)
}

// SetPostProcessingModelMap writes the tenant/account-level post-processing
// map, replacing whatever was stored before.
func (c *AccountConfigResource) SetPostProcessingModelMap(ctx context.Context, m PostProcessingModelMap) error {
	return c.Set(ctx, PostProcessingModelMapKey, m)
}

// DeletePostProcessingModelMap removes the account-level map so the
// resolver cascade falls through to the system defaults.
func (c *AccountConfigResource) DeletePostProcessingModelMap(ctx context.Context) error {
	return c.Delete(ctx, PostProcessingModelMapKey)
}
