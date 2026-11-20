package custd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type siteMonitorDestination struct {
	BaseURL    string
	ReportPath string
	WriteToken string
}

func submitSiteMonitorRequest(ctx context.Context, destination siteMonitorDestination, prefix, method string, body []byte) (*http.Response, error) {
	endpoint, err := destination.endpoint(prefix)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("custd: build monitor request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+destination.WriteToken)
	req.Header.Set("X-Custd-Sdk", sdkIdentity)
	client := &http.Client{
		Timeout:       defaultHTTPTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custd: send monitor request: %w", err)
	}
	return response, nil
}

func (destination siteMonitorDestination) endpoint(prefix string) (string, error) {
	base, err := url.Parse(destination.BaseURL)
	if err != nil || base.Hostname() == "" || base.User != nil || base.RawQuery != "" ||
		base.ForceQuery || base.Fragment != "" || (base.Path != "" && base.Path != "/") ||
		!isSecureOrLocalHTTP(destination.BaseURL) {
		return "", fmt.Errorf("custd: site monitor requires an HTTPS API origin")
	}
	suffix, prefixed := strings.CutPrefix(destination.ReportPath, prefix)
	tenant, monitor, split := strings.Cut(suffix, "/")
	if !prefixed || !split || !validSiteMonitorSegment(tenant) ||
		!validSiteMonitorSegment(monitor) || len(destination.WriteToken) != 64 {
		return "", fmt.Errorf("custd: site monitor requires a report path and write credential")
	}
	return strings.TrimRight(destination.BaseURL, "/") + destination.ReportPath, nil
}

func validSiteMonitorSegment(segment string) bool {
	decoded, err := url.PathUnescape(segment)
	return err == nil && decoded != "" && decoded != "." && decoded != ".." &&
		!strings.ContainsAny(decoded, "/\\?#")
}
