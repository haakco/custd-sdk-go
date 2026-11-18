package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteRegionalTrendMatch is public feed evidence that matched a saved Site term.
type AdminSiteRegionalTrendMatch struct {
	Topic         string    `json:"topic"`
	SourceURL     string    `json:"sourceUrl"`
	PublishedAt   time.Time `json:"publishedAt"`
	DiscussionURL string    `json:"discussionUrl,omitempty"`
	MatchedTerm   string    `json:"matchedTerm"`
}

// AdminSiteRegionalTrendMarket is the evidence for one configured Site market.
type AdminSiteRegionalTrendMarket struct {
	Supported bool                          `json:"supported"`
	Country   string                        `json:"country"`
	CheckedAt *time.Time                    `json:"checkedAt,omitempty"`
	Stale     bool                          `json:"stale"`
	Matches   []AdminSiteRegionalTrendMatch `json:"matches"`
}

// AdminSiteTechnicalTrends is global technical discussion evidence for a Site that opted in.
type AdminSiteTechnicalTrends struct {
	Source    string                        `json:"source"`
	CheckedAt *time.Time                    `json:"checkedAt,omitempty"`
	Stale     bool                          `json:"stale"`
	Matches   []AdminSiteRegionalTrendMatch `json:"matches"`
}

// AdminSiteRegionalTrends is the Site's public regional and optional technical discussion evidence.
type AdminSiteRegionalTrends struct {
	SiteUUID      string                         `json:"siteUuid"`
	Source        string                         `json:"source"`
	HasSavedTerms bool                           `json:"hasSavedTerms"`
	Markets       []AdminSiteRegionalTrendMarket `json:"markets"`
	Technical     *AdminSiteTechnicalTrends      `json:"technical,omitempty"`
}

// GetRegionalTrends reads public trends matching a Site's declared audience and saved terms.
// A site-bound credential may pass an empty siteUUID for its own Site.
func (c *SiteDomainAdminClient) GetRegionalTrends(ctx context.Context, siteUUID string) (*AdminSiteRegionalTrends, error) {
	path := siteDomainsPath + "/regional-trends"
	if siteUUID != "" {
		path += "?" + url.Values{"siteUuid": {siteUUID}}.Encode()
	}
	var trends AdminSiteRegionalTrends
	err := c.request(ctx, http.MethodGet, path, nil, &trends)
	return &trends, err
}
