package custd

import (
	"context"
	"net/http"
	"net/url"
)

// AdminSiteBacklinkSource is a source of referring-domain data.
type AdminSiteBacklinkSource string

// The sources a referring domain can come from (AdminSiteReferringDomain.Sources and
// AdminSiteLinkSourceTotal.Source).
const (
	AdminSiteLinkSourceBing        AdminSiteBacklinkSource = "bing"
	AdminSiteLinkSourceCommonCrawl AdminSiteBacklinkSource = "commoncrawl"
	AdminSiteLinkSourceCrawl       AdminSiteBacklinkSource = "crawl"
)

// AdminSiteLinkStatus is whether a link source has read a domain.
type AdminSiteLinkStatus string

// The states of AdminSiteLinkSourceTotal.Status and AdminSiteLostLinks.Status. Unavailable means
// the source cannot say anything for the domain and pending that it has not read yet; neither is
// zero links.
const (
	AdminSiteLinkStatusAvailable   AdminSiteLinkStatus = "available"
	AdminSiteLinkStatusPending     AdminSiteLinkStatus = "pending"
	AdminSiteLinkStatusFailed      AdminSiteLinkStatus = "failed"
	AdminSiteLinkStatusUnavailable AdminSiteLinkStatus = "unavailable"
)

// AdminSiteLinkSourceTotal is what one source says about a domain, over Scope as of AsOf (a day,
// or the Common Crawl release). ReferringDomains and Links are nil when the source has no number,
// never 0 for missing data.
type AdminSiteLinkSourceTotal struct {
	Source           AdminSiteBacklinkSource `json:"source"`
	Status           AdminSiteLinkStatus     `json:"status"`
	Scope            string                  `json:"scope"`
	AsOf             string                  `json:"asOf,omitempty"`
	ReferringDomains *int64                  `json:"referringDomains,omitempty"`
	Links            *int64                  `json:"links,omitempty"`
	Listed           int                     `json:"listed"`
	Complete         bool                    `json:"complete"`
	LastFailure      string                  `json:"lastFailure,omitempty"`
}

// AdminSiteDomainAuthority is a registrable domain's place in one Common Crawl domain graph. The
// ranks are nil when InGraph is false.
type AdminSiteDomainAuthority struct {
	Release            string   `json:"release"`
	GraphDomain        string   `json:"graphDomain"`
	InGraph            bool     `json:"inGraph"`
	HarmonicRank       *int64   `json:"harmonicRank,omitempty"`
	HarmonicCentrality *float64 `json:"harmonicCentrality,omitempty"`
	PageRankRank       *int64   `json:"pageRankRank,omitempty"`
	PageRank           *float64 `json:"pageRank,omitempty"`
	Hosts              *int     `json:"hosts,omitempty"`
}

// AdminSiteReferringDomain is one registrable domain linking to the domain, with what each source
// says about it. A nil count or Followed means that source did not see it.
type AdminSiteReferringDomain struct {
	Domain       string   `json:"domain"`
	Sources      []string `json:"sources"`
	BingLinks    *int     `json:"bingLinks,omitempty"`
	CrawlLinks   *int     `json:"crawlLinks,omitempty"`
	Followed     *bool    `json:"followed,omitempty"`
	AnchorText   string   `json:"anchorText,omitempty"`
	HarmonicRank *int64   `json:"harmonicRank,omitempty"`
	PageRankRank *int64   `json:"pageRankRank,omitempty"`
}

// AdminSiteLostReferringDomain is a domain the previous Common Crawl release listed as linking
// that the current one has no link from; HarmonicRank is its previous position.
type AdminSiteLostReferringDomain struct {
	Domain       string `json:"domain"`
	HarmonicRank *int64 `json:"harmonicRank,omitempty"`
}

// AdminSiteLostLinks is the quarter-to-quarter diff of Common Crawl referring domains.
type AdminSiteLostLinks struct {
	Status          AdminSiteLinkStatus            `json:"status"`
	Release         string                         `json:"release,omitempty"`
	PreviousRelease string                         `json:"previousRelease,omitempty"`
	Scope           string                         `json:"scope"`
	Domains         []AdminSiteLostReferringDomain `json:"domains"`
}

// AdminSiteReferringDomains is who links to one of a Site's domains: each source's totals, its
// Common Crawl authority (nil until a release is read), the referring domains and those lost.
type AdminSiteReferringDomains struct {
	SiteUUID         string                     `json:"siteUuid"`
	Domain           string                     `json:"domain"`
	Role             AdminSiteWebRole           `json:"role"`
	Sources          []AdminSiteLinkSourceTotal `json:"sources"`
	Authority        *AdminSiteDomainAuthority  `json:"authority,omitempty"`
	ReferringDomains []AdminSiteReferringDomain `json:"referringDomains"`
	Lost             AdminSiteLostLinks         `json:"lost"`
}

// GetReferringDomains returns who links to domain, one of the Site's own or competitor DNS names.
// A site-bound credential may pass an empty siteUUID for its own Site.
func (c *SiteDomainAdminClient) GetReferringDomains(
	ctx context.Context, siteUUID, domain string,
) (*AdminSiteReferringDomains, error) {
	var referring AdminSiteReferringDomains
	path := siteScopedPath("/referring-domains", siteUUID, url.Values{"domain": {domain}})
	err := c.request(ctx, http.MethodGet, path, nil, &referring)
	return &referring, err
}
