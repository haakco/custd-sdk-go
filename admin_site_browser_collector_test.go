package custd

import (
	"net/http"
	"testing"
)

func TestSiteBrowserCollectorConvergesOwnedDomain(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"siteUuid":"s-1","mode":"enabled","browserKey":"fixture-key",
		"environment":"production","allowedOrigin":"https://www.example.com"}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	on, err := domains.SetBrowserCollector(t.Context(), "d-1", SiteBrowserCollectorRequest{
		Mode: SiteBrowserCollectorEnabled, Environment: "production",
	})
	if err != nil || on.BrowserKey != "fixture-key" || on.AllowedOrigin != "https://www.example.com" {
		t.Fatalf("enable did not return the collector: %v", err)
	}

	doer.body = `{"siteUuid":"s-1","mode":"disabled"}`
	off, err := domains.SetBrowserCollector(t.Context(), "d-1", SiteBrowserCollectorRequest{Mode: SiteBrowserCollectorDisabled})
	if err != nil || off.Mode != SiteBrowserCollectorDisabled || off.BrowserKey != "" || off.AllowedOrigin != "" {
		t.Fatalf("disable did not clear key and origin: %v", err)
	}

	const route = "PUT http://localhost:8080/api/v1/site-domains/d-1/browser-collector"
	assertSiteRequests(t, doer.requests, []string{route, route})
	if string(doer.requests[0].Body) != `{"mode":"enabled","environment":"production"}` ||
		string(doer.requests[1].Body) != `{"mode":"disabled"}` {
		t.Fatal("collector configuration request did not match the contract")
	}
}
