package sonzai

import (
	"context"
)

// OrgMembership represents a user's membership in an organization.
type OrgMembership struct {
	OrgID string `json:"org_id"`
	Role  string `json:"role"`
	Name  string `json:"name,omitempty"`
}

// MeResponse represents the response from the /api/v1/me endpoint.
type MeResponse struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Orgs   []OrgMembership `json:"orgs"`
}

// Me returns the authenticated user's profile information including user ID,
// email, and organization memberships.
//
//	user, err := client.Me(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(user.UserID, user.Email)
func (c *Client) Me(ctx context.Context) (*MeResponse, error) {
	var result MeResponse
	if err := c.http.Get(ctx, "/api/v1/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
