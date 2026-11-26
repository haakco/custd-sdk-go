package custd

import (
	"context"
	"time"
)

var siteDomainSurfaceRows = []surfaceRow{
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/incidents/{incidentUuid}/acknowledgment",
		path: "/api/v1/site-domains/a1/incidents/a2/acknowledgment",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.AcknowledgeIncident(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/campaign-links/{linkUuid}/archive",
		path: "/api/v1/site-domains/a1/campaign-links/a2/archive",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ArchiveCampaignLink(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/attestation",
		path: "/api/v1/site-domains/a1/attestation",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.Attest(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/maintenance/{windowUuid}",
		path: "/api/v1/site-domains/a1/maintenance/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CancelMaintenance(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/page-check",
		path: "/api/v1/site-domains/a1/page-check",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CheckPage(context.Background(), "a1", AdminSitePageCheckRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/page-wording",
		path: "/api/v1/site-domains/a1/page-wording",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ComparePageWording(context.Background(), "a1", AdminSitePageWordingQuery{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/wording-compare",
		path: "/api/v1/site-domains/a1/wording-compare",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CompareWording(context.Background(), "a1", AdminSiteWordingCompareRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/analytics/ga4",
		path: "/api/v1/site-domains/a1/analytics/ga4",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ConnectAnalytics(context.Background(), "a1", AdminSiteAnalyticsConnectionRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/search/{provider}",
		path: "/api/v1/site-domains/a1/search/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ConnectSearch(context.Background(), "a1", AdminSiteSearchProvider("a2"), AdminSiteSearchConnectionRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains",
		path: "/api/v1/site-domains",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.Create(context.Background(), AdminSiteDomainCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/campaign-links",
		path: "/api/v1/site-domains/a1/campaign-links",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CreateCampaignLink(context.Background(), "a1", AdminSiteCampaignLinkRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/heartbeats",
		path: "/api/v1/site-domains/a1/heartbeats",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CreateHeartbeat(context.Background(), "a1", AdminSiteHeartbeatCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/private-uptime",
		path: "/api/v1/site-domains/a1/private-uptime",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CreatePrivateUptime(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/report-shares",
		path: "/api/v1/site-domains/a1/report-shares",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CreateReportShare(context.Background(), "a1", AdminSiteReportShareRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}",
		path: "/api/v1/site-domains/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.Delete(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/heartbeats/{monitorUuid}",
		path: "/api/v1/site-domains/a1/heartbeats/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DeleteHeartbeat(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/keyword-lists",
		path: "/api/v1/site-domains/keyword-lists",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DeleteKeywordList(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/private-uptime/{monitorUuid}",
		path: "/api/v1/site-domains/a1/private-uptime/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DeletePrivateUptime(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/saved-reports/{reportUuid}",
		path: "/api/v1/site-domains/a1/saved-reports/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DeleteSavedReport(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/indexnow",
		path: "/api/v1/site-domains/a1/indexnow",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DisableIndexNow(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/analytics/ga4",
		path: "/api/v1/site-domains/a1/analytics/ga4",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DisconnectAnalytics(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/search/{provider}",
		path: "/api/v1/site-domains/a1/search/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.DisconnectSearch(context.Background(), "a1", AdminSiteSearchProvider("a2"))
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/finding-dismissals",
		path: "/api/v1/site-domains/a1/finding-dismissals",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.DismissFinding(context.Background(), "a1", AdminSiteFindingDismissalRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/indexnow",
		path: "/api/v1/site-domains/a1/indexnow",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.EnableIndexNow(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/export/search-metrics",
		path: "/api/v1/site-domains/a1/export/search-metrics",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ExportSearchMetrics(context.Background(), "a1", 2, 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/export",
		path: "/api/v1/site-domains/a1/export",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ExportSiteInsights(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}",
		path: "/api/v1/site-domains/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/ai-settings",
		path: "/api/v1/site-domains/ai-settings",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetAISettings(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/ai-owner-handoffs",
		path: "/api/v1/site-domains/ai-owner-handoffs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.CreateAIOwnerHandoff(context.Background(),
				SiteAIOwnerHandoffRequest{TenantSlug: "a1", OwnerSubject: "a2"})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/ai-usage",
		path: "/api/v1/site-domains/ai-usage",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetAIUsage(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/ai-visibility",
		path: "/api/v1/site-domains/a1/ai-visibility",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetAIVisibility(context.Background(), "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/alert-destination",
		path: "/api/v1/site-domains/alert-destination",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetAlertDestination(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/analytics",
		path: "/api/v1/site-domains/a1/analytics",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetAnalytics(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/audit-delta",
		path: "/api/v1/site-domains/a1/audit-delta",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetAuditDelta(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/content-gap",
		path: "/api/v1/site-domains/content-gap",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetContentGap(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/findings/{findingUuid}/search-impact",
		path: "/api/v1/site-domains/a1/findings/a2/search-impact",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetFindingSearchImpact(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/indexnow",
		path: "/api/v1/site-domains/a1/indexnow",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetIndexNow(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/keyword-clusters",
		path: "/api/v1/site-domains/keyword-clusters",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetKeywordClusters(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/lighthouse",
		path: "/api/v1/site-domains/a1/lighthouse",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetLighthouseHistory(context.Background(), "a1", 2)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/lighthouse",
		path: "/api/v1/site-domains/a1/lighthouse",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetLighthousePageHistory(context.Background(), "a1", 2, "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/lighthouse/pages",
		path: "/api/v1/site-domains/a1/lighthouse/pages",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetLighthousePages(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/page-facts",
		path: "/api/v1/site-domains/a1/page-facts",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetPageFacts(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/page-rechecks/{recheckUuid}",
		path: "/api/v1/site-domains/a1/page-rechecks/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetPageRecheck(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/page-search-queries",
		path: "/api/v1/site-domains/a1/page-search-queries",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetPageSearchQueries(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/private-uptime",
		path: "/api/v1/site-domains/a1/private-uptime",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetPrivateUptime(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/project-context",
		path: "/api/v1/site-domains/project-context",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetProjectContext(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/rankings",
		path: "/api/v1/site-domains/a1/rankings",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetRankings(context.Background(), "a1", 2)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/referring-domains",
		path: "/api/v1/site-domains/referring-domains",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetReferringDomains(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/regional-trends",
		path: "/api/v1/site-domains/regional-trends",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetRegionalTrends(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/report",
		path: "/api/v1/site-domains/a1/report",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetReport(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/report",
		path: "/api/v1/site-domains/a1/report",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetReportWithOptions(context.Background(), "a1", AdminSiteReportOptions{Template: "a2", Days: 7})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/saved-reports/{reportUuid}",
		path: "/api/v1/site-domains/a1/saved-reports/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetSavedReport(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/search-account",
		path: "/api/v1/site-domains/search-account",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetSearchAccount(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/search",
		path: "/api/v1/site-domains/a1/search",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetSearchPerformance(context.Background(), "a1", 2)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/search-pool",
		path: "/api/v1/site-domains/search-pool",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetSearchPool(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/search-value",
		path: "/api/v1/site-domains/a1/search-value",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetSearchValue(context.Background(), "a1", []string{"a2"}, 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/status-page",
		path: "/api/v1/site-domains/a1/status-page",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetStatusPageSettings(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/topic-trends",
		path: "/api/v1/site-domains/topic-trends",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetTopicTrends(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/traffic",
		path: "/api/v1/site-domains/a1/traffic",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetTraffic(context.Background(), "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/uptime-history",
		path: "/api/v1/site-domains/a1/uptime-history",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetUptimeHistory(context.Background(), "a1", 2)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/web-domains",
		path: "/api/v1/site-domains/web-domains",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.GetWebDomains(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains",
		path: "/api/v1/site-domains",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.List(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/audit-pages",
		path: "/api/v1/site-domains/a1/audit-pages",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListAuditPages(context.Background(), "a1", 2, 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/campaign-links",
		path: "/api/v1/site-domains/a1/campaign-links",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListCampaignLinks(context.Background(), "a1", true)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/finding-dismissals",
		path: "/api/v1/site-domains/a1/finding-dismissals",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListFindingDismissals(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/findings",
		path: "/api/v1/site-domains/a1/findings",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListFindings(context.Background(), "a1", AdminSiteFindingState("a2"))
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/heartbeats",
		path: "/api/v1/site-domains/a1/heartbeats",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListHeartbeats(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/incidents",
		path: "/api/v1/site-domains/a1/incidents",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListIncidents(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/keyword-lists",
		path: "/api/v1/site-domains/keyword-lists",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListKeywordLists(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/maintenance",
		path: "/api/v1/site-domains/a1/maintenance",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListMaintenance(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/page-rechecks",
		path: "/api/v1/site-domains/a1/page-rechecks",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListPageRechecks(context.Background(), "a1", nil)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/probes",
		path: "/api/v1/site-domains/a1/probes",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListProbes(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/report-shares",
		path: "/api/v1/site-domains/a1/report-shares",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListReportShares(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/saved-reports",
		path: "/api/v1/site-domains/a1/saved-reports",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListSavedReports(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/search/inspections",
		path: "/api/v1/site-domains/a1/search/inspections",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListSearchInspections(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/search-pool/queries",
		path: "/api/v1/site-domains/search-pool/queries",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListSearchPoolQueries(context.Background(), AdminSiteSearchPoolQueriesQuery{Search: "a1", Limit: 5})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/search/rows",
		path: "/api/v1/site-domains/a1/search/rows",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListSearchRows(context.Background(), "a1", AdminSiteSearchRowsQuery{Dimension: "a2", Provider: AdminSiteSearchProviderGoogle, Days: 7, Offset: 1, Limit: 5})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/web-pages",
		path: "/api/v1/site-domains/web-pages",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ListWebPages(context.Background(), "a1", AdminSiteWebPagesQuery{SiteUUID: "a2", After: "a3", Limit: 5})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/{domainUuid}/next-actions",
		path: "/api/v1/site-domains/a1/next-actions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.NextActions(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-domains/portfolio-summary",
		path: "/api/v1/site-domains/portfolio-summary",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.PortfolioSummary(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/status-page/updates",
		path: "/api/v1/site-domains/a1/status-page/updates",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.PostStatusPageUpdate(context.Background(), "a1", AdminSiteStatusUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/status-page/preview",
		path: "/api/v1/site-domains/a1/status-page/preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.PreviewStatusPage(context.Background(), "a1", AdminSiteStatusDraft{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/status-page",
		path: "/api/v1/site-domains/a1/status-page",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.PublishStatusPage(context.Background(), "a1", AdminSiteStatusDraft{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/saved-reports/{reportUuid}",
		path: "/api/v1/site-domains/a1/saved-reports/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ReplaceSavedReport(context.Background(), "a1", "a2", AdminSiteSavedReportRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/access-log-hours/replay",
		path: "/api/v1/admin/access-log-hours/replay",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ReplayAccessLogHour(context.Background(), time.Time{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/audit",
		path: "/api/v1/site-domains/a1/audit",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RequestAudit(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/page-rechecks",
		path: "/api/v1/site-domains/a1/page-rechecks",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RequestPageRecheck(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/private-uptime/check",
		path: "/api/v1/site-domains/a1/private-uptime/check",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RequestPrivateUptime(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/search/inspections",
		path: "/api/v1/site-domains/a1/search/inspections",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RequestSearchInspection(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/uptime-check",
		path: "/api/v1/site-domains/a1/uptime-check",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RequestUptimeCheck(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/finding-dismissals/{dismissalUuid}",
		path: "/api/v1/site-domains/a1/finding-dismissals/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.RestoreFinding(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/search/{provider}/retry",
		path: "/api/v1/site-domains/a1/search/a2/retry",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RetrySearchImport(context.Background(), "a1", AdminSiteSearchProvider("a2"))
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/report-shares/{shareUuid}",
		path: "/api/v1/site-domains/a1/report-shares/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.RevokeReportShare(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/heartbeats/{monitorUuid}/credential",
		path: "/api/v1/site-domains/a1/heartbeats/a2/credential",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RotateHeartbeat(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/private-uptime/{monitorUuid}/credential",
		path: "/api/v1/site-domains/a1/private-uptime/a2/credential",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RotatePrivateUptime(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/keyword-lists",
		path: "/api/v1/site-domains/keyword-lists",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SaveKeywordList(context.Background(), "a1", AdminSiteKeywordListRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/saved-reports",
		path: "/api/v1/site-domains/a1/saved-reports",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SaveReport(context.Background(), "a1", AdminSiteSavedReportRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/maintenance",
		path: "/api/v1/site-domains/a1/maintenance",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.ScheduleMaintenance(context.Background(), "a1", AdminSiteMaintenanceRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/ai-settings/consent",
		path: "/api/v1/site-domains/ai-settings/consent",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetAIConsent(context.Background(), SiteAIConsentRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/ai-settings",
		path: "/api/v1/site-domains/ai-settings",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetAISettings(context.Background(), SiteAISettingsRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/alert-destination",
		path: "/api/v1/site-domains/alert-destination",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetAlertDestination(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/browser-collector",
		path: "/api/v1/site-domains/a1/browser-collector",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetBrowserCollector(context.Background(), "a1", SiteBrowserCollectorRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/checks",
		path: "/api/v1/site-domains/a1/checks",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetCheckMode(context.Background(), "a1", AdminSiteCheckMode("a2"))
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/checks",
		path: "/api/v1/site-domains/a1/checks",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetCheckSettings(context.Background(), "a1", AdminSiteCheckSettings{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/lighthouse/pages",
		path: "/api/v1/site-domains/a1/lighthouse/pages",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetLighthousePages(context.Background(), "a1", nil)
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/probes",
		path: "/api/v1/site-domains/a1/probes",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetProbes(context.Background(), "a1", nil)
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/project-context",
		path: "/api/v1/site-domains/project-context",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetProjectContext(context.Background(), "a1", AdminSiteProjectContextRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/search-pool",
		path: "/api/v1/site-domains/search-pool",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SetSearchPool(context.Background(), false, "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/indexnow/submissions",
		path: "/api/v1/site-domains/a1/indexnow/submissions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.SubmitIndexNow(context.Background(), "a1", nil)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/robots-test",
		path: "/api/v1/site-domains/a1/robots-test",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.TestRobots(context.Background(), "a1", AdminSiteRobotsTestRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/rankings/keywords",
		path: "/api/v1/site-domains/a1/rankings/keywords",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.TrackKeyword(context.Background(), "a1", AdminSiteTrackKeywordRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/attestation",
		path: "/api/v1/site-domains/a1/attestation",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.Unattest(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/status-page",
		path: "/api/v1/site-domains/a1/status-page",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.UnpublishStatusPage(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/rankings/keywords/{keywordUuid}",
		path: "/api/v1/site-domains/a1/rankings/keywords/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SiteDomains.UntrackKeyword(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-domains/{domainUuid}/rankings/keywords/{keywordUuid}",
		path: "/api/v1/site-domains/a1/rankings/keywords/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.UpdateTrackedKeyword(context.Background(), "a1", "a2", AdminSiteTrackedKeywordUpdate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/verify",
		path: "/api/v1/site-domains/a1/verify",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.Verify(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/indexnow/verify",
		path: "/api/v1/site-domains/a1/indexnow/verify",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.VerifyIndexNowKey(context.Background(), "a1")
			return err
		},
	},
}
