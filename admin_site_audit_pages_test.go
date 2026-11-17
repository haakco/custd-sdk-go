package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteAuditPagesFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-audit-pages.json", key)
}

func TestSiteAuditPagesRoutesAndAbsentMeasurements(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteAuditPagesFixture(t, "auditPages"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	list, err := domains.ListAuditPages(ctx, "d-1", 0, 0)
	if err != nil || list.AuditedAt == nil || list.Total != 3 || len(list.Pages) != 3 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	home, redirected, orphan := list.Pages[0], list.Pages[1], list.Pages[2]
	if home.Indexability != AdminSiteAuditPageIndexable || home.Depth == nil || *home.Depth != 0 ||
		redirected.Indexability != AdminSiteAuditPageRedirected || redirected.FinalURL != "https://example.com/new" ||
		orphan.Depth != nil || orphan.InboundLinks == nil || *orphan.InboundLinks != 0 {
		t.Fatalf("pages = %+v", list.Pages)
	}
	doer.body = siteAuditPagesFixture(t, "auditPagesTruncated")
	truncated, err := domains.ListAuditPages(ctx, "d-1", 50, 50)
	broken := truncated.Pages[0]
	if err != nil || !truncated.Truncated || broken.Status != 0 || broken.Indexability != AdminSiteAuditPageError ||
		broken.Depth != nil || broken.InboundLinks != nil {
		t.Fatalf("truncated = %+v, %v", truncated, err)
	}
	doer.body = siteAuditPagesFixture(t, "auditPagesNotAudited")
	none, err := domains.ListAuditPages(ctx, "d-1", 0, AdminSiteAuditPagesMaxLimit)
	if err != nil || none.AuditedAt != nil || len(none.Pages) != 0 {
		t.Fatalf("not audited = %+v, %v", none, err)
	}
	base := siteTestBase + "/audit-pages"
	assertSiteRequests(t, doer.requests, []string{"GET " + base, "GET " + base + "?limit=50&offset=50", "GET " + base + "?limit=200"})
}

func TestSiteAuditDeltaDecodesEachRun(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteAuditPagesFixture(t, "auditDelta"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	delta, err := domains.GetAuditDelta(ctx, "d-1")
	if err != nil || delta.Run != AdminSiteAuditRunComplete || len(delta.New) != 1 || len(delta.Reopened) != 1 ||
		delta.New[0].Rule != SiteRuleBrokenFragment || delta.Reopened[0].Rule != SiteRuleSitemapBlockedURL ||
		!delta.Fixed[0].History.Fixed || delta.Fixed[0].History.FixedAt == nil {
		t.Fatalf("delta = %+v, %v", delta, err)
	}
	doer.body = siteAuditPagesFixture(t, "auditDeltaFailed")
	failed, err := domains.GetAuditDelta(ctx, "d-1")
	if err != nil || failed.Run != AdminSiteAuditRunFailed || failed.Failure == "" || len(failed.New) != 0 {
		t.Fatalf("failed = %+v, %v", failed, err)
	}
	doer.body = siteAuditPagesFixture(t, "auditDeltaNotAudited")
	none, err := domains.GetAuditDelta(ctx, "d-1")
	if err != nil || none.Run != AdminSiteAuditRunNotAudited || none.AuditedAt != nil {
		t.Fatalf("not audited = %+v, %v", none, err)
	}
	assertSiteRequests(t, doer.requests[:1], []string{"GET " + siteTestBase + "/audit-delta"})
}

func TestSiteLighthousePagesAndPageHistory(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteAuditPagesFixture(t, "lighthousePageHistory"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	history, err := domains.GetLighthousePageHistory(ctx, "d-1", 7, "https://example.com/pricing")
	if err != nil || history.URL != "https://example.com/pricing" || len(history.Runs) != 1 || len(history.Pages) != 2 {
		t.Fatalf("history = %+v, %v", history, err)
	}
	home, pricing := history.Pages[0], history.Pages[1]
	if home.PreviousScore == nil || *home.PreviousScore != 95 || home.RegressionOpen ||
		pricing.PreviousAt != nil || pricing.PreviousScore != nil || !pricing.RegressionOpen {
		t.Fatalf("pages = %+v", history.Pages)
	}
	doer.body = siteAuditPagesFixture(t, "lighthousePagesEmpty")
	empty, err := domains.GetLighthousePages(ctx, "d-1")
	if err != nil || len(empty.URLs) != 0 || empty.MaxPages != AdminSiteLighthouseMaxExtraPages {
		t.Fatalf("empty = %+v, %v", empty, err)
	}
	doer.body = siteAuditPagesFixture(t, "lighthousePages")
	set, err := domains.SetLighthousePages(ctx, "d-1", []string{"https://example.com/pricing", "https://example.com/blog"})
	if err != nil || len(set.URLs) != 2 || set.HomePage != "https://example.com/" {
		t.Fatalf("set = %+v, %v", set, err)
	}
	if _, err := domains.SetLighthousePages(ctx, "d-1", nil); err != nil {
		t.Fatal(err)
	}
	pages := siteTestBase + "/lighthouse/pages"
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/lighthouse?days=7&url=https%3A%2F%2Fexample.com%2Fpricing",
		"GET " + pages, "PUT " + pages, "PUT " + pages,
	})
	assertSiteBody(t, doer.requests[2], siteAuditPagesFixture(t, "lighthousePagesRequest"))
	assertSiteBody(t, doer.requests[3], siteAuditPagesFixture(t, "lighthousePagesClearRequest"))
	doer.status, doer.body = http.StatusBadRequest, siteAuditPagesFixture(t, "lighthousePageNotExtraProblem")
	_, err = domains.SetLighthousePages(ctx, "d-1", []string{"https://example.com/"})
	assertSiteProblemCode(t, err, http.StatusBadRequest, SiteLighthousePageNotExtraCode)
}
