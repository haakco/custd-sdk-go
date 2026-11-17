package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteSavedReportsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-saved-reports.json", key)
}

func TestSiteReportTemplateAndDaysQuery(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSavedReportsFixture(t, "executiveReport"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains

	report, err := domains.GetReportWithOptions(context.Background(), "d-1",
		AdminSiteReportOptions{Template: AdminSiteReportTemplateExecutive, Days: 14})
	if err != nil || report.Template != AdminSiteReportTemplateExecutive || report.Search.Days != 14 {
		t.Fatalf("report = %+v, %v", report, err)
	}
	assertSiteRequests(t, doer.requests, []string{"GET " + siteTestBase + "/report?days=14&template=executive"})
}

func TestSiteSavedReportRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSavedReportsFixture(t, "savedReportList"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	req := AdminSiteSavedReportRequest{Name: "November client summary", Template: AdminSiteReportTemplateExecutive, Days: 14}

	list, err := domains.ListSavedReports(ctx, "d-1")
	if err != nil || len(list.Reports) != 2 || list.MaxReports != AdminSiteMaxSavedReports || list.Reports[1].LiveShares != 0 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	doer.status, doer.body = http.StatusCreated, siteSavedReportsFixture(t, "savedReport")
	saved, err := domains.SaveReport(ctx, "d-1", req)
	if err != nil || saved.ReportUUID != "c9f0f895-fb98-4b91-a2f1-2b0d6e3c5a33" || saved.LiveShares != 1 ||
		saved.Report.Template != AdminSiteReportTemplateExecutive || saved.Report.Host != "example.com" {
		t.Fatalf("saved = %+v, %v", saved, err)
	}
	doer.status = http.StatusOK
	if _, err := domains.GetSavedReport(ctx, "d-1", "r-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := domains.ReplaceSavedReport(ctx, "d-1", "r-1", req); err != nil {
		t.Fatal(err)
	}
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DeleteSavedReport(ctx, "d-1", "r-1"); err != nil {
		t.Fatal(err)
	}
	base := siteTestBase + "/saved-reports"
	assertSiteRequests(t, doer.requests, []string{
		"GET " + base, "POST " + base, "GET " + base + "/r-1", "PUT " + base + "/r-1", "DELETE " + base + "/r-1",
	})
	assertSiteBody(t, doer.requests[1], siteSavedReportsFixture(t, "savedReportRequest"))
	assertSiteBody(t, doer.requests[3], siteSavedReportsFixture(t, "savedReportRequest"))
	if _, err := domains.SaveReport(ctx, "d-1", AdminSiteSavedReportRequest{Name: "Plain"}); err != nil {
		t.Fatal(err)
	}
	assertSiteBody(t, doer.requests[5], `{"name":"Plain"}`)
	for key, code := range map[string]string{
		"savedReportNameTakenProblem": SavedSiteReportNameTakenCode, "savedReportLimitProblem": SavedSiteReportLimitCode,
	} {
		doer.status, doer.body = http.StatusConflict, siteSavedReportsFixture(t, key)
		_, err := domains.SaveReport(ctx, "d-1", req)
		assertSiteProblemCode(t, err, http.StatusConflict, code)
	}
}

func TestSiteSavedReportShares(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSavedReportsFixture(t, "reportShares"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	shares, err := domains.ListReportShares(ctx, "d-1")
	if err != nil || shares.Shares[0].ReportUUID != "c9f0f895-fb98-4b91-a2f1-2b0d6e3c5a33" || shares.Shares[1].ReportUUID != "" {
		t.Fatalf("shares = %+v, %v", shares, err)
	}
	doer.status, doer.body = http.StatusCreated, siteSavedReportsFixture(t, "reportShareCreated")
	created, err := domains.CreateReportShare(ctx, "d-1",
		AdminSiteReportShareRequest{ExpiresInDays: 14, ReportUUID: "c9f0f895-fb98-4b91-a2f1-2b0d6e3c5a33"})
	if err != nil || created.Share.ReportUUID == "" || created.Token == "" {
		t.Fatalf("created = %+v, %v", created, err)
	}
	assertSiteBody(t, doer.requests[1], siteSavedReportsFixture(t, "reportShareRequest"))
}
