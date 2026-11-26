package custd

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AdminSiteSearchPoolMembership is whether the tenant has joined the shared search pool.
type AdminSiteSearchPoolMembership string

// The values of AdminSiteRankings.Pool: whether the tenant has joined the shared search pool.
const (
	AdminSiteSearchPoolJoined    AdminSiteSearchPoolMembership = "joined"
	AdminSiteSearchPoolNotJoined AdminSiteSearchPoolMembership = "not_joined"
)

// AdminSiteSearchPoolDemandBand is a query's pooled impressions band.
type AdminSiteSearchPoolDemandBand string

// The demand bands of AdminSiteSearchPoolBand.DemandBand: a query's pooled impressions over the
// pool's window.
const (
	AdminSiteSearchPoolDemand100To999  AdminSiteSearchPoolDemandBand = "100_999"
	AdminSiteSearchPoolDemand1KTo10K   AdminSiteSearchPoolDemandBand = "1k_10k"
	AdminSiteSearchPoolDemand10KTo100K AdminSiteSearchPoolDemandBand = "10k_100k"
	AdminSiteSearchPoolDemand100KPlus  AdminSiteSearchPoolDemandBand = "100k_plus"
)

// The shared search pool's rules and the bounds of ListSearchPoolQueries.
const (
	AdminSiteSearchPoolMinContributors     = 5
	AdminSiteSearchPoolMinImpressions      = 100
	AdminSiteSearchPoolWindowDays          = 28
	AdminSiteSearchPoolQueriesDefaultLimit = 50
	AdminSiteSearchPoolQueriesMaxLimit     = 200
	AdminSiteSearchPoolSearchMaxLength     = 100
)

// AdminSiteSearchPoolTermsVersion is the shared search pool terms version current when this SDK
// was released (https://custd.com/docs/terms/shared-search-pool/). A join must name the version
// the server currently requires, which GetSearchPool reports as TermsVersion; prefer that live
// value, since this constant goes stale when the terms change.
const AdminSiteSearchPoolTermsVersion = "2026-11-18"

// SiteSearchPoolNotJoinedCode is the problem code a 403 carries when a tenant that has not joined
// the shared search pool reads it.
const SiteSearchPoolNotJoinedCode = "site_search_pool_not_joined"

// SiteSearchPoolSiteBoundCode is the problem code a 403 carries when a credential bound to one Site
// reads or changes the tenant's search pool setting.
const SiteSearchPoolSiteBoundCode = "site_search_pool_site_bound"

// SiteSearchPoolSettingInvalidCode is the problem code a 400 carries when a setting change does
// not say joined true or false.
const SiteSearchPoolSettingInvalidCode = "site_search_pool_setting_invalid"

// SiteSearchPoolQueryInvalidCode is the problem code a 400 carries unless search is at most 100
// characters and limit 1 to 200.
const SiteSearchPoolQueryInvalidCode = "site_search_pool_query_invalid"

// AdminSiteSearchPoolSetting is whether the tenant has joined the shared search pool, with the
// rules the pool applies and the terms a join accepts. JoinedAt and AcceptedTermsVersion are empty
// while not joined; TermsVersion is the version a join must name.
type AdminSiteSearchPoolSetting struct {
	Joined               bool       `json:"joined"`
	JoinedAt             *time.Time `json:"joinedAt,omitempty"`
	AcceptedTermsVersion string     `json:"acceptedTermsVersion,omitempty"`
	TermsVersion         string     `json:"termsVersion"`
	MinContributors      int        `json:"minContributors"`
	MinImpressions       int        `json:"minImpressions"`
	WindowDays           int        `json:"windowDays"`
}

// AdminSiteSearchPoolPositions is the share, in percent rounded to 5, of a query's pooled
// impressions at average positions 1-3, 4-10, 11-20 and after 20.
type AdminSiteSearchPoolPositions struct {
	Top3     int `json:"top3"`
	Top10    int `json:"top10"`
	Top20    int `json:"top20"`
	Beyond20 int `json:"beyond20"`
}

// AdminSiteSearchPoolBand is one pooled query's bands, over every country and device.
type AdminSiteSearchPoolBand struct {
	DemandBand AdminSiteSearchPoolDemandBand `json:"demandBand"`
	Positions  AdminSiteSearchPoolPositions  `json:"positions"`
}

// AdminSiteSearchPoolQuery is one pooled query and its bands.
type AdminSiteSearchPoolQuery struct {
	Query string `json:"query"`
	AdminSiteSearchPoolBand
}

// AdminSiteSearchPoolQueryPage is a page of the pool, busiest demand band first. From, To and
// ComputedAt are empty before the first rollup.
type AdminSiteSearchPoolQueryPage struct {
	From       string                     `json:"from,omitempty"`
	To         string                     `json:"to,omitempty"`
	ComputedAt *time.Time                 `json:"computedAt,omitempty"`
	Queries    []AdminSiteSearchPoolQuery `json:"queries"`
}

// AdminSiteSearchPoolQueriesQuery narrows a pool read to queries containing Search, at most Limit
// (1..200); empty Search and 0 Limit use the server's defaults (every query, 50).
type AdminSiteSearchPoolQueriesQuery struct {
	Search string
	Limit  int
}

// adminSiteSearchPoolRequest is the setting change body; joined is always sent, termsVersion only
// on a join.
type adminSiteSearchPoolRequest struct {
	Joined       bool   `json:"joined"`
	TermsVersion string `json:"termsVersion,omitempty"`
}

// ErrSearchPoolTermsVersionRequired refuses a join that names no terms version, before any request.
var ErrSearchPoolTermsVersionRequired = errors.New(
	"custd: joining the shared search pool requires the terms version the caller accepts")

const siteSearchPoolPath = siteDomainsPath + "/search-pool"

// GetSearchPool returns whether the tenant has joined the shared search pool.
func (c *SiteDomainAdminClient) GetSearchPool(ctx context.Context) (*AdminSiteSearchPoolSetting, error) {
	var setting AdminSiteSearchPoolSetting
	err := c.request(ctx, http.MethodGet, siteSearchPoolPath, nil, &setting)
	return &setting, err
}

// SetSearchPool joins (true) or leaves (false) the shared search pool for the whole tenant.
// Joining shares the tenant's Search Console query numbers, pooled anonymously, and is the
// organisation's acceptance of the pool terms: termsVersion is required to join (the setting's
// TermsVersion, or AdminSiteSearchPoolTermsVersion) and ignored when leaving. A stale version is
// refused with a 400.
func (c *SiteDomainAdminClient) SetSearchPool(
	ctx context.Context, joined bool, termsVersion string,
) (*AdminSiteSearchPoolSetting, error) {
	body := adminSiteSearchPoolRequest{Joined: joined}
	if joined {
		if strings.TrimSpace(termsVersion) == "" {
			return nil, ErrSearchPoolTermsVersionRequired
		}
		body.TermsVersion = termsVersion
	}
	var setting AdminSiteSearchPoolSetting
	err := c.request(ctx, http.MethodPut, siteSearchPoolPath, body, &setting)
	return &setting, err
}

// ListSearchPoolQueries reads the shared search pool; a tenant that has not joined is refused
// with SiteSearchPoolNotJoinedCode.
func (c *SiteDomainAdminClient) ListSearchPoolQueries(
	ctx context.Context, query AdminSiteSearchPoolQueriesQuery,
) (*AdminSiteSearchPoolQueryPage, error) {
	var page AdminSiteSearchPoolQueryPage
	path := siteQueryPath(siteSearchPoolPath+"/queries", url.Values{
		"search": {query.Search}, "limit": {sitePositiveInt(query.Limit)},
	})
	err := c.request(ctx, http.MethodGet, path, nil, &page)
	return &page, err
}
