package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSiteHeartbeatManagement(t *testing.T) {
	const monitorJSON = `{"monitorUuid":"m-1","name":"Backup","schedule":{"cron":"0 6 * * *","timezone":"Etc/UTC","graceMinutes":15},"startedAt":"2026-11-20T00:00:00Z","nextExpectedAt":"2026-11-21T06:00:00Z","status":"on-time","latest":{"report":{"scheduledFor":"2026-11-20T06:00:00Z","outcome":"succeeded"},"receivedAt":"2026-11-20T06:05:00Z"}}`
	const credentialJSON = `{"monitor":` + monitorJSON + `,"reportPath":"/api/v1/site-heartbeat-reports/t-1/m-1","token":"fixture"}`
	doer := newCaptureDoer(http.StatusCreated, credentialJSON)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	created, err := domains.CreateHeartbeat(ctx, "d-1", AdminSiteHeartbeatCreate{
		Name: "Backup", Schedule: AdminSiteHeartbeatSchedule{Cron: "0 6 * * *", Timezone: "Etc/UTC", GraceMinutes: 15},
	})
	if err != nil || created.Token != "fixture" || created.ReportPath != "/api/v1/site-heartbeat-reports/t-1/m-1" {
		t.Fatal("create did not return the monitor credential", err)
	}

	doer.status, doer.body = http.StatusOK, `{"monitors":[`+monitorJSON+`]}`
	list, err := domains.ListHeartbeats(ctx, "d-1")
	if err != nil || len(list.Monitors) != 1 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	monitor := list.Monitors[0]
	if monitor.MonitorUUID != "m-1" || monitor.Status != "on-time" || monitor.NextExpectedAt.IsZero() ||
		monitor.Latest == nil || monitor.Latest.Report.Outcome != "succeeded" {
		t.Fatalf("monitor = %+v", monitor)
	}

	doer.body = credentialJSON
	rotated, err := domains.RotateHeartbeat(ctx, "d-1", "m-1")
	if err != nil || rotated.Token != "fixture" {
		t.Fatal("rotate did not return a credential", err)
	}

	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DeleteHeartbeat(ctx, "d-1", "m-1"); err != nil {
		t.Fatal(err)
	}

	base := siteTestBase + "/heartbeats"
	assertSiteRequests(t, doer.requests, []string{"POST " + base, "GET " + base,
		"POST " + base + "/m-1/credential", "DELETE " + base + "/m-1"})
	assertSiteBody(t, doer.requests[0], `{"name":"Backup","schedule":{"cron":"0 6 * * *","timezone":"Etc/UTC","graceMinutes":15}}`)
}
