package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSiteStatusPageLifecycle(t *testing.T) {
	const selection = `{"title":"Service status","components":[{"domainUuid":"d-1","label":"Portal"}]}`
	const live = `{"publicKey":"fixture","publishedAt":"2026-11-20T06:00:00Z"}`
	doer := newCaptureDoer(http.StatusOK, "{}")
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	draft := AdminSiteStatusDraft{
		Title: "Service status", Components: []AdminSiteStatusSelection{{DomainUUID: "d-1", Label: "Portal"}},
	}

	settings, err := domains.GetStatusPageSettings(ctx, "d-1")
	if err != nil || settings.Draft != nil || settings.Publication != nil {
		t.Fatalf("settings = %+v, %v", settings, err)
	}

	doer.body = `{"title":"Service status","components":[{"label":"Portal","status":"maintenance"}],"updates":[]}`
	preview, err := domains.PreviewStatusPage(ctx, "d-1", draft)
	if err != nil || len(preview.Components) != 1 || preview.Components[0].Status != "maintenance" {
		t.Fatalf("preview = %+v, %v", preview, err)
	}

	doer.body = live
	publication, err := domains.PublishStatusPage(ctx, "d-1", draft)
	if err != nil || publication.PublicKey != "fixture" || publication.PublishedAt.IsZero() {
		t.Fatalf("publication = %+v, %v", publication, err)
	}

	doer.body = `{"updates":[{"message":"Service restored.","publishedAt":"2026-11-20T06:05:00Z"}]}`
	history, err := domains.PostStatusPageUpdate(ctx, "d-1", AdminSiteStatusUpdateRequest{
		RequestUUID: "update-1", Message: "Service restored.",
	})
	if err != nil || len(history.Updates) != 1 || history.Updates[0].Message != "Service restored." {
		t.Fatalf("updates = %+v, %v", history, err)
	}

	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.UnpublishStatusPage(ctx, "d-1"); err != nil {
		t.Fatal(err)
	}
	base := siteTestBase + "/status-page"
	assertSiteRequests(t, doer.requests, []string{
		"GET " + base, "POST " + base + "/preview", "PUT " + base, "POST " + base + "/updates", "DELETE " + base,
	})
	assertSiteBody(t, doer.requests[1], selection)
	assertSiteBody(t, doer.requests[2], selection)
	assertSiteBody(t, doer.requests[3], `{"requestUuid":"update-1","message":"Service restored."}`)
}
