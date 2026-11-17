package custd

import (
	"context"
	"net/http"
	"testing"
)

const siteDomainsTestBase = "http://localhost:8080/api/v1/site-domains"

func siteWebDataFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-web-data.json", key)
}

func TestSiteWebDomainsAndPagesRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteWebDataFixture(t, "webDomains"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	web, err := domains.GetWebDomains(ctx, "s-1")
	if err != nil || len(web.Domains) != 3 {
		t.Fatalf("web = %+v, %v", web, err)
	}
	own, rival, fresh := web.Domains[0], web.Domains[1], web.Domains[2]
	if own.Role != AdminSiteWebRoleOwn || own.TrancoBand != AdminSiteTrancoBandTop10K || own.TrancoRank == nil ||
		rival.TrancoBand != AdminSiteTrancoBandUnranked || rival.TrancoRank != nil || rival.SitemapStatus != AdminSiteWebStatusFailed ||
		fresh.TrancoBand != AdminSiteTrancoBandUnknown || fresh.InventoryAt != nil || fresh.PageCount != 0 {
		t.Fatalf("domains = %+v", web.Domains)
	}
	doer.body = siteWebDataFixture(t, "webPages")
	pages, err := domains.ListWebPages(ctx, "rival.example", AdminSiteWebPagesQuery{})
	if err != nil || pages.Next != "https://rival.example/pricing" || pages.Pages[0].Sources[1] != AdminSiteWebSourceWayback {
		t.Fatalf("pages = %+v, %v", pages, err)
	}
	doer.body = siteWebDataFixture(t, "webPagesLast")
	last, err := domains.ListWebPages(ctx, "rival.example",
		AdminSiteWebPagesQuery{SiteUUID: "s-1", After: pages.Next, Limit: 10})
	if err != nil || last.Next != "" || len(last.Pages) != 0 {
		t.Fatalf("last = %+v, %v", last, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteDomainsTestBase + "/web-domains?siteUuid=s-1",
		"GET " + siteDomainsTestBase + "/web-pages?domain=rival.example",
		"GET " + siteDomainsTestBase + "/web-pages?after=https%3A%2F%2Frival.example%2Fpricing&domain=rival.example&limit=10&siteUuid=s-1",
	})
}

func TestSiteTopicTrendsAndContentGap(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteWebDataFixture(t, "topicTrends"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	trends, err := domains.GetTopicTrends(ctx, "")
	if err != nil || len(trends.Trends) != 3 {
		t.Fatalf("trends = %+v, %v", trends, err)
	}
	found, failed := trends.Trends[0], trends.Trends[2]
	if found.Status != AdminSiteTrendStatusFound || found.Source != AdminSiteTrendSourceBrand || found.Months[1].Views != 0 ||
		failed.Status != AdminSiteTrendStatusFailed || failed.CheckedAt != nil || failed.LastFailure == "" {
		t.Fatalf("trends = %+v", trends.Trends)
	}
	doer.body = siteWebDataFixture(t, "contentGap")
	gap, err := domains.GetContentGap(ctx, "s-1", "seo")
	if err != nil || gap.Seed != "seo" || gap.OwnCoverage != AdminSiteGapCoverageAvailable || !gap.Partial ||
		gap.Topics[0].Pages[0].Domain != "rival.example" {
		t.Fatalf("gap = %+v, %v", gap, err)
	}
	doer.body = siteWebDataFixture(t, "contentGapUnavailable")
	unavailable, err := domains.GetContentGap(ctx, "", "")
	if err != nil || unavailable.OwnCoverage != AdminSiteGapCoverageUnavailable || unavailable.Seed != "" {
		t.Fatalf("unavailable = %+v, %v", unavailable, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteDomainsTestBase + "/topic-trends",
		"GET " + siteDomainsTestBase + "/content-gap?seed=seo&siteUuid=s-1",
		"GET " + siteDomainsTestBase + "/content-gap",
	})
}

func TestSiteKeywordListRoutesAndProblems(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteWebDataFixture(t, "keywordLists"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	req := AdminSiteKeywordListRequest{Name: "Pricing", SeedTerm: "pricing", Keywords: []string{"pricing tool", "price checker", "seo audit"}}

	lists, err := domains.ListKeywordLists(ctx, "s-1")
	if err != nil || len(lists.Lists) != 2 || lists.Lists[0].SeedTerm != "" || lists.Lists[1].SeedTerm != "pricing" {
		t.Fatalf("lists = %+v, %v", lists, err)
	}
	doer.body = siteWebDataFixture(t, "keywordList")
	if list, err := domains.SaveKeywordList(ctx, "s-1", req); err != nil || len(list.Keywords) != 3 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	doer.body = siteWebDataFixture(t, "keywordClusters")
	clusters, err := domains.GetKeywordClusters(ctx, "", "Pricing")
	measured, unmeasured := clusters.Clusters[0], clusters.Clusters[1]
	if err != nil || measured.Clicks == nil || *measured.Clicks != 12 || measured.AvgPosition == nil ||
		unmeasured.Clicks != nil || unmeasured.Impressions != nil || unmeasured.AvgPosition != nil ||
		clusters.StrikingDistance[0].Position != 6.5 {
		t.Fatalf("clusters = %+v, %v", clusters, err)
	}
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DeleteKeywordList(ctx, "s-1", "Pricing"); err != nil {
		t.Fatal(err)
	}
	lists2 := siteDomainsTestBase + "/keyword-lists"
	assertSiteRequests(t, doer.requests, []string{
		"GET " + lists2 + "?siteUuid=s-1", "PUT " + lists2 + "?siteUuid=s-1",
		"GET " + siteDomainsTestBase + "/keyword-clusters?name=Pricing", "DELETE " + lists2 + "?name=Pricing&siteUuid=s-1",
	})
	assertSiteBody(t, doer.requests[1], siteWebDataFixture(t, "keywordListRequest"))
	doer.status, doer.body = http.StatusNotFound, siteWebDataFixture(t, "keywordListNotFoundProblem")
	assertSiteProblemCode(t, domains.DeleteKeywordList(ctx, "s-1", "Gone"), http.StatusNotFound, SiteKeywordListNotFoundCode)
	doer.status, doer.body = http.StatusConflict, siteWebDataFixture(t, "keywordListLimitProblem")
	_, err = domains.SaveKeywordList(ctx, "s-1", AdminSiteKeywordListRequest{Name: "More"})
	assertSiteProblemCode(t, err, http.StatusConflict, SiteKeywordListLimitCode)
	assertSiteBody(t, doer.requests[5], `{"name":"More","keywords":[]}`)
}
