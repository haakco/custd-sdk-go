package custd

import (
	"context"
	"net/http"
	"time"
)

// AdminSiteAnalyticsProvider is a web analytics provider a domain can connect.
type AdminSiteAnalyticsProvider string

// AdminSiteAnalyticsProviderGA4 names Google Analytics 4, the one analytics provider. Custd reads
// the property as its Search Console service account, which the tenant adds as a Viewer.
const AdminSiteAnalyticsProviderGA4 AdminSiteAnalyticsProvider = "ga4"

// AdminSiteAnalyticsConnectionStatus is the state of a domain's analytics connection.
type AdminSiteAnalyticsConnectionStatus string

// The values of AdminSiteAnalyticsConnection.Status. Revoked means GA4 refused Custd's service
// account: add it as a Viewer and connect again. Failing is retried on schedule.
const (
	AdminSiteAnalyticsPending AdminSiteAnalyticsConnectionStatus = "pending"
	AdminSiteAnalyticsActive  AdminSiteAnalyticsConnectionStatus = "active"
	AdminSiteAnalyticsFailing AdminSiteAnalyticsConnectionStatus = "failing"
	AdminSiteAnalyticsRevoked AdminSiteAnalyticsConnectionStatus = "revoked"
)

// AdminSiteAnalyticsUnavailableReport is an analytics report that can be unavailable.
type AdminSiteAnalyticsUnavailableReport string

// The values of AdminSiteAnalyticsUnavailable.Report: GA4 reports Custd does not import, each said
// as unavailable rather than empty.
const (
	AdminSiteAnalyticsEcommerce AdminSiteAnalyticsUnavailableReport = "ecommerce"
	AdminSiteAnalyticsAudience  AdminSiteAnalyticsUnavailableReport = "audience"
)

// SiteAnalyticsPropertyInvalidCode is the problem code a 400 carries when propertyId is not a GA4
// numeric property ID (123456789 or properties/123456789).
const SiteAnalyticsPropertyInvalidCode = "site_analytics_property_invalid"

// SiteAnalyticsConnectionNotFoundCode is the problem code a 404 carries when disconnecting a
// domain that has no GA4 property.
const SiteAnalyticsConnectionNotFoundCode = "site_analytics_connection_not_found"

// AdminSiteAnalyticsConnectionRequest connects a domain to one GA4 property by its numeric ID, as
// GA4's Property details show it; properties/ID is accepted.
type AdminSiteAnalyticsConnectionRequest struct {
	PropertyID string `json:"propertyId"`
}

// AdminSiteAnalyticsConnection is one domain's GA4 property and where its import stands.
type AdminSiteAnalyticsConnection struct {
	Provider       AdminSiteAnalyticsProvider         `json:"provider"`
	PropertyID     string                             `json:"propertyId"`
	Status         AdminSiteAnalyticsConnectionStatus `json:"status"`
	LastImportedAt *time.Time                         `json:"lastImportedAt,omitempty"`
	LastFailure    string                             `json:"lastFailure,omitempty"`
}

// AdminSiteAnalyticsRow is one row of a GA4 report over the window: a landing page path, key event
// name or channel group. A measure GA4 did not report is nil, so missing conversions never read
// as zero.
type AdminSiteAnalyticsRow struct {
	Key             string   `json:"key"`
	Sessions        *int64   `json:"sessions,omitempty"`
	EngagedSessions *int64   `json:"engagedSessions,omitempty"`
	KeyEvents       *float64 `json:"keyEvents,omitempty"`
}

// AdminSiteAnalyticsUnavailable is a GA4 report Custd does not read, and why.
type AdminSiteAnalyticsUnavailable struct {
	Report AdminSiteAnalyticsUnavailableReport `json:"report"`
	Reason string                              `json:"reason"`
}

// AdminSiteAnalyticsReport is what Custd imported from a domain's GA4 property over the newest 28
// complete days: Organic Search landing pages and key events, and sessions by channel for all
// traffic. Connection is nil when the domain has no GA4 property, and the window empty until the
// first import. KeyEventsRecorded is false when the property recorded no organic key event.
type AdminSiteAnalyticsReport struct {
	DomainUUID        string                          `json:"domainUuid"`
	Connection        *AdminSiteAnalyticsConnection   `json:"connection,omitempty"`
	WindowFrom        string                          `json:"windowFrom,omitempty"`
	WindowTo          string                          `json:"windowTo,omitempty"`
	KeyEventsRecorded bool                            `json:"keyEventsRecorded"`
	LandingPages      []AdminSiteAnalyticsRow         `json:"landingPages"`
	KeyEvents         []AdminSiteAnalyticsRow         `json:"keyEvents"`
	Channels          []AdminSiteAnalyticsRow         `json:"channels"`
	Unavailable       []AdminSiteAnalyticsUnavailable `json:"unavailable"`
}

// GetAnalytics returns what Custd imported from the domain's GA4 property.
func (c *SiteDomainAdminClient) GetAnalytics(ctx context.Context, domainUUID string) (*AdminSiteAnalyticsReport, error) {
	var report AdminSiteAnalyticsReport
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/analytics", nil, &report)
	return &report, err
}

// ConnectAnalytics connects the domain to a GA4 property, or replaces it; the first import follows
// on the scheduler's next run.
func (c *SiteDomainAdminClient) ConnectAnalytics(
	ctx context.Context, domainUUID string, req AdminSiteAnalyticsConnectionRequest,
) (*AdminSiteAnalyticsConnection, error) {
	var connection AdminSiteAnalyticsConnection
	err := c.request(ctx, http.MethodPut, siteAnalyticsGA4Path(domainUUID), req, &connection)
	return &connection, err
}

// DisconnectAnalytics removes the domain's GA4 property and its imported rows. A domain with no
// property answers 404 with SiteAnalyticsConnectionNotFoundCode.
func (c *SiteDomainAdminClient) DisconnectAnalytics(ctx context.Context, domainUUID string) error {
	return c.request(ctx, http.MethodDelete, siteAnalyticsGA4Path(domainUUID), nil, nil)
}

func siteAnalyticsGA4Path(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/analytics/" + string(AdminSiteAnalyticsProviderGA4)
}
