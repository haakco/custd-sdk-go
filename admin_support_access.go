package custd

import (
	"context"
	"net/http"
	"net/url"
)

// SupportAccessAdminClient owns operator support-access grants.
//
// A grant is one operator, one company, one reason and one bounded window. The
// operator is taken from the authenticated session by the server, never from
// the request, so a caller cannot widen its own authority by naming someone
// else here.
type SupportAccessAdminClient struct {
	admin *AdminClient
}

// SupportAccessGrant is one operator's bounded permission to act on a company.
type SupportAccessGrant struct {
	SupportAccessGrantUUID string  `json:"supportAccessGrantUuid"`
	CompanySlug            string  `json:"companySlug"`
	OperatorUserUUID       string  `json:"operatorUserUuid"`
	Reason                 string  `json:"reason"`
	StartsAt               string  `json:"startsAt"`
	ExpiresAt              string  `json:"expiresAt"`
	RevokedAt              *string `json:"revokedAt"`
}

// SupportAccessGrantCreateRequest asks the control plane for a grant. The
// operator is server-derived; callers must not pre-fill it.
type SupportAccessGrantCreateRequest struct {
	CompanySlug string `json:"companySlug"`
	Reason      string `json:"reason"`
	ExpiresAt   string `json:"expiresAt"`
}

// SupportAccessGrantListResponse lists the caller's grants.
type SupportAccessGrantListResponse struct {
	Grants []SupportAccessGrant `json:"grants"`
}

const supportAccessPath = "/support-access"

// Create asks for a support-access grant and reports the request identifier the
// server logged, so a client can quote it if the grant is refused.
func (c *SupportAccessAdminClient) Create(
	ctx context.Context,
	req SupportAccessGrantCreateRequest,
) (*SupportAccessGrant, ResponseMetadata, error) {
	var out SupportAccessGrant
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, supportAccessPath, req, &out)
	return &out, metadata, err
}

// List returns the caller's grants, including any that have expired.
func (c *SupportAccessAdminClient) List(
	ctx context.Context,
) (*SupportAccessGrantListResponse, ResponseMetadata, error) {
	var out SupportAccessGrantListResponse
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, supportAccessPath, nil, &out)
	return &out, metadata, err
}

// Revoke ends a grant early. Revocation is forward-only: the server records
// when it happened and keeps the grant readable.
func (c *SupportAccessAdminClient) Revoke(
	ctx context.Context,
	grantUUID string,
) (*SupportAccessGrant, ResponseMetadata, error) {
	var out SupportAccessGrant
	path := supportAccessPath + "/" + url.PathEscape(grantUUID) + "/revoke"
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, &out)
	return &out, metadata, err
}
