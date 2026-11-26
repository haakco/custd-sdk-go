package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// AuthDomainMonitoringKeyHeader carries the external checker's monitoring key.
const AuthDomainMonitoringKeyHeader = "X-Custd-Auth-Domain-Monitoring-Key"

// AppDiscoverySignIn is a hosted app's browser sign-in registration. Pages use
// the authorization-code flow with PKCE and send the access token to the app
// item routes as X-Custd-User-Token.
type AppDiscoverySignIn struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	TokenEndpoint         string `json:"tokenEndpoint"`
	EndSessionEndpoint    string `json:"endSessionEndpoint"`
	ClientID              string `json:"clientId"`
	Audience              string `json:"audience"`
	RedirectURI           string `json:"redirectUri"`
	PostLogoutRedirectURI string `json:"postLogoutRedirectUri"`
}

// AppDiscoveryDocument is a hosted app's public /.well-known/custd.json: where
// its API is, its public Site key and what it can do.
type AppDiscoveryDocument struct {
	APIBaseURL            string              `json:"apiBaseUrl"`
	Host                  string              `json:"host"`
	SiteUUID              string              `json:"siteUuid"`
	BrowserKey            string              `json:"browserKey"`
	Environment           string              `json:"environment"`
	Capabilities          []string            `json:"capabilities"`
	SignIn                *AppDiscoverySignIn `json:"signIn,omitempty"`
	IdentityMode          string              `json:"identityMode"`
	AllowedOrigins        []string            `json:"allowedOrigins"`
	PageViewSchemaVersion string              `json:"pageViewSchemaVersion,omitempty"`
}

// AuthDomainMonitoringInventory is the realm hostnames the external checker
// probes. A host that leaves Hosts without appearing in RetiredHosts was lost,
// not retired.
type AuthDomainMonitoringInventory struct {
	Hosts        []string `json:"hosts"`
	RetiredHosts []string `json:"retiredHosts"`
}

// GetAppDiscovery reads a hosted app's public discovery document by host. It
// sends no credential; a host that serves no app answers 404.
func (c *CustdClient) GetAppDiscovery(ctx context.Context, host string) (*AppDiscoveryDocument, error) {
	var out AppDiscoveryDocument
	err := c.publicRead(ctx, "/api/v1/app-hosts/"+url.PathEscape(host)+"/discovery", nil, &out)
	return &out, err
}

// GetSharedSiteReport reads the site report a share link opens. The share
// token is the only authority, so no client credential is sent.
func (c *CustdClient) GetSharedSiteReport(ctx context.Context, token string) (*AdminSiteReport, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("custd: shared site report requires a share token")
	}
	var out AdminSiteReport
	err := c.publicRead(ctx, "/api/v1/shared-site-reports/"+url.PathEscape(token), nil, &out)
	return &out, err
}

// GetAuthDomainMonitoringInventory returns the realm hostnames the external
// checker may probe. The monitoring key is its only credential.
func (c *CustdClient) GetAuthDomainMonitoringInventory(
	ctx context.Context, monitoringKey string,
) (*AuthDomainMonitoringInventory, error) {
	if strings.TrimSpace(monitoringKey) == "" {
		return nil, fmt.Errorf("custd: auth domain monitoring requires a monitoring key")
	}
	var out AuthDomainMonitoringInventory
	headers := map[string]string{AuthDomainMonitoringKeyHeader: monitoringKey}
	err := c.publicRead(ctx, "/api/v1/auth-domain-monitoring/inventory", headers, &out)
	return &out, err
}

// publicRead issues a GET that carries none of the client's own credentials.
func (c *CustdClient) publicRead(ctx context.Context, path string, headers map[string]string, out any) error {
	request := map[string]string{"X-Custd-Sdk": sdkIdentity}
	for name, value := range headers {
		request[name] = value
	}
	response, err := c.requestApplication(ctx, &bufferedRequest{
		Method: http.MethodGet, URL: strings.TrimRight(c.config.BaseURL, "/") + path, Headers: request,
	})
	if err != nil {
		return err
	}
	if err := c.checkStatus(response.StatusCode, response.Body); err != nil {
		return err
	}
	if err := unmarshalResponse(response.Body, out); err != nil {
		return validationWrap("custd: decode public response", err)
	}
	return nil
}

// SiteStatusSubscriptionRequest asks a public status page to email updates to one address.
type SiteStatusSubscriptionRequest struct {
	Email string `json:"email"`
}

// SiteStatusSubscriptionToken is the single-use token from a status page subscription email.
type SiteStatusSubscriptionToken struct {
	Token string `json:"token"`
}

// GetPublicStatusPage reads a published status page by its public key. It
// sends no credential; an unpublished key answers 404.
func (c *CustdClient) GetPublicStatusPage(ctx context.Context, key string) (*AdminSiteStatusPage, error) {
	var out AdminSiteStatusPage
	err := c.publicRead(ctx, publicStatusPagePath(key), nil, &out)
	return &out, err
}

// SubscribeToStatusPage asks for a confirmation email. The server answers the
// same way for an unknown key, so success does not prove the page exists.
func (c *CustdClient) SubscribeToStatusPage(ctx context.Context, key string, request SiteStatusSubscriptionRequest) error {
	return c.publicWrite(ctx, publicStatusPagePath(key)+"/subscribe", request)
}

// ConfirmStatusPageSubscription confirms a subscription with the emailed token.
func (c *CustdClient) ConfirmStatusPageSubscription(ctx context.Context, key string, token SiteStatusSubscriptionToken) error {
	return c.publicWrite(ctx, publicStatusPagePath(key)+"/confirm-subscription", token)
}

// UnsubscribeFromStatusPage ends a subscription with the emailed token.
func (c *CustdClient) UnsubscribeFromStatusPage(ctx context.Context, key string, token SiteStatusSubscriptionToken) error {
	return c.publicWrite(ctx, publicStatusPagePath(key)+"/unsubscribe", token)
}

func publicStatusPagePath(key string) string {
	return "/api/v1/public-status-pages/" + url.PathEscape(key)
}

// publicWrite posts a JSON body to a route whose only authority is the body or
// path, answering 204 with no body.
func (c *CustdClient) publicWrite(ctx context.Context, path string, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("custd: encode public request: %w", err)
	}
	response, err := c.requestApplication(ctx, &bufferedRequest{
		Method: http.MethodPost, URL: strings.TrimRight(c.config.BaseURL, "/") + path,
		Headers: map[string]string{"Content-Type": "application/json", "X-Custd-Sdk": sdkIdentity},
		Body:    payload,
	})
	if err != nil {
		return err
	}
	return c.checkStatus(response.StatusCode, response.Body)
}
