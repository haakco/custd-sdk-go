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

// AdminSiteKeywordList is one named list of keywords a Site tracks. SeedTerm is empty when the
// list was not built from a seed.
type AdminSiteKeywordList struct {
	Name      string    `json:"name"`
	SeedTerm  string    `json:"seedTerm,omitempty"`
	Keywords  []string  `json:"keywords"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
}

// AdminSiteKeywordCluster is one group of a list's keywords and their search numbers. Clicks,
// Impressions and AvgPosition are nil when no keyword in it was measured.
type AdminSiteKeywordCluster struct {
	Cluster          string   `json:"cluster"`
	Keywords         []string `json:"keywords"`
	KeywordCount     int      `json:"keywordCount"`
	MeasuredKeywords int      `json:"measuredKeywords"`
	Clicks           *int64   `json:"clicks,omitempty"`
	Impressions      *int64   `json:"impressions,omitempty"`
	AvgPosition      *float64 `json:"avgPosition,omitempty"`
}

// AdminSiteKeywordStriking is one list keyword close to the first page of results.
type AdminSiteKeywordStriking struct {
	Keyword     string  `json:"keyword"`
	Clicks      int64   `json:"clicks"`
	Impressions int64   `json:"impressions"`
	Position    float64 `json:"position"`
}

// AdminSiteKeywordClusters is a keyword list grouped into clusters with the search numbers of the
// last MetricsDays days from MetricsSource.
type AdminSiteKeywordClusters struct {
	SiteUUID         string                     `json:"siteUuid"`
	Name             string                     `json:"name"`
	MetricsSource    string                     `json:"metricsSource"`
	MetricsDays      int                        `json:"metricsDays"`
	TotalKeywords    int                        `json:"totalKeywords"`
	Clusters         []AdminSiteKeywordCluster  `json:"clusters"`
	StrikingDistance []AdminSiteKeywordStriking `json:"strikingDistance"`
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
