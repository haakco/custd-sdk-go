package custd

import (
	"context"
	"net/http"
)

// ApplicationPrincipalTarget addresses one application identity inside a directory.
// Custd validates the project, environment, pool and provider mapping before acting.
type ApplicationPrincipalTarget struct {
	ProjectID       string
	EnvironmentID   string
	DirectoryID     string
	ProviderSubject string
}

// ApplicationPrincipalSuspension is the server's suspension or restoration receipt.
type ApplicationPrincipalSuspension struct {
	ProjectID       string `json:"projectId"`
	EnvironmentID   string `json:"environmentId"`
	DirectoryID     string `json:"directoryId"`
	PrincipalID     string `json:"principalId"`
	Enabled         bool   `json:"enabled"`
	SessionsRevoked int64  `json:"sessionsRevoked"`
	Revision        int64  `json:"revision"`
	Replayed        bool   `json:"replayed"`
}

// ApplicationPrincipalErasureRequest requires explicit confirmation of terminal erasure.
type ApplicationPrincipalErasureRequest struct {
	Confirm bool `json:"confirm"`
}

// ApplicationPrincipalErasure reports the operation outcome and actual deletion counts.
// CountsRecorded is false when a replay carries status but not the original counts.
type ApplicationPrincipalErasure struct {
	ProjectID      string `json:"projectId"`
	EnvironmentID  string `json:"environmentId"`
	DirectoryID    string `json:"directoryId"`
	PrincipalID    string `json:"principalId"`
	Status         string `json:"status"`
	Invitations    int64  `json:"invitations"`
	MagicLinks     int64  `json:"magicLinks"`
	Memberships    int64  `json:"memberships"`
	Mappings       int64  `json:"mappings"`
	ProfileValues  int64  `json:"profileValues"`
	Principals     int64  `json:"principals"`
	Revision       int64  `json:"revision"`
	Replayed       bool   `json:"replayed"`
	CountsRecorded bool   `json:"countsRecorded"`
}

// ErasePrincipal invokes the server-owned durable erasure of the full principal mapping set.
func (c *AuthProjectAdminClient) ErasePrincipal(
	ctx context.Context, target ApplicationPrincipalTarget, body ApplicationPrincipalErasureRequest, options AuthProjectRequestOptions,
) (*ApplicationPrincipalErasure, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out ApplicationPrincipalErasure
	if err := c.admin.requestWithHeaders(ctx, http.MethodPost, target.path()+"/erase", body, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

func (target ApplicationPrincipalTarget) path() string {
	return applicationPrincipalPath(target.ProjectID, target.EnvironmentID, target.DirectoryID, target.ProviderSubject)
}

// SuspendPrincipal withdraws authority and revokes every principal-owned identity's sessions.
func (c *AuthProjectAdminClient) SuspendPrincipal(
	ctx context.Context, target ApplicationPrincipalTarget, options AuthProjectRequestOptions,
) (*ApplicationPrincipalSuspension, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out ApplicationPrincipalSuspension
	if err := c.admin.requestWithHeaders(ctx, http.MethodPost, target.path()+"/suspend", nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// RestorePrincipal restores authority without changing factors; pending withdrawals are refused.
func (c *AuthProjectAdminClient) RestorePrincipal(
	ctx context.Context, target ApplicationPrincipalTarget, options AuthProjectRequestOptions,
) (*ApplicationPrincipalSuspension, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out ApplicationPrincipalSuspension
	if err := c.admin.requestWithHeaders(ctx, http.MethodPost, target.path()+"/restore", nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}
