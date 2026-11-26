package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteReportingFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-insights-reporting.json", key)
}

func TestSiteReportDecodesAbsentMeasurementsAsNil(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteReportingFixture(t, "report"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains

	report, err := domains.GetReport(context.Background(), "d-1")
	if err != nil {
		t.Fatal(err)
	}
	windows := report.Uptime.Windows
	if report.Host != "example.com" || report.Template != AdminSiteReportTemplateFull || report.Uptime.Status != AdminSiteUptimeUnknown || report.Uptime.ResponseMs != nil ||
		windows.Last24h.UpPercent != nil || windows.Last7d.UpPercent == nil || *windows.Last7d.UpPercent != 90 {
		t.Fatalf("uptime = %+v", report.Uptime)
	}
	speed := report.Speed[0]
	if speed.Scores.Performance == nil || *speed.Scores.Performance != 83 || speed.Scores.Accessibility != nil ||
		speed.Metrics.TotalBlockingTimeMs != nil || speed.Metrics.CumulativeLayoutShift == nil {
		t.Fatalf("speed = %+v", speed)
	}
	if !report.Search.Connected || len(report.Search.StrikingDistance) != 1 || report.Certificate.ValidTo != nil ||
		report.Audit.AuditedAt != nil || report.Audit.Findings[0].Severity != "high" {
		t.Fatalf("report = %+v", report)
	}
	assertSiteRequests(t, doer.requests, []string{"GET " + siteTestBase + "/report"})
}

func TestSiteUptimeAndLighthouseHistoryRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteReportingFixture(t, "uptime"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	uptime, err := domains.GetUptimeHistory(ctx, "d-1", 7)
	if err != nil || uptime.ResponseDays != 7 || uptime.MonitoredSince != nil || uptime.Hours[0].AvgMs != nil ||
		uptime.Days[0].UpPercent != nil {
		t.Fatalf("uptime = %+v, %v", uptime, err)
	}
	doer.body = siteReportingFixture(t, "lighthouse")
	lighthouse, err := domains.GetLighthouseHistory(ctx, "d-1", 0)
	if err != nil || len(lighthouse.Runs) != 2 || lighthouse.URL != "https://example.com/" || lighthouse.Pages == nil {
		t.Fatalf("lighthouse = %+v, %v", lighthouse, err)
	}
	ok, failed := lighthouse.Runs[0], lighthouse.Runs[1]
	if ok.Outcome != AdminSiteLighthouseSuccess || ok.Scores.SEO == nil || *ok.Scores.SEO != 95 ||
		failed.Outcome != AdminSiteLighthouseFailed || failed.Failure != "timeout" ||
		failed.Scores != (AdminSiteLighthouseScores{}) || failed.Metrics != (AdminSiteLighthouseMetrics{}) {
		t.Fatalf("runs = %+v", lighthouse.Runs)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/uptime-history?responseDays=7", "GET " + siteTestBase + "/lighthouse",
	})
}

func TestSiteReportShareRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteReportingFixture(t, "shares"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	shares, err := domains.ListReportShares(ctx, "d-1")
	if err != nil || len(shares.Shares) != 1 || shares.Shares[0].ShareUUID != "share-1" {
		t.Fatalf("shares = %+v, %v", shares, err)
	}
	doer.status, doer.body = http.StatusCreated, siteReportingFixture(t, "created")
	created, err := domains.CreateReportShare(ctx, "d-1", AdminSiteReportShareRequest{ExpiresInDays: 14})
	if err != nil || created.Token != "test-only-share-token" || created.Share.ExpiresAt.IsZero() {
		t.Fatalf("created = %+v, %v", created, err)
	}
	assertSiteBody(t, doer.requests[1], `{"expiresInDays":14}`)
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.RevokeReportShare(ctx, "d-1", "share-1"); err != nil {
		t.Fatal(err)
	}
	shares2 := siteTestBase + "/report-shares"
	assertSiteRequests(t, doer.requests, []string{"GET " + shares2, "POST " + shares2, "DELETE " + shares2 + "/share-1"})
}
