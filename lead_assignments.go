package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// LeadAssignmentsResource provides operations for the Lead Assignment
// service: the tenant-generic work-distribution primitive (any vertical —
// leads, tickets, shifts) that offers a unit of work to one rep from a
// candidate roster, structurally dedups (at most one active assignment per
// lead_ref), and re-offers on SLA expiry.
type LeadAssignmentsResource struct {
	http *httpClient
}

// LeadAssignment is one row of the assignment ledger.
type LeadAssignment struct {
	AssignmentID string `json:"assignment_id"`
	// LeadRef is the caller-owned external key for the unit of work (a CRM
	// lead id, a ticket id, a shift id, ...).
	LeadRef   string `json:"lead_ref"`
	RepUserID string `json:"rep_user_id"`
	// State is one of: offered, claimed, expired, reassigned, released, completed.
	State string `json:"state"`
	// Policy is the distribution policy that chose RepUserID (e.g.
	// round_robin, load_balanced).
	Policy       string          `json:"policy"`
	Propensity   *float64        `json:"propensity,omitempty"`
	Features     json.RawMessage `json:"features,omitempty"`
	OfferedAt    string          `json:"offered_at"`
	SLAExpiresAt string          `json:"sla_expires_at"`
	ClaimedAt    *string         `json:"claimed_at,omitempty"`
	CompletedAt  *string         `json:"completed_at,omitempty"`
	// PriorAssignmentID links a re-offered assignment back to the expired one
	// it replaced.
	PriorAssignmentID *string `json:"prior_assignment_id,omitempty"`
}

// OfferLeadAssignmentOptions configures a POST /lead-assignments/offer call.
type OfferLeadAssignmentOptions struct {
	// LeadRef is the caller-owned external key for the unit of work.
	LeadRef string `json:"lead_ref"`
	// Candidates lists the eligible rep user ids to distribute among.
	// Required — at least one candidate.
	Candidates []string `json:"candidates"`
	// Policy is the distribution policy: round_robin (default) or
	// load_balanced.
	Policy string `json:"policy,omitempty"`
	// Features carries optional context/ML signals captured at offer time.
	Features map[string]interface{} `json:"features,omitempty"`
	// SLASeconds is the offer window in seconds before re-offer to the next
	// candidate (platform default 900 when omitted).
	SLASeconds int `json:"sla_seconds,omitempty"`
}

// OfferLeadAssignmentResult is the outcome of an offer call.
type OfferLeadAssignmentResult struct {
	Assignment LeadAssignment `json:"assignment"`
	// Deduplicated is true when the lead already had an active assignment;
	// Assignment is then the pre-existing one, not a new one.
	Deduplicated bool `json:"deduplicated"`
}

// ListLeadAssignmentsOptions filters a GET /lead-assignments call.
type ListLeadAssignmentsOptions struct {
	// RepUserID, when set, restricts the list to assignments offered to this rep.
	RepUserID string
	// State, when set, restricts the list to this state (offered, claimed,
	// expired, reassigned, released, completed).
	State string
	// Limit caps the number of rows returned (platform default 50, max 200).
	Limit int
}

// ListLeadAssignmentsResult is the response of a list call.
type ListLeadAssignmentsResult struct {
	Assignments []LeadAssignment `json:"assignments"`
}

// Offer distributes a unit of work (lead_ref) to one rep from the candidate
// roster, chosen by the named policy. At most one active assignment can
// exist per lead: offering a lead that already has an active
// (offered/claimed) assignment returns the existing assignment with
// Deduplicated=true instead of creating a second one.
func (l *LeadAssignmentsResource) Offer(ctx context.Context, opts OfferLeadAssignmentOptions) (*OfferLeadAssignmentResult, error) {
	var result OfferLeadAssignmentResult
	if err := l.http.Post(ctx, "/api/v1/lead-assignments/offer", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns the project's assignment ledger rows, newest offer first,
// optionally filtered by rep and/or state.
func (l *LeadAssignmentsResource) List(ctx context.Context, opts *ListLeadAssignmentsOptions) (*ListLeadAssignmentsResult, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.RepUserID != "" {
			params["rep_user_id"] = opts.RepUserID
		}
		if opts.State != "" {
			params["state"] = opts.State
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}
	var result ListLeadAssignmentsResult
	if err := l.http.Get(ctx, "/api/v1/lead-assignments", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get reads one lead assignment by id.
func (l *LeadAssignmentsResource) Get(ctx context.Context, assignmentID string) (*LeadAssignment, error) {
	var result LeadAssignment
	path := fmt.Sprintf("/api/v1/lead-assignments/%s", url.PathEscape(assignmentID))
	if err := l.http.Get(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Claim transitions an offered assignment to claimed — the rep accepted the
// work before the SLA lapsed. Returns an error (409) when the assignment is
// not in the offered state.
func (l *LeadAssignmentsResource) Claim(ctx context.Context, assignmentID string) (*LeadAssignment, error) {
	var result LeadAssignment
	path := fmt.Sprintf("/api/v1/lead-assignments/%s/claim", url.PathEscape(assignmentID))
	if err := l.http.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Release transitions an offered or claimed assignment to released — the
// rep (or an operator) hands the work back, freeing it to be offered again.
// Returns an error (409) when the assignment is already terminal.
func (l *LeadAssignmentsResource) Release(ctx context.Context, assignmentID string) (*LeadAssignment, error) {
	var result LeadAssignment
	path := fmt.Sprintf("/api/v1/lead-assignments/%s/release", url.PathEscape(assignmentID))
	if err := l.http.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Complete transitions a claimed assignment to completed — the work is
// finished. Returns an error (409) when the assignment is not claimed.
func (l *LeadAssignmentsResource) Complete(ctx context.Context, assignmentID string) (*LeadAssignment, error) {
	var result LeadAssignment
	path := fmt.Sprintf("/api/v1/lead-assignments/%s/complete", url.PathEscape(assignmentID))
	if err := l.http.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
