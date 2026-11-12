package custd

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

const siteDomainFixture = `{"domainUuid":"d-1","siteUuid":"s-1","host":"www.example.com",
	"ownership":{"verified":true,"verifiedAt":"2026-11-10T08:00:00Z","verifiedBy":"dns",
		"instructions":{"dnsRecordName":"_custd-verification.www.example.com",
			"dnsRecordValue":"custd-verification=tok","fileUrl":"https://www.example.com/.well-known/custd-verification.txt",
			"fileContent":"tok"}},
	"audit":{"lastAuditedAt":"2026-11-11T08:00:00Z","lastSummary":{"high":1,"medium":2,"low":0,"info":3,
		"pagesChecked":4,"requests":9,"truncated":false}},
	"createdAt":"2026-11-09T08:00:00Z"}`

func TestSiteDomainOperationsSendTheirRoutesWithActingCompany(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteDomainFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "acme"
	domains := client.Admin.SiteDomains
	ctx := context.Background()

	if _, err := domains.Create(ctx, AdminSiteDomainCreate{SiteUUID: "s-1", Host: "www.example.com"}); err != nil {
		t.Fatal(err)
	}
	if got := string(doer.requests[0].Body); got != `{"siteUuid":"s-1","host":"www.example.com"}` {
		t.Fatalf("create body = %s", got)
	}
	calls := []func() error{
		func() error { _, err := domains.List(ctx, "s-1"); return err },
		func() error { _, err := domains.Get(ctx, "d-1"); return err },
		func() error { _, err := domains.Verify(ctx, "d-1"); return err },
		func() error { _, err := domains.RequestAudit(ctx, "d-1"); return err },
		func() error { _, err := domains.ListFindings(ctx, "d-1", AdminSiteFindingStateAll); return err },
		func() error { return domains.Delete(ctx, "d-1") },
	}
	for _, call := range calls {
		if err := call(); err != nil {
			t.Fatal(err)
		}
	}
	base := "http://localhost:8080/api/v1/site-domains"
	assertSiteRequests(t, doer.requests, []string{
		"POST " + base + "?companySlug=acme",
		"GET " + base + "?companySlug=acme&siteUuid=s-1",
		"GET " + base + "/d-1?companySlug=acme",
		"POST " + base + "/d-1/verify?companySlug=acme",
		"POST " + base + "/d-1/audit?companySlug=acme",
		"GET " + base + "/d-1/findings?companySlug=acme&state=all",
		"DELETE " + base + "/d-1?companySlug=acme",
	})
}

// A site-bound credential names no Site: neither the body nor the query carries one.
func TestSiteBoundCredentialOmitsItsSite(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteDomainFixture)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	if _, err := domains.Create(ctx, AdminSiteDomainCreate{Host: "shop.example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := domains.List(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if got := string(doer.requests[0].Body); got != `{"host":"shop.example.com"}` {
		t.Fatalf("create body = %s", got)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/site-domains", "GET http://localhost:8080/api/v1/site-domains",
	})
}

func TestSiteDomainResponsesDecode(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteDomainFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	domain, err := client.Admin.SiteDomains.Get(context.Background(), "d-1")
	if err != nil {
		t.Fatal(err)
	}
	summary := domain.Audit.LastSummary
	if !domain.Ownership.Verified || domain.Ownership.Instructions.FileContent != "tok" ||
		summary == nil || summary.High != 1 || summary.PagesChecked != 4 || domain.CreatedAt.IsZero() {
		t.Fatalf("domain = %+v", domain)
	}

	doer.body = `{"findings":[{"findingUuid":"f-1","rule":"tls.expiring","severity":"high","title":"Certificate expires soon",
		"url":"https://www.example.com/","evidence":"9 days","fix":"Renew it","source":"tls",
		"history":{"firstSeenAt":"2026-11-10T08:00:00Z","lastSeenAt":"2026-11-11T08:00:00Z","fixed":false}}]}`
	findings, err := client.Admin.SiteDomains.ListFindings(context.Background(), "d-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := doer.requests[1].URL; got != "http://localhost:8080/api/v1/site-domains/d-1/findings" {
		t.Fatalf("findings URL = %s", got)
	}
	if len(findings.Findings) != 1 || findings.Findings[0].Severity != "high" || findings.Findings[0].History.Fixed {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSiteDomainAuditOfUnverifiedDomainSurfacesProblemCode(t *testing.T) {
	doer := newCaptureDoer(http.StatusConflict, `{"type":"about:blank","title":"Conflict","status":409,
		"detail":"site domain: verify the domain before asking for an audit","code":"site_domain_not_verified"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.SiteDomains.RequestAudit(context.Background(), "d-1")
	var requestErr *RequestError
	if !errors.As(err, &requestErr) || requestErr.StatusCode != http.StatusConflict ||
		requestErr.Problem == nil || requestErr.Problem.Code != SiteDomainNotVerifiedCode {
		t.Fatalf("err = %#v", err)
	}
}
