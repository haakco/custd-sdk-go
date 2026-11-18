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
	if !finding.Recheckable {
		t.Fatal("canonical.missing finding is not recheckable")
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

// ListPageRechecks sends the ids comma-joined, as the server reads them, and decodes the list.
func TestSitePageRechecksReadSendsCommaJoinedIDs(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "pageRecheckList"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains

	list, err := domains.ListPageRechecks(context.Background(), "d-1", []string{"r-1", "r-2"})
	if err != nil || len(list.Rechecks) != 2 || list.Rechecks[0].State != AdminSitePageRecheckChecked ||
		list.Rechecks[1].State != AdminSitePageRecheckFailed {
		t.Fatalf("list = %+v, %v", list, err)
	}
	assertSiteRequests(t, doer.requests, []string{"GET " + siteTestBase + "/page-rechecks?ids=r-1%2Cr-2"})
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

// The draft tools ride on the page check: the request names previousUrl and phrase, and the
// response adds a slug, the redirect and the phrase's use, with the page's share image and FAQ
// markup. The older pageCheck shape above decodes without any of them.
func TestSitePageCheckDraftTools(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "pageCheckDraftTools"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	var req AdminSitePageCheckRequest
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "pageCheckDraftToolsRequest")), &req); err != nil {
		t.Fatal(err)
	}

	check, err := domains.CheckPage(context.Background(), "d-1", req)
	if err != nil {
		t.Fatal(err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "pageCheckDraftToolsRequest"))
	phrase := check.Phrase
	if check.Slug != "cafe-hosting-plans-for-small-teams" || check.Redirect == nil ||
		check.Redirect.Format != AdminSiteGeneratedFormatRedirect || phrase == nil || phrase.Count != 2 ||
		!phrase.InDescription || phrase.InH1 || phrase.DensityPercent == nil || *phrase.DensityPercent != 3.85 {
		t.Fatalf("draft tools = %q %+v %+v", check.Slug, check.Redirect, phrase)
	}
	if check.Page.ShareImage != "https://example.com/img/team.jpg" || check.Page.FAQSchema == nil ||
		check.Page.FAQSchema.Format != AdminSiteGeneratedFormatHTML {
		t.Fatalf("page = %+v", check.Page)
	}
	var older AdminSitePageCheck
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "pageCheck")), &older); err != nil ||
		older.Slug != "" || older.Redirect != nil || older.Phrase != nil || older.Page.ShareImage != "" ||
		older.Page.FAQSchema != nil {
		t.Fatalf("older = %+v, %v", older, err)
	}
	var empty AdminSitePhraseUse
	if err := json.Unmarshal([]byte(`{"phrase":"x","count":0,"wordCount":0,"densityPercent":0}`), &empty); err != nil ||
		empty.DensityPercent != nil {
		t.Fatalf("no words = %+v, %v", empty, err)
	}
}

func TestSiteWordingCompareRouteAndNotCompared(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "wordingCompare"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	var req AdminSiteWordingCompareRequest
	if err := json.Unmarshal([]byte(sitePageToolsFixture(t, "wordingCompareRequest")), &req); err != nil {
		t.Fatal(err)
	}

	compared, err := domains.CompareWording(context.Background(), "d-1", req)
	if err != nil || !compared.Compared || compared.SharedPercent == nil || *compared.SharedPercent != 91 ||
		!compared.NearDuplicate || compared.ThresholdPercent != 90 {
		t.Fatalf("compared = %+v, %v", compared, err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "wordingCompareRequest"))
	doer.body = sitePageToolsFixture(t, "wordingCompareNotCompared")
	short, err := domains.CompareWording(context.Background(), "d-1", AdminSiteWordingCompareRequest{First: "a", Second: "b"})
	if err != nil || short.Compared || short.SharedPercent != nil || short.MinWords != 200 || short.FirstWords != 5 {
		t.Fatalf("short = %+v, %v", short, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST " + siteTestBase + "/wording-compare", "POST " + siteTestBase + "/wording-compare",
	})
}

func TestSiteObservedWordingRouteAndEvidenceStates(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "observedWording"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	query := AdminSitePageWordingQuery{
		FirstURL: "https://example.com/plans?locale=en&view=full", SecondURL: "https://example.com/support",
	}
	ctx := context.Background()
	ready, err := domains.ComparePageWording(ctx, "d-1", query)
	if err != nil || ready.First.State != AdminSiteWordingReady || ready.First.ObservedAt == nil ||
		ready.Second.State != AdminSiteWordingReady || ready.Second.ObservedAt == nil ||
		ready.Comparison == nil || !ready.Comparison.Compared || ready.Comparison.SharedPercent == nil ||
		*ready.Comparison.SharedPercent != 96 {
		t.Fatalf("ready evidence = %+v, %v", ready, err)
	}

	doer.body = sitePageToolsFixture(t, "observedWordingUnavailable")
	unavailable, err := domains.ComparePageWording(ctx, "d-1", query)
	if err != nil || unavailable.First.State != AdminSiteWordingUnavailable || unavailable.First.ObservedAt == nil ||
		unavailable.Second.State != AdminSiteWordingUnread || unavailable.Second.ObservedAt != nil ||
		unavailable.Comparison != nil {
		t.Fatalf("unavailable evidence = %+v, %v", unavailable, err)
	}

	doer.body = sitePageToolsFixture(t, "observedWordingShort")
	short, err := domains.ComparePageWording(ctx, "d-1", query)
	if err != nil || short.First.State != AdminSiteWordingShort || short.Comparison == nil || short.Comparison.Compared ||
		short.Comparison.SharedPercent != nil || short.Comparison.FirstWords != 80 {
		t.Fatalf("short evidence = %+v, %v", short, err)
	}
	route := "GET " + siteTestBase + "/page-wording?firstUrl=https%3A%2F%2Fexample.com%2Fplans%3Flocale%3Den%26view%3Dfull&secondUrl=https%3A%2F%2Fexample.com%2Fsupport"
	assertSiteRequests(t, doer.requests, []string{route, route, route})
}

// A content-scope draft sends its scope and lists the rules it checked; a page nothing has read is
// a 404 the client can tell from a missing domain.
func TestSitePageCheckContentScopeAndPageNotRead(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, sitePageToolsFixture(t, "pageCheckContent"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	req := AdminSitePageCheckRequest{
		URL:   "https://shop.example.com/blog/pricing-guide",
		HTML:  `<h2>Pricing guide</h2><p>Short body copy about pricing.</p><img src="/chart.png">`,
		Scope: AdminSitePageCheckScopeContent,
	}
	check, err := domains.CheckPage(ctx, "d-1", req)
	if err != nil || check.Scope != AdminSitePageCheckScopeContent || len(check.CheckedRules) != 6 ||
		check.CheckedRules[0] != SiteRuleHeadingSkip || check.Page.Title != "" || len(check.Findings) != 4 {
		t.Fatalf("check = %+v, %v", check, err)
	}
	assertSiteBody(t, doer.requests[0], sitePageToolsFixture(t, "pageCheckContentRequest"))
	doer.status, doer.body = http.StatusNotFound, sitePageToolsFixture(t, "pageNotReadProblem")
	_, err = domains.GetPageFacts(ctx, "d-1", "https://shop.example.com/new")
	assertSiteProblemCode(t, err, http.StatusNotFound, SitePageNotReadCode)
}
