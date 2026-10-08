package custd

import (
	"context"
	"encoding/json"
	"net/http"
)

// ApplicationIdentityTraitsStatus states whether one identity's public traits are complete.
type ApplicationIdentityTraitsStatus string

const (
	ApplicationIdentityTraitsIncluded    ApplicationIdentityTraitsStatus = "included"
	ApplicationIdentityTraitsIncomplete  ApplicationIdentityTraitsStatus = "incomplete"
	ApplicationIdentityTraitsAbsent      ApplicationIdentityTraitsStatus = "identity_absent"
	ApplicationIdentityTraitsUnavailable ApplicationIdentityTraitsStatus = "directory_unavailable"
	ApplicationIdentityTraitsUnreadable  ApplicationIdentityTraitsStatus = "provider_unreadable"
)

// ApplicationIdentityTrait preserves the provider's JSON value without coercion.
type ApplicationIdentityTrait struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

// ApplicationPrincipalIdentityExport is one mapped identity and its own completeness status.
type ApplicationPrincipalIdentityExport struct {
	ProviderIssuer  string                          `json:"providerIssuer"`
	ProviderSubject string                          `json:"providerSubject"`
	TraitsStatus    ApplicationIdentityTraitsStatus `json:"traitsStatus"`
	Traits          []ApplicationIdentityTrait      `json:"traits"`
}

// ApplicationPrincipalMembershipView includes active and removed memberships.
type ApplicationPrincipalMembershipView struct {
	OrganisationID   string `json:"organisationId"`
	OrganisationSlug string `json:"organisationSlug"`
	OrganisationName string `json:"organisationName"`
	Role             string `json:"role"`
	CreatedAt        string `json:"createdAt"`
	RemovedAt        string `json:"removedAt,omitempty"`
}

type ApplicationPrincipalProfileValueView struct {
	FieldKey string `json:"fieldKey"`
	Value    string `json:"value"`
}

// ApplicationPrincipalExport contains public identity traits and application data,
// never credential, token, code or session material. Each identity states its own completeness.
type ApplicationPrincipalExport struct {
	ProjectID     string                                 `json:"projectId"`
	EnvironmentID string                                 `json:"environmentId"`
	DirectoryID   string                                 `json:"directoryId"`
	PrincipalID   string                                 `json:"principalId"`
	Enabled       bool                                   `json:"enabled"`
	CreatedAt     string                                 `json:"createdAt"`
	Identities    []ApplicationPrincipalIdentityExport   `json:"identities"`
	Memberships   []ApplicationPrincipalMembershipView   `json:"memberships"`
	ProfileValues []ApplicationPrincipalProfileValueView `json:"profileValues"`
}

func (c *AuthProjectAdminClient) ExportPrincipal(
	ctx context.Context, target ApplicationPrincipalTarget, options AuthProjectRequestOptions,
) (*ApplicationPrincipalExport, error) {
	var out ApplicationPrincipalExport
	if err := c.admin.requestWithHeaders(ctx, http.MethodGet, target.path()+"/export", nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}
