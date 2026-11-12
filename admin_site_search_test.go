package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSiteSearchOperationsSendTheirRoutesAndDecode(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"googleConfigured":true,
		"googleServiceAccountEmail":"custd@project.iam.gserviceaccount.com","bingConfigured":false}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	account, err := domains.GetSearchAccount(ctx)
	if err != nil || !account.GoogleConfigured || account.BingConfigured ||
		account.GoogleServiceAccountEmail != "custd@project.iam.gserviceaccount.com" {
		t.Fatalf("account = %+v err=%v", account, err)
	}

	doer.body = `{"domainUuid":"d-1","days":7,
		"connections":[{"provider":"google","property":"sc-domain:example.com","status":"active",
			"importedThrough":"2026-11-10","lastImportedAt":"2026-11-12T06:00:00Z"}],
		"sources":[{"provider":"google","source":"Google Search Console","granularity":"daily","from":"2026-11-04",
			"to":"2026-11-10","days":7,"totals":{"clicks":42,"impressions":1200,"ctr":0.035,"position":8.4}}],
		"topPages":[{"provider":"google","key":"https://www.example.com/","totals":{"clicks":30,"impressions":800,"ctr":0.0375}}],
		"topQueries":[{"provider":"google","key":"example shop","totals":{"clicks":12,"impressions":90,"ctr":0.13,"position":2.1}}]}`
	performance, err := domains.GetSearchPerformance(ctx, "d-1", 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(performance.Connections) != 1 || performance.Connections[0].Status != AdminSiteSearchConnectionActive ||
		performance.Connections[0].LastImportedAt == nil || len(performance.Sources) != 1 ||
		performance.Sources[0].Totals.Clicks != 42 || performance.Sources[0].Totals.Position != 8.4 ||
		len(performance.TopPages) != 1 || performance.TopPages[0].Totals.Position != 0 ||
		len(performance.TopQueries) != 1 || performance.TopQueries[0].Key != "example shop" {
		t.Fatalf("performance = %+v", performance)
	}

	doer.body = `{"provider":"bing","property":"https://www.example.com/","status":"pending"}`
	connection, err := domains.ConnectSearch(ctx, "d-1", AdminSiteSearchProviderBing,
		AdminSiteSearchConnectionRequest{Property: "https://www.example.com/", APIKey: "bing-api-key-0123456789"})
	if err != nil || connection.Status != AdminSiteSearchConnectionPending || connection.Provider != AdminSiteSearchProviderBing {
		t.Fatalf("connection = %+v err=%v", connection, err)
	}
	if got := string(doer.requests[2].Body); got != `{"property":"https://www.example.com/","apiKey":"bing-api-key-0123456789"}` {
		t.Fatalf("connect body = %s", got)
	}

	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DisconnectSearch(ctx, "d-1", AdminSiteSearchProviderGoogle); err != nil {
		t.Fatal(err)
	}
	base := "http://localhost:8080/api/v1/site-domains"
	assertSiteRequests(t, doer.requests, []string{
		"GET " + base + "/search-account", "GET " + base + "/d-1/search?days=7",
		"PUT " + base + "/d-1/search/bing", "DELETE " + base + "/d-1/search/google",
	})
}
