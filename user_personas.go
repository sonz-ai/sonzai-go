package sonzai

import (
	"context"
	"fmt"
)

// UserPersonasResource provides user persona CRUD operations.
type UserPersonasResource struct {
	http *httpClient
}

// UserPersona represents a user persona template.
type UserPersona struct {
	PersonaID   string `json:"persona_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Style       string `json:"style,omitempty"`
	IsDefault   bool   `json:"is_default"`
	TenantID    string `json:"tenant_id,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// UserPersonaListResponse is the response from listing user personas.
type UserPersonaListResponse struct {
	Personas []UserPersona `json:"personas"`
}

// CreateUserPersonaOptions configures a user persona creation request.
type CreateUserPersonaOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Style       string `json:"style,omitempty"`
}

// UpdateUserPersonaOptions configures a user persona update request.
type UpdateUserPersonaOptions struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Style       string `json:"style,omitempty"`
}

// List returns all user personas for the tenant.
func (u *UserPersonasResource) List(ctx context.Context) (*UserPersonaListResponse, error) {
	var result UserPersonaListResponse
	if err := u.http.Get(ctx, "/api/v1/user-personas", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single user persona by ID.
func (u *UserPersonasResource) Get(ctx context.Context, personaID string) (*UserPersona, error) {
	var result UserPersona
	if err := u.http.Get(ctx, fmt.Sprintf("/api/v1/user-personas/%s", personaID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new user persona.
func (u *UserPersonasResource) Create(ctx context.Context, opts CreateUserPersonaOptions) (*UserPersona, error) {
	var result UserPersona
	if err := u.http.Post(ctx, "/api/v1/user-personas", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a user persona's fields.
func (u *UserPersonasResource) Update(ctx context.Context, personaID string, opts UpdateUserPersonaOptions) (*UserPersona, error) {
	var result UserPersona
	if err := u.http.Put(ctx, fmt.Sprintf("/api/v1/user-personas/%s", personaID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete permanently deletes a user persona.
func (u *UserPersonasResource) Delete(ctx context.Context, personaID string) error {
	return u.http.Delete(ctx, fmt.Sprintf("/api/v1/user-personas/%s", personaID), nil)
}
