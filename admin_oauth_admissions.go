package custd

import (
	"context"
	"net/http"
	"net/url"
)

// OAuthClientAdmissionAdminClient owns a company's external-application admissions.
//
// An admission is a company's authority to admit one already-registered OAuth
// client. It is not the client's native registration: approval pins the reviewed
// revision and security metadata, and it never replaces the server's current
// membership and metadata checks at token time. Revocation fences new grants and
// refreshes without deleting another company's admission.
type OAuthClientAdmissionAdminClient struct {
	admin *AdminClient
}

// AdminOAuthClientAdmissionGrant is the requested or approved ceiling for one
// configured resource.
type AdminOAuthClientAdmissionGrant struct {
	ResourceURI string   `json:"resourceUri"`
	Scopes      []string `json:"scopes"`
}

// AdminOAuthClientAdmission is the company-owned admission snapshot. ApprovedAt
// and RevokedAt are server-stamped timestamps; the SDK carries them as the
// strings the server sends, as it does for every other admin timestamp.
type AdminOAuthClientAdmission struct {
	AdmissionUUID             string                           `json:"admissionUuid"`
	CompanyUUID               string                           `json:"companyUuid"`
	ClientID                  string                           `json:"clientId"`
	RequestedSecurityRevision string                           `json:"requestedSecurityRevision"`
	ApprovedSecurityRevision  string                           `json:"approvedSecurityRevision,omitempty"`
	RequestedGrants           []AdminOAuthClientAdmissionGrant `json:"requestedGrants"`
	ApprovedGrants            []AdminOAuthClientAdmissionGrant `json:"approvedGrants"`
	Revision                  int64                            `json:"revision"`
	ApprovedAt                *string                          `json:"approvedAt,omitempty"`
	RevokedAt                 *string                          `json:"revokedAt,omitempty"`
}

// AdminOAuthClientAdmissionMetadata is the public client metadata a review needs
// to be informed: the stored name and redirects, the security revision they hash
// with, and when the registration expires. Only Review returns it.
type AdminOAuthClientAdmissionMetadata struct {
	ClientName       string   `json:"clientName"`
	RedirectURIs     []string `json:"redirectUris"`
	SecurityRevision string   `json:"securityRevision"`
	ExpiresAt        string   `json:"expiresAt"`
}

// AdminOAuthClientAdmissionList is the selected company's admissions. It carries
// no client metadata and never widens to another company's inventory.
type AdminOAuthClientAdmissionList struct {
	Admissions []AdminOAuthClientAdmission `json:"admissions"`
}

// AdminOAuthClientAdmissionReviewRequest selects the company whose stored
// admission should be revalidated against its current client metadata.
type AdminOAuthClientAdmissionReviewRequest struct {
	CompanySlug string `json:"companySlug"`
}

// AdminOAuthClientAdmissionReviewResponse is the exact stored admission plus the
// current metadata a reviewer must see before approving it.
type AdminOAuthClientAdmissionReviewResponse struct {
	Admission AdminOAuthClientAdmission         `json:"admission"`
	Metadata  AdminOAuthClientAdmissionMetadata `json:"metadata"`
}

// AdminOAuthClientAdmissionApproveRequest pins the reviewed snapshot: the
// revision and security revision the reviewer saw, and the grants accepted. The
// server refuses a changed request with 409, so it must be reviewed again.
type AdminOAuthClientAdmissionApproveRequest struct {
	CompanySlug      string                           `json:"companySlug"`
	ClientID         string                           `json:"clientId"`
	ExpectedRevision int64                            `json:"expectedRevision"`
	SecurityRevision string                           `json:"securityRevision"`
	Grants           []AdminOAuthClientAdmissionGrant `json:"grants"`
}

// AdminOAuthClientAdmissionRevokeRequest fences the admission at the revision the
// caller last saw.
type AdminOAuthClientAdmissionRevokeRequest struct {
	CompanySlug      string `json:"companySlug"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

// AdminOAuthClientAdmissionMutationResponse is the admission after an approve or
// revoke, carrying the revision the server applied.
type AdminOAuthClientAdmissionMutationResponse struct {
	Admission AdminOAuthClientAdmission `json:"admission"`
}

const oauthClientAdmissionsPath = "/oauth-client-admissions"

// List returns the selected company's admissions without fetching client
// metadata. companySlug is required and checked by the server against the
// caller's current human membership.
func (c *OAuthClientAdmissionAdminClient) List(
	ctx context.Context,
	companySlug string,
) (AdminOAuthClientAdmissionList, error) {
	var out AdminOAuthClientAdmissionList
	query := url.Values{companySlugQueryParameter: {companySlug}}.Encode()
	err := c.admin.request(ctx, http.MethodGet, oauthClientAdmissionsPath+"?"+query, nil, &out)
	return out, err
}

// Review revalidates the stored admission against the client's current metadata
// and returns both, so approval is a decision about what the server holds now.
func (c *OAuthClientAdmissionAdminClient) Review(
	ctx context.Context,
	admissionUUID string,
	req AdminOAuthClientAdmissionReviewRequest,
) (AdminOAuthClientAdmissionReviewResponse, error) {
	var out AdminOAuthClientAdmissionReviewResponse
	path := oauthClientAdmissionActionPath(admissionUUID, "review")
	err := c.admin.request(ctx, http.MethodPost, path, req, &out)
	return out, err
}

// Approve accepts exactly the reviewed snapshot. A changed admission answers 409;
// the SDK leaves that conflict visible so the caller reviews again instead of
// retrying a decision the server has already rejected.
func (c *OAuthClientAdmissionAdminClient) Approve(
	ctx context.Context,
	admissionUUID string,
	req AdminOAuthClientAdmissionApproveRequest,
) (AdminOAuthClientAdmissionMutationResponse, error) {
	var out AdminOAuthClientAdmissionMutationResponse
	path := oauthClientAdmissionActionPath(admissionUUID, "approve")
	err := c.admin.request(ctx, http.MethodPost, path, req, &out)
	return out, err
}

// Revoke withdraws this company's admission at expectedRevision. It fences new
// grants and refreshes; it does not delete the client's native metadata.
func (c *OAuthClientAdmissionAdminClient) Revoke(
	ctx context.Context,
	admissionUUID string,
	req AdminOAuthClientAdmissionRevokeRequest,
) (AdminOAuthClientAdmissionMutationResponse, error) {
	var out AdminOAuthClientAdmissionMutationResponse
	path := oauthClientAdmissionActionPath(admissionUUID, "revoke")
	err := c.admin.request(ctx, http.MethodPost, path, req, &out)
	return out, err
}

// oauthClientAdmissionActionPath escapes the caller-supplied identifier into the
// action route, so a value containing a separator cannot address another route.
func oauthClientAdmissionActionPath(admissionUUID, action string) string {
	return oauthClientAdmissionsPath + "/" + url.PathEscape(admissionUUID) + "/" + action
}
