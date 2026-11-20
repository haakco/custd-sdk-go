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
	doer.status, doer.body = http.StatusAccepted, `{"domainUuid":"d-1","checks":{"uptime":{"status":"down"}}}`
	queued, err := domains.RequestUptimeCheck(context.Background(), "d-1")
	if err != nil || queued.DomainUUID != "d-1" || queued.Checks.Uptime.Status != AdminSiteUptimeDown {
		t.Fatalf("queued check snapshot = %+v, %v", queued, err)
	}
	base := "http://localhost:8080/api/v1/site-domains/d-1"
	assertSiteRequests(t, doer.requests, []string{"PUT " + base + "/checks", "GET " + base + "/incidents", "POST " + base + "/uptime-check"})
	if len(doer.requests[2].Body) != 0 {
		t.Fatal("queue uptime check unexpectedly sends a body")
	}
}

func TestSiteChromeUXAvailability(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"checks":{"chromeUx":{"checkedAt":"2026-11-20T07:00:00Z",
		"observation":{"firstDate":"2026-10-21T00:00:00Z","lastDate":"2026-11-17T00:00:00Z","lcpMs":1700,"cls":0}}}}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	measured, err := domains.Get(ctx, "d-1")
	if err != nil || measured.Checks.ChromeUX == nil || measured.Checks.ChromeUX.Observation == nil {
		t.Fatalf("measured CrUX = %+v, %v", measured, err)
	}
	observation := measured.Checks.ChromeUX.Observation
	if observation.CLS == nil || *observation.CLS != 0 || observation.INPMs != nil ||
		observation.LCPMs == nil || *observation.LCPMs != 1700 || !observation.LastDate.After(observation.FirstDate) {
		t.Fatalf("observation = %+v", observation)
	}
	doer.body = `{"checks":{"chromeUx":{"checkedAt":"2026-11-20T07:00:00Z"}}}`
	empty, err := domains.Get(ctx, "d-1")
	if err != nil || empty.Checks.ChromeUX == nil || empty.Checks.ChromeUX.Observation != nil {
		t.Fatalf("no-data CrUX = %+v, %v", empty, err)
	}
	doer.body = `{"checks":{}}`
	unchecked, err := domains.Get(ctx, "d-1")
	if err != nil || unchecked.Checks.ChromeUX != nil {
		t.Fatalf("unchecked CrUX = %+v, %v", unchecked, err)
	}
}

func TestSiteIncidentAcknowledgmentDoesNotResolveAnOutage(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"incidentUuid":"i-1","acknowledgedAt":"2026-11-20T07:00:00Z","cause":"HTTP 503","failedChecks":2}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	incident, err := domains.AcknowledgeIncident(context.Background(), "d-1", "i-1")
	if err != nil || incident.IncidentUUID != "i-1" || incident.AcknowledgedAt == nil || incident.ResolvedAt != nil {
		t.Fatalf("acknowledged incident = %+v, %v", incident, err)
	}
	assertSiteRequests(t, doer.requests, []string{"POST " + siteTestBase + "/incidents/i-1/acknowledgment"})
	if len(doer.requests[0].Body) != 0 {
		t.Fatal("acknowledgment unexpectedly sends a body")
	}
}

func TestSiteHTTPExpectationsSetAndClear(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"checks":{"mode":"uptime","httpExpectations":{"statusCode":200,"maxResponseMs":750,"json":{"field":"status","value":"ok"}}}}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	settings := AdminSiteCheckSettings{Mode: AdminSiteCheckModeUptime, HTTPExpectations: &AdminSiteHTTPExpectations{
		StatusCode: 200, MaxResponseMs: 750, JSON: &AdminSiteHTTPJSONExpectation{Field: "status", Value: "ok"},
	}}
	domain, err := domains.SetCheckSettings(context.Background(), "d-1", settings)
	if err != nil || domain.Checks.HTTPExpectations == nil || domain.Checks.HTTPExpectations.JSON == nil ||
		domain.Checks.HTTPExpectations.MaxResponseMs != 750 || domain.Checks.HTTPExpectations.JSON.Value != "ok" {
		t.Fatalf("settings result = %+v, %v", domain, err)
	}
	doer.body = `{"checks":{"mode":"uptime"}}`
	settings.HTTPExpectations = &AdminSiteHTTPExpectations{}
	cleared, err := domains.SetCheckSettings(context.Background(), "d-1", settings)
	if err != nil || cleared.Checks.HTTPExpectations != nil {
		t.Fatalf("cleared settings = %+v, %v", cleared, err)
	}
	target := "PUT " + siteTestBase + "/checks"
	assertSiteRequests(t, doer.requests, []string{target, target})
	assertSiteBody(t, doer.requests[0],
		`{"mode":"uptime","httpExpectations":{"statusCode":200,"maxResponseMs":750,"json":{"field":"status","value":"ok"}}}`)
	assertSiteBody(t, doer.requests[1], `{"mode":"uptime","httpExpectations":{}}`)
}

func TestSiteRegistrationDistinguishesUnknownExpiry(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"checks":{"registration":{"domain":"example.com","source":"rdap.example.com","readAt":"2026-11-20T07:00:00Z","expiresAt":"2027-01-01T00:00:00Z"}}}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	read, err := domains.Get(ctx, "d-1")
	if err != nil || read.Checks.Registration == nil || read.Checks.Registration.ExpiresAt == nil ||
		read.Checks.Registration.Source != "rdap.example.com" {
		t.Fatalf("registry lookup = %+v, %v", read, err)
	}
	doer.body = `{"checks":{"registration":{"domain":"example.com","readAt":"2026-11-20T07:00:00Z"}}}`
	read, err = domains.Get(ctx, "d-1")
	if err != nil || read.Checks.Registration == nil || read.Checks.Registration.ReadAt.IsZero() ||
		read.Checks.Registration.ExpiresAt != nil {
		t.Fatalf("unknown expiry = %+v, %v", read, err)
	}
}
