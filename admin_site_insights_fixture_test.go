package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// siteInsightsFixture returns one recorded admin-api response from the shared
// contract-fixtures/site-insights.json, which Custd's siteaudit types encoded.
func siteInsightsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-insights.json", key)
}

// recordedFixture returns one recorded value from a shared contract fixture keyed by name.
func recordedFixture(t *testing.T, name, key string) string {
	t.Helper()
	var recorded map[string]json.RawMessage
	if err := json.Unmarshal(readContractFixture(t, name), &recorded); err != nil {
		t.Fatal(err)
	}
	body, ok := recorded[key]
	if !ok {
		t.Fatalf("%s has no %q", name, key)
	}
	return string(body)
}

// A broker attests and withdraws on the domain's attestation route; the recorded domain
// decodes its attestation, and a subdomain names the parent that verifies it.
func TestSiteDomainAttestationRoutesAndOwnershipDecode(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteInsightsFixture(t, "domain"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	attested, err := domains.Attest(ctx, "d-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := domains.Unattest(ctx, "d-1"); err != nil {
		t.Fatal(err)
	}
	base := "http://localhost:8080/api/v1/site-domains/d-1/attestation"
	assertSiteRequests(t, doer.requests, []string{"POST " + base, "DELETE " + base})
	attestation := attested.Ownership.Attestation
	if attested.Ownership.VerifiedBy != "attested" || attestation == nil ||
		attestation.ClientID != "awthy-site-broker" || attestation.AttestedAt.IsZero() {
		t.Fatalf("ownership = %+v", attested.Ownership)
	}

	doer.body = siteInsightsFixture(t, "parentVerifiedDomain")
	child, err := domains.Get(ctx, "d-3")
	if err != nil {
		t.Fatal(err)
	}
	if child.Ownership.VerifiedBy != "parent" || child.Ownership.Attestation != nil ||
		child.Ownership.VerifiedViaDomainUUID != "0198a3b2-0000-7000-8000-000000000001" {
		t.Fatalf("ownership = %+v", child.Ownership)
	}
}

func TestSiteSearchPerformanceDecodesComparisonDailyAndStrikingDistance(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteInsightsFixture(t, "searchPerformance"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains

	performance, err := domains.GetSearchPerformance(context.Background(), "d-1", 28)
	if err != nil {
		t.Fatal(err)
	}
	source := performance.Sources[0]
	if source.Previous == nil || source.Previous.Clicks != 2 || source.Previous.Impressions != 100 ||
		len(source.Daily) != 1 || source.Daily[0] != (AdminSiteSearchDay{Day: "2026-11-15", Clicks: 4, Impressions: 200}) {
		t.Fatalf("source = %+v", source)
	}
	striking := performance.StrikingDistance
	if len(striking) != 1 || striking[0].Key != "cheap widgets" || striking[0].Totals.Position != 11.2 {
		t.Fatalf("strikingDistance = %+v", striking)
	}
}

func TestSiteFindingRuleMatchesRuleConstant(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteInsightsFixture(t, "findingList"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains

	findings, err := domains.ListFindings(context.Background(), "d-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if findings.Findings[0].Rule != SiteRuleTitleMissing || findings.Findings[1].Rule != SiteRuleUptimeDown {
		t.Fatalf("findings = %+v", findings.Findings)
	}
	title, uptime := findings.Findings[0].Remedy, findings.Findings[1].Remedy
	if title != (AdminSiteRemedy{Kind: AdminSiteRemedyKindContent, Element: "<title>"}) ||
		uptime.Kind != AdminSiteRemedyKindServer || uptime.Generated != nil {
		t.Fatalf("remedies = %+v, %+v", title, uptime)
	}
}
