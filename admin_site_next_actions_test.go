package custd

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func siteNextActionsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-next-actions.json", key)
}

// NextActions and PortfolioSummary read the server's recorded shapes, and a kind or place this
// SDK does not name still decodes.
func TestSiteNextActionsAndPortfolio(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteNextActionsFixture(t, "nextActions"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	next, err := domains.NextActions(ctx, "d-1")
	if err != nil || next.Status != AdminSiteUptimeDown || next.ActionCount != 6 || len(next.Actions) != AdminSiteNextActionsMax ||
		next.Actions[0].Kind != AdminSiteNextActionOutage || next.Actions[2].Target.Place != AdminSiteNextActionPlaceSearchSettings ||
		next.CheckedAt == nil || len(next.NotChecked) != 1 || next.NotChecked[0] != AdminSiteNotCheckedSearch {
		t.Fatalf("next = %+v, %v", next, err)
	}
	doer.body = siteNextActionsFixture(t, "nextActionsFindings")
	next, err = domains.NextActions(ctx, "d-1")
	if err != nil || next.Actions[0].AffectedCount != 2 || next.Actions[0].Target.Place != AdminSiteNextActionPlaceFinding ||
		next.Actions[1].Target.Place != AdminSiteNextActionPlacePage || !next.Actions[1].Recheckable ||
		next.Actions[1].Rule != "title.missing" {
		t.Fatalf("findings = %+v, %v", next, err)
	}
	doer.body = siteNextActionsFixture(t, "portfolioSummary")
	summary, err := domains.PortfolioSummary(ctx)
	if err != nil || len(summary.Domains) != 4 || summary.Domains[0].FirstAction.Kind != AdminSiteNextActionVerify ||
		summary.Domains[1].OpenFindings.High != 4 || summary.Domains[3].FirstAction != nil {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	doer.body = strings.Replace(siteNextActionsFixture(t, "nextActionsPending"), `"verify"`, `"suggestion-ready"`, 1)
	if next, err = domains.NextActions(ctx, "d-1"); err != nil || next.Actions[0].Kind != "suggestion-ready" {
		t.Fatalf("unknown kind = %+v, %v", next, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/next-actions", "GET " + siteTestBase + "/next-actions",
		"GET http://localhost:8080/api/v1/site-domains/portfolio-summary", "GET " + siteTestBase + "/next-actions",
	})
}

// The report carries the same first three actions as the read; an older report without them
// decodes with none.
func TestSiteReportNextActions(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteNextActionsFixture(t, "reportWithNextActions"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	report, err := domains.GetReport(context.Background(), "d-1")
	if err != nil || len(report.NextActions) != AdminSiteNextActionsMax || report.NextActions[0].Kind != AdminSiteNextActionOutage {
		t.Fatalf("report = %+v, %v", report.NextActions, err)
	}
	read := siteNextActionsFixture(t, "nextActions")
	if !strings.Contains(read, `"title": "`+report.NextActions[1].Title+`"`) {
		t.Fatalf("report action %q is not the read's", report.NextActions[1].Title)
	}
}
