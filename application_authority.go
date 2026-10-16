package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ApplicationAuthorityRequest asks about one application access token, not a
// platform credential. Consumers must not log this value.
type ApplicationAuthorityRequest struct {
	AccessToken string `json:"accessToken"`
}

// ApplicationAuthorityDecision is an online decision, including HTTP-200 denial.
// It does not replace signature, issuer, carrier-audience, group or expiry checks.
type ApplicationAuthorityDecision struct {
	Authorized bool   `json:"authorized"`
	Reason     string `json:"reason"`
}

// CheckApplicationAuthority reads current authority using only the supplied
// application credential. It never mints or forwards the client's machine token.
// Consumers own the exact-token/realm binding and hard freshness deadline.
func (c *CustdClient) CheckApplicationAuthority(ctx context.Context, request ApplicationAuthorityRequest) (ApplicationAuthorityDecision, error) {
	var denied ApplicationAuthorityDecision
	if strings.TrimSpace(request.AccessToken) == "" {
		return denied, fmt.Errorf("custd: application access token is required")
	}
	if !isSecureOrLocalHTTP(c.config.BaseURL) {
		return denied, fmt.Errorf("custd: base URL must use https unless it targets localhost")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return denied, fmt.Errorf("custd: encode application authority request: %w", err)
	}
	response, err := c.requestApplicationAuthority(ctx, &HTTPRequest{
		Method: http.MethodPost,
		URL:    strings.TrimRight(c.config.BaseURL, "/") + "/oauth/authority-check",
		Headers: map[string]string{
			"Content-Type": "application/json", "X-Custd-Sdk": sdkIdentity,
			"Authorization": "Bearer " + request.AccessToken,
		},
		Body: body,
	})
	if err != nil {
		return denied, err
	}
	if err := c.checkStatus(response.StatusCode, response.Body); err != nil {
		return denied, err
	}
	var decision struct {
		Authorized *bool   `json:"authorized"`
		Reason     *string `json:"reason"`
	}
	if err := json.Unmarshal(response.Body, &decision); err != nil {
		return denied, fmt.Errorf("custd: decode application authority decision: %w", err)
	}
	if decision.Authorized == nil || decision.Reason == nil {
		return denied, fmt.Errorf("custd: application authority decision requires authorized and reason")
	}
	return ApplicationAuthorityDecision{Authorized: *decision.Authorized, Reason: *decision.Reason}, nil
}

func (c *CustdClient) requestApplicationAuthority(ctx context.Context, request *HTTPRequest) (*HTTPResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.config.HTTPClient != nil {
		return c.config.HTTPClient.Do(request)
	}
	req, err := http.NewRequestWithContext(ctx, request.Method, request.URL, bytes.NewReader(request.Body))
	if err != nil {
		return nil, fmt.Errorf("custd: create application authority request: %w", err)
	}
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	// Do not use c.httpClient: its OAuth transport may mint and replace the
	// application Authorization header with a platform machine credential.
	client := &http.Client{Timeout: defaultHTTPTimeout}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custd: application authority request failed: %w", err)
	}
	body, err := readResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("custd: read application authority decision: %w", err)
	}
	return &HTTPResponse{StatusCode: response.StatusCode, Body: body}, nil
}
