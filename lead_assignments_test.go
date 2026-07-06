package sonzai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestLeadAssignmentsOffer(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/lead-assignments/offer":
			var body OfferLeadAssignmentOptions
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.LeadRef != "lead-1" || len(body.Candidates) != 2 || body.Policy != "load_balanced" || body.SLASeconds != 600 {
				t.Fatalf("unexpected offer body: %+v", body)
			}
			jsonResponse(w, 200, OfferLeadAssignmentResult{
				Assignment: LeadAssignment{
					AssignmentID: "assign-1",
					LeadRef:      "lead-1",
					RepUserID:    "rep-1",
					State:        "offered",
					Policy:       "load_balanced",
					OfferedAt:    "2026-07-06T10:00:00Z",
					SLAExpiresAt: "2026-07-06T10:10:00Z",
				},
				Deduplicated: false,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.LeadAssignments.Offer(context.Background(), OfferLeadAssignmentOptions{
		LeadRef:    "lead-1",
		Candidates: []string{"rep-1", "rep-2"},
		Policy:     "load_balanced",
		SLASeconds: 600,
	})
	if err != nil {
		t.Fatalf("Offer: %v", err)
	}
	if result.Deduplicated || result.Assignment.AssignmentID != "assign-1" || result.Assignment.State != "offered" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestLeadAssignmentsList(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/lead-assignments":
			q := r.URL.Query()
			if q.Get("rep_user_id") != "rep-1" || q.Get("state") != "offered" || q.Get("limit") != "10" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			jsonResponse(w, 200, ListLeadAssignmentsResult{
				Assignments: []LeadAssignment{{AssignmentID: "assign-1", RepUserID: "rep-1", State: "offered"}},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.LeadAssignments.List(context.Background(), &ListLeadAssignmentsOptions{
		RepUserID: "rep-1",
		State:     "offered",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Assignments) != 1 || result.Assignments[0].AssignmentID != "assign-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestLeadAssignmentsGet(t *testing.T) {
	server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/lead-assignments/assign-1":
			jsonResponse(w, 200, LeadAssignment{AssignmentID: "assign-1", State: "offered"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	result, err := client.LeadAssignments.Get(context.Background(), "assign-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if result.AssignmentID != "assign-1" || result.State != "offered" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestLeadAssignmentsTransitions(t *testing.T) {
	cases := []struct {
		name string
		path string
		call func(ctx context.Context, c *Client) (*LeadAssignment, error)
	}{
		{"claim", "/api/v1/lead-assignments/assign-1/claim", func(ctx context.Context, c *Client) (*LeadAssignment, error) {
			return c.LeadAssignments.Claim(ctx, "assign-1")
		}},
		{"release", "/api/v1/lead-assignments/assign-1/release", func(ctx context.Context, c *Client) (*LeadAssignment, error) {
			return c.LeadAssignments.Release(ctx, "assign-1")
		}},
		{"complete", "/api/v1/lead-assignments/assign-1/complete", func(ctx context.Context, c *Client) (*LeadAssignment, error) {
			return c.LeadAssignments.Complete(ctx, "assign-1")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, client := testServer(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != tc.path {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				jsonResponse(w, 200, LeadAssignment{AssignmentID: "assign-1", State: tc.name + "d"})
			})
			defer server.Close()

			result, err := tc.call(context.Background(), client)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if result.AssignmentID != "assign-1" {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}
