package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteIndexNowFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-indexnow.json", key)
}

func TestSiteIndexNowStateRoutesAndAbsentFields(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteIndexNowFixture(t, "disabled"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	disabled, err := domains.GetIndexNow(ctx, "d-1")
	if err != nil || disabled.Enabled || disabled.Key != "" || disabled.VerifiedAt != nil || disabled.DailyURLLimit != 1000 {
		t.Fatalf("disabled = %+v, %v", disabled, err)
	}
	doer.body = siteIndexNowFixture(t, "enabledUnverified")
	enabled, err := domains.EnableIndexNow(ctx, "d-1")
	if err != nil || !enabled.Enabled || enabled.Verified || enabled.KeyLocation == "" || enabled.LastCheckFailure == "" {
		t.Fatalf("enabled = %+v, %v", enabled, err)
	}
	doer.body = siteIndexNowFixture(t, "verified")
	verified, err := domains.VerifyIndexNowKey(ctx, "d-1")
	if err != nil || !verified.Verified || verified.VerifiedAt == nil || verified.URLsToday != 3 {
		t.Fatalf("verified = %+v, %v", verified, err)
	}
	queued, accepted, failed := verified.Submissions[0], verified.Submissions[1], verified.Submissions[2]
	if queued.Status != AdminSiteIndexNowQueued || queued.ProviderStatus != nil || queued.SubmittedAt != nil ||
		accepted.Status != AdminSiteIndexNowAccepted || accepted.ProviderStatus == nil || *accepted.ProviderStatus != 202 ||
		accepted.SubmittedAt == nil || failed.Status != AdminSiteIndexNowFailed || failed.Failure == "" {
		t.Fatalf("submissions = %+v", verified.Submissions)
	}
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.DisableIndexNow(ctx, "d-1"); err != nil {
		t.Fatal(err)
	}
	base := siteTestBase + "/indexnow"
	assertSiteRequests(t, doer.requests, []string{"GET " + base, "PUT " + base, "POST " + base + "/verify", "DELETE " + base})
	for i, req := range doer.requests {
		if len(req.Body) != 0 {
			t.Fatalf("request[%d] body = %s, want none", i, req.Body)
		}
	}
}

func TestSiteIndexNowSubmitAndProblems(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, siteIndexNowFixture(t, "submissionQueued"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	urls := []string{"https://example.com/plans", "https://example.com/about"}

	submission, err := domains.SubmitIndexNow(ctx, "d-1", urls)
	if err != nil || submission.Status != AdminSiteIndexNowQueued || len(submission.URLs) != 2 ||
		submission.ProviderStatus != nil || submission.SubmittedAt != nil {
		t.Fatalf("submission = %+v, %v", submission, err)
	}
	assertSiteRequests(t, doer.requests, []string{"POST " + siteTestBase + "/indexnow/submissions"})
	assertSiteBody(t, doer.requests[0], siteIndexNowFixture(t, "submissionRequest"))
	for _, tc := range []struct {
		key    string
		status int
		code   string
		call   func() error
	}{
		{"domainUnverifiedProblem", http.StatusConflict, SiteIndexNowDomainUnverifiedCode, func() error {
			_, err := domains.EnableIndexNow(ctx, "d-1")
			return err
		}},
		{"notEnabledProblem", http.StatusConflict, SiteIndexNowNotEnabledCode, func() error {
			_, err := domains.VerifyIndexNowKey(ctx, "d-1")
			return err
		}},
		{"keyUnverifiedProblem", http.StatusConflict, SiteIndexNowKeyUnverifiedCode, func() error {
			_, err := domains.SubmitIndexNow(ctx, "d-1", urls)
			return err
		}},
		{"urlsInvalidProblem", http.StatusBadRequest, SiteIndexNowURLsInvalidCode, func() error {
			_, err := domains.SubmitIndexNow(ctx, "d-1", nil)
			return err
		}},
		{"limitProblem", http.StatusTooManyRequests, SiteIndexNowLimitCode, func() error {
			_, err := domains.SubmitIndexNow(ctx, "d-1", urls)
			return err
		}},
	} {
		doer.status, doer.body = tc.status, siteIndexNowFixture(t, tc.key)
		assertSiteProblemCode(t, tc.call(), tc.status, tc.code)
	}
}
