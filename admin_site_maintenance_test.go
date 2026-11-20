package custd

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestSiteMaintenanceLifecycle(t *testing.T) {
	const start = "2026-11-20T06:00:00Z"
	const end = "2026-11-20T07:00:00Z"
	const window = `{"windowUuid":"w-1","startsAt":"` + start + `","endsAt":"` + end + `"}`
	doer := newCaptureDoer(http.StatusOK, window)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	startsAt := time.Date(2026, 11, 20, 6, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(time.Hour)

	scheduled, err := domains.ScheduleMaintenance(ctx, "d-1", AdminSiteMaintenanceRequest{
		WindowUUID: "w-1", StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil || scheduled.WindowUUID != "w-1" || !scheduled.StartsAt.Equal(startsAt) || scheduled.CancelledAt != nil {
		t.Fatalf("scheduled = %+v, %v", scheduled, err)
	}

	doer.body = `{"windows":[` + window + `]}`
	list, err := domains.ListMaintenance(ctx, "d-1")
	if err != nil || len(list.Windows) != 1 || !list.Windows[0].EndsAt.Equal(endsAt) {
		t.Fatalf("list = %+v, %v", list, err)
	}

	doer.body = `{"windowUuid":"w-1","startsAt":"` + start + `","endsAt":"` + end + `","cancelledAt":"2026-11-20T06:15:00Z"}`
	cancelled, err := domains.CancelMaintenance(ctx, "d-1", "w-1")
	if err != nil || cancelled.CancelledAt == nil || !cancelled.CancelledAt.Equal(startsAt.Add(15*time.Minute)) {
		t.Fatalf("cancelled = %+v, %v", cancelled, err)
	}

	base := siteTestBase + "/maintenance"
	assertSiteRequests(t, doer.requests, []string{"POST " + base, "GET " + base, "DELETE " + base + "/w-1"})
	assertSiteBody(t, doer.requests[0], window)
}
