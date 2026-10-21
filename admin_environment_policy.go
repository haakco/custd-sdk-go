package custd

import (
	"context"
	"net/http"
	"net/url"
)

// EnvironmentPolicyAdminClient owns the tenant environment setting.
//
// The policy is read-before-write material: it reports the vocabulary, the
// active values and any reason a new environment would currently be refused,
// so a caller states intent against the real state rather than guessing.
type EnvironmentPolicyAdminClient struct {
	admin *AdminClient
}

// EnvironmentPolicyReason explains a state that would degrade or refuse a new
// environment. Code is the stable identifier; Summary and Remedy are for a
// person.
type EnvironmentPolicyReason struct {
	Code    string `json:"code"`
	Summary string `json:"summary"`
	Remedy  string `json:"remedy"`
}

// EnvironmentPolicy is the tenant's current environment setting.
type EnvironmentPolicy struct {
	AutoCreateEnvironments bool                      `json:"autoCreateEnvironments"`
	DefinitionPresent      bool                      `json:"definitionPresent"`
	DefinitionEnabled      bool                      `json:"definitionEnabled"`
	ActiveValueLimit       int                       `json:"activeValueLimit"`
	ActiveCount            int                       `json:"activeCount"`
	ActiveValues           []string                  `json:"activeValues"`
	DisabledValues         []string                  `json:"disabledValues"`
	Reasons                []EnvironmentPolicyReason `json:"reasons"`
}

// EnvironmentPolicyUpdateRequest sets automatic environment creation. A nil
// field leaves the current setting untouched, so a caller can change one
// behaviour without restating the others.
type EnvironmentPolicyUpdateRequest struct {
	AutoCreateEnvironments *bool `json:"autoCreateEnvironments"`
}

const environmentPolicyPath = "/environment-policy"

// Get reads the policy. CompanySlug is optional and only narrows the answer for
// a system administrator; the tenant otherwise comes from the token.
func (c *EnvironmentPolicyAdminClient) Get(
	ctx context.Context,
	companySlug string,
) (*EnvironmentPolicy, ResponseMetadata, error) {
	var out EnvironmentPolicy
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, environmentPolicyPath+optionalCompanySlug(companySlug), nil, &out,
	)
	return &out, metadata, err
}

// Update sets automatic creation and returns the resulting policy, so a caller
// sees the state it just applied rather than assuming it.
func (c *EnvironmentPolicyAdminClient) Update(
	ctx context.Context,
	companySlug string,
	req EnvironmentPolicyUpdateRequest,
) (*EnvironmentPolicy, ResponseMetadata, error) {
	var out EnvironmentPolicy
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPatch, environmentPolicyPath+optionalCompanySlug(companySlug), req, &out,
	)
	return &out, metadata, err
}

// optionalCompanySlug renders "?companySlug=..." only when a slug is given.
func optionalCompanySlug(companySlug string) string {
	if companySlug == "" {
		return ""
	}
	return "?" + companySlugQueryParameter + "=" + url.QueryEscape(companySlug)
}
