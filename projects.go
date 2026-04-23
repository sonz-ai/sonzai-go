package sonzai

import (
	"context"
	"fmt"
)

// ProjectsResource provides project management operations.
type ProjectsResource struct {
	http *httpClient
}

// Project represents a Sonzai project.
type Project struct {
	ProjectID   string `json:"project_id"`
	TenantID    string `json:"tenant_id,omitempty"`
	Name        string `json:"name"`
	Environment string `json:"environment,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// CreateProjectOptions configures a project creation request.
type CreateProjectOptions struct {
	Name        string `json:"name"`
	Environment string `json:"environment,omitempty"`
}

// List returns all projects for the authenticated tenant.
// If the tenant has no projects, the platform auto-creates a default project and returns it.
func (p *ProjectsResource) List(ctx context.Context) ([]Project, error) {
	var result []Project
	if err := p.http.Get(ctx, "/api/v1/projects", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get returns a single project by ID.
func (p *ProjectsResource) Get(ctx context.Context, projectID string) (*Project, error) {
	var result Project
	if err := p.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new project for the authenticated tenant.
func (p *ProjectsResource) Create(ctx context.Context, opts CreateProjectOptions) (*Project, error) {
	var result Project
	if err := p.http.Post(ctx, "/api/v1/projects", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EnsureProject lists projects and returns the first one whose name matches the given name.
// If none is found, it creates a new project with that name.
// This is the idempotent way to ensure a named project exists.
func (p *ProjectsResource) EnsureProject(ctx context.Context, name string) (*Project, error) {
	projects, err := p.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	for i := range projects {
		if projects[i].Name == name {
			return &projects[i], nil
		}
	}
	// Not found — create it.
	created, err := p.Create(ctx, CreateProjectOptions{Name: name})
	if err != nil {
		return nil, fmt.Errorf("create project %q: %w", name, err)
	}
	return created, nil
}
