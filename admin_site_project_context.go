package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

const siteProjectContextPath = siteDomainsPath + "/project-context"

// AdminSiteMarket is one market: an ISO 3166-1 alpha-2 country such as ZA and an ISO 639
// language such as en.
type AdminSiteMarket struct {
	Country  string `json:"country"`
	Language string `json:"language"`
}

// AdminSiteProjectContext is one Site's brand names, named competitors and markets, shared by
// every domain of the Site. UpdatedAt is nil until it is first set.
type AdminSiteProjectContext struct {
	SiteUUID          string            `json:"siteUuid"`
	BrandNames        []string          `json:"brandNames"`
	CompetitorDomains []string          `json:"competitorDomains"`
	Markets           []AdminSiteMarket `json:"markets"`
	HackerNewsEnabled bool              `json:"hackerNewsEnabled,omitempty"`
	UpdatedAt         *time.Time        `json:"updatedAt,omitempty"`
}

// AdminSiteProjectContextRequest replaces a Site's project context: up to 5 brand names, 10
// competitor DNS names and 10 markets. Nil lists are sent as empty, clearing them.
type AdminSiteProjectContextRequest struct {
	// Nil preserves the saved choice. A pointer to false disables technical discussion trends.
	HackerNewsEnabled *bool             `json:"hackerNewsEnabled,omitempty"`
	BrandNames        []string          `json:"brandNames"`
	CompetitorDomains []string          `json:"competitorDomains"`
	Markets           []AdminSiteMarket `json:"markets"`
}

// GetProjectContext returns a Site's project context. A site-bound credential may pass an empty
// siteUUID for its own Site.
func (c *SiteDomainAdminClient) GetProjectContext(ctx context.Context, siteUUID string) (*AdminSiteProjectContext, error) {
	var project AdminSiteProjectContext
	err := c.request(ctx, http.MethodGet, siteProjectContextPathFor(siteUUID), nil, &project)
	return &project, err
}

// SetProjectContext replaces a Site's project context. A site-bound credential may pass an empty
// siteUUID for its own Site.
func (c *SiteDomainAdminClient) SetProjectContext(
	ctx context.Context, siteUUID string, req AdminSiteProjectContextRequest,
) (*AdminSiteProjectContext, error) {
	body := AdminSiteProjectContextRequest{
		HackerNewsEnabled: req.HackerNewsEnabled,
		BrandNames:        nonNilSlice(req.BrandNames),
		CompetitorDomains: nonNilSlice(req.CompetitorDomains),
		Markets:           nonNilSlice(req.Markets),
	}
	var project AdminSiteProjectContext
	err := c.request(ctx, http.MethodPut, siteProjectContextPathFor(siteUUID), body, &project)
	return &project, err
}

func nonNilSlice[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}

func siteProjectContextPathFor(siteUUID string) string {
	if siteUUID == "" {
		return siteProjectContextPath
	}
	return siteProjectContextPath + "?" + url.Values{"siteUuid": {siteUUID}}.Encode()
}
