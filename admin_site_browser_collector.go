package custd

import (
	"context"
	"net/http"
)

// SiteBrowserCollectorMode specifies the desired browser collector state.
type SiteBrowserCollectorMode string

const (
	SiteBrowserCollectorEnabled  SiteBrowserCollectorMode = "enabled"
	SiteBrowserCollectorDisabled SiteBrowserCollectorMode = "disabled"
)

// SiteBrowserCollectorRequest converges the collector to a desired state.
// Environment is required when enabling the collector.
type SiteBrowserCollectorRequest struct {
	Mode        SiteBrowserCollectorMode `json:"mode"`
	Environment string                   `json:"environment,omitempty"`
}

// SiteBrowserCollector is the collector configuration for an owned Site.
// BrowserKey is omitted when disabled. Treat it as a write credential.
type SiteBrowserCollector struct {
	SiteUUID      string                   `json:"siteUuid"`
	Mode          SiteBrowserCollectorMode `json:"mode"`
	BrowserKey    string                   `json:"browserKey,omitempty"`
	Environment   string                   `json:"environment,omitempty"`
	AllowedOrigin string                   `json:"allowedOrigin,omitempty"`
}

// SetBrowserCollector enables or disables the collector of a domain the Site credential owns.
// Repeating an identical enable request returns the same browser key.
func (c *SiteDomainAdminClient) SetBrowserCollector(
	ctx context.Context, domainUUID string, request SiteBrowserCollectorRequest,
) (*SiteBrowserCollector, error) {
	var collector SiteBrowserCollector
	err := c.request(ctx, http.MethodPut, siteDomainPath(domainUUID)+"/browser-collector", request, &collector)
	return &collector, err
}
