package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Account is the authenticated caller's own account, as the API publishes it.
type Account struct {
	CompanySlug          string   `json:"companySlug"`
	EffectiveCompanySlug string   `json:"effectiveCompanySlug,omitempty"`
	Email                string   `json:"email"`
	Roles                []string `json:"roles"`
	Permissions          []string `json:"permissions"`
}

// AuthorizationServerMetadata is the trusted discovery subset an OAuth
// authorization server publishes.
type AuthorizationServerMetadata struct {
	Issuer                      string `json:"issuer"`
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	JWKSURI                     string `json:"jwks_uri"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
}

// JWK is one JSON Web Key an issuer publishes.
type JWK struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
}

// JWKS is the set of signing keys an issuer publishes.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// IssuerInspection is an issuer's discovery document together with the keys
// it publishes.
type IssuerInspection struct {
	Metadata AuthorizationServerMetadata `json:"metadata"`
	JWKS     JWKS                        `json:"jwks"`
}

// GetHealth reads the API's liveness answer.
//
// The route is the root of the configured base URL, not a path under
// /api/v1: the liveness probe is served next to the API root and is public.
// It therefore carries no credential, so a caller can tell "the API is
// unreachable" apart from "this caller is not authorized".
func (c *AdminClient) GetHealth(ctx context.Context) (*Health, ResponseMetadata, error) {
	if !isSecureOrLocalHTTP(c.client.config.BaseURL) {
		return nil, ResponseMetadata{}, errors.New("custd: base URL must use https unless it targets localhost")
	}
	url := strings.TrimRight(c.client.config.BaseURL, "/") + "/health"
	headers := map[string]string{"Accept": "application/json", "X-Custd-Sdk": sdkIdentity}
	metadata, body, err := c.requestOutsideAdmin(ctx, http.MethodGet, url, headers, nil)
	if err != nil {
		return nil, metadata, err
	}
	health := &Health{}
	if len(bytes.TrimSpace(body)) > 0 {
		if err := decodeAdminResponse(body, health); err != nil {
			return nil, metadata, err
		}
	} else {
		// The liveness route answers 200 with an empty body, so an empty body
		// is the healthy answer, as the operator CLI reports it.
		health.Status = "ok"
	}
	return health, metadata, nil
}

// GetAccount reads the authenticated caller's account document.
//
// The route is /api/v1/account/me, which is NOT under /api/v1/admin, so it is
// requested by its full URL instead of through the admin-relative path helper.
func (c *AdminClient) GetAccount(ctx context.Context) (*Account, ResponseMetadata, error) {
	if !isSecureOrLocalHTTP(c.client.config.BaseURL) {
		return nil, ResponseMetadata{}, errors.New("custd: base URL must use https unless it targets localhost")
	}
	url := strings.TrimRight(c.client.config.BaseURL, "/") + "/api/v1/account/me"
	headers := c.client.headers(false)
	c.applyActingCompany(headers)
	metadata, body, err := c.requestOutsideAdmin(ctx, http.MethodGet, url, headers, nil)
	if err != nil {
		return nil, metadata, err
	}
	var out Account
	if err := decodeAdminResponse(body, &out); err != nil {
		return nil, metadata, err
	}
	return &out, metadata, nil
}

// InspectIssuer reads an OAuth authorization server's discovery document and
// the signing keys it publishes.
//
// Both routes come from the issuer itself: the discovery document at
// /.well-known/oauth-authorization-server names the key set to read. The Custd
// credential is never sent to an issuer.
func (c *AdminClient) InspectIssuer(ctx context.Context, issuerURL string) (*IssuerInspection, error) {
	issuer, err := trustedIssuerURL(issuerURL)
	if err != nil {
		return nil, err
	}
	discoveryURL := issuer + "/.well-known/oauth-authorization-server"
	var metadata AuthorizationServerMetadata
	if err := c.fetchIssuerDocument(ctx, discoveryURL, &metadata); err != nil {
		return nil, err
	}
	if metadata.Issuer != issuer {
		return nil, errors.New("custd: issuer metadata does not match the requested issuer")
	}
	keysURL, err := trustedIssuerEndpoint(metadata.JWKSURI, issuer)
	if err != nil {
		return nil, err
	}
	var keys JWKS
	if err := c.fetchIssuerDocument(ctx, keysURL, &keys); err != nil {
		return nil, err
	}
	if len(keys.Keys) == 0 {
		return nil, errors.New("custd: issuer published no signing keys")
	}
	return &IssuerInspection{Metadata: metadata, JWKS: keys}, nil
}

// trustedIssuerURL accepts an issuer origin the SDK is willing to talk to and
// returns it without a trailing slash.
func trustedIssuerURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return "", errors.New("custd: issuer must be a trusted HTTPS origin")
	}
	if !isSecureOrLocalHTTP(parsed.String()) {
		return "", errors.New("custd: issuer must be a trusted HTTPS origin")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

// trustedIssuerEndpoint accepts an issuer-published endpoint only when it stays
// on the issuer's own origin, so a hostile discovery document cannot point the
// key fetch somewhere else.
func trustedIssuerEndpoint(raw, issuer string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return "", errors.New("custd: issuer endpoint must be an absolute URL on the issuer origin")
	}
	if !isSecureOrLocalHTTP(parsed.String()) {
		return "", errors.New("custd: issuer endpoint must use https unless it targets localhost")
	}
	origin, err := url.Parse(issuer)
	if err != nil || origin.Host == "" {
		return "", errors.New("custd: issuer must be a trusted HTTPS origin")
	}
	// Compare the whole origin, not just the scheme. A discovery document is untrusted input,
	// so an endpoint on a different host would let it redirect the key fetch to keys of its own
	// choosing while still looking like a well-formed https URL.
	if !strings.EqualFold(parsed.Scheme, origin.Scheme) || !strings.EqualFold(parsed.Host, origin.Host) {
		return "", errors.New("custd: issuer endpoint must be on the issuer origin")
	}
	return parsed.String(), nil
}

// fetchIssuerDocument reads one JSON document from an issuer.
func (c *AdminClient) fetchIssuerDocument(ctx context.Context, rawURL string, out any) error {
	if c.client.config.HTTPClient != nil {
		response, err := c.client.config.HTTPClient.Do(&HTTPRequest{
			Method:  http.MethodGet,
			URL:     rawURL,
			Headers: map[string]string{"Accept": "application/json"},
		})
		if err != nil {
			return fmt.Errorf("custd: issuer request failed: %w", err)
		}
		if err := c.client.checkStatus(response.StatusCode, response.Body); err != nil {
			return err
		}
		return decodeAdminResponse(response.Body, out)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("custd: create issuer request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	// The issuer is a third party: the request carries no Custd credential, so
	// the platform's own client is not used for it.
	client := &http.Client{Timeout: defaultHTTPTimeout, CheckRedirect: sameOriginRedirect}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("custd: issuer request failed: %w", err)
	}
	body, err := readResponseBody(response.Body)
	if err != nil {
		return fmt.Errorf("custd: read issuer response: %w", err)
	}
	if err := c.client.checkStatus(response.StatusCode, body); err != nil {
		return err
	}
	return decodeAdminResponse(body, out)
}

// sameOriginRedirect refuses a redirect that leaves the origin the request
// started from.
func sameOriginRedirect(request *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	if !strings.EqualFold(request.URL.Scheme, via[0].URL.Scheme) ||
		!strings.EqualFold(request.URL.Host, via[0].URL.Host) {
		return errors.New("custd: issuer request redirected to another origin")
	}
	return nil
}

// requestOutsideAdmin issues one request to a URL outside the /api/v1/admin
// namespace and returns the server's request identifier with the raw answer.
//
// The admin request helper always prefixes /api/v1/admin, so the routes that
// live elsewhere (the public liveness probe and /api/v1/account/me) are
// requested by their full URL instead.
func (c *AdminClient) requestOutsideAdmin(
	ctx context.Context, method, url string, headers map[string]string, payload any,
) (ResponseMetadata, []byte, error) {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return ResponseMetadata{}, nil, fmt.Errorf("custd: marshal request: %w", err)
		}
	}
	if c.client.config.HTTPClient != nil {
		return c.requestOutsideAdminViaDoer(ctx, method, url, headers, body)
	}
	return c.requestOutsideAdminViaHTTP(ctx, method, url, headers, body)
}

// requestOutsideAdminViaDoer sends the request through the configured transport.
func (c *AdminClient) requestOutsideAdminViaDoer(
	ctx context.Context, method, url string, headers map[string]string, body []byte,
) (ResponseMetadata, []byte, error) {
	response, err := c.client.config.HTTPClient.Do(&HTTPRequest{
		Method: method, URL: url, Headers: headers, Body: body,
	})
	if err != nil {
		return ResponseMetadata{}, nil, fmt.Errorf("custd: request failed: %w", err)
	}
	metadata := ResponseMetadata{RequestID: responseHeader(response.Headers, requestIDHeader)}
	recordMetadata(ctx, metadata.RequestID)
	if err := c.client.checkStatus(response.StatusCode, response.Body); err != nil {
		return metadata, nil, err
	}
	return metadata, response.Body, nil
}

// requestOutsideAdminViaHTTP sends the request through the SDK's own client.
func (c *AdminClient) requestOutsideAdminViaHTTP(
	ctx context.Context, method, url string, headers map[string]string, body []byte,
) (ResponseMetadata, []byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return ResponseMetadata{}, nil, fmt.Errorf("custd: create request: %w", err)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := c.client.httpClient.Do(request)
	if err != nil {
		return ResponseMetadata{}, nil, fmt.Errorf("custd: request failed: %w", err)
	}
	responseBody, err := readResponseBody(response.Body)
	if err != nil {
		return ResponseMetadata{}, nil, fmt.Errorf("custd: read response: %w", err)
	}
	metadata := ResponseMetadata{RequestID: response.Header.Get(requestIDHeader)}
	recordMetadata(ctx, metadata.RequestID)
	if err := c.client.checkStatus(response.StatusCode, responseBody); err != nil {
		return metadata, nil, err
	}
	return metadata, responseBody, nil
}
