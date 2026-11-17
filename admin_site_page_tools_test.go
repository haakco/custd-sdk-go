package custd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

const siteTestBase = "http://localhost:8080/api/v1/site-domains/d-1"

func sitePageToolsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-page-tools.json", key)
}

// assertSiteBody checks a request body equals the recorded request, compared as JSON.
func assertSiteBody(t *testing.T, req *HTTPRequest, recorded string) {
	t.Helper()
	var got, want any
	if err := json.Unmarshal(req.Body, &got); err != nil {
		t.Fatalf("body %s: %v", req.Body, err)
	}
	if err := json.Unmarshal([]byte(recorded), &want); err != nil {
		t.Fatal(err)
	}
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("body = %s, want %s", gotJSON, wantJSON)
	}
}

func assertSiteProblemCode(t *testing.T, err error, status int, code string) {
	t.Helper()
	var requestErr *RequestError
	if !errors.As(err, &requestErr) || requestErr.StatusCode != status ||
		requestErr.Problem == nil || requestErr.Problem.Code != code {
		t.Fatalf("err = %#v, want %d %s", err, status, code)
	}
}

// A finding with a generator carries code to copy; an unknown rule decodes an empty remedy.
func TestSiteRemedyDecodesGeneratedAndUnknown(t *testing.T) {
	var finding AdminSiteFinding
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "findingWithGeneratedFix")), &finding); err != nil {
		t.Fatal(err)
	}
	generated := finding.Remedy.Generated
	if finding.Remedy.Kind != AdminSiteRemedyKindMarkup || generated == nil ||
		generated.Format != AdminSiteGeneratedFormatHTML || generated.Code != `<link rel="canonical" href="https://example.com/plans">` {
		t.Fatalf("remedy = %+v", finding.Remedy)
	}
	var unknown AdminSiteRemedy
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "remedyUnknownRule")), &unknown); err != nil ||
		unknown != (AdminSiteRemedy{}) {
		t.Fatalf("unknown remedy = %+v, %v", unknown, err)
	}
}

func TestSitePageFactsRouteAndReadingEase(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "pageFacts"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	facts, err := domains.GetPageFacts(ctx, "d-1", "https://example.com/plans")
	if err != nil {
		t.Fatal(err)
	}
	ease := facts.Readability.FleschReadingEase
	if facts.Status != 200 || facts.Lang != "en" || len(facts.Headings) != 2 || facts.Headings[0].Level != 1 ||
		len(facts.Hreflang) != 1 || facts.OpenGraph.Title != "Hosting plans" || ease == nil || *ease != 57.1 {
		t.Fatalf("facts = %+v", facts)
	}
	doer.body = sitePageToolsFixture(t, "pageFactsWithoutReadingEase")
	other, err := domains.GetPageFacts(ctx, "d-1", "https://example.com/af/planne")
	if err != nil || other.Readability.FleschReadingEase != nil || other.Readability.Sentences != 1 {
		t.Fatalf("readability = %+v, %v", other.Readability, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + siteTestBase + "/page-facts?url=https%3A%2F%2Fexample.com%2Fplans",
		"GET " + siteTestBase + "/page-facts?url=https%3A%2F%2Fexample.com%2Faf%2Fplanne",
	})
}

func TestSitePageRecheckRoutesAndStates(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, sitePageToolsFixture(t, "pageRecheckQueued"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	queued, err := domains.RequestPageRecheck(ctx, "d-1", "https://example.com/plans")
	if err != nil || queued.State != AdminSitePageRecheckQueued || queued.CheckedAt != nil ||
		queued.RecheckUUID != "0198a3b2-0000-7000-8000-000000000040" {
		t.Fatalf("queued = %+v, %v", queued, err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "pageRecheckRequest"))
	doer.status, doer.body = http.StatusOK, sitePageToolsFixture(t, "pageRecheckChecked")
	checked, err := domains.GetPageRecheck(ctx, "d-1", queued.RecheckUUID)
	if err != nil || checked.State != AdminSitePageRecheckChecked || checked.CheckedAt == nil ||
		checked.OpenRules[0] != SiteRuleCanonicalMissing || checked.FixedRules[0] != SiteRuleTitleMissing {
		t.Fatalf("checked = %+v, %v", checked, err)
	}
	doer.body = sitePageToolsFixture(t, "pageRecheckFailed")
	failed, err := domains.GetPageRecheck(ctx, "d-1", "r-2")
	if err != nil || failed.State != AdminSitePageRecheckFailed || failed.Failure == "" {
		t.Fatalf("failed = %+v, %v", failed, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST " + siteTestBase + "/page-rechecks",
		"GET " + siteTestBase + "/page-rechecks/" + queued.RecheckUUID,
		"GET " + siteTestBase + "/page-rechecks/r-2",
	})
	doer.status, doer.body = http.StatusTooManyRequests, sitePageToolsFixture(t, "pageRecheckLimitProblem")
	_, err = domains.RequestPageRecheck(ctx, "d-1", "https://example.com/plans")
	assertSiteProblemCode(t, err, http.StatusTooManyRequests, SitePageRecheckLimitCode)
	doer.status, doer.body = http.StatusBadRequest, sitePageToolsFixture(t, "pageNotOnDomainProblem")
	_, err = domains.RequestPageRecheck(ctx, "d-1", "https://other.example/")
	assertSiteProblemCode(t, err, http.StatusBadRequest, SitePageNotOnDomainCode)
}

func TestSitePageCheckSendsDraftAndDecodesFindings(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "pageCheck"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	var req AdminSitePageCheckRequest
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "pageCheckRequest")), &req); err != nil {
		t.Fatal(err)
	}

	check, err := domains.CheckPage(context.Background(), "d-1", req)
	if err != nil {
		t.Fatal(err)
	}
	assertSiteRequests(t, doer.requests, []string{"POST " + siteTestBase + "/page-check"})
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "pageCheckRequest"))
	if check.Page.URL != req.URL || len(check.Findings) == 0 {
		t.Fatalf("check = %+v", check)
	}
	var canonical *AdminSitePageFinding
	for i := range check.Findings {
		if check.Findings[i].Rule == SiteRuleCanonicalMissing {
			canonical = &check.Findings[i]
		}
	}
	if canonical == nil || canonical.Remedy.Generated == nil || canonical.Remedy.Kind != AdminSiteRemedyKindMarkup {
		t.Fatalf("canonical finding = %+v", canonical)
	}
}

func TestSiteRobotsTestPastedAndLive(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "robotsTest"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	var req AdminSiteRobotsTestRequest
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "robotsTestRequest")), &req); err != nil {
		t.Fatal(err)
	}

	pasted, err := domains.TestRobots(ctx, "d-1", req)
	if err != nil || pasted.Allowed || pasted.Source != AdminSiteRobotsSourcePasted || pasted.RobotsStatus != nil ||
		pasted.Rule == nil || *pasted.Rule != (AdminSiteRobotsRule{Line: 2, Directive: "disallow", Pattern: "/private/"}) {
		t.Fatalf("pasted = %+v, %v", pasted, err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "robotsTestRequest"))
	doer.body = sitePageToolsFixture(t, "robotsTestLiveMissing")
	live, err := domains.TestRobots(ctx, "d-1", AdminSiteRobotsTestRequest{URL: "/", UserAgent: "GPTBot"})
	if err != nil || !live.Allowed || live.Source != AdminSiteRobotsSourceLive || live.RobotsStatus == nil ||
		*live.RobotsStatus != 404 || live.Rule != nil || live.Note == "" {
		t.Fatalf("live = %+v, %v", live, err)
	}
	assertSiteBody(t, doer.requests[1], `{"url":"/","userAgent":"GPTBot"}`)
	assertSiteRequests(t, doer.requests, []string{
		"POST " + siteTestBase + "/robots-test", "POST " + siteTestBase + "/robots-test",
	})
}
