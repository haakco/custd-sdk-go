package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	siteAIOwnerHandoffPath   = siteDomainsPath + "/ai-owner-handoffs"
	siteAIOwnerBrowserPath   = "/api/v1/site-ai-owner"
	siteAIOwnerCookieName    = "__Host-custd-site-ai-owner"
	siteAIOwnerBindingHeader = "X-Custd-Site-AI-Session"
)

// SiteAIOwnerHandoffRequest asks for a one-time handoff that lets one owner of
// a data space the calling broker manages change that tenant's AI settings.
type SiteAIOwnerHandoffRequest struct {
	TenantSlug   string `json:"tenantSlug"`
	OwnerSubject string `json:"ownerSubject"`
}

// SiteAIOwnerHandoff is the single-use code (valid for one minute) and the
// address it is redeemed at.
type SiteAIOwnerHandoff struct {
	Code      string `json:"code"`
	RedeemURL string `json:"redeemUrl"`
}

// SiteAIOwnerSettings is the owner session's tenant, its write binding and the
// tenant's current AI settings.
type SiteAIOwnerSettings struct {
	TenantSlug     string         `json:"tenantSlug"`
	SessionBinding string         `json:"sessionBinding"`
	Settings       SiteAISettings `json:"settings"`
}

// CreateAIOwnerHandoff issues a handoff for an approved data-space broker. The
// tenant comes from the request body, so no acting company is sent.
func (c *SiteDomainAdminClient) CreateAIOwnerHandoff(
	ctx context.Context, body SiteAIOwnerHandoffRequest,
) (*SiteAIOwnerHandoff, error) {
	var out SiteAIOwnerHandoff
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, siteAIOwnerHandoffPath, body, &out)
	return &out, err
}

// SiteAIOwnerSession is a signed-in AI settings session opened by redeeming a
// handoff. It carries the session cookie and write binding itself; it never
// sends the client's credential or an acting company.
type SiteAIOwnerSession struct {
	client  *CustdClient
	origin  string
	handle  string
	binding string
}

// RedeemSiteAIOwnerHandoff opens an owner session from a handoff. hubOrigin is
// the broker's registered hub origin the handoff was issued for.
func (c *CustdClient) RedeemSiteAIOwnerHandoff(
	ctx context.Context, handoff SiteAIOwnerHandoff, hubOrigin string,
) (*SiteAIOwnerSession, error) {
	origin, err := siteAIOwnerOrigin(handoff.RedeemURL)
	if err != nil {
		return nil, err
	}
	if handoff.Code == "" || hubOrigin == "" {
		return nil, fmt.Errorf("custd: AI owner handoff requires a code and the hub origin")
	}
	response, err := c.requestApplication(ctx, &bufferedRequest{
		Method: http.MethodPost, URL: handoff.RedeemURL,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded", "Origin": hubOrigin, "X-Custd-Sdk": sdkIdentity,
		},
		Body: []byte(url.Values{"code": {handoff.Code}}.Encode()),
	})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusSeeOther {
		if err := c.checkStatus(response.StatusCode, response.Body); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("custd: AI owner redeem returned HTTP %d, want 303", response.StatusCode)
	}
	cookie, err := http.ParseSetCookie(response.Header.Get("Set-Cookie"))
	if err != nil || cookie.Name != siteAIOwnerCookieName || cookie.Value == "" {
		return nil, fmt.Errorf("custd: AI owner redeem returned no session cookie")
	}
	return &SiteAIOwnerSession{client: c, origin: origin, handle: cookie.Value}, nil
}

// GetSettings reads the session tenant's AI settings and refreshes the write binding.
func (s *SiteAIOwnerSession) GetSettings(ctx context.Context) (*SiteAIOwnerSettings, error) {
	var out SiteAIOwnerSettings
	if err := s.send(ctx, http.MethodGet, "/settings", nil, &out); err != nil {
		return nil, err
	}
	s.binding = out.SessionBinding
	return &out, nil
}

// SetSettings replaces the session tenant's AI mode and cap.
func (s *SiteAIOwnerSession) SetSettings(ctx context.Context, body SiteAISettingsRequest) (*SiteAISettings, error) {
	return s.write(ctx, "/settings", body)
}

// SetConsent records or revokes the session tenant's AI terms consent.
func (s *SiteAIOwnerSession) SetConsent(ctx context.Context, body SiteAIConsentRequest) (*SiteAISettings, error) {
	return s.write(ctx, "/consent", body)
}

// write reads the binding first when the session has none, so a caller can
// write straight after redeeming.
func (s *SiteAIOwnerSession) write(ctx context.Context, path string, body any) (*SiteAISettings, error) {
	if s.binding == "" {
		if _, err := s.GetSettings(ctx); err != nil {
			return nil, err
		}
	}
	var out SiteAISettings
	if err := s.send(ctx, http.MethodPut, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SiteAIOwnerSession) send(ctx context.Context, method, path string, body, out any) error {
	headers := map[string]string{
		"Cookie":      (&http.Cookie{Name: siteAIOwnerCookieName, Value: s.handle}).String(),
		"X-Custd-Sdk": sdkIdentity,
	}
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("custd: encode AI owner request: %w", err)
		}
		payload = encoded
		headers["Content-Type"] = "application/json"
		headers["Origin"] = s.origin
		headers[siteAIOwnerBindingHeader] = s.binding
	}
	response, err := s.client.requestApplication(ctx, &bufferedRequest{
		Method: method, URL: s.origin + siteAIOwnerBrowserPath + path, Headers: headers, Body: payload,
	})
	if err != nil {
		return err
	}
	if err := s.client.checkStatus(response.StatusCode, response.Body); err != nil {
		return err
	}
	if err := unmarshalResponse(response.Body, out); err != nil {
		return validationWrap("custd: decode AI owner response", err)
	}
	return nil
}

func siteAIOwnerOrigin(redeemURL string) (string, error) {
	parsed, err := url.Parse(redeemURL)
	if err != nil || !isSecureOrLocalHTTP(redeemURL) || parsed.User != nil || parsed.RawQuery != "" ||
		parsed.Fragment != "" || parsed.Path != siteAIOwnerBrowserPath+"/redeem" {
		return "", fmt.Errorf("custd: AI owner handoff requires the Custd redeem URL")
	}
	return strings.TrimSuffix(parsed.Scheme+"://"+parsed.Host, "/"), nil
}
