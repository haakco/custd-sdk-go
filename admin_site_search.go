package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// AdminSiteSearchProvider is a search engine Custd imports a domain's search numbers from.
type AdminSiteSearchProvider string

const (
	// AdminSiteSearchProviderGoogle imports from Google Search Console through Custd's service
	// account, which the client adds as a (Restricted) user on its property.
	AdminSiteSearchProviderGoogle AdminSiteSearchProvider = "google"
	// AdminSiteSearchProviderBing imports from Bing Webmaster Tools with the client's API key.
	AdminSiteSearchProviderBing AdminSiteSearchProvider = "bing"
)

// AdminSiteSearchConnectionStatus is the import state of one provider connection.
type AdminSiteSearchConnectionStatus string

const (
	// AdminSiteSearchConnectionPending is connected but not yet imported.
	AdminSiteSearchConnectionPending AdminSiteSearchConnectionStatus = "pending"
	AdminSiteSearchConnectionActive  AdminSiteSearchConnectionStatus = "active"
	// AdminSiteSearchConnectionFailing is a connection whose latest import failed; LastFailure says why.
	AdminSiteSearchConnectionFailing AdminSiteSearchConnectionStatus = "failing"
	// AdminSiteSearchConnectionRevoked is a connection the provider no longer lets Custd read.
	AdminSiteSearchConnectionRevoked AdminSiteSearchConnectionStatus = "revoked"
)

// SiteSearchNotConfiguredCode is the problem code a 409 carries when this Custd deployment has
// no account or key for the provider being connected.
const SiteSearchNotConfiguredCode = "site_search_not_configured"

// SiteSearchConnectionNotFoundCode is the problem code a 404 carries when disconnecting a
// provider the domain is not connected to.
const SiteSearchConnectionNotFoundCode = "site_search_connection_not_found"

// AdminSiteSearchAccount says which providers this Custd deployment can import from, and the
// Google service account email the client must add as a user on its Search Console property.
type AdminSiteSearchAccount struct {
	GoogleConfigured          bool   `json:"googleConfigured"`
	GoogleServiceAccountEmail string `json:"googleServiceAccountEmail,omitempty"`
	BingConfigured            bool   `json:"bingConfigured"`
}

// AdminSiteSearchConnectionRequest connects a domain to one provider. Property is
// "sc-domain:example.com" or a URL prefix such as "https://www.example.com/" for Google, and a
// URL prefix ending in "/" for Bing. APIKey is Bing only; it is write-only and never returned.
type AdminSiteSearchConnectionRequest struct {
	Property string `json:"property"`
	APIKey   string `json:"apiKey,omitempty"`
}

// AdminSiteSearchConnection is one provider connection and its import state.
type AdminSiteSearchConnection struct {
	Provider        AdminSiteSearchProvider         `json:"provider"`
	Property        string                          `json:"property"`
	Status          AdminSiteSearchConnectionStatus `json:"status"`
	ImportedThrough string                          `json:"importedThrough,omitempty"`
	LastImportedAt  *time.Time                      `json:"lastImportedAt,omitempty"`
	LastFailure     string                          `json:"lastFailure,omitempty"`
}

// AdminSiteSearchTotals are summed search numbers. CTR is 0..1; Position is 0 when unknown.
type AdminSiteSearchTotals struct {
	Clicks      int64   `json:"clicks"`
	Impressions int64   `json:"impressions"`
	CTR         float64 `json:"ctr"`
	Position    float64 `json:"position,omitempty"`
}

// AdminSiteSearchGranularity is how finely a search source reports numbers.
type AdminSiteSearchGranularity string

const (
	AdminSiteSearchGranularityDaily  AdminSiteSearchGranularity = "daily"
	AdminSiteSearchGranularityWeekly AdminSiteSearchGranularity = "weekly"
)

// AdminSiteSearchSource is one provider's totals for the window, with the dates it covers.
// Granularity is "daily" or "weekly".
type AdminSiteSearchSource struct {
	Provider    AdminSiteSearchProvider    `json:"provider"`
	Source      string                     `json:"source"`
	Granularity AdminSiteSearchGranularity `json:"granularity"`
	From        string                     `json:"from,omitempty"`
	To          string                     `json:"to,omitempty"`
	Days        int                        `json:"days"`
	ImportedAt  *time.Time                 `json:"importedAt,omitempty"`
	Totals      AdminSiteSearchTotals      `json:"totals"`
	// Previous are the totals for the same number of days before the window, when there are any.
	Previous *AdminSiteSearchTotals `json:"previous,omitempty"`
	// Daily is each imported day's totals, oldest first; Bing's are week buckets.
	Daily []AdminSiteSearchDay `json:"daily"`
}

// AdminSiteSearchDay is one provider's totals for one imported day (YYYY-MM-DD).
type AdminSiteSearchDay struct {
	Day         string `json:"day"`
	Clicks      int64  `json:"clicks"`
	Impressions int64  `json:"impressions"`
}

// AdminSiteSearchRow is one page URL or query text and its totals.
type AdminSiteSearchRow struct {
	Provider AdminSiteSearchProvider `json:"provider"`
	Key      string                  `json:"key"`
	Totals   AdminSiteSearchTotals   `json:"totals"`
}

// AdminSiteSearchPerformance is a domain's imported search numbers over the last Days days.
type AdminSiteSearchPerformance struct {
	DomainUUID  string                      `json:"domainUuid"`
	Days        int                         `json:"days"`
	Connections []AdminSiteSearchConnection `json:"connections"`
	Sources     []AdminSiteSearchSource     `json:"sources"`
	TopPages    []AdminSiteSearchRow        `json:"topPages"`
	TopQueries  []AdminSiteSearchRow        `json:"topQueries"`
	// StrikingDistance are queries averaging position 5 to 20: already shown and close to the top
	// results, so improving their page is the cheapest gain. Busiest first.
	StrikingDistance []AdminSiteSearchRow `json:"strikingDistance"`
}

// GetSearchAccount returns which providers this Custd deployment can import from.
func (c *SiteDomainAdminClient) GetSearchAccount(ctx context.Context) (*AdminSiteSearchAccount, error) {
	var account AdminSiteSearchAccount
	err := c.request(ctx, http.MethodGet, siteDomainsPath+"/search-account", nil, &account)
	return &account, err
}

// GetSearchPerformance returns a domain's search numbers over the last days days (1..90); 0
// lets the server default to 28.
func (c *SiteDomainAdminClient) GetSearchPerformance(
	ctx context.Context, domainUUID string, days int,
) (*AdminSiteSearchPerformance, error) {
	var performance AdminSiteSearchPerformance
	path := siteDomainPath(domainUUID) + "/search"
	if days > 0 {
		path += "?" + url.Values{"days": {strconv.Itoa(days)}}.Encode()
	}
	err := c.request(ctx, http.MethodGet, path, nil, &performance)
	return &performance, err
}

// ConnectSearch connects a domain to a provider, or replaces its property or key. A deployment
// without an account or key for the provider refuses it with a 409 carrying
// SiteSearchNotConfiguredCode.
func (c *SiteDomainAdminClient) ConnectSearch(
	ctx context.Context, domainUUID string, provider AdminSiteSearchProvider, req AdminSiteSearchConnectionRequest,
) (*AdminSiteSearchConnection, error) {
	var connection AdminSiteSearchConnection
	err := c.request(ctx, http.MethodPut, siteSearchConnectionPath(domainUUID, provider), req, &connection)
	return &connection, err
}

// DisconnectSearch removes a provider connection and the numbers imported through it. A provider
// that is not connected answers 404 with SiteSearchConnectionNotFoundCode.
func (c *SiteDomainAdminClient) DisconnectSearch(
	ctx context.Context, domainUUID string, provider AdminSiteSearchProvider,
) error {
	return c.request(ctx, http.MethodDelete, siteSearchConnectionPath(domainUUID, provider), nil, nil)
}

// SiteSearchRetryNotFailingCode is the problem code a 409 carries when RetrySearchImport names a
// connection that is not failing: connect a revoked one again, and a pending or active one is
// already due on schedule.
const SiteSearchRetryNotFailingCode = "site_search_retry_not_failing"

// RetrySearchImport makes a failing provider import due on the scheduler's next tick and returns
// the connection as it stands.
func (c *SiteDomainAdminClient) RetrySearchImport(
	ctx context.Context, domainUUID string, provider AdminSiteSearchProvider,
) (*AdminSiteSearchConnection, error) {
	var connection AdminSiteSearchConnection
	err := c.request(ctx, http.MethodPost, siteSearchConnectionPath(domainUUID, provider)+"/retry", nil, &connection)
	return &connection, err
}

func siteSearchConnectionPath(domainUUID string, provider AdminSiteSearchProvider) string {
	return siteDomainPath(domainUUID) + "/search/" + url.PathEscape(string(provider))
}

// AdminSiteSearchDimension is the dimension imported search rows are broken down by.
type AdminSiteSearchDimension string

// The dimensions AdminSiteSearchRowsQuery.Dimension reads.
const (
	AdminSiteSearchDimensionPage    AdminSiteSearchDimension = "page"
	AdminSiteSearchDimensionQuery   AdminSiteSearchDimension = "query"
	AdminSiteSearchDimensionCountry AdminSiteSearchDimension = "country"
	AdminSiteSearchDimensionDevice  AdminSiteSearchDimension = "device"
)

// AdminSiteSearchRowsQuery pages through one dimension of a domain's imported search rows. Zero
// values let the server default: the query dimension, every provider, 28 days, offset 0 and 100
// rows (Limit at most 1000).
type AdminSiteSearchRowsQuery struct {
	Dimension AdminSiteSearchDimension
	Provider  AdminSiteSearchProvider
	Days      int
	Offset    int
	Limit     int
}

func (q AdminSiteSearchRowsQuery) values() url.Values {
	values := url.Values{}
	if q.Dimension != "" {
		values.Set("dimension", string(q.Dimension))
	}
	if q.Provider != "" {
		values.Set("provider", string(q.Provider))
	}
	for name, value := range map[string]int{"days": q.Days, "offset": q.Offset, "limit": q.Limit} {
		if value > 0 {
			values.Set(name, strconv.Itoa(value))
		}
	}
	return values
}

// AdminSiteSearchUnavailable names a provider that cannot report the dimension at all, so its
// missing rows are not zero traffic.
type AdminSiteSearchUnavailable struct {
	Provider AdminSiteSearchProvider `json:"provider"`
	Reason   string                  `json:"reason"`
}

// AdminSiteSearchRowPage is one page of a dimension's rows. Total counts the rows before paging;
// Provider is empty when every provider was read.
type AdminSiteSearchRowPage struct {
	DomainUUID  string                       `json:"domainUuid"`
	Dimension   AdminSiteSearchDimension     `json:"dimension"`
	Provider    AdminSiteSearchProvider      `json:"provider,omitempty"`
	Days        int                          `json:"days"`
	Offset      int                          `json:"offset"`
	Limit       int                          `json:"limit"`
	Total       int                          `json:"total"`
	Rows        []AdminSiteSearchRow         `json:"rows"`
	Unavailable []AdminSiteSearchUnavailable `json:"unavailable"`
}

// ListSearchRows reads one page of a domain's imported rows for one dimension.
func (c *SiteDomainAdminClient) ListSearchRows(
	ctx context.Context, domainUUID string, query AdminSiteSearchRowsQuery,
) (*AdminSiteSearchRowPage, error) {
	var page AdminSiteSearchRowPage
	path := siteDomainPath(domainUUID) + "/search/rows"
	if values := query.values(); len(values) > 0 {
		path += "?" + values.Encode()
	}
	err := c.request(ctx, http.MethodGet, path, nil, &page)
	return &page, err
}
