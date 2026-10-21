package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	// AcceptedPrincipalID is the server-recorded acceptance linkage, never an
	// address match. Nil means unavailable, including historical accepted rows.
	AcceptedPrincipalID *string `json:"acceptedPrincipalId"`
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

// The two by-id invitation writes. Each addresses one invitation and changes
// it; the verb names the route and the answer's shape.
const (
	// AuthProjectInvitationVerbRevoke withdraws a live invitation.
	AuthProjectInvitationVerbRevoke = "revoke"
	// AuthProjectInvitationVerbRedeliver replaces one invitation's link.
	AuthProjectInvitationVerbRedeliver = "redeliver"
)

// AuthProjectInvitationByIDTarget addresses one invitation by id. The
// addressed project, environment and invitation id are the whole address: the
// body carries only the invitation id, and the retry identity travels in the
// request options.
type AuthProjectInvitationByIDTarget struct {
	// Verb is AuthProjectInvitationVerbRevoke or
	// AuthProjectInvitationVerbRedeliver.
	Verb          string
	ProjectID     string
	EnvironmentID string
	InvitationID  string
}

// AuthProjectInvitationByIDResult is the answer to one by-id invitation
// write.
//
// The two verbs answer with different shapes: a revocation answers with the
// withdrawn invitation, and a redelivery answers with the invitation and its
// fresh link. One envelope carries both, so InviteURL and Replayed are
// meaningful only for a redelivery.
type AuthProjectInvitationByIDResult struct {
	Invitation AuthProjectInvitation `json:"invitation"`
	InviteURL  string                `json:"inviteUrl,omitempty"`
	Replayed   bool                  `json:"replayed,omitempty"`
}

// InvitationByID sends one by-id invitation write: withdrawing a live
// invitation, or replacing its link so an address whose link never arrived can
// still be admitted.
//
// Custd makes both writes retry-safe per Idempotency-Key, so the key is
// required and a repeated attempt is answered as a replay rather than a second
// change.
func (c *AuthProjectAdminClient) InvitationByID(
	ctx context.Context,
	target AuthProjectInvitationByIDTarget,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitationByIDResult, ResponseMetadata, error) {
	if target.Verb != AuthProjectInvitationVerbRevoke && target.Verb != AuthProjectInvitationVerbRedeliver {
		return nil, ResponseMetadata{}, fmt.Errorf("custd: unknown auth project invitation verb %q", target.Verb)
	}
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, ResponseMetadata{}, err
	}
	path := authProjectEnvironmentPath(target.ProjectID, target.EnvironmentID) +
		"/invitations/" + url.PathEscape(target.Verb)
	if target.Verb == AuthProjectInvitationVerbRevoke {
		return c.revokeInvitationByID(ctx, path, target.InvitationID, options)
	}
	return c.redeliverInvitationByID(ctx, path, target.InvitationID, options)
}

// revokeInvitationByID withdraws one invitation and reports the withdrawn
// invitation the server returns.
func (c *AuthProjectAdminClient) revokeInvitationByID(
	ctx context.Context,
	path string,
	invitationID string,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitationByIDResult, ResponseMetadata, error) {
	var view AuthProjectInvitation
	req := AuthProjectInvitationRevocationRequest{InvitationID: invitationID}
	metadata, err := c.admin.requestWithMetadataHeaders(
		ctx, http.MethodPost, path, req, &view, authProjectHeaders(options),
	)
	if err != nil {
		return nil, metadata, err
	}
	return &AuthProjectInvitationByIDResult{Invitation: view}, metadata, nil
}

// redeliverInvitationByID replaces one invitation's link and reports the
// fresh link, which is carried here and nowhere else.
func (c *AuthProjectAdminClient) redeliverInvitationByID(
	ctx context.Context,
	path string,
	invitationID string,
	options AuthProjectRequestOptions,
) (*AuthProjectInvitationByIDResult, ResponseMetadata, error) {
	var issued AuthProjectInvitationIssued
	req := AuthProjectInvitationRedeliveryRequest{InvitationID: invitationID}
	metadata, err := c.admin.requestWithMetadataHeaders(
		ctx, http.MethodPost, path, req, &issued, authProjectHeaders(options),
	)
	if err != nil {
		return nil, metadata, err
	}
	return &AuthProjectInvitationByIDResult{
		Invitation: issued.Invitation,
		InviteURL:  issued.InviteURL,
		Replayed:   issued.Replayed,
	}, metadata, nil
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
