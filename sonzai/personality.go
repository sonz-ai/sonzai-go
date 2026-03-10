package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// PersonalityResource provides personality operations for an agent.
type PersonalityResource struct {
	http *httpClient
}

// Get returns the personality profile and evolution history.
func (p *PersonalityResource) Get(ctx context.Context, agentID string, opts *PersonalityGetOptions) (*PersonalityResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.HistoryLimit > 0 {
			params["history_limit"] = strconv.Itoa(opts.HistoryLimit)
		}
		if opts.Since != "" {
			params["since"] = opts.Since
		}
	}

	var result PersonalityResponse
	err := p.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/personality", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
