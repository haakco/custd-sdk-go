package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type PurchaseHandoffCredential struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	App          string `json:"app"`
	Scope        string `json:"scope"`
}

type PurchaseHandoffCredentialSummary struct {
	ClientID  string    `json:"clientId"`
	CreatedAt time.Time `json:"createdAt"`
}

type PurchaseHandoffCredentialList struct {
	Credentials []PurchaseHandoffCredentialSummary `json:"credentials"`
}

func (c *AuthProjectAdminClient) ListPurchaseHandoffCredentials(ctx context.Context, slug string, options AuthProjectRequestOptions) (*PurchaseHandoffCredentialList, error) {
	var out PurchaseHandoffCredentialList
	err := c.admin.requestWithHeaders(ctx, http.MethodGet, purchaseHandoffCredentialsPath(slug), nil, &out, authProjectHeaders(options))
	return &out, err
}

// CreatePurchaseHandoffCredential returns the client secret once.
func (c *AuthProjectAdminClient) CreatePurchaseHandoffCredential(ctx context.Context, slug string, options AuthProjectRequestOptions) (*PurchaseHandoffCredential, error) {
	var out PurchaseHandoffCredential
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, purchaseHandoffCredentialsPath(slug), nil, &out, authProjectHeaders(options))
	return &out, err
}

func (c *AuthProjectAdminClient) DeletePurchaseHandoffCredential(ctx context.Context, slug, clientID string, options AuthProjectRequestOptions) error {
	return c.admin.requestWithHeaders(ctx, http.MethodDelete, purchaseHandoffCredentialsPath(slug)+"/"+url.PathEscape(clientID), nil, nil, authProjectHeaders(options))
}

func purchaseHandoffCredentialsPath(slug string) string {
	return appsPath + "/" + url.PathEscape(slug) + "/purchase-handoff-credentials"
}
