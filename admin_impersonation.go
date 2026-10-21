package custd

import (
	"context"
	"net/http"
	"net/url"
)

// ImpersonationAdminClient owns impersonation sessions.
//
// An impersonation session lets an operator act as a named user for a bounded
// window, with the user's consent recorded. Consent is server-owned: a caller
// cannot start an unconsented session by asking, and must not assume one is
// live because the create call returned.
type ImpersonationAdminClient struct {
	admin *AdminClient
}

// ImpersonationSessionCreateRequest asks for a session. The starter is taken
// from the authenticated session, never from the request.
type ImpersonationSessionCreateRequest struct {
	CompanySlug    string `json:"companySlug"`
	TargetUserUUID string `json:"targetUserUuid"`
	Reason         string `json:"reason"`
	ExpiresAt      string `json:"expiresAt"`
}

// ImpersonationSession is one operator acting as one user.
type ImpersonationSession struct {
	SessionUUID     string  `json:"sessionUuid"`
	CompanySlug     string  `json:"companySlug"`
	TargetUserUUID  string  `json:"targetUserUuid"`
	TargetUserEmail string  `json:"targetUserEmail,omitempty"`
	Reason          string  `json:"reason"`
	ConsentState    string  `json:"consentState"`
	StartedAt       string  `json:"startedAt"`
	ExpiresAt       string  `json:"expiresAt"`
	StoppedAt       *string `json:"stoppedAt"`
	StopReason      string  `json:"stopReason,omitempty"`
}

// ImpersonationSessionListResponse lists the caller's sessions.
type ImpersonationSessionListResponse struct {
	Sessions []ImpersonationSession `json:"sessions"`
}

const impersonationSessionsPath = "/impersonation-sessions"

// Start asks for an impersonation session. Consent is server-owned, so a
// returned ConsentState is the answer, not a formality.
func (c *ImpersonationAdminClient) Start(
	ctx context.Context,
	req ImpersonationSessionCreateRequest,
) (*ImpersonationSession, ResponseMetadata, error) {
	var out ImpersonationSession
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, impersonationSessionsPath, req, &out,
	)
	return &out, metadata, err
}

// List returns the caller's impersonation sessions.
func (c *ImpersonationAdminClient) List(
	ctx context.Context,
) (*ImpersonationSessionListResponse, ResponseMetadata, error) {
	var out ImpersonationSessionListResponse
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, impersonationSessionsPath, nil, &out,
	)
	return &out, metadata, err
}

// Stop ends a session early. Stopping is forward-only; the server records when
// it happened and why.
func (c *ImpersonationAdminClient) Stop(
	ctx context.Context,
	sessionUUID string,
) (*ImpersonationSession, ResponseMetadata, error) {
	return c.sessionAction(ctx, sessionUUID, "/stop")
}

// Terminate ends a session as an operator action, distinct from the session
// holder stopping it. The server records the difference.
func (c *ImpersonationAdminClient) Terminate(
	ctx context.Context,
	sessionUUID string,
) (*ImpersonationSession, ResponseMetadata, error) {
	return c.sessionAction(ctx, sessionUUID, "/terminate")
}

func (c *ImpersonationAdminClient) sessionAction(
	ctx context.Context,
	sessionUUID string,
	action string,
) (*ImpersonationSession, ResponseMetadata, error) {
	var out ImpersonationSession
	path := impersonationSessionsPath + "/" + url.PathEscape(sessionUUID) + action
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, &out)
	return &out, metadata, err
}
