package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// The roles of AdminSiteWebDomain.Role.
const (
	AdminSiteWebRoleOwn        = "own"
	AdminSiteWebRoleCompetitor = "competitor"
)

// The sources a page was listed in (AdminSiteWebPage.Sources).
const (
	AdminSiteWebSourceSitemap = "sitemap"
	AdminSiteWebSourceWayback = "wayback"
)

// The inventory states of AdminSiteWebDomain.SitemapStatus and WaybackStatus. Compare against
// these rather than assuming the set is closed.
const (
	AdminSiteWebStatusPending     = "pending"
	AdminSiteWebStatusDone        = "done"
	AdminSiteWebStatusPartial     = "partial"
	AdminSiteWebStatusUnavailable = "unavailable"
	AdminSiteWebStatusFailed      = "failed"
)

// The Tranco rank bands of AdminSiteWebDomain.TrancoBand. Unranked means the domain is outside
// the top million; unknown means no list has been read yet.
const (
	AdminSiteTrancoBandTop1K    = "top-1k"
	AdminSiteTrancoBandTop10K   = "top-10k"
	AdminSiteTrancoBandTop100K  = "top-100k"
	AdminSiteTrancoBandTop1M    = "top-1m"
	AdminSiteTrancoBandUnranked = "unranked"
	AdminSiteTrancoBandUnknown  = "unknown"
)

// The states of AdminSiteTopicTrend.Status.
const (
	AdminSiteTrendStatusPending   = "pending"
	AdminSiteTrendStatusFound     = "found"
	AdminSiteTrendStatusNoArticle = "no_article"
	AdminSiteTrendStatusFailed    = "failed"
)

// Where an AdminSiteTopicTrend's term came from.
const (
	AdminSiteTrendSourceBrand = "brand"
	AdminSiteTrendSourceSeed  = "seed"
)

// The values of AdminSiteContentGap.OwnCoverage: whether the Site's own pages could be read.
const (
	AdminSiteGapCoverageAvailable   = "available"
	AdminSiteGapCoverageUnavailable = "unavailable"
)

// AdminSiteWebDomain is one of the Site's own or competitor domains, its Tranco rank and its
// page inventory. TrancoRank and InventoryAt are nil when not known.
type AdminSiteWebDomain struct {
	Domain          string     `json:"domain"`
	Role            string     `json:"role"`
	TrancoBand      string     `json:"trancoBand"`
	TrancoRank      *int       `json:"trancoRank,omitempty"`
	TrancoListDay   string     `json:"trancoListDay,omitempty"`
	SitemapStatus   string     `json:"sitemapStatus"`
	WaybackStatus   string     `json:"waybackStatus"`
	PageCount       int        `json:"pageCount"`
	LastFailure     string     `json:"lastFailure,omitempty"`
	InventoryAt     *time.Time `json:"inventoryAt,omitempty"`
	NextInventoryAt time.Time  `json:"nextInventoryAt"`
}

// AdminSiteWebDomains lists a Site's own and competitor domains.
type AdminSiteWebDomains struct {
	SiteUUID string               `json:"siteUuid"`
	Domains  []AdminSiteWebDomain `json:"domains"`
}

// AdminSiteWebPage is one public URL of a domain and where it was listed (YYYY-MM-DD days).
type AdminSiteWebPage struct {
	URL       string   `json:"url"`
	Sources   []string `json:"sources"`
	FirstSeen string   `json:"firstSeen"`
	LastSeen  string   `json:"lastSeen"`
}

// AdminSiteWebPages is one page of a domain's URLs; pass Next as the next query's After, and an
// empty Next means this was the last page.
type AdminSiteWebPages struct {
	SiteUUID string             `json:"siteUuid"`
	Domain   string             `json:"domain"`
	Pages    []AdminSiteWebPage `json:"pages"`
	Next     string             `json:"next,omitempty"`
}

// AdminSiteWebPagesQuery pages a domain's URLs. SiteUUID may be empty for a site-bound
// credential; After is the previous page's Next; Limit 0 uses the server's 100.
type AdminSiteWebPagesQuery struct {
	SiteUUID string
	After    string
	Limit    int
}

// AdminSiteTrendMonth is one month's (YYYY-MM) page views.
type AdminSiteTrendMonth struct {
	Month string `json:"month"`
	Views int64  `json:"views"`
}

// AdminSiteTopicTrend is one brand or seed term's Wikipedia article views by month. CheckedAt is
// nil before the first check.
type AdminSiteTopicTrend struct {
	Term        string                `json:"term"`
	Source      string                `json:"source"`
	Project     string                `json:"project"`
	Article     string                `json:"article"`
	Status      string                `json:"status"`
	LastFailure string                `json:"lastFailure,omitempty"`
	CheckedAt   *time.Time            `json:"checkedAt,omitempty"`
	Months      []AdminSiteTrendMonth `json:"months"`
}

// AdminSiteTopicTrends lists a Site's topic trends.
type AdminSiteTopicTrends struct {
	SiteUUID string                `json:"siteUuid"`
	Trends   []AdminSiteTopicTrend `json:"trends"`
}

// AdminSiteGapPage is one competitor page on a gap topic.
type AdminSiteGapPage struct {
	Domain    string   `json:"domain"`
	URL       string   `json:"url"`
	Sources   []string `json:"sources"`
	FirstSeen string   `json:"firstSeen"`
	LastSeen  string   `json:"lastSeen"`
}

// AdminSiteGapTopic is one topic competitors cover and the Site does not.
type AdminSiteGapTopic struct {
	Topic       string             `json:"topic"`
	Label       string             `json:"label"`
	Competitors []string           `json:"competitors"`
	PageCount   int                `json:"pageCount"`
	Pages       []AdminSiteGapPage `json:"pages"`
}

// AdminSiteGapCompetitor is one competitor's inventory as the gap read it.
type AdminSiteGapCompetitor struct {
	Domain        string `json:"domain"`
	SitemapStatus string `json:"sitemapStatus"`
	WaybackStatus string `json:"waybackStatus"`
	PageCount     int    `json:"pageCount"`
}

// AdminSiteContentGap is the topics competitors cover that the Site does not. Partial means an
// inventory was incomplete.
type AdminSiteContentGap struct {
	SiteUUID    string                   `json:"siteUuid"`
	Seed        string                   `json:"seed,omitempty"`
	OwnCoverage string                   `json:"ownCoverage"`
	OwnPages    int                      `json:"ownPages"`
	OwnQueries  int                      `json:"ownQueries"`
	Partial     bool                     `json:"partial"`
	Competitors []AdminSiteGapCompetitor `json:"competitors"`
	Topics      []AdminSiteGapTopic      `json:"topics"`
}

// GetWebDomains returns the Site's own and competitor domains. A site-bound credential may pass
// an empty siteUUID for its own Site.
func (c *SiteDomainAdminClient) GetWebDomains(ctx context.Context, siteUUID string) (*AdminSiteWebDomains, error) {
	var domains AdminSiteWebDomains
	err := c.request(ctx, http.MethodGet, siteScopedPath("/web-domains", siteUUID, nil), nil, &domains)
	return &domains, err
}

// ListWebPages returns one page of a domain's public URLs.
func (c *SiteDomainAdminClient) ListWebPages(
	ctx context.Context, domain string, query AdminSiteWebPagesQuery,
) (*AdminSiteWebPages, error) {
	var pages AdminSiteWebPages
	values := url.Values{"domain": {domain}, "after": {query.After}, "limit": {sitePositiveInt(query.Limit)}}
	err := c.request(ctx, http.MethodGet, siteScopedPath("/web-pages", query.SiteUUID, values), nil, &pages)
	return &pages, err
}

// GetTopicTrends returns the Site's brand and seed term trends.
func (c *SiteDomainAdminClient) GetTopicTrends(ctx context.Context, siteUUID string) (*AdminSiteTopicTrends, error) {
	var trends AdminSiteTopicTrends
	err := c.request(ctx, http.MethodGet, siteScopedPath("/topic-trends", siteUUID, nil), nil, &trends)
	return &trends, err
}

// GetContentGap returns the topics competitors cover and the Site does not, narrowed to seed when
// it is not empty.
func (c *SiteDomainAdminClient) GetContentGap(ctx context.Context, siteUUID, seed string) (*AdminSiteContentGap, error) {
	var gap AdminSiteContentGap
	path := siteScopedPath("/content-gap", siteUUID, url.Values{"seed": {seed}})
	err := c.request(ctx, http.MethodGet, path, nil, &gap)
	return &gap, err
}

// siteScopedPath is a Site-scoped site-domains route with the Site (omitted when empty) and the
// other non-empty values in one query.
func siteScopedPath(route, siteUUID string, values url.Values) string {
	if values == nil {
		values = url.Values{}
	}
	values.Set("siteUuid", siteUUID)
	return siteQueryPath(siteDomainsPath+route, values)
}
