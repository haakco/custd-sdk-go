package custd

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func siteAccessLogsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-access-logs.json", key)
}

func TestSiteTrafficDecodesCoverageAndAbsentDays(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteAccessLogsFixture(t, "traffic"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	traffic, err := domains.GetTraffic(ctx, "d-1", "2026-11-15", "2026-11-16")
	if err != nil || traffic.Totals == nil || traffic.Totals.P95Ms == nil || *traffic.Totals.P95Ms != 750 ||
		traffic.Coverage.Complete || traffic.Coverage.HoursMissing != 24 {
		t.Fatalf("traffic = %+v, %v", traffic, err)
	}
	missing, partial := traffic.Days[0], traffic.Days[1]
	if missing.Coverage != AdminSiteTrafficCoverageMissing || missing.Requests != nil || missing.NotFound != nil ||
		partial.Coverage != AdminSiteTrafficCoveragePartial || partial.NotFound == nil || *partial.NotFound != 0 {
		t.Fatalf("days = %+v", traffic.Days)
	}
	if traffic.NotFoundPaths[1].Key != AdminSiteTrafficOtherKey || traffic.SlowPaths[0].MaxMs != 4100 ||
		traffic.ReferrerClasses[1].Class != AdminSiteReferrerClassAI || traffic.Crawlers[1].Identity != AdminSiteCrawlerClaimsToBe {
		t.Fatalf("traffic lists = %+v", traffic)
	}
	doer.body = siteAccessLogsFixture(t, "trafficEmpty")
	empty, err := domains.GetTraffic(ctx, "d-1", "", "")
	if err != nil || empty.Totals != nil || len(empty.Crawlers) != 0 {
		t.Fatalf("empty = %+v, %v", empty, err)
	}
	doer.body = siteAccessLogsFixture(t, "aiVisibility")
	ai, err := domains.GetAIVisibility(ctx, "d-1", "2026-11-15", "")
	if err != nil || ai.Crawlers[0].Identity != AdminSiteCrawlerVerified || ai.Assistants[0].Key != "chatgpt.com" ||
		ai.Pages[1].Path != AdminSiteTrafficOtherKey {
		t.Fatalf("ai = %+v, %v", ai, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/traffic?from=2026-11-15&to=2026-11-16", "GET " + siteTestBase + "/traffic",
		"GET " + siteTestBase + "/ai-visibility?from=2026-11-15",
	})
}

func TestAccessLogHourReplayRouteAndProblem(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, siteAccessLogsFixture(t, "accessLogHourReplay"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	hour := time.Date(2026, 11, 17, 12, 0, 0, 0, time.FixedZone("SAST", 2*60*60))

	replay, err := domains.ReplayAccessLogHour(ctx, hour)
	if err != nil || replay.PreviousStatus != AdminAccessLogHourFailed || !replay.HourStart.Equal(hour) {
		t.Fatalf("replay = %+v, %v", replay, err)
	}
	assertSiteRequests(t, doer.requests, []string{"POST http://localhost:8080/api/v1/admin/access-log-hours/replay"})
	assertSiteBody(t, doer.requests[0], siteAccessLogsFixture(t, "accessLogHourReplayRequest"))
	doer.status, doer.body = http.StatusBadRequest, siteAccessLogsFixture(t, "accessLogHourInvalidProblem")
	_, err = domains.ReplayAccessLogHour(ctx, hour.Add(time.Minute))
	assertSiteProblemCode(t, err, http.StatusBadRequest, AccessLogHourInvalidCode)
}
