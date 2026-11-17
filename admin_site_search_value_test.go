package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteSearchValueFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-search-value.json", key)
}

func TestSiteFindingSearchImpactKeepsZeroAndAbsentApart(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSearchValueFixture(t, "findingImpact"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	impact, err := domains.GetFindingSearchImpact(ctx, "d-1", "f-fixed", 0)
	if err != nil || impact.Status != AdminSiteSearchImpactMeasured || impact.FixedAt == nil || impact.Before == nil ||
		impact.After.To != "2026-11-12" || len(impact.Sources) != 2 {
		t.Fatalf("impact = %+v, %v", impact, err)
	}
	google, bing := impact.Sources[0], impact.Sources[1]
	if google.Provider != AdminSiteSearchProviderGoogle || google.After.Clicks == nil || *google.After.Clicks != 0 ||
		bing.Status != AdminSiteSearchImpactInsufficient || bing.After.Clicks != nil || bing.After.Impressions != nil ||
		bing.After.CTR != nil || bing.After.Reason == "" {
		t.Fatalf("sources = %+v", impact.Sources)
	}
	doer.body = siteSearchValueFixture(t, "findingImpactOpen")
	open, err := domains.GetFindingSearchImpact(ctx, "d-1", "f-open", AdminSiteSearchImpactMinDays)
	if err != nil || open.Status != AdminSiteSearchImpactOpen || open.FixedAt != nil || open.Before != nil ||
		open.After != nil || len(open.Sources) != 0 {
		t.Fatalf("open = %+v, %v", open, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/findings/f-fixed/search-impact", "GET " + siteTestBase + "/findings/f-open/search-impact?days=7",
	})
}

func TestSiteSearchValueGoalsQueryAndAbsentConversions(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteSearchValueFixture(t, "searchValue"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	value, err := domains.GetSearchValue(ctx, "d-1", []string{"signup", "purchase"}, 0)
	if err != nil || value.GoalStatus != AdminSiteSearchValuePartial || value.GoalReason == "" || len(value.Pages) != 3 {
		t.Fatalf("value = %+v, %v", value, err)
	}
	pricing, blog, old := value.Pages[0], value.Pages[1], value.Pages[2]
	if pricing.GoalConversions == nil || *pricing.GoalConversions != 9 || pricing.Search.Position != 4.2 ||
		blog.GoalConversions == nil || *blog.GoalConversions != 0 ||
		old.Provider != AdminSiteSearchProviderBing || old.Landings != nil || old.GoalConversions != nil {
		t.Fatalf("pages = %+v", value.Pages)
	}
	doer.body = siteSearchValueFixture(t, "searchValueUnavailable")
	unavailable, err := domains.GetSearchValue(ctx, "d-1", []string{"signup"}, 7)
	if err != nil || unavailable.GoalStatus != AdminSiteSearchValueUnavailable ||
		unavailable.Pages[0].Landings != nil || unavailable.Pages[0].GoalConversions != nil {
		t.Fatalf("unavailable = %+v, %v", unavailable, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/search-value?goal=signup&goal=purchase",
		"GET " + siteTestBase + "/search-value?days=7&goal=signup",
	})
}
