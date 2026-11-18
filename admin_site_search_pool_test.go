package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteSearchPoolFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-search-pool.json", key)
}

func TestSiteSearchPoolSettingRoutesAndBodies(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSearchPoolFixture(t, "settingNotJoined"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	setting, err := domains.GetSearchPool(ctx)
	if err != nil || setting.Joined || setting.JoinedAt != nil || setting.MinContributors != AdminSiteSearchPoolMinContributors {
		t.Fatalf("setting = %+v, %v", setting, err)
	}
	doer.body = siteSearchPoolFixture(t, "settingJoined")
	joined, err := domains.SetSearchPool(ctx, true)
	if err != nil || !joined.Joined || joined.JoinedAt == nil || joined.WindowDays != AdminSiteSearchPoolWindowDays {
		t.Fatalf("joined = %+v, %v", joined, err)
	}
	doer.body = siteSearchPoolFixture(t, "settingNotJoined")
	if _, err := domains.SetSearchPool(ctx, false); err != nil {
		t.Fatal(err)
	}
	pool := siteDomainsTestBase + "/search-pool"
	assertSiteRequests(t, doer.requests, []string{"GET " + pool, "PUT " + pool, "PUT " + pool})
	assertSiteBody(t, doer.requests[1], siteSearchPoolFixture(t, "joinRequest"))
	assertSiteBody(t, doer.requests[2], siteSearchPoolFixture(t, "leaveRequest"))
	doer.status, doer.body = http.StatusForbidden, siteSearchPoolFixture(t, "siteBoundProblem")
	_, err = domains.SetSearchPool(ctx, true)
	assertSiteProblemCode(t, err, http.StatusForbidden, SiteSearchPoolSiteBoundCode)
	doer.status, doer.body = http.StatusBadRequest, siteSearchPoolFixture(t, "settingInvalidProblem")
	_, err = domains.SetSearchPool(ctx, true)
	assertSiteProblemCode(t, err, http.StatusBadRequest, SiteSearchPoolSettingInvalidCode)
}

func TestSiteSearchPoolQueriesBandsAndProblems(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSearchPoolFixture(t, "queries"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	page, err := domains.ListSearchPoolQueries(ctx, AdminSiteSearchPoolQueriesQuery{})
	if err != nil || page.ComputedAt == nil || page.From != "2026-10-18" || len(page.Queries) != 2 {
		t.Fatalf("page = %+v, %v", page, err)
	}
	busy, rare := page.Queries[0], page.Queries[1]
	if busy.Query != "pricing tool" || busy.DemandBand != AdminSiteSearchPoolDemand10KTo100K || busy.Positions.Top3 != 35 ||
		rare.DemandBand != AdminSiteSearchPoolDemand100To999 || rare.Positions.Top3 != 0 || rare.Positions.Beyond20 != 75 {
		t.Fatalf("queries = %+v", page.Queries)
	}
	doer.body = siteSearchPoolFixture(t, "queriesBeforeRollup")
	empty, err := domains.ListSearchPoolQueries(ctx, AdminSiteSearchPoolQueriesQuery{Search: "pricing", Limit: 10})
	if err != nil || empty.ComputedAt != nil || empty.From != "" || len(empty.Queries) != 0 {
		t.Fatalf("empty = %+v, %v", empty, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteDomainsTestBase + "/search-pool/queries",
		"GET " + siteDomainsTestBase + "/search-pool/queries?limit=10&search=pricing",
	})
	doer.status, doer.body = http.StatusForbidden, siteSearchPoolFixture(t, "notJoinedProblem")
	_, err = domains.ListSearchPoolQueries(ctx, AdminSiteSearchPoolQueriesQuery{})
	assertSiteProblemCode(t, err, http.StatusForbidden, SiteSearchPoolNotJoinedCode)
	doer.status, doer.body = http.StatusBadRequest, siteSearchPoolFixture(t, "queryInvalidProblem")
	_, err = domains.ListSearchPoolQueries(ctx, AdminSiteSearchPoolQueriesQuery{Limit: 500})
	assertSiteProblemCode(t, err, http.StatusBadRequest, SiteSearchPoolQueryInvalidCode)
}
