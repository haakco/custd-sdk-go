package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// AdminSiteInsightsExportSchema names the export document's version.
const AdminSiteInsightsExportSchema = "custd.site-insights-export.v1"

// AdminSiteSearchMetricsExportMaxLimit bounds one ExportSearchMetrics page, and is its default.
const AdminSiteSearchMetricsExportMaxLimit = 10000

// The dimensions of AdminSiteSearchMetricDay.Dimension; total has an empty Key.
const (
	AdminSiteSearchMetricTotal   = "total"
	AdminSiteSearchMetricPage    = "page"
	AdminSiteSearchMetricQuery   = "query"
	AdminSiteSearchMetricCountry = "country"
	AdminSiteSearchMetricDevice  = "device"
)

// AdminSiteInsightsExport is everything Site Insights holds for one domain and its Site, as the
// typed reads return it. Daily search metrics are paged separately with ExportSearchMetrics;
// Search.MetricRows says how many there are. Secrets and other tenants' pooled data are never
// part of it.
type AdminSiteInsightsExport struct {
	Schema     string                            `json:"schema"`
	ExportedAt time.Time                         `json:"exportedAt"`
	Domain     AdminSiteDomain                   `json:"domain"`
	Monitoring AdminSiteInsightsMonitoringExport `json:"monitoring"`
	Search     AdminSiteInsightsSearchExport     `json:"search"`
	Reports    AdminSiteInsightsReportsExport    `json:"reports"`
	Site       AdminSiteInsightsSiteExport       `json:"site"`
}

// AdminSiteInsightsMonitoringExport is the domain's findings (at most 10000, most severe first;
// FindingRows is how many there are), incidents, uptime and one Lighthouse history per measured
// page, home page first.
type AdminSiteInsightsMonitoringExport struct {
	Findings    AdminSiteFindingList         `json:"findings"`
	FindingRows int                          `json:"findingRows"`
	Incidents   AdminSiteIncidentList        `json:"incidents"`
	Uptime      AdminSiteUptimeHistory       `json:"uptime"`
	Lighthouse  []AdminSiteLighthouseHistory `json:"lighthouse"`
}

// AdminSiteInsightsSearchExport is the domain's rankings, URL inspections, IndexNow state and the
// count of daily search metric rows.
type AdminSiteInsightsSearchExport struct {
	Rankings    AdminSiteRankings             `json:"rankings"`
	Inspections AdminSiteSearchInspectionList `json:"inspections"`
	IndexNow    AdminSiteIndexNowState        `json:"indexNow"`
	MetricRows  int                           `json:"metricRows"`
}

// AdminSiteInsightsReportsExport is every saved report with its content and every campaign link,
// archived ones included.
type AdminSiteInsightsReportsExport struct {
	SavedReports  []AdminSiteSavedReport    `json:"savedReports"`
	CampaignLinks AdminSiteCampaignLinkList `json:"campaignLinks"`
}

// AdminSiteInsightsSiteExport is what the domain's Site holds: project context, keyword lists,
// own and competitor web domains and topic trends.
type AdminSiteInsightsSiteExport struct {
	ProjectContext AdminSiteProjectContext `json:"projectContext"`
	KeywordLists   AdminSiteKeywordLists   `json:"keywordLists"`
	WebDomains     AdminSiteWebDomains     `json:"webDomains"`
	TopicTrends    AdminSiteTopicTrends    `json:"topicTrends"`
}

// AdminSiteSearchMetricDay is one provider's figures for one day and one dimension key.
type AdminSiteSearchMetricDay struct {
	Provider    AdminSiteSearchProvider `json:"provider"`
	Day         string                  `json:"day"`
	Dimension   string                  `json:"dimension"`
	Key         string                  `json:"key"`
	Clicks      int64                   `json:"clicks"`
	Impressions int64                   `json:"impressions"`
	AvgPosition float64                 `json:"avgPosition"`
}

// AdminSiteSearchMetricsExportPage is one page of a domain's stored daily search metrics, ordered
// by provider, day, dimension and key; read on while Offset+len(Rows) is under Total.
type AdminSiteSearchMetricsExportPage struct {
	DomainUUID string                     `json:"domainUuid"`
	Offset     int                        `json:"offset"`
	Limit      int                        `json:"limit"`
	Total      int                        `json:"total"`
	Rows       []AdminSiteSearchMetricDay `json:"rows"`
}

// ExportSiteInsights returns everything Site Insights holds for the domain as one typed document.
func (c *SiteDomainAdminClient) ExportSiteInsights(ctx context.Context, domainUUID string) (*AdminSiteInsightsExport, error) {
	var export AdminSiteInsightsExport
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/export", nil, &export)
	return &export, err
}

// ExportSearchMetrics reads one page of the domain's stored daily search metrics. A zero limit lets
// the server send AdminSiteSearchMetricsExportMaxLimit rows; a negative offset or a limit past the
// maximum is a 400.
func (c *SiteDomainAdminClient) ExportSearchMetrics(
	ctx context.Context, domainUUID string, offset, limit int,
) (*AdminSiteSearchMetricsExportPage, error) {
	values := url.Values{}
	if offset != 0 {
		values.Set("offset", strconv.Itoa(offset))
	}
	if limit != 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	var page AdminSiteSearchMetricsExportPage
	err := c.request(ctx, http.MethodGet, siteQueryPath(siteDomainPath(domainUUID)+"/export/search-metrics", values), nil, &page)
	return &page, err
}
