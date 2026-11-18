package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSiteAIClient(t *testing.T) {
	const tenantBase = "http://localhost:8080/api/v1/site-domains"
	var fixture struct {
		Settings, Revoked, TenantUsage, PlatformUsage json.RawMessage
	}
	if err := json.Unmarshal(readContractFixture(t, "site-ai.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	doer := newCaptureDoer(http.StatusOK, string(fixture.Settings))
	admin := newAdminTestClient(t, doer, "http://localhost:8080").Admin
	ctx := context.Background()
	settings, err := admin.SiteDomains.GetAISettings(ctx)
	if err != nil || settings.CurrentTermsVersion != "terms-v1" || settings.ConsentAcceptedAt == nil || !settings.KeyRefused {
		t.Fatalf("settings = %+v, %v", settings, err)
	}
	settings, err = admin.SiteDomains.SetAISettings(ctx, SiteAISettingsRequest{Mode: SiteAIOwnKey, MonthlyCapUSDMicros: 10_000_000, APIKey: "test-only-key"})
	if err != nil || !settings.KeyConfigured {
		t.Fatalf("updated settings = %+v, %v", settings, err)
	}
	_, err = admin.SiteDomains.SetAIConsent(ctx, SiteAIConsentRequest{Action: SiteAIGrantConsent, TermsVersion: "terms-v1"})
	if err != nil {
		t.Fatal(err)
	}
	doer.body = string(fixture.Revoked)
	settings, err = admin.SiteDomains.SetAIConsent(ctx, SiteAIConsentRequest{Action: SiteAIRevokeConsent})
	if err != nil || settings.KeyRefused {
		t.Fatalf("revoked settings = %+v, %v", settings, err)
	}
	doer.body = string(fixture.TenantUsage)
	tenant, err := admin.SiteDomains.GetAIUsage(ctx)
	if err != nil || len(tenant.Rows) != 1 || tenant.Rows[0].CapAccountedUSDMicros != 61 || tenant.Rows[0].SamplingProduct != "" {
		t.Fatalf("tenant usage = %+v, %v", tenant, err)
	}
	doer.body = string(fixture.PlatformUsage)
	platform, _, err := admin.System.GetSiteAIUsage(ctx, SiteAIUsagePreviousMonth)
	if err != nil || platform.Period != SiteAIUsagePreviousMonth || len(platform.Rows) != 1 ||
		platform.Rows[0].SamplingProduct != SiteAIProductHosting || platform.Rows[0].EstimatedCostUSDMicros != 27 {
		t.Fatalf("platform usage = %+v, %v", platform, err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET " + tenantBase + "/ai-settings",
		"PUT " + tenantBase + "/ai-settings",
		"PUT " + tenantBase + "/ai-settings/consent",
		"PUT " + tenantBase + "/ai-settings/consent",
		"GET " + tenantBase + "/ai-usage",
		"GET http://localhost:8080/api/v1/admin/site-insights/ai-usage?period=previous-month",
	})
	if body := string(doer.requests[1].Body); body != `{"aiMode":"own-key","monthlyCapUsdMicros":10000000,"apiKey":"test-only-key"}` {
		t.Fatalf("settings body = %s", body)
	}
	if body := string(doer.requests[2].Body); body != `{"action":"grant","termsVersion":"terms-v1"}` {
		t.Fatalf("consent body = %s", body)
	}
}

func TestSiteAIProductRejectsUnknownSampleAttribution(t *testing.T) {
	var row SiteAIUsageRow
	if err := json.Unmarshal([]byte(`{"samplingProduct":"unregistered"}`), &row); err == nil {
		t.Fatal("unsupported sampling product was accepted")
	}
}
