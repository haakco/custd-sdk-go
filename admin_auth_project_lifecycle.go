package custd

import (
	"context"
	"net/http"
	"net/url"
)

// ApplicationProfileField is one field of a principal's profile as the policy
// discloses it. Editable reports whether this administrator surface may write the
// field; a provider-owned field is disclosed but not writable here.
type ApplicationProfileField struct {
	FieldKey   string `json:"fieldKey"`
	Value      string `json:"value,omitempty"`
	Present    bool   `json:"present"`
	Required   bool   `json:"required"`
	EditableBy string `json:"editableBy"`
	Editable   bool   `json:"editable"`
}

// ApplicationPrincipalProfile is one principal's profile. It carries only the
// fields the environment's policy marks visible, and it names the required fields
// that have no stored value as MissingRequired rather than inventing a default.
type ApplicationPrincipalProfile struct {
	ProjectID       string                    `json:"projectId"`
	EnvironmentID   string                    `json:"environmentId"`
	DirectoryID     string                    `json:"directoryId"`
	PrincipalID     string                    `json:"principalId"`
	Fields          []ApplicationProfileField `json:"fields"`
	MissingRequired []string                  `json:"missingRequired"`
}

// ApplicationProfileValue is one field value a caller asks to store.
type ApplicationProfileValue struct {
	FieldKey string `json:"fieldKey"`
	Value    string `json:"value"`
}

// ApplicationProfileValuesRequest is one principal's requested profile values.
// The environment, directory and principal come from the address, so the body can
// name no boundary of its own. A write is an upsert, so repeating it converges.
type ApplicationProfileValuesRequest struct {
	Values []ApplicationProfileValue `json:"values"`
}

// ApplicationMembershipChoice is one organisation an application principal may act
// for, with the role it holds there. The role is scoped to this organisation
// alone: the same principal may hold a different role in another.
type ApplicationMembershipChoice struct {
	OrganisationID   string `json:"organisationId"`
	OrganisationSlug string `json:"organisationSlug"`
	OrganisationName string `json:"organisationName"`
	Role             string `json:"role"`
}

// ApplicationMembershipContext is the set of organisations one application
// principal may act for. It is a management read, not the login path, so a paused
// environment still lists its memberships.
type ApplicationMembershipContext struct {
	ProjectID     string                        `json:"projectId"`
	EnvironmentID string                        `json:"environmentId"`
	DirectoryID   string                        `json:"directoryId"`
	PrincipalID   string                        `json:"principalId"`
	Memberships   []ApplicationMembershipChoice `json:"memberships"`
}

// GetPrincipalProfile reads one application principal's profile values. It returns
// only the fields the policy marks visible and reports each required field with no
// stored value as missing rather than defaulted.
func (c *AuthProjectAdminClient) GetPrincipalProfile(
	ctx context.Context,
	target ApplicationPrincipalTarget,
	options AuthProjectRequestOptions,
) (*ApplicationPrincipalProfile, error) {
	var out ApplicationPrincipalProfile
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, target.path()+"/profile", nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetPrincipalProfile sets one application principal's profile values. Custd makes
// it retry-safe per Idempotency-Key, so the key is required. A write refuses a
// field the policy does not configure and a provider-owned field.
func (c *AuthProjectAdminClient) SetPrincipalProfile(
	ctx context.Context,
	target ApplicationPrincipalTarget,
	req ApplicationProfileValuesRequest,
	options AuthProjectRequestOptions,
) (*ApplicationPrincipalProfile, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out ApplicationPrincipalProfile
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, target.path()+"/profile", req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPrincipalMemberships reports the organisations one application principal may
// currently act for. It is an operator read and does not require the credential
// store to be serving.
func (c *AuthProjectAdminClient) ListPrincipalMemberships(
	ctx context.Context,
	target ApplicationPrincipalTarget,
	options AuthProjectRequestOptions,
) (*ApplicationMembershipContext, error) {
	var out ApplicationMembershipContext
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, target.path()+"/memberships", nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// AuthProjectDeletionRequest deletes one project. ConfirmSlug must equal the
// stored project slug exactly: a caller that cannot name the project cannot
// delete it, which stops a mistyped identifier from removing the wrong boundary.
type AuthProjectDeletionRequest struct {
	ConfirmSlug string `json:"confirmSlug"`
}

// AuthProjectReleasedResources is what deleting a project released, grouped so
// the receipt documents the cascade instead of returning a bare count.
type AuthProjectReleasedResources struct {
	Environments     int `json:"environments"`
	Directories      int `json:"directories"`
	Organisations    int `json:"organisations"`
	Memberships      int `json:"memberships"`
	Invitations      int `json:"invitations"`
	AudienceBindings int `json:"audienceBindings"`
	Operations       int `json:"operations"`
}

// AuthProjectDeletion is the receipt for a deleted project.
type AuthProjectDeletion struct {
	ProjectID string                       `json:"projectId"`
	Slug      string                       `json:"slug"`
	Name      string                       `json:"name"`
	Released  AuthProjectReleasedResources `json:"released"`
}

// AuthProjectEnvironmentDeletionRequest deletes one environment. ConfirmSlug must
// equal the stored environment slug exactly, so a mistyped identifier cannot
// remove the wrong boundary.
type AuthProjectEnvironmentDeletionRequest struct {
	ConfirmSlug string `json:"confirmSlug"`
}

// AuthProjectEnvironmentReleasedResources is what deleting one environment
// released, grouped so the receipt documents the cascade.
type AuthProjectEnvironmentReleasedResources struct {
	Directories      int `json:"directories"`
	Organisations    int `json:"organisations"`
	Memberships      int `json:"memberships"`
	Invitations      int `json:"invitations"`
	AudienceBindings int `json:"audienceBindings"`
	Operations       int `json:"operations"`
	ProfileValues    int `json:"profileValues"`
}

// AuthProjectEnvironmentDeletion is the receipt for a deleted environment. It
// names both the project and the one environment it removed, so a caller can tell
// an environment deletion from a project deletion without reading the request back.
type AuthProjectEnvironmentDeletion struct {
	ProjectID     string                                  `json:"projectId"`
	ProjectSlug   string                                  `json:"projectSlug"`
	EnvironmentID string                                  `json:"environmentId"`
	Slug          string                                  `json:"slug"`
	Name          string                                  `json:"name"`
	Released      AuthProjectEnvironmentReleasedResources `json:"released"`
}

// DeleteProject removes the project and everything scoped to it in one
// transaction. It is owner-only and refuses with project_holds_directory_store
// while any directory still holds a credential store that was not released.
func (c *AuthProjectAdminClient) DeleteProject(
	ctx context.Context,
	projectID string,
	req AuthProjectDeletionRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectDeletion, error) {
	var out AuthProjectDeletion
	path := "/auth-projects/" + url.PathEscape(projectID) + "/delete"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEnvironment removes only the addressed environment and everything scoped
// to it, leaving the project's other environments untouched. It refuses with
// environment_holds_directory_store while a directory still holds a credential
// store slot.
func (c *AuthProjectAdminClient) DeleteEnvironment(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectEnvironmentDeletionRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectEnvironmentDeletion, error) {
	var out AuthProjectEnvironmentDeletion
	path := authProjectEnvironmentPath(projectID, environmentID) + "/delete"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// AuthProjectEnvironmentOffboardRequest offboards one environment. ConfirmSlug must
// equal the stored environment slug exactly: a caller that cannot name the
// environment cannot retire it, which stops a mistyped identifier from
// withdrawing the wrong boundary.
type AuthProjectEnvironmentOffboardRequest struct {
	ConfirmSlug string `json:"confirmSlug"`
}

// AuthEnvironmentClientCleanup reports the provider half of one offboard. Status
// is one of complete, pending or unavailable; ClientsRemoved counts the owned
// registrations this run deleted; ErrorCategory is the bounded reason a pending
// run did not finish.
type AuthEnvironmentClientCleanup struct {
	Status         string `json:"status"`
	ClientsRemoved int    `json:"clientsRemoved"`
	ErrorCategory  string `json:"errorCategory,omitempty"`
}

// AuthEnvironmentRetainedResources is what an offboard kept, so the receipt states
// plainly that nothing was destroyed and how much credential-store capacity stays
// reserved.
type AuthEnvironmentRetainedResources struct {
	Directories   int `json:"directories"`
	Organisations int `json:"organisations"`
	Memberships   int `json:"memberships"`
	Invitations   int `json:"invitations"`
	ProfileValues int `json:"profileValues"`
	ReservedSlots int `json:"reservedSlots"`
}

// AuthProjectEnvironmentOffboarding is the receipt for one offboarded environment.
// It names both the project and the one environment it withdrew, so a caller can
// tell it from a physical delete without reading the request back.
type AuthProjectEnvironmentOffboarding struct {
	ProjectID          string                           `json:"projectId"`
	ProjectSlug        string                           `json:"projectSlug"`
	EnvironmentID      string                           `json:"environmentId"`
	Slug               string                           `json:"slug"`
	Name               string                           `json:"name"`
	AuthorityWithdrawn bool                             `json:"authorityWithdrawn"`
	LoginPaused        bool                             `json:"loginPaused"`
	Revision           int64                            `json:"revision"`
	ClientCleanup      AuthEnvironmentClientCleanup     `json:"clientCleanup"`
	Retained           AuthEnvironmentRetainedResources `json:"retained"`
}

// OffboardEnvironment withdraws one environment and reconciles its owned provider
// clients away while keeping the environment, its pools, its directory-store
// reservations and its identity mappings. It is the safe step for an environment
// whose directory still holds a credential store, which a physical delete refuses.
//
// The route has no Idempotency-Key: the withdrawal is an addressed marker and a
// retry resumes the provider stage rather than duplicating it. A deployment with no
// provider surface reports clientCleanup.status as unavailable instead of failing.
func (c *AuthProjectAdminClient) OffboardEnvironment(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectEnvironmentOffboardRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectEnvironmentOffboarding, error) {
	var out AuthProjectEnvironmentOffboarding
	path := authProjectEnvironmentPath(projectID, environmentID) + "/offboard"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}
