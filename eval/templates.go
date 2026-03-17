package eval

import (
	"context"
	"fmt"
)

// TemplatesResource provides eval template CRUD operations.
type TemplatesResource struct {
	backend Backend
}

// List returns all eval templates, optionally filtered by type.
func (t *TemplatesResource) List(ctx context.Context, templateType string) (*TemplateListResponse, error) {
	params := map[string]string{}
	if templateType != "" {
		params["type"] = templateType
	}

	var result TemplateListResponse
	err := t.backend.Get(ctx, "/api/v1/eval-templates", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a specific eval template.
func (t *TemplatesResource) Get(ctx context.Context, templateID string) (*Template, error) {
	var result Template
	err := t.backend.Get(ctx, fmt.Sprintf("/api/v1/eval-templates/%s", templateID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new eval template.
func (t *TemplatesResource) Create(ctx context.Context, opts TemplateCreateOptions) (*Template, error) {
	var result Template
	err := t.backend.Post(ctx, "/api/v1/eval-templates", opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an eval template.
func (t *TemplatesResource) Update(ctx context.Context, templateID string, opts TemplateUpdateOptions) (*Template, error) {
	var result Template
	err := t.backend.Put(ctx, fmt.Sprintf("/api/v1/eval-templates/%s", templateID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an eval template.
func (t *TemplatesResource) Delete(ctx context.Context, templateID string) error {
	return t.backend.Delete(ctx, fmt.Sprintf("/api/v1/eval-templates/%s", templateID), nil)
}
