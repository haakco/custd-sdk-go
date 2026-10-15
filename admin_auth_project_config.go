package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AuthProjectCapability reports which configured modes this environment may use,
// so a caller can show unavailable options honestly rather than offering them.
// The legal vocabularies are server-owned and surfaced unchanged.
type AuthProjectCapability struct {
	IdentityModes        []string `json:"identityModes"`
	RegistrationPolicies []string `json:"registrationPolicies"`
	ProfileFieldKeys     []string `json:"profileFieldKeys"`
	AudienceConstraints  string   `json:"audienceConstraints"`
	SubjectUniqueness    string   `json:"subjectUniqueness"`
}

// AuthProjectDirectoryView is one directory's credential-store state as reported.
type AuthProjectDirectoryView struct {
	PoolID            string `json:"poolId"`
	PoolSlug          string `json:"poolSlug"`
	State             string `json:"state"`
	SlotRef           string `json:"slotRef,omitempty"`
	RuntimePublicURL  string `json:"runtimePublicUrl,omitempty"`
	Revision          int64  `json:"revision"`
	AttemptCount      int    `json:"attemptCount"`
	LastErrorCategory string `json:"lastErrorCategory,omitempty"`
	ReadyAt           string `json:"readyAt,omitempty"`
}

// AuthProjectDirectoryCapacity reports the deployment's declared credential-store
// pool: how many stores it declares, how many directories hold, and how many are
// free to allocate. The pool is deployment-wide, so every environment's report
// carries the same numbers.
type AuthProjectDirectoryCapacity struct {
	Declared int `json:"declared"`
	InUse    int `json:"inUse"`
	Free     int `json:"free"`
}

// AuthProjectDirectoryStatus is the environment's identity directories and their
// credential-store readiness. Ready is true only when every directory that has a
// store row is serving; a directory with no store row is a draft and does not make
// the environment unready.
type AuthProjectDirectoryStatus struct {
	ProjectID     string                       `json:"projectId"`
	EnvironmentID string                       `json:"environmentId"`
	Revision      int64                        `json:"revision"`
	Directories   []AuthProjectDirectoryView   `json:"directories"`
	Ready         bool                         `json:"ready"`
	Capacity      AuthProjectDirectoryCapacity `json:"capacity"`
}

// AuthProjectDirectoryAction names one directory credential-store action. Custd
// owns the four legal values and the SDK rejects anything else before sending.
type AuthProjectDirectoryAction string

const (
	// AuthProjectDirectoryEnable allocates a declared store and verifies it serves.
	AuthProjectDirectoryEnable AuthProjectDirectoryAction = "enable"
	// AuthProjectDirectoryPause stops a directory authenticating without releasing it.
	AuthProjectDirectoryPause AuthProjectDirectoryAction = "pause"
	// AuthProjectDirectoryResume restarts a paused directory's store.
	AuthProjectDirectoryResume AuthProjectDirectoryAction = "resume"
	// AuthProjectDirectoryRelease stops the directory authenticating and keeps the slot.
	AuthProjectDirectoryRelease AuthProjectDirectoryAction = "release"
)

var authProjectDirectoryActions = map[AuthProjectDirectoryAction]struct{}{
	AuthProjectDirectoryEnable:  {},
	AuthProjectDirectoryPause:   {},
	AuthProjectDirectoryResume:  {},
	AuthProjectDirectoryRelease: {},
}

// AuthProjectDirectoryRequest is the body of a directory action. It cannot name a
// store, a runtime URL or a state: those are server-owned. ExpectedRevision is the
// revision the caller read.
type AuthProjectDirectoryRequest struct {
	PoolID           string `json:"poolId"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

// AuthProjectEnvironmentOrganisation is one organisation an environment's
// directory holds: the stable id an invitation names, the slug and name a person
// sees, and the identity pool it belongs to. It carries no member data.
type AuthProjectEnvironmentOrganisation struct {
	OrganisationID string `json:"organisationId"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	IdentityPoolID string `json:"identityPoolId"`
}

// AuthProjectEnvironmentOrganisationList is the named list an environment's
// organisation read answers with; it echoes the addressed pair so a caller can
// confirm which environment it read.
type AuthProjectEnvironmentOrganisationList struct {
	ProjectID     string                               `json:"projectId"`
	EnvironmentID string                               `json:"environmentId"`
	Organisations []AuthProjectEnvironmentOrganisation `json:"organisations"`
}

// AuthProjectClientSyncRequest is the body of clients/reconcile. The expected
// revision is required and refused unless it matches, so a caller acting on an
// older read cannot register clients for configuration that has already changed.
type AuthProjectClientSyncRequest struct {
	ExpectedRevision int64 `json:"expectedRevision"`
}

// AuthProjectClientSyncRegistration is one audience's reconcile outcome. It
// reports what the provider was observed to hold, not what the caller asked for.
type AuthProjectClientSyncRegistration struct {
	Audience string `json:"audience"`
	ClientID string `json:"clientId"`
	Action   string `json:"action"`
	Reason   string `json:"reason,omitempty"`
}

// AuthProjectClientSyncReport is one reconcile's outcome. Superseded reports that
// the environment's configuration moved while this run was writing, so the result
// describes a revision that is no longer current.
type AuthProjectClientSyncReport struct {
	EnvironmentID string                              `json:"environmentId"`
	Revision      int64                               `json:"revision"`
	CheckedAt     string                              `json:"checkedAt"`
	ErrorCategory string                              `json:"errorCategory,omitempty"`
	Superseded    bool                                `json:"superseded"`
	Clients       []AuthProjectClientSyncRegistration `json:"clients"`
}

// GetEnvironmentCapability reads the environment's supported auth configuration.
func (c *AuthProjectAdminClient) GetEnvironmentCapability(
	ctx context.Context,
	projectID string,
	environmentID string,
	options AuthProjectRequestOptions,
) (*AuthProjectCapability, error) {
	var out AuthProjectCapability
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, authProjectEnvironmentPath(projectID, environmentID)+"/capability",
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDirectories reports every directory in the environment and whether the
// environment is fully stored.
func (c *AuthProjectAdminClient) ListDirectories(
	ctx context.Context,
	projectID string,
	environmentID string,
	options AuthProjectRequestOptions,
) (*AuthProjectDirectoryStatus, error) {
	var out AuthProjectDirectoryStatus
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, authProjectEnvironmentPath(projectID, environmentID)+"/directories",
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ActOnDirectory enables, pauses, resumes or releases one directory credential
// store. Custd makes it retry-safe per Idempotency-Key, so the key is required.
func (c *AuthProjectAdminClient) ActOnDirectory(
	ctx context.Context,
	projectID string,
	environmentID string,
	action AuthProjectDirectoryAction,
	req AuthProjectDirectoryRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectOperation, error) {
	if _, ok := authProjectDirectoryActions[action]; !ok {
		return nil, fmt.Errorf("custd: auth-project directory action %q is not one of enable, pause, resume or release", action)
	}
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectOperation
	path := authProjectEnvironmentPath(projectID, environmentID) + "/directories/" + url.PathEscape(string(action))
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, path, req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEnvironmentOrganisations reports the organisations one environment's
// directory holds. The read is environment-scoped and fails closed for a foreign
// or unknown environment rather than answering with an empty list.
func (c *AuthProjectAdminClient) ListEnvironmentOrganisations(
	ctx context.Context,
	projectID string,
	environmentID string,
	options AuthProjectRequestOptions,
) (*AuthProjectEnvironmentOrganisationList, error) {
	var out AuthProjectEnvironmentOrganisationList
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, authProjectEnvironmentPath(projectID, environmentID)+"/organisations",
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReconcileEnvironmentClients registers the environment's configured application
// audiences with the shared identity provider and removes registration the
// configuration no longer declares. Custd makes it retry-safe per
// Idempotency-Key, so the key is required; a run that loses a race with a
// configuration change reports Superseded so the caller reconciles again.
func (c *AuthProjectAdminClient) ReconcileEnvironmentClients(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectClientSyncRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectClientSyncReport, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectClientSyncReport
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/clients/reconcile",
		req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}
