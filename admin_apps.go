package custd

import (
	"context"
	"net/http"
	"net/url"
)

// An app is a hosted application served at <slug>.<apps domain>: an auth
// project whose production environment is hosted. It is addressed by its slug,
// which comes from its name. An app the caller cannot see answers
// app_not_found, exactly like one that does not exist. Releases are deployed
// through the environment's release routes (ListAppReleases and friends).

// AppStatus is an app's one status. A list shows the worst first.
type AppStatus string

const (
	AppStatusNeedsAttention    AppStatus = "needs_attention"
	AppStatusRemovalUnfinished AppStatus = "removal_unfinished"
	AppStatusLive              AppStatus = "live"
	AppStatusNotDeployed       AppStatus = "not_deployed"
	AppStatusPaused            AppStatus = "paused"
	AppStatusOffboarded        AppStatus = "offboarded"
)

// AppReasonCode explains a needs_attention status. The set is closed.
type AppReasonCode string

const (
	AppReasonSiteRegistrationIncomplete AppReasonCode = "site_registration_incomplete"
	AppReasonSignInNotReady             AppReasonCode = "sign_in_not_ready"
)

// AppTrafficReporting says whether a created app's traffic reports are published.
type AppTrafficReporting string

const (
	AppTrafficReportingReady       AppTrafficReporting = "ready"
	AppTrafficReportingUnavailable AppTrafficReporting = "unavailable"
)

// AppCreateRequest is a new app: its name, from which the address is derived,
// and whether people sign in to it. Sign-in starts invite-only.
type AppCreateRequest struct {
	Name   string `json:"name"`
	SignIn bool   `json:"signIn,omitempty"`
}

// AppReason is why an app needs attention.
type AppReason struct {
	Code    AppReasonCode `json:"code"`
	Message string        `json:"message"`
}

// AppSummary is one app as the list shows it.
type AppSummary struct {
	AppID           string     `json:"appId"`
	Slug            string     `json:"slug"`
	Name            string     `json:"name"`
	EnvironmentID   string     `json:"environmentId"`
	Host            string     `json:"host"`
	URL             string     `json:"url"`
	SiteUUID        string     `json:"siteUuid,omitempty"`
	Status          AppStatus  `json:"status"`
	Reason          *AppReason `json:"reason,omitempty"`
	SignIn          bool       `json:"signIn"`
	ActiveReleaseID string     `json:"activeReleaseId,omitempty"`
	ServingSince    string     `json:"servingSince,omitempty"`
}

// AppCreation is the receipt for one create. Replayed is true when the same
// Idempotency-Key had already created the app.
type AppCreation struct {
	App              AppSummary          `json:"app"`
	Replayed         bool                `json:"replayed"`
	TrafficReporting AppTrafficReporting `json:"trafficReporting"`
}

// AppListResponse is one page of the caller's apps, worst status first. Pass
// NextAfter as the next call's after.
type AppListResponse struct {
	Apps      []AppSummary `json:"apps"`
	NextAfter string       `json:"nextAfter,omitempty"`
}

// AppOverview is one app with its latest release, how many releases it keeps
// and how many live owner and share links its items have.
type AppOverview struct {
	App           AppSummary         `json:"app"`
	LatestRelease *AppReleaseSummary `json:"latestRelease,omitempty"`
	ReleaseCount  int                `json:"releaseCount"`
	OwnerLinks    int                `json:"ownerLinks"`
	ShareLinks    int                `json:"shareLinks"`
}

// AppDeployCredential is a CI credential that can deploy only this app. The
// secret is in this response only; hand it to the caller's secret store.
type AppDeployCredential struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	App          string `json:"app"`
	Scope        string `json:"scope"`
}

const appsPath = "/apps"

// CreateApp creates an app with hosting, traffic, error capture and app data
// on; SignIn also turns on invite-only sign-in. Repeating the same body and
// Idempotency-Key finishes a create that did not complete
// (app_setup_incomplete) or reads the same app back, so the key is required.
func (c *AuthProjectAdminClient) CreateApp(
	ctx context.Context, req AppCreateRequest, options AuthProjectRequestOptions,
) (*AppCreation, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AppCreation
	if err := c.admin.requestWithHeaders(ctx, http.MethodPost, appsPath, req, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListApps returns one page of the caller's apps. An empty after reads the first page.
func (c *AuthProjectAdminClient) ListApps(
	ctx context.Context, after string, options AuthProjectRequestOptions,
) (*AppListResponse, error) {
	path := appsPath
	if after != "" {
		path += "?after=" + url.QueryEscape(after)
	}
	var out AppListResponse
	if err := c.admin.requestWithHeaders(ctx, http.MethodGet, path, nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetApp reads one app by its slug.
func (c *AuthProjectAdminClient) GetApp(
	ctx context.Context, slug string, options AuthProjectRequestOptions,
) (*AppOverview, error) {
	var out AppOverview
	if err := c.admin.requestWithHeaders(ctx, http.MethodGet, appPath(slug), nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAppDeployCredential creates a client-credentials OAuth client holding
// apps.releases.write that reaches only this app's release routes and its
// read, acting for the person who created it.
func (c *AuthProjectAdminClient) CreateAppDeployCredential(
	ctx context.Context, slug string, options AuthProjectRequestOptions,
) (*AppDeployCredential, error) {
	var out AppDeployCredential
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, appPath(slug)+"/deploy-credentials", nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAppDeployCredential removes a deploy credential. Repeating a delete
// that failed part-way finishes it.
func (c *AuthProjectAdminClient) DeleteAppDeployCredential(
	ctx context.Context, slug, clientID string, options AuthProjectRequestOptions,
) error {
	path := appPath(slug) + "/deploy-credentials/" + url.PathEscape(clientID)
	return c.admin.requestWithHeaders(ctx, http.MethodDelete, path, nil, nil, authProjectHeaders(options))
}

func appPath(slug string) string {
	return appsPath + "/" + url.PathEscape(slug)
}
