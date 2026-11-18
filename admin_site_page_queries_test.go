package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSitePageSearchQueriesRouteAndStates(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, recordedFixture(t, "site-page-queries.json", "pageSearchQueries"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	queries, err := domains.GetPageSearchQueries(ctx, "d-1", "https://shop.example.com/pricing")
	if err != nil || queries.State != AdminSitePageQueriesAvailable || queries.Window == nil || queries.Window.From != "2026-10-18" ||
		len(queries.Queries) != 2 || !queries.Queries[0].InDescription || queries.Queries[0].Position != 4.2 {
		t.Fatalf("queries = %+v, %v", queries, err)
	}
	if len(queries.TitleOptions) != 4 || queries.TitleOptions[1].Template != AdminSiteTitleTemplateGuide ||
		len(queries.LinkSources) != 1 || queries.LinkSources[0].Where != AdminSiteLinkSourceInTitle {
		t.Fatalf("tools = %+v %+v", queries.TitleOptions, queries.LinkSources)
	}
	doer.body = recordedFixture(t, "site-page-queries.json", "pageSearchQueriesNotConnected")
	none, err := domains.GetPageSearchQueries(ctx, "d-1", "https://shop.example.com/pricing")
	if err != nil || none.State != AdminSitePageQueriesNotConnected || none.Window != nil || len(none.Queries) != 0 {
		t.Fatalf("not connected = %+v, %v", none, err)
	}
	assertSiteRequests(t, doer.requests[:1], []string{
		"GET " + siteTestBase + "/page-search-queries?url=https%3A%2F%2Fshop.example.com%2Fpricing",
	})
}
