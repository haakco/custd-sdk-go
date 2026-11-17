package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSiteSearchRowsQueryAndUnavailable(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "searchRows"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	page, err := domains.ListSearchRows(ctx, "d-1", AdminSiteSearchRowsQuery{Dimension: AdminSiteSearchDimensionCountry})
	if err != nil {
		t.Fatal(err)
	}
	if page.Provider != "" || page.Total != 1 || page.Rows[0].Key != "zaf" || page.Rows[0].Totals.Position != 6.5 ||
		len(page.Unavailable) != 1 || page.Unavailable[0].Provider != AdminSiteSearchProviderBing {
		t.Fatalf("page = %+v", page)
	}
	doer.body = sitePageToolsFixture(t, "searchRowsNoPosition")
	bing, err := domains.ListSearchRows(ctx, "d-1", AdminSiteSearchRowsQuery{
		Dimension: AdminSiteSearchDimensionQuery, Provider: AdminSiteSearchProviderBing, Days: 7, Offset: 100, Limit: 50,
	})
	if err != nil || bing.Provider != AdminSiteSearchProviderBing || bing.Offset != 100 || len(bing.Unavailable) != 0 {
		t.Fatalf("bing = %+v, %v", bing, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/search/rows?dimension=country",
		"GET " + siteTestBase + "/search/rows?days=7&dimension=query&limit=50&offset=100&provider=bing",
	})
}

func TestSiteSearchInspectionRoutesAndAbsentResults(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, sitePageToolsFixture(t, "searchInspectionQueued"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	queued, err := domains.RequestSearchInspection(ctx, "d-1", "https://example.com/plans")
	if err != nil || queued.Status != AdminSiteSearchInspectionQueued || queued.InspectedAt != nil || queued.Verdict != "" {
		t.Fatalf("queued = %+v, %v", queued, err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "searchInspectionRequest"))
	doer.status, doer.body = http.StatusOK, sitePageToolsFixture(t, "searchInspectionList")
	list, err := domains.ListSearchInspections(ctx, "d-1")
	if err != nil || list.DailyLimit != 200 || list.UsedToday != 3 || len(list.Inspections) != 3 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	done, unknown, failed := list.Inspections[0], list.Inspections[1], list.Inspections[2]
	if done.Verdict != "PASS" || done.LastCrawlAt == nil || done.CrawledAs != "MOBILE" ||
		unknown.LastCrawlAt != nil || unknown.Verdict != "NEUTRAL" ||
		failed.Status != AdminSiteSearchInspectionFailed || failed.Failure == "" {
		t.Fatalf("inspections = %+v", list.Inspections)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST " + siteTestBase + "/search/inspections", "GET " + siteTestBase + "/search/inspections",
	})
	for _, tc := range []struct {
		key    string
		status int
		code   string
	}{
		{"searchInspectionUrlInvalidProblem", http.StatusBadRequest, SiteSearchInspectionURLInvalidCode},
		{"searchInspectionNotConnectedProblem", http.StatusConflict, SiteSearchInspectionNotConnectedCode},
		{"searchInspectionLimitProblem", http.StatusTooManyRequests, SiteSearchInspectionLimitCode},
	} {
		doer.status, doer.body = tc.status, sitePageToolsFixture(t, tc.key)
		_, err := domains.RequestSearchInspection(ctx, "d-1", "https://example.com/plans")
		assertSiteProblemCode(t, err, tc.status, tc.code)
	}
}

func TestSiteProjectContextRoutesAndEmptyLists(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "projectContext"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	var req AdminSiteProjectContextRequest
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "projectContextRequest")), &req); err != nil {
		t.Fatal(err)
	}

	set, err := domains.SetProjectContext(ctx, "s-1", req)
	if err != nil || set.UpdatedAt == nil || set.Markets[0] != (AdminSiteMarket{Country: "ZA", Language: "en"}) {
		t.Fatalf("set = %+v, %v", set, err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "projectContextRequest"))
	doer.body = sitePageToolsFixture(t, "projectContextEmpty")
	empty, err := domains.GetProjectContext(ctx, "")
	if err != nil || empty.UpdatedAt != nil || empty.BrandNames == nil || len(empty.Markets) != 0 {
		t.Fatalf("empty = %+v, %v", empty, err)
	}
	if _, err := domains.SetProjectContext(ctx, "", AdminSiteProjectContextRequest{}); err != nil {
		t.Fatal(err)
	}
	assertSiteBody(t, doer.requests[2], `{"brandNames":[],"competitorDomains":[],"markets":[]}`)
	base := "http://localhost:8080/api/v1/site-domains/project-context"
	assertSiteRequests(t, doer.requests, []string{"PUT " + base + "?siteUuid=s-1", "GET " + base, "PUT " + base})
}
