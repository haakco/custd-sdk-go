package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// AuthProjectIdentityMode names how an environment's identity boundary is
// shared. Custd owns the meaning of both legal values and the SDK surfaces them
// unchanged.
type AuthProjectIdentityMode string

const (
	// AuthProjectIdentityIsolated gives the environment its own identity pool.
	AuthProjectIdentityIsolated AuthProjectIdentityMode = "isolated"
	// AuthProjectIdentityShared lets the environment share an identity pool.
	AuthProjectIdentityShared AuthProjectIdentityMode = "shared"
)

// authProjectOwningUserHeader names the platform user a machine credential acts
// for (services/admin-api/cmd/api/admin_auth_project.go:76). Custd validates the
// named user as a live member of the machine caller's own company, so a machine
// caller must send it; a human administrator's token subject is the actor and
// leaves it unset.
const authProjectOwningUserHeader = "X-Custd-Owning-User-UUID"

// AuthProjectRequestOptions carries what every project-auth control-plane call
// needs beyond its typed request body.
//
// OwningUserUUID is sent as X-Custd-Owning-User-UUID. A machine credential must
// name the platform user it acts for or Custd refuses the call with 403; a human
// administrator leaves it empty.
//
// IdempotencyKey is sent as Idempotency-Key on the operations Custd makes
// retry-safe (create project, revoke one session, revoke all sessions). Custd
// requires it on those routes, so the SDK rejects an empty key before sending.
type AuthProjectRequestOptions struct {
	OwningUserUUID string
	IdempotencyKey string
}

// AuthProjectCreateRequest is the body of POST /api/v1/admin/auth-projects. It
// creates the named owning user's project ownership, identity pool and first
// paused environment atomically.
type AuthProjectCreateRequest struct {
	Slug            string                  `json:"slug"`
	Name            string                  `json:"name"`
	EnvironmentSlug string                  `json:"environmentSlug"`
	IdentityMode    AuthProjectIdentityMode `json:"identityMode"`
}

// AuthProjectEnvironmentCreateRequest is the body of
// POST /api/v1/admin/auth-projects/{projectId}/environments.
type AuthProjectEnvironmentCreateRequest struct {
	Slug         string                  `json:"slug"`
	Name         string                  `json:"name"`
	IdentityMode AuthProjectIdentityMode `json:"identityMode"`
}

// AuthProjectSummary is one project and the environment the call is scoped to.
type AuthProjectSummary struct {
	ProjectID       string                  `json:"projectId"`
	Slug            string                  `json:"slug"`
	Name            string                  `json:"name"`
	EnvironmentID   string                  `json:"environmentId"`
	EnvironmentSlug string                  `json:"environmentSlug"`
	IdentityMode    AuthProjectIdentityMode `json:"identityMode"`
}

// AuthProjectCreation is the response to creating a project. Replayed reports
// that the same body and Idempotency-Key returned the original operation;
// RuntimeReady reports whether the environment is serving, not whether
// application login is enabled.
type AuthProjectCreation struct {
	Project      AuthProjectSummary `json:"project"`
	OperationID  string             `json:"operationId"`
	Replayed     bool               `json:"replayed"`
	RuntimeReady bool               `json:"runtimeReady"`
}

// AuthProjectListResponse is the page of GET /api/v1/admin/auth-projects.
// NextAfter is empty on the last page.
type AuthProjectListResponse struct {
	Projects  []AuthProjectSummary `json:"projects"`
	NextAfter string               `json:"nextAfter,omitempty"`
}

// ApplicationSession is one session a directory holds for an application
// principal. The list carries no session token or credential.
type ApplicationSession struct {
	SessionID                   string `json:"sessionId"`
	Active                      bool   `json:"active"`
	AuthenticatedAt             string `json:"authenticatedAt,omitempty"`
	AuthenticatorAssuranceLevel string `json:"authenticatorAssuranceLevel,omitempty"`
	ExpiresAt                   string `json:"expiresAt,omitempty"`
	IssuedAt                    string `json:"issuedAt,omitempty"`
}

// ApplicationSessionInventory is the response to listing one application
// principal's live sessions.
type ApplicationSessionInventory struct {
	ProjectID     string               `json:"projectId"`
	EnvironmentID string               `json:"environmentId"`
	DirectoryID   string               `json:"directoryId"`
	PrincipalID   string               `json:"principalId"`
	Sessions      []ApplicationSession `json:"sessions"`
}

// ApplicationSessionRevokeRequest is the body of
// POST .../principals/{providerSubject}/sessions/revoke. The session identifier
// is checked against the principal's own sessions before it is revoked.
type ApplicationSessionRevokeRequest struct {
	SessionID string `json:"sessionId"`
}

// ApplicationSessionRevocation is the response to revoking sessions. Revoked is
// the number of sessions ended.
type ApplicationSessionRevocation struct {
	ProjectID   string `json:"projectId"`
	DirectoryID string `json:"directoryId"`
	PrincipalID string `json:"principalId"`
	Revoked     int64  `json:"revoked"`
	SessionID   string `json:"sessionId,omitempty"`
}

// ApplicationSessionsRevokeAllRequest is the body of
// POST .../principals/{providerSubject}/sessions/revoke-all. Confirm must be
// true: a missing field is never read as the strongest action.
type ApplicationSessionsRevokeAllRequest struct {
	Confirm bool `json:"confirm"`
}

// ApplicationMembershipRevokeRequest is the body of
// POST .../principals/{providerSubject}/memberships/revoke. The project,
// environment, directory and provider subject come from the address, so the body
// names only the organisation.
type ApplicationMembershipRevokeRequest struct {
	OrganisationID string `json:"organisationId"`
	Reason         string `json:"reason,omitempty"`
}

// ApplicationMembershipRevocation is the response to ending a membership. The
// membership row is kept with RemovedAt stamped rather than deleted.
type ApplicationMembershipRevocation struct {
	ProjectID      string `json:"projectId"`
	EnvironmentID  string `json:"environmentId"`
	DirectoryID    string `json:"directoryId"`
	PrincipalID    string `json:"principalId"`
	OrganisationID string `json:"organisationId"`
	RemovedAt      string `json:"removedAt"`
	Removed        bool   `json:"removed"`
}

// AuthProjectAudienceBinding is one application audience the environment should
// admit and its provider registration settings. This is the binding a consumer's
// edge admits against; the contract derives the audience group
// custd-group-<environmentID>-<audienceSlug> from it.
type AuthProjectAudienceBinding struct {
	Audience               string   `json:"audience"`
	PublicClient           bool     `json:"publicClient"`
	RedirectURIs           []string `json:"redirectUris"`
	PostLogoutRedirectURIs []string `json:"postLogoutRedirectUris"`
	AllowedOrigins         []string `json:"allowedOrigins"`
}

// AuthProjectProfileField is one profile field's policy inside the environment's
// desired state.
type AuthProjectProfileField struct {
	Key                  string `json:"key"`
	Required             bool   `json:"required"`
	VisibleToApplication bool   `json:"visibleToApplication"`
	EditableBy           string `json:"editableBy"`
}

// AuthProjectDesiredState is the environment configuration an apply writes. A
// consumer builds it from these fields alone: Audiences carries the audience
// binding, and IdentityMode, RegistrationPolicy, LoginPaused and ProfileFields
// carry the rest. The legal scalar vocabularies are read from the capability
// operation Custd exposes for the environment, not fixed here.
type AuthProjectDesiredState struct {
	IdentityMode       AuthProjectIdentityMode      `json:"identityMode"`
	RegistrationPolicy string                       `json:"registrationPolicy"`
	LoginPaused        bool                         `json:"loginPaused"`
	Audiences          []AuthProjectAudienceBinding `json:"audiences"`
	ProfileFields      []AuthProjectProfileField    `json:"profileFields"`
}

// AuthProjectDesiredStateRequest is the body of both the apply and preview
// operations. ExpectedRevision is the revision the caller read; Custd refuses an
// apply when the stored revision has moved.
type AuthProjectDesiredStateRequest struct {
	DesiredState     AuthProjectDesiredState `json:"desiredState"`
	ExpectedRevision int64                   `json:"expectedRevision"`
}

// AuthProjectOperation is the receipt an apply returns. Replayed reports that
// the same body and Idempotency-Key returned the original operation.
type AuthProjectOperation struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	Status         string `json:"status"`
	Revision       int64  `json:"revision"`
	IdempotencyKey string `json:"idempotencyKey"`
	AppliedAt      string `json:"appliedAt"`
	Replayed       bool   `json:"replayed"`
}

// AuthProjectChange is one field-level change a preview reports.
type AuthProjectChange struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// AuthProjectPreview is the field-level change set an apply would write. It
// writes nothing.
type AuthProjectPreview struct {
	ProjectID     string              `json:"projectId"`
	EnvironmentID string              `json:"environmentId"`
	Revision      int64               `json:"revision"`
	Changes       []AuthProjectChange `json:"changes"`
	SideEffects   []string            `json:"sideEffects"`
	NoOp          bool                `json:"noOp"`
}

// AuthProjectClientRegistration is one audience's provider registration read
// back from status. ClientID is the derived custd-app-<environmentID>-<audienceSlug>
// the audience binding resolves to.
type AuthProjectClientRegistration struct {
	Audience string `json:"audience"`
	ClientID string `json:"clientId"`
	Observed bool   `json:"observed"`
}

// AuthProjectClientSync is the registration half of status: which application
// clients the provider is known to hold and whether they are current.
type AuthProjectClientSync struct {
	ClientIDs     []string                        `json:"clientIds"`
	Revision      int64                           `json:"revision"`
	CheckedAt     string                          `json:"checkedAt"`
	Current       bool                            `json:"current"`
	Registrations []AuthProjectClientRegistration `json:"registrations"`
	Issuer        string                          `json:"issuer,omitempty"`
	ErrorCategory string                          `json:"errorCategory,omitempty"`
}

// AuthProjectStatus is the environment's configured state and applied revision.
// Desired.Audiences is the audience binding the caller applied, so the
// project/environment/audience mapping the contract requires is readable here
// after an apply. Reconciled describes the configuration and LoginReady
// describes execution; neither is inferred from the other. ClientSync is absent
// until registration has run.
type AuthProjectStatus struct {
	ProjectID     string                  `json:"projectId"`
	EnvironmentID string                  `json:"environmentId"`
	Revision      int64                   `json:"revision"`
	Desired       AuthProjectDesiredState `json:"desired"`
	Reconciled    bool                    `json:"reconciled"`
	ReconcileNote string                  `json:"reconcileNote"`
	LoginReady    bool                    `json:"loginReady"`
	LoginNote     string                  `json:"loginNote"`
	ClientSync    *AuthProjectClientSync  `json:"clientSync,omitempty"`
}

// AuthProjectAdminClient manages Custd projects, their environments, and the
// application principals that directory holds inside an environment. Every call
// is an admin control-plane call under /api/v1/admin/auth-projects and carries
// the owning-user header from AuthProjectRequestOptions.
type AuthProjectAdminClient struct {
	admin *AdminClient
}

// ListProjects returns up to 100 environments the named owning user owns or
// operates. Pass the previous page's NextAfter as after for the next page.
func (c *AuthProjectAdminClient) ListProjects(
	ctx context.Context,
	after string,
	options AuthProjectRequestOptions,
) (*AuthProjectListResponse, error) {
	path := "/auth-projects"
	if trimmed := strings.TrimSpace(after); trimmed != "" {
		path += "?after=" + url.QueryEscape(trimmed)
	}
	var out AuthProjectListResponse
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, path, nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProject creates the named owning user's project ownership, identity pool
// and first paused environment atomically.
func (c *AuthProjectAdminClient) CreateProject(
	ctx context.Context,
	req AuthProjectCreateRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectCreation, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectCreation
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, "/auth-projects", req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateEnvironment adds a further environment to an existing project. The new
// environment starts paused and invite-only.
func (c *AuthProjectAdminClient) CreateEnvironment(
	ctx context.Context,
	projectID string,
	req AuthProjectEnvironmentCreateRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectSummary, error) {
	var out AuthProjectSummary
	path := "/auth-projects/" + url.PathEscape(projectID) + "/environments"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPrincipalSessions reports the sessions the directory currently holds for
// one application principal.
func (c *AuthProjectAdminClient) ListPrincipalSessions(
	ctx context.Context,
	projectID string,
	environmentID string,
	directoryID string,
	providerSubject string,
	options AuthProjectRequestOptions,
) (*ApplicationSessionInventory, error) {
	var out ApplicationSessionInventory
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, applicationPrincipalPath(projectID, environmentID, directoryID, providerSubject)+"/sessions",
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokePrincipalSession ends exactly one of an application principal's
// sessions and leaves its other sessions untouched.
func (c *AuthProjectAdminClient) RevokePrincipalSession(
	ctx context.Context,
	projectID string,
	environmentID string,
	directoryID string,
	providerSubject string,
	req ApplicationSessionRevokeRequest,
	options AuthProjectRequestOptions,
) (*ApplicationSessionRevocation, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out ApplicationSessionRevocation
	path := applicationPrincipalPath(projectID, environmentID, directoryID, providerSubject) + "/sessions/revoke"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokePrincipalSessions ends every session the directory holds for one
// application principal. The request must confirm explicitly.
func (c *AuthProjectAdminClient) RevokePrincipalSessions(
	ctx context.Context,
	projectID string,
	environmentID string,
	directoryID string,
	providerSubject string,
	req ApplicationSessionsRevokeAllRequest,
	options AuthProjectRequestOptions,
) (*ApplicationSessionRevocation, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out ApplicationSessionRevocation
	path := applicationPrincipalPath(projectID, environmentID, directoryID, providerSubject) + "/sessions/revoke-all"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokePrincipalMembership ends one application identity's membership of one
// organisation.
func (c *AuthProjectAdminClient) RevokePrincipalMembership(
	ctx context.Context,
	projectID string,
	environmentID string,
	directoryID string,
	providerSubject string,
	req ApplicationMembershipRevokeRequest,
	options AuthProjectRequestOptions,
) (*ApplicationMembershipRevocation, error) {
	var out ApplicationMembershipRevocation
	path := applicationPrincipalPath(projectID, environmentID, directoryID, providerSubject) + "/memberships/revoke"
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ApplyEnvironmentDesiredState writes the environment's desired state and
// returns the operation receipt. Custd makes it retry-safe per Idempotency-Key,
// so the key is required. The audience binding the consumer's edge admits
// against is written here and read back through GetEnvironmentStatus.
func (c *AuthProjectAdminClient) ApplyEnvironmentDesiredState(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectDesiredStateRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectOperation, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectOperation
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, environmentDesiredStatePath(projectID, environmentID)+"/apply",
		req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// PreviewEnvironmentDesiredState reports the field-level changes an apply would
// make and writes nothing.
func (c *AuthProjectAdminClient) PreviewEnvironmentDesiredState(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectDesiredStateRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectPreview, error) {
	var out AuthProjectPreview
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, environmentDesiredStatePath(projectID, environmentID)+"/preview",
		req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEnvironmentStatus reads the environment's configured state and its applied
// revision. Status.Desired.Audiences carries the audience binding and Status
// carries the environment binding, so a caller that cannot create or read the
// binding anywhere else reads it here.
func (c *AuthProjectAdminClient) GetEnvironmentStatus(
	ctx context.Context,
	projectID string,
	environmentID string,
	options AuthProjectRequestOptions,
) (*AuthProjectStatus, error) {
	var out AuthProjectStatus
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, environmentDesiredStatePath(projectID, environmentID)+"/status",
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// environmentDesiredStatePath addresses one environment under one project. Every
// segment is escaped so a caller-supplied identifier cannot reshape the path.
func environmentDesiredStatePath(projectID, environmentID string) string {
	return "/auth-projects/" + url.PathEscape(projectID) +
		"/environments/" + url.PathEscape(environmentID)
}

// applicationPrincipalPath addresses one application principal inside one
// directory. Every segment is escaped so a caller-supplied identifier cannot
// reshape the path.
func applicationPrincipalPath(projectID, environmentID, directoryID, providerSubject string) string {
	return "/auth-projects/" + url.PathEscape(projectID) +
		"/environments/" + url.PathEscape(environmentID) +
		"/directories/" + url.PathEscape(directoryID) +
		"/principals/" + url.PathEscape(providerSubject)
}

// authProjectHeaders builds the headers every project-auth call sends. A blank
// owning user is omitted because a human administrator's own token subject is
// the actor.
func authProjectHeaders(options AuthProjectRequestOptions) map[string]string {
	headers := map[string]string{}
	if trimmed := strings.TrimSpace(options.OwningUserUUID); trimmed != "" {
		headers[authProjectOwningUserHeader] = trimmed
	}
	if key := strings.TrimSpace(options.IdempotencyKey); key != "" {
		headers["Idempotency-Key"] = key
	}
	return headers
}

// requireAuthProjectIdempotencyKey rejects an empty key on the routes Custd
// makes retry-safe, so a caller gets one clear error instead of a service 400.
func requireAuthProjectIdempotencyKey(options AuthProjectRequestOptions) error {
	if strings.TrimSpace(options.IdempotencyKey) == "" {
		return fmt.Errorf("custd: auth-project idempotency key is required")
	}
	return nil
}
