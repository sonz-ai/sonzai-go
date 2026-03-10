package sonzai

import (
	"context"
	"fmt"
)

// EvalTemplatesResource provides eval template operations.
type EvalTemplatesResource struct {
	http *httpClient
}

func newEvalTemplatesResource(http *httpClient) *EvalTemplatesResource {
	return &EvalTemplatesResource{http: http}
}

// EvalTemplateCreateOptions configures a template creation request.
type EvalTemplateCreateOptions struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	TemplateType  string                 `json:"template_type,omitempty"`
	JudgeModel    string                 `json:"judge_model,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	ScoringRubric string                 `json:"scoring_rubric,omitempty"`
	Categories    []EvalTemplateCategory `json:"categories,omitempty"`
}

// EvalTemplateUpdateOptions configures a template update request.
type EvalTemplateUpdateOptions struct {
	Name          *string                `json:"name,omitempty"`
	Description   *string                `json:"description,omitempty"`
	TemplateType  *string                `json:"template_type,omitempty"`
	JudgeModel    *string                `json:"judge_model,omitempty"`
	Temperature   *float64               `json:"temperature,omitempty"`
	MaxTokens     *int                   `json:"max_tokens,omitempty"`
	ScoringRubric *string                `json:"scoring_rubric,omitempty"`
	Categories    []EvalTemplateCategory `json:"categories,omitempty"`
}

// List returns all eval templates.
func (e *EvalTemplatesResource) List(ctx context.Context, templateType string) (*EvalTemplateListResponse, error) {
	params := map[string]string{}
	if templateType != "" {
		params["type"] = templateType
	}

	var result EvalTemplateListResponse
	err := e.http.get(ctx, "/api/v1/eval-templates", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a specific eval template.
func (e *EvalTemplatesResource) Get(ctx context.Context, templateID string) (*EvalTemplate, error) {
	var result EvalTemplate
	err := e.http.get(ctx, fmt.Sprintf("/api/v1/eval-templates/%s", templateID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new eval template.
func (e *EvalTemplatesResource) Create(ctx context.Context, opts EvalTemplateCreateOptions) (*EvalTemplate, error) {
	var result EvalTemplate
	err := e.http.post(ctx, "/api/v1/eval-templates", opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an eval template.
func (e *EvalTemplatesResource) Update(ctx context.Context, templateID string, opts EvalTemplateUpdateOptions) (*EvalTemplate, error) {
	var result EvalTemplate
	err := e.http.put(ctx, fmt.Sprintf("/api/v1/eval-templates/%s", templateID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an eval template.
func (e *EvalTemplatesResource) Delete(ctx context.Context, templateID string) error {
	return e.http.del(ctx, fmt.Sprintf("/api/v1/eval-templates/%s", templateID), nil)
}
