package sonzai

import "context"

// OrgResource provides organization-level billing, contract, and usage operations.
type OrgResource struct {
	http *httpClient
}

// OrgBillingCheckoutOptions configures a billing checkout session.
type OrgBillingCheckoutOptions struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency,omitempty"`
}

// OrgSubscribeOptions configures a contract subscription.
type OrgSubscribeOptions struct {
	ContractID string `json:"contractId"`
}

// GetBilling returns the org's current billing status.
func (o *OrgResource) GetBilling(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/billing", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateBillingPortal creates a billing management portal session.
func (o *OrgResource) CreateBillingPortal(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/billing/portal", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateBillingCheckout creates a checkout session for purchasing credits.
func (o *OrgResource) CreateBillingCheckout(ctx context.Context, opts OrgBillingCheckoutOptions) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/billing/checkout", opts, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetContract returns the org's current contract details.
func (o *OrgResource) GetContract(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/contract", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Subscribe subscribes the org to a contract plan.
func (o *OrgResource) Subscribe(ctx context.Context, opts OrgSubscribeOptions) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/contract/subscribe", opts, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetLedger returns the org's credit ledger.
func (o *OrgResource) GetLedger(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/ledger", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetModelPricing returns current model pricing for the org.
func (o *OrgResource) GetModelPricing(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/model-pricing", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetServiceUsage returns current service usage metrics.
func (o *OrgResource) GetServiceUsage(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/service-usage", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUsageSummary returns an aggregated usage summary.
func (o *OrgResource) GetUsageSummary(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/usage-summary", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListServiceAgreements returns the org's service agreements.
func (o *OrgResource) ListServiceAgreements(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/service-agreements", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListActiveCharacters returns currently active characters for the org.
func (o *OrgResource) ListActiveCharacters(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/characters", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetContextEngineEvents returns context engine events for the org.
func (o *OrgResource) GetContextEngineEvents(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/events", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RedeemVoucher redeems a voucher code for credits.
func (o *OrgResource) RedeemVoucher(ctx context.Context, code string) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/vouchers/redeem", map[string]string{"code": code}, &result); err != nil {
		return nil, err
	}
	return result, nil
}
