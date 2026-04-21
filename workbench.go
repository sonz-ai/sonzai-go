package sonzai

import "context"

// WorkbenchResource provides internal simulation and debugging operations.
// These endpoints are intended for development and testing workflows.
type WorkbenchResource struct {
	http *httpClient
}

// Chat runs a workbench chat request.
func (w *WorkbenchResource) Chat(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/chat", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Prepare prepares a workbench session.
func (w *WorkbenchResource) Prepare(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/prepare", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetState returns the current workbench state.
func (w *WorkbenchResource) GetState(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/state", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// AdvanceTime advances the workbench simulation clock.
func (w *WorkbenchResource) AdvanceTime(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/advance-time", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResetAgent resets an agent within a workbench session.
func (w *WorkbenchResource) ResetAgent(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/reset-agent", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SessionEnd ends the workbench session.
func (w *WorkbenchResource) SessionEnd(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/session-end", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SimulateUser sends a simulated user message in a workbench session.
func (w *WorkbenchResource) SimulateUser(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/simulate-user", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateBio runs bio generation in the workbench.
func (w *WorkbenchResource) GenerateBio(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/generate-bio", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateCharacter runs character generation in the workbench.
func (w *WorkbenchResource) GenerateCharacter(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/generate-character", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateSeedMemories runs seed memory generation in the workbench.
func (w *WorkbenchResource) GenerateSeedMemories(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/generate-seed-memories", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
