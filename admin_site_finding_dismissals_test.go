package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSiteFindingDismissalLifecycle(t *testing.T) {
	const dismissal = `{"dismissalUuid":"dismissal-1","rule":"description.length","url":"https://example.test/","title":"Short description","reason":"Intentional","dismissedAt":"2026-11-20T00:00:00Z"}`
	doer := newCaptureDoer(http.StatusOK, dismissal)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	result, err := domains.DismissFinding(ctx, "d-1", AdminSiteFindingDismissalRequest{
		FindingUUID: "finding-1", Scope: AdminSiteFindingDismissalURL, Reason: "Intentional",
	})
	if err != nil || result.DismissalUUID != "dismissal-1" || result.URL != "https://example.test/" || result.DismissedAt.IsZero() {
		t.Fatalf("dismissal = %+v, %v", result, err)
	}
	doer.body = `{"dismissals":[` + dismissal + `]}`
	list, err := domains.ListFindingDismissals(ctx, "d-1")
	if err != nil || len(list.Dismissals) != 1 || list.Dismissals[0].Reason != "Intentional" {
		t.Fatalf("dismissals = %+v, %v", list, err)
	}
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.RestoreFinding(ctx, "d-1", "dismissal-1"); err != nil {
		t.Fatal(err)
	}
	base := siteTestBase + "/finding-dismissals"
	assertSiteRequests(t, doer.requests, []string{"POST " + base, "GET " + base, "DELETE " + base + "/dismissal-1"})
	assertSiteBody(t, doer.requests[0], `{"findingUuid":"finding-1","scope":"url","reason":"Intentional"}`)
}
