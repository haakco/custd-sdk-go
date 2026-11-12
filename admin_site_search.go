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

// AdminSiteSearchSource is one provider's totals for the window, with the dates it covers.
// Granularity is "daily" or "weekly".
type AdminSiteSearchSource struct {
	Provider    AdminSiteSearchProvider `json:"provider"`
	Source      string                  `json:"source"`
	Granularity string                  `json:"granularity"`
	From        string                  `json:"from,omitempty"`
	To          string                  `json:"to,omitempty"`
	Days        int                     `json:"days"`
	ImportedAt  *time.Time              `json:"importedAt,omitempty"`
	Totals      AdminSiteSearchTotals   `json:"totals"`
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

func siteSearchConnectionPath(domainUUID string, provider AdminSiteSearchProvider) string {
	return siteDomainPath(domainUUID) + "/search/" + url.PathEscape(string(provider))
}
