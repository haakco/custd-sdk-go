package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// SiteKeywordListNotFoundCode is the problem code a 404 carries when the Site has no keyword list
// of that name.
const SiteKeywordListNotFoundCode = "site_keyword_list_not_found"

// SiteKeywordListLimitCode is the problem code a 409 carries once the Site keeps its maximum
// number of keyword lists.
const SiteKeywordListLimitCode = "site_keyword_list_limit"

// AdminSiteKeywordDemandSourceBing names where keyword demand and related keywords come from:
// Bing's keyword research for the Site's first market, read with the Site's own Bing key.
const AdminSiteKeywordDemandSourceBing = "bing_keyword_research"

// The values of AdminSiteKeywordDemandSet.State: why a list's keywords have demand numbers or not.
const (
	AdminSiteKeywordDemandAvailable = "available"
	AdminSiteKeywordDemandNoMarket  = "no_market"
	AdminSiteKeywordDemandNoBing    = "no_bing"
)

// AdminSiteKeywordList is one named list of keywords a Site tracks. SeedTerm is empty when the
// list was not built from a seed. Related is what Bing relates to the seed term, nil until read,
// with no seed, or from an older server.
type AdminSiteKeywordList struct {
	Name      string                    `json:"name"`
	SeedTerm  string                    `json:"seedTerm,omitempty"`
	Keywords  []string                  `json:"keywords"`
	CreatedAt time.Time                 `json:"createdAt"`
	UpdatedAt time.Time                 `json:"updatedAt"`
	Related   *AdminSiteRelatedKeywords `json:"related,omitempty"`
	// MarketSampleEnabled allows sampling the public seed topic; MarketSample
	// is the latest sample, nil before the first one.
	MarketSampleEnabled bool                   `json:"marketSampleEnabled"`
	MarketSample        *AdminSiteMarketSample `json:"marketSample,omitempty"`
}

// AdminSiteMarketSample is one sample of the public search results for a seed topic.
type AdminSiteMarketSample struct {
	ObservedAt time.Time                  `json:"observedAt"`
	Hits       []AdminSiteMarketSearchHit `json:"hits"`
}

// AdminSiteMarketSearchHit is one result of a market sample. Authority is the
// sampler's authority tier, -1 when unknown; Date is nil when unpublished.
type AdminSiteMarketSearchHit struct {
	Title     string     `json:"title"`
	URL       string     `json:"url"`
	Authority int        `json:"authority"`
	Date      *time.Time `json:"date,omitempty"`
}

// AdminSiteRelatedKeyword is one keyword Bing relates to a seed term, with its searches over the
// period.
type AdminSiteRelatedKeyword struct {
	Keyword          string `json:"keyword"`
	Impressions      int64  `json:"impressions"`
	BroadImpressions int64  `json:"broadImpressions"`
}

// AdminSiteRelatedKeywords is what Bing relates to a list's seed term in Market, the busiest 50,
// over From to To.
type AdminSiteRelatedKeywords struct {
	Source    string                    `json:"source"`
	Market    string                    `json:"market"`
	From      string                    `json:"from"`
	To        string                    `json:"to"`
	CheckedAt time.Time                 `json:"checkedAt"`
	Keywords  []AdminSiteRelatedKeyword `json:"keywords"`
}

// AdminSiteKeywordDemand is one keyword's Bing searches in the latest week Bing reported. The
// numbers are nil when Bing reported none, never zero; LastFailure says why the latest read
// failed (it is retried a day later).
type AdminSiteKeywordDemand struct {
	Keyword           string    `json:"keyword"`
	WeeklyImpressions *int64    `json:"weeklyImpressions,omitempty"`
	BroadImpressions  *int64    `json:"broadImpressions,omitempty"`
	WeekOf            string    `json:"weekOf,omitempty"`
	CheckedAt         time.Time `json:"checkedAt"`
	LastFailure       string    `json:"lastFailure,omitempty"`
}

// AdminSiteKeywordDemandSet is the Bing demand Custd has read for a list's keywords, apart from the
// Site's own numbers. Keywords holds only keywords Bing has been asked about; the rest are read
// within a week.
type AdminSiteKeywordDemandSet struct {
	State    string                   `json:"state"`
	Source   string                   `json:"source"`
	Market   string                   `json:"market,omitempty"`
	Keywords []AdminSiteKeywordDemand `json:"keywords"`
}

// AdminSiteKeywordLists lists a Site's keyword lists.
type AdminSiteKeywordLists struct {
	SiteUUID string                 `json:"siteUuid"`
	Lists    []AdminSiteKeywordList `json:"lists"`
}

// AdminSiteKeywordListRequest creates or replaces the list called Name; SeedTerm is omitted when
// empty.
type AdminSiteKeywordListRequest struct {
	Name     string   `json:"name"`
	SeedTerm string   `json:"seedTerm,omitempty"`
	Keywords []string `json:"keywords"`
	// MarketSampleEnabled allows sampling the public seed topic; it needs a SeedTerm.
	MarketSampleEnabled bool `json:"marketSampleEnabled,omitempty"`
}

// AdminSiteKeywordCluster is one group of a list's keywords and their search numbers. Clicks,
// Impressions and AvgPosition are nil when no keyword in it was measured; BingWeeklyImpressions
// sums Bing's latest week of the keywords Bing reported, nil when none.
type AdminSiteKeywordCluster struct {
	Cluster          string   `json:"cluster"`
	Keywords         []string `json:"keywords"`
	KeywordCount     int      `json:"keywordCount"`
	MeasuredKeywords int      `json:"measuredKeywords"`
	Clicks           *int64   `json:"clicks,omitempty"`
	Impressions      *int64   `json:"impressions,omitempty"`
	AvgPosition      *float64 `json:"avgPosition,omitempty"`
	// BingWeeklyImpressions is Bing's latest weekly searches, summed.
	BingWeeklyImpressions *int64 `json:"bingWeeklyImpressions,omitempty"`
}

// AdminSiteKeywordStriking is one list keyword close to the first page of results.
type AdminSiteKeywordStriking struct {
	Keyword     string  `json:"keyword"`
	Clicks      int64   `json:"clicks"`
	Impressions int64   `json:"impressions"`
	Position    float64 `json:"position"`
}

// AdminSiteKeywordClusters is a keyword list grouped into clusters with the search numbers of the
// last MetricsDays days from MetricsSource. Demand is Bing keyword research for the list, nil from
// an older server.
type AdminSiteKeywordClusters struct {
	SiteUUID         string                     `json:"siteUuid"`
	Name             string                     `json:"name"`
	MetricsSource    string                     `json:"metricsSource"`
	MetricsDays      int                        `json:"metricsDays"`
	TotalKeywords    int                        `json:"totalKeywords"`
	Clusters         []AdminSiteKeywordCluster  `json:"clusters"`
	StrikingDistance []AdminSiteKeywordStriking `json:"strikingDistance"`
	Demand           *AdminSiteKeywordDemandSet `json:"demand,omitempty"`
}

// ListKeywordLists returns the Site's keyword lists. A site-bound credential may pass an empty
// siteUUID for its own Site.
func (c *SiteDomainAdminClient) ListKeywordLists(ctx context.Context, siteUUID string) (*AdminSiteKeywordLists, error) {
	var lists AdminSiteKeywordLists
	err := c.request(ctx, http.MethodGet, siteScopedPath(siteKeywordListsRoute, siteUUID, nil), nil, &lists)
	return &lists, err
}

// SaveKeywordList creates or replaces the Site's keyword list named in req.
func (c *SiteDomainAdminClient) SaveKeywordList(
	ctx context.Context, siteUUID string, req AdminSiteKeywordListRequest,
) (*AdminSiteKeywordList, error) {
	body := AdminSiteKeywordListRequest{Name: req.Name, SeedTerm: req.SeedTerm, Keywords: nonNilSlice(req.Keywords)}
	var list AdminSiteKeywordList
	err := c.request(ctx, http.MethodPut, siteScopedPath(siteKeywordListsRoute, siteUUID, nil), body, &list)
	return &list, err
}

// DeleteKeywordList deletes the Site's keyword list called name.
func (c *SiteDomainAdminClient) DeleteKeywordList(ctx context.Context, siteUUID, name string) error {
	path := siteScopedPath(siteKeywordListsRoute, siteUUID, url.Values{"name": {name}})
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

// GetKeywordClusters groups the keyword list called name into clusters with their search numbers.
func (c *SiteDomainAdminClient) GetKeywordClusters(
	ctx context.Context, siteUUID, name string,
) (*AdminSiteKeywordClusters, error) {
	var clusters AdminSiteKeywordClusters
	path := siteScopedPath("/keyword-clusters", siteUUID, url.Values{"name": {name}})
	err := c.request(ctx, http.MethodGet, path, nil, &clusters)
	return &clusters, err
}

const siteKeywordListsRoute = "/keyword-lists"
