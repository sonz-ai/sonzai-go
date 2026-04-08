package sonzai

// LLM provider identifiers and model ID constants.
//
// Use these when specifying a provider or model in chat requests, or when
// building a model picker UI. At runtime you can also call
// [Client.ListModels] to get the live list of providers and models enabled
// on the current deployment.
//
//	result, _ := client.ListModels(ctx)
//	for _, p := range result.Providers {
//	    fmt.Println(p.ProviderName, p.DefaultModel)
//	}
const (
	// ---------------------------------------------------------------------------
	// Provider identifiers
	// ---------------------------------------------------------------------------

	// ProviderGemini is the provider ID for Google Gemini.
	ProviderGemini = "gemini"

	// ProviderZhipu is the provider ID for Zhipu AI (GLM-4 family).
	ProviderZhipu = "zhipu"

	// ProviderVolcEngine is the provider ID for VolcEngine (Doubao family).
	ProviderVolcEngine = "volcengine"

	// ProviderOpenRouter is the provider ID for OpenRouter (multi-model gateway, fallback).
	ProviderOpenRouter = "openrouter"

	// ProviderCustom is the provider ID for a project-configured custom LLM (BYOM).
	ProviderCustom = "custom"

	// ---------------------------------------------------------------------------
	// Model IDs — Google Gemini
	// ---------------------------------------------------------------------------

	// ModelGeminiFlashLite is the fast, cost-efficient Gemini model.
	// This is the platform's recommended default for most use cases.
	ModelGeminiFlashLite = "gemini-3.1-flash-lite-preview"

	// ---------------------------------------------------------------------------
	// Model IDs — Zhipu AI
	// ---------------------------------------------------------------------------

	// ModelZhipuGLM4Flash is the lightweight, zero-cost flash model for
	// high-throughput workloads.
	ModelZhipuGLM4Flash = "glm-4-flash"

	// ModelZhipuGLM4Plus is the highest-capability GLM-4 model.
	ModelZhipuGLM4Plus = "glm-4-plus"

	// ---------------------------------------------------------------------------
	// Model IDs — VolcEngine (Doubao)
	// ---------------------------------------------------------------------------

	// ModelDoubaoCharacter is a long-context character model optimised for
	// roleplay and dialogue.
	ModelDoubaoCharacter = "doubao-1-5-pro-32k-character"

	// ---------------------------------------------------------------------------
	// Model IDs — OpenRouter (fallback)
	// ---------------------------------------------------------------------------

	ModelOpenRouterClaudeHaiku  = "anthropic/claude-3-haiku"
	ModelOpenRouterClaudeSonnet = "anthropic/claude-3.5-sonnet"

	// ---------------------------------------------------------------------------
	// Default
	// ---------------------------------------------------------------------------

	// DefaultModel is the platform's overall default model ID.
	DefaultModel = ModelGeminiFlashLite
)
