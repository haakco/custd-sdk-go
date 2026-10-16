package custd

import (
	"context"
	"net/http"
	"net/url"
)

// PrincipalAudience is one audience's admission state. Managed means the environment requires explicit grants.
type PrincipalAudience struct {
	Audience string `json:"audience"`
	Granted  bool   `json:"granted"`
	Managed  bool   `json:"managed"`
}

// PrincipalAudiences lists the addressed principal's audience admission states.
type PrincipalAudiences struct {
	AudienceEntitlementMode string              `json:"audienceEntitlementMode"`
	Audiences               []PrincipalAudience `json:"audiences"`
}

// AudienceGrantRequest names the directory, organisation and exact registered audience to change.
type AudienceGrantRequest struct {
	DirectoryID    string `json:"directoryId"`
	Audience       string `json:"audience"`
	OrganisationID string `json:"organisationId"`
}

// AudienceGrant reports one audience's resulting grant state.
type AudienceGrant struct {
	PrincipalID    string `json:"principalId"`
	OrganisationID string `json:"organisationId"`
	Audience       string `json:"audience"`
	Granted        bool   `json:"granted"`
	Managed        bool   `json:"managed"`
}

// ListPrincipalAudiences reads admission state using the existing directory/provider-subject address.
func (c *AuthProjectAdminClient) ListPrincipalAudiences(ctx context.Context, target ApplicationPrincipalTarget, options AuthProjectRequestOptions) (*PrincipalAudiences, error) {
	var out PrincipalAudiences
	path := principalAudiencesPath(target) + "?" + url.Values{"directoryId": {target.DirectoryID}}.Encode()
	if err := c.admin.requestWithHeaders(ctx, http.MethodGet, path, nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// GrantPrincipalAudience grants exactly one audience; it is idempotent and audited server-side.
func (c *AuthProjectAdminClient) GrantPrincipalAudience(ctx context.Context, target ApplicationPrincipalTarget, req AudienceGrantRequest, options AuthProjectRequestOptions) (*AudienceGrant, error) {
	return c.changePrincipalAudience(ctx, target, req, "grant", options)
}

// RevokePrincipalAudience revokes only this audience, leaving sibling audiences and organisation
// membership untouched. It is idempotent and audited server-side; no replacement audience is required.
func (c *AuthProjectAdminClient) RevokePrincipalAudience(ctx context.Context, target ApplicationPrincipalTarget, req AudienceGrantRequest, options AuthProjectRequestOptions) (*AudienceGrant, error) {
	return c.changePrincipalAudience(ctx, target, req, "revoke", options)
}

func (c *AuthProjectAdminClient) changePrincipalAudience(ctx context.Context, target ApplicationPrincipalTarget, req AudienceGrantRequest, action string, options AuthProjectRequestOptions) (*AudienceGrant, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AudienceGrant
	if err := c.admin.requestWithHeaders(ctx, http.MethodPost, principalAudiencesPath(target)+"/"+action, req, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Audience routes use the same provider-subject identity, with directoryId outside the path.
func principalAudiencesPath(target ApplicationPrincipalTarget) string {
	return authProjectEnvironmentPath(target.ProjectID, target.EnvironmentID) + "/principals/" + url.PathEscape(target.ProviderSubject) + "/audiences"
}
