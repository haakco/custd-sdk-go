package custd

import (
	"context"
	"net/http"
	"testing"
)

// The export decodes every part into the SDK's own read types, and search metrics page with
// offset and limit sent only when set.
func TestSiteInsightsExportAndSearchMetricPages(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, recordedFixture(t, "site-export.json", "export"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	export, err := domains.ExportSiteInsights(ctx, "d-1")
	if err != nil || export.Schema != AdminSiteInsightsExportSchema || export.Domain.Host == "" ||
		export.Monitoring.FindingRows != len(export.Monitoring.Findings.Findings) || len(export.Monitoring.Lighthouse) != 1 ||
		len(export.Search.Rankings.Keywords) == 0 || export.Search.MetricRows != 2 || len(export.Reports.SavedReports) != 1 ||
		len(export.Reports.CampaignLinks.Links) == 0 || len(export.Site.KeywordLists.Lists) == 0 {
		t.Fatalf("export = %+v, %v", export, err)
	}
	doer.body = recordedFixture(t, "site-export.json", "searchMetrics")
	page, err := domains.ExportSearchMetrics(ctx, "d-1", 0, 0)
	if err != nil || page.Total != 2 || len(page.Rows) != 2 || page.Rows[0].Dimension != AdminSiteSearchMetricTotal ||
		page.Rows[0].Key != "" || page.Rows[1].Provider != AdminSiteSearchProviderGoogle || page.Rows[1].AvgPosition != 4.2 {
		t.Fatalf("page = %+v, %v", page, err)
	}
	if _, err := domains.ExportSearchMetrics(ctx, "d-1", 10000, 500); err != nil {
		t.Fatal(err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/export", "GET " + siteTestBase + "/export/search-metrics",
		"GET " + siteTestBase + "/export/search-metrics?limit=500&offset=10000",
	})
}
