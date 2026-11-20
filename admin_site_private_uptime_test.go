package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSitePrivateUptimeManagement(t *testing.T) {
	const credentialJSON = `{"monitor":{"monitorUuid":"m-1","networkLabel":"Office","status":"unknown"},"reportPath":"/api/v1/site-private-uptime-reports/t-1/m-1","token":"fixture"}`
	doer := newCaptureDoer(http.StatusCreated, credentialJSON)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	created, err := domains.CreatePrivateUptime(ctx, "d-1", "Office")
	if err != nil || created.Monitor.MonitorUUID != "m-1" || created.Monitor.NetworkLabel != "Office" ||
		created.Monitor.Status != "unknown" || created.ReportPath != "/api/v1/site-private-uptime-reports/t-1/m-1" || created.Token != "fixture" {
		t.Fatal("create did not return the monitor and credential", err)
	}

	doer.status, doer.body = http.StatusAccepted, `{"requestedAt":"2026-11-20T00:18:41Z"}`
	work, err := domains.RequestPrivateUptime(ctx, "d-1")
	if err != nil || work.RequestedAt == nil {
		t.Fatal("request did not return pending work", err)
	}

	doer.status = http.StatusOK
	doer.body = `{"monitorUuid":"m-1","networkLabel":"Office","status":"up","latest":{"reportUuid":"r-1","observedAt":"2026-11-20T00:18:42Z","statusCode":200,"responseMs":419}}`
	monitor, err := domains.GetPrivateUptime(ctx, "d-1")
	if err != nil || monitor.Status != "up" || monitor.RequestedAt != nil || monitor.Latest == nil || monitor.Latest.StatusCode != 200 {
		t.Fatalf("monitor = %+v, %v", monitor, err)
	}

	doer.body = credentialJSON
	rotated, err := domains.RotatePrivateUptime(ctx, "d-1", "m-1")
	if err != nil || rotated.Token != "fixture" {
		t.Fatal("rotate did not return a credential", err)
	}

	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DeletePrivateUptime(ctx, "d-1", "m-1"); err != nil {
		t.Fatal(err)
	}

	base := siteTestBase + "/private-uptime"
	assertSiteRequests(t, doer.requests, []string{"POST " + base, "POST " + base + "/check", "GET " + base,
		"POST " + base + "/m-1/credential", "DELETE " + base + "/m-1"})
	assertSiteBody(t, doer.requests[0], `{"networkLabel":"Office"}`)
}
