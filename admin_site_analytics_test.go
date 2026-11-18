package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteAnalyticsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-analytics.json", key)
}

// GA4 connects, reads and disconnects on the domain's analytics routes; a measure GA4 did not
// report stays nil, and a domain with no property reads an empty report with no connection.
func TestSiteAnalyticsRoutesAndMissingMeasures(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteAnalyticsFixture(t, "connectionPending"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	connection, err := domains.ConnectAnalytics(ctx, "d-1", AdminSiteAnalyticsConnectionRequest{PropertyID: "properties/123456789"})
	if err != nil || connection.Status != AdminSiteAnalyticsPending || connection.Provider != AdminSiteAnalyticsProviderGA4 {
		t.Fatalf("connection = %+v, %v", connection, err)
	}
	doer.body = siteAnalyticsFixture(t, "report")
	report, err := domains.GetAnalytics(ctx, "d-1")
	if err != nil || report.Connection == nil || report.Connection.Status != AdminSiteAnalyticsActive || !report.KeyEventsRecorded ||
		*report.LandingPages[0].KeyEvents != 34 || report.LandingPages[1].KeyEvents != nil ||
		report.KeyEvents[0].Sessions != nil || report.Unavailable[0].Report != AdminSiteAnalyticsEcommerce {
		t.Fatalf("report = %+v, %v", report, err)
	}
	doer.body = siteAnalyticsFixture(t, "reportNoKeyEvents")
	quiet, err := domains.GetAnalytics(ctx, "d-1")
	if err != nil || quiet.KeyEventsRecorded || quiet.LandingPages[0].KeyEvents != nil || len(quiet.KeyEvents) != 0 {
		t.Fatalf("no key events = %+v, %v", quiet, err)
	}
	doer.body = siteAnalyticsFixture(t, "reportNotConnected")
	none, err := domains.GetAnalytics(ctx, "d-1")
	if err != nil || none.Connection != nil || none.WindowFrom != "" || len(none.LandingPages) != 0 {
		t.Fatalf("not connected = %+v, %v", none, err)
	}
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DisconnectAnalytics(ctx, "d-1"); err != nil {
		t.Fatal(err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"PUT " + siteTestBase + "/analytics/ga4", "GET " + siteTestBase + "/analytics",
		"GET " + siteTestBase + "/analytics", "GET " + siteTestBase + "/analytics",
		"DELETE " + siteTestBase + "/analytics/ga4",
	})
	assertSiteBody(t, doer.requests[0], siteAnalyticsFixture(t, "connectionRequest"))
	doer.status, doer.body = http.StatusBadRequest, siteAnalyticsFixture(t, "propertyInvalidProblem")
	_, err = domains.ConnectAnalytics(ctx, "d-1", AdminSiteAnalyticsConnectionRequest{PropertyID: "G-ABC123"})
	assertSiteProblemCode(t, err, http.StatusBadRequest, SiteAnalyticsPropertyInvalidCode)
	doer.status, doer.body = http.StatusNotFound, siteAnalyticsFixture(t, "connectionNotFoundProblem")
	assertSiteProblemCode(t, domains.DisconnectAnalytics(ctx, "d-1"), http.StatusNotFound, SiteAnalyticsConnectionNotFoundCode)
}
