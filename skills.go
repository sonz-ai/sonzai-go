package sonzai

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// SkillsResource provides project-scoped skill library + per-agent
// enablement operations. A skill is a markdown playbook the agent loads
// on demand via the sonzai_load_skill tool. Developers manage the
// library at the project level; each agent opts into individual skills
// via the enablement surface.
type SkillsResource struct {
	http *httpClient
}

// Skill is a project-scoped playbook entry.
type Skill struct {
	TenantID      string    `json:"tenant_id"`
	ProjectID     string    `json:"project_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	WhenToUse     string    `json:"when_to_use,omitempty"`
	Content       string    `json:"content"`
	Author        string    `json:"author"` // "developer" | "agent"
	AuthorUserID  string    `json:"author_user_id,omitempty"`
	AuthorAgentID string    `json:"author_agent_id,omitempty"`
	Version       int64     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProjectSkillBody is the create-skill request payload.
type ProjectSkillBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WhenToUse   string `json:"when_to_use,omitempty"`
	Content     string `json:"content"`
}

// UpdateProjectSkillInputBody is the patch payload. Fields left empty are
// preserved from the existing row; Version and CreatedAt are server-managed.
type UpdateProjectSkillInputBody struct {
	Description string `json:"description,omitempty"`
	WhenToUse   string `json:"when_to_use,omitempty"`
	Content     string `json:"content,omitempty"`
}

// ListProjectSkillsResponse is returned by ListProjectSkills.
type ListProjectSkillsResponse struct {
	Skills []Skill `json:"skills"`
}

// ListEnabledSkillsResponse lists the skill names currently enabled for an agent.
type ListEnabledSkillsResponse struct {
	Skills []string `json:"skills"`
}

// ToggleEnabledSkillInput is the enable/disable request.
type ToggleEnabledSkillInput struct {
	SkillName string `json:"skill_name"`
	Enabled   bool   `json:"enabled"`
}

// ToggleEnabledSkillResponse echoes the applied state.
type ToggleEnabledSkillResponse struct {
	SkillName string `json:"skill_name"`
	Enabled   bool   `json:"enabled"`
}

// GetSkillLoadCountResponse is the per-agent load-count read.
type GetSkillLoadCountResponse struct {
	SkillName string `json:"skill_name"`
	Count     int64  `json:"count"`
}

// ListProjectSkills returns every skill in the project library.
func (r *SkillsResource) ListProjectSkills(ctx context.Context, projectID string) (*ListProjectSkillsResponse, error) {
	var out ListProjectSkillsResponse
	path := fmt.Sprintf("/api/v1/projects/%s/skills", url.PathEscape(projectID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProjectSkill adds a new skill to the project library.
func (r *SkillsResource) CreateProjectSkill(ctx context.Context, projectID string, body ProjectSkillBody) (*Skill, error) {
	var out Skill
	path := fmt.Sprintf("/api/v1/projects/%s/skills", url.PathEscape(projectID))
	if err := r.http.Post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetProjectSkill returns a single skill by name.
func (r *SkillsResource) GetProjectSkill(ctx context.Context, projectID, name string) (*Skill, error) {
	var out Skill
	path := fmt.Sprintf("/api/v1/projects/%s/skills/%s", url.PathEscape(projectID), url.PathEscape(name))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProjectSkill patches a skill. Only non-empty fields overwrite;
// absent fields are preserved from the existing row on the server side.
func (r *SkillsResource) UpdateProjectSkill(ctx context.Context, projectID, name string, body UpdateProjectSkillInputBody) (*Skill, error) {
	var out Skill
	path := fmt.Sprintf("/api/v1/projects/%s/skills/%s", url.PathEscape(projectID), url.PathEscape(name))
	if err := r.http.Put(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProjectSkill removes a skill from the library.
func (r *SkillsResource) DeleteProjectSkill(ctx context.Context, projectID, name string) error {
	path := fmt.Sprintf("/api/v1/projects/%s/skills/%s", url.PathEscape(projectID), url.PathEscape(name))
	return r.http.Delete(ctx, path, nil)
}

// ListEnabledSkills returns the skills currently enabled for an agent.
// The result is a list of skill names — fetch full skills via
// ListProjectSkills or GetProjectSkill when metadata is needed.
func (r *SkillsResource) ListEnabledSkills(ctx context.Context, agentID string) (*ListEnabledSkillsResponse, error) {
	var out ListEnabledSkillsResponse
	path := fmt.Sprintf("/api/v1/agents/%s/skills/enabled", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ToggleSkillEnabled flips the enablement flag for (agentID, skillName).
func (r *SkillsResource) ToggleSkillEnabled(ctx context.Context, agentID string, input ToggleEnabledSkillInput) (*ToggleEnabledSkillResponse, error) {
	var out ToggleEnabledSkillResponse
	path := fmt.Sprintf("/api/v1/agents/%s/skills/enabled", url.PathEscape(agentID))
	if err := r.http.Post(ctx, path, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSkillLoadCount reads the per-agent counter that bumps on every
// sonzai_load_skill call with the matching name.
func (r *SkillsResource) GetSkillLoadCount(ctx context.Context, agentID, skillName string) (*GetSkillLoadCountResponse, error) {
	var out GetSkillLoadCountResponse
	path := fmt.Sprintf("/api/v1/agents/%s/skills/enabled/%s/load-count", url.PathEscape(agentID), url.PathEscape(skillName))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
