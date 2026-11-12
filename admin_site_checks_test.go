package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSiteCheckOperationsSendTheirRoutesAndDecode(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"domainUuid":"d-1","siteUuid":"s-1","host":"www.example.com",
		"ownership":{"verified":true,"instructions":{}},"audit":{},"createdAt":"2026-11-09T08:00:00Z",
		"checks":{"mode":"paused","uptime":{"status":"down","statusCode":503,"responseMs":120,"lastFailure":"HTTP 503",
			"consecutiveFailures":2},"tls":{"validTo":"2027-01-30T00:00:00Z","issuer":"R11"},
			"warming":{"unwarmedResponseMs":820},
			"openIncident":{"incidentUuid":"i-1","startedAt":"2026-11-12T10:00:00Z","openedAt":"2026-11-12T10:05:00Z",
				"cause":"HTTP 503","failedChecks":2}}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	domains := client.Admin.SiteDomains

	domain, err := domains.SetCheckMode(context.Background(), "d-1", AdminSiteCheckModePaused)
	if err != nil {
		t.Fatal(err)
	}
	checks := domain.Checks
	if checks.Mode != AdminSiteCheckModePaused || checks.Uptime.Status != AdminSiteUptimeDown ||
		checks.OpenIncident == nil || checks.OpenIncident.FailedChecks != 2 ||
		checks.Warming.UnwarmedResponseMs == nil || *checks.Warming.UnwarmedResponseMs != 820 || checks.Warming.WarmedResponseMs != nil {
		t.Fatalf("checks = %+v", checks)
	}
	if got := string(doer.requests[0].Body); got != `{"mode":"paused"}` {
		t.Fatalf("set mode body = %s", got)
	}

	doer.body = `{"incidents":[{"incidentUuid":"i-1","startedAt":"2026-11-12T10:00:00Z","openedAt":"2026-11-12T10:05:00Z",
		"resolvedAt":"2026-11-12T10:20:00Z","cause":"HTTP 503","failedChecks":3}]}`
	incidents, err := domains.ListIncidents(context.Background(), "d-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(incidents.Incidents) != 1 || incidents.Incidents[0].ResolvedAt == nil {
		t.Fatalf("incidents = %+v", incidents)
	}
	base := "http://localhost:8080/api/v1/site-domains/d-1"
	assertSiteRequests(t, doer.requests, []string{"PUT " + base + "/checks", "GET " + base + "/incidents"})
}
