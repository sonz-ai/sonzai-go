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
	// GameName is the legacy alias of BusinessName. The API emits both
	// keys with the same value; new code should read BusinessName.
	//
	// Deprecated: use BusinessName.
	GameName     string `json:"game_name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	Environment  string `json:"environment,omitempty"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at,omitempty"`

	// DefaultAgentKBWrite is the project-level default for whether agents
	// in this project may autonomously edit the knowledge base. The
	// platform OR-resolves this with each agent's own knowledgeBaseWrite
	// capability — agent flag true wins immediately; only when off does
	// the project default apply. nil = not configured (no default).
	DefaultAgentKBWrite *bool `json:"default_agent_kb_write,omitempty"`
}

// CreateProjectOptions configures a project creation request.
type CreateProjectOptions struct {
	Name                string `json:"name"`
	Environment         string `json:"environment,omitempty"`
	DefaultAgentKBWrite *bool  `json:"default_agent_kb_write,omitempty"`
}

// UpdateProjectOptions configures a project update request. All fields are
// optional — only non-nil/non-empty values are sent. Pass a *bool pointer
// for DefaultAgentKBWrite to set true/false explicitly.
type UpdateProjectOptions struct {
	Name string `json:"name,omitempty"`
	// GameName is the legacy alias of BusinessName. Either is accepted on
	// the wire; BusinessName wins when both are set.
	//
	// Deprecated: use BusinessName.
	GameName            string `json:"game_name,omitempty"`
	BusinessName        string `json:"business_name,omitempty"`
	Environment         string `json:"environment,omitempty"`
	DefaultAgentKBWrite *bool  `json:"default_agent_kb_write,omitempty"`
}

// Update modifies an existing project's settings. Org-admin only on the
// platform side; non-admin callers get 403.
func (p *ProjectsResource) Update(ctx context.Context, projectID string, opts UpdateProjectOptions) (*Project, error) {
	var result Project
	if err := p.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s", projectID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns all projects for the authenticated tenant.
// If the tenant has no projects, the platform auto-creates a default project and returns it.
func (p *ProjectsResource) List(ctx context.Context) ([]Project, error) {
	var result struct {
		Items     []Project `json:"items"`
		NextCursor string   `json:"next_cursor,omitempty"`
		HasMore    bool     `json:"has_more,omitempty"`
	}
	if err := p.http.Get(ctx, "/api/v1/projects", nil, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
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
