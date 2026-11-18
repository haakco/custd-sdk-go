package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteLinksFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-links.json", key)
}

func TestSiteReferringDomainsKeepZeroAndAbsentApart(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteLinksFixture(t, "referringDomains"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	own, err := domains.GetReferringDomains(ctx, "s-1", "example.com")
	if err != nil || own.Role != AdminSiteWebRoleOwn || len(own.Sources) != 3 || own.Authority == nil ||
		own.Authority.HarmonicRank == nil || own.Lost.Status != AdminSiteLinkStatusAvailable || len(own.Lost.Domains) != 2 {
		t.Fatalf("own = %+v, %v", own, err)
	}
	graph := own.Sources[1]
	if graph.Source != AdminSiteLinkSourceCommonCrawl || graph.Complete || graph.ReferringDomains == nil ||
		*graph.ReferringDomains != 1843 || graph.Links != nil || graph.Listed != 3 {
		t.Fatalf("commoncrawl = %+v", graph)
	}
	news, blog := own.ReferringDomains[0], own.ReferringDomains[1]
	if len(news.Sources) != 3 || news.Followed == nil || *news.Followed || news.CrawlLinks == nil ||
		blog.Followed != nil || blog.CrawlLinks != nil || blog.BingLinks == nil || own.Lost.Domains[1].HarmonicRank != nil {
		t.Fatalf("referring = %+v, lost = %+v", own.ReferringDomains, own.Lost.Domains)
	}
	doer.body = siteLinksFixture(t, "referringDomainsPending")
	pending, err := domains.GetReferringDomains(ctx, "", "rival.example")
	if err != nil || pending.Role != AdminSiteWebRoleCompetitor || pending.Authority != nil ||
		pending.Sources[0].Status != AdminSiteLinkStatusUnavailable || pending.Sources[1].Status != AdminSiteLinkStatusPending ||
		pending.Sources[1].ReferringDomains != nil || len(pending.ReferringDomains) != 0 || pending.Lost.Status != AdminSiteLinkStatusPending {
		t.Fatalf("pending = %+v, %v", pending, err)
	}
	doer.body = siteLinksFixture(t, "referringDomainsOutsideGraph")
	outside, err := domains.GetReferringDomains(ctx, "s-1", "shop.example.dev")
	bing, crawl := outside.Sources[0], outside.Sources[2]
	if err != nil || outside.Authority == nil || outside.Authority.InGraph || outside.Authority.HarmonicRank != nil ||
		bing.Status != AdminSiteLinkStatusFailed || bing.LastFailure == "" || bing.ReferringDomains != nil ||
		crawl.ReferringDomains == nil || *crawl.ReferringDomains != 0 || crawl.Links == nil || *crawl.Links != 0 {
		t.Fatalf("outside = %+v, %v", outside, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteDomainsTestBase + "/referring-domains?domain=example.com&siteUuid=s-1",
		"GET " + siteDomainsTestBase + "/referring-domains?domain=rival.example",
		"GET " + siteDomainsTestBase + "/referring-domains?domain=shop.example.dev&siteUuid=s-1",
	})
}
