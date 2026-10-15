package custd

import (
	"context"
	"net/http"
)

// AuthProjectInvitation is one invitation as a caller sees it. It never carries
// the link token or its hash. Delivered is only meaningful on the response to
// issuing an invitation: nothing stores it, so a list cannot report it.
type AuthProjectInvitation struct {
	InvitationID   string `json:"invitationId"`
	Email          string `json:"email"`
	MembershipRole string `json:"membershipRole"`
	PoolID         string `json:"poolId"`
	OrganisationID string `json:"organisationId"`
	ExpiresAt      string `json:"expiresAt"`
	CreatedAt      string `json:"createdAt"`
	// State is the one word an operator needs: open, accepted, revoked or expired.
	State string `json:"state"`
	// Delivered reports whether the link reached the delivery owner. It is only
	// meaningful on the response to issuing an invitation and absent elsewhere.
	Delivered bool `json:"delivered,omitempty"`
}

// AuthProjectInvitationIssueRequest invites one address into a directory. It
// cannot name a project, environment, token or state: the path resolves the
// scope, the configuration resolves the directory, and the token is the
// server's own randomness.
type AuthProjectInvitationIssueRequest struct {
	PoolID           string `json:"poolId"`
	OrganisationID   string `json:"organisationId"`
	Email            string `json:"email"`
	MembershipRole   string `json:"membershipRole"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// AuthProjectInvitationIssued is the response to issuing one invitation.
// InviteURL is returned exactly once, here: only the token's hash is stored, so a
// later read cannot recover it.
type AuthProjectInvitationIssued struct {
	Invitation AuthProjectInvitation `json:"invitation"`
	InviteURL  string                `json:"inviteUrl"`
	Replayed   bool                  `json:"replayed"`
}

// AuthProjectInvitationListResponse is an environment's invitations, newest first.
type AuthProjectInvitationListResponse struct {
	Invitations []AuthProjectInvitation `json:"invitations"`
}

// AuthProjectInvitationRevocationRequest addresses one invitation to withdraw.
type AuthProjectInvitationRevocationRequest struct {
	InvitationID string `json:"invitationId"`
}

// AuthProjectInvitationRedeliveryRequest names one invitation whose link the
// caller wants delivered again. It cannot carry a link: the server mints a fresh
// one, so a caller never chooses where a person is sent.
type AuthProjectInvitationRedeliveryRequest struct {
	InvitationID string `json:"invitationId"`
}

// ListInvitations reports the environment's invitations.
func (c *AuthProjectAdminClient) ListInvitations(
	ctx context.Context,
	projectID string,
	environmentID string,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitationListResponse, error) {
	var out AuthProjectInvitationListResponse
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet, authProjectEnvironmentPath(projectID, environmentID)+"/invitations",
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// IssueInvitation records one invitation and returns its link. Custd makes it
// retry-safe per Idempotency-Key, so the key is required; a retry of the same
// request is answered from the recorded invitation rather than minting a second
// link for the same address.
func (c *AuthProjectAdminClient) IssueInvitation(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectInvitationIssueRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitationIssued, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectInvitationIssued
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/invitations",
		req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokeInvitation withdraws a live invitation before it is used. Custd makes it
// retry-safe per Idempotency-Key, so the key is required.
func (c *AuthProjectAdminClient) RevokeInvitation(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectInvitationRevocationRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitation, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectInvitation
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/invitations/revoke",
		req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// RedeliverInvitation replaces one live invitation's link on the same invitation
// and returns the fresh link. Custd makes it retry-safe per Idempotency-Key, so
// the key is required; repeating the same key reports Replayed and carries no
// link, because the first attempt's link is the live one.
func (c *AuthProjectAdminClient) RedeliverInvitation(
	ctx context.Context,
	projectID string,
	environmentID string,
	req AuthProjectInvitationRedeliveryRequest,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitationIssued, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AuthProjectInvitationIssued
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/invitations/redeliver",
		req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}
