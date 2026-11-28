package custd

import (
	"net/http"
	"testing"
)

func TestStorefrontPartnerInvitationUsesScopedBody(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"sellerCompanyName":"Fixture Seller"}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "partner-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	result, err := client.Storefronts.AcceptBillingPartnerInvitation(t.Context(), AcceptBillingPartnerInvitationRequest{Code: "fixture"})
	if err != nil || result.SellerCompanyName != "Fixture Seller" {
		t.Fatalf("partner acceptance did not return the seller: %v", err)
	}
	request := doer.requests[0]
	if request.Method != http.MethodPost || request.URL != baseURL+billingPartnersPath+"/invitations/accept" ||
		request.Headers["Authorization"] != "Bearer partner-token" || string(request.Body) != `{"code":"fixture"}` {
		t.Fatal("partner acceptance used an incorrect route, identity or body")
	}
}

func TestStorefrontPartnerSettlementMatchesRoleScopedRoutes(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"mode":"set_off","nextAfter":"next"}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "partner-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx := t.Context()
	page, err := client.Storefronts.ListBillingPartnerSellers(ctx, BillingPartnerPageRequest{After: "previous"})
	if err != nil || page.NextAfter != "next" {
		t.Fatalf("seller page: %v", err)
	}
	if _, err := client.Storefronts.GetSellerSettlementPreference(ctx, "seller"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Storefronts.GetPartnerSettlementPreference(ctx, "partner"); err != nil {
		t.Fatal(err)
	}
	preference, err := client.Storefronts.RecordPartnerSettlementPreference(ctx, "seller", RecordPartnerSettlementPreferenceRequest{
		PreferenceUUID: "revision", Mode: PartnerSettlementSetOff,
	})
	if err != nil || preference.Mode != PartnerSettlementSetOff {
		t.Fatalf("preference: %v", err)
	}
	paths := []string{"/sellers?after=previous", "/sellers/seller/settlement-preference", "/partner/settlement-preference", "/sellers/seller/settlement-preference"}
	for index, path := range paths {
		request := doer.requests[index]
		method, body := http.MethodGet, ""
		if index == len(paths)-1 {
			method, body = http.MethodPost, `{"preferenceUuid":"revision","mode":"set_off"}`
		}
		if request.Method != method || request.URL != baseURL+billingPartnersPath+path || string(request.Body) != body {
			t.Fatalf("wrong request for %s", path)
		}
	}
}

func TestStorefrontPartnerTaxDisclosurePreservesExplicitNonregistration(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"status":"not_registered","jurisdiction":"ZA"}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "partner-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx := t.Context()
	selection, err := client.Storefronts.GetSellerTaxDisclosure(ctx, "seller")
	if err != nil || selection.Current != nil {
		t.Fatalf("missing tax declaration was not preserved: %+v, %v", selection, err)
	}
	if _, err := client.Storefronts.GetPartnerTaxDisclosure(ctx, "partner"); err != nil {
		t.Fatal(err)
	}
	disclosure, err := client.Storefronts.RecordPartnerTaxDisclosure(ctx, "seller", RecordPartnerTaxDisclosureRequest{
		DisclosureUUID: "revision", Status: PartnerTaxNotRegistered, Jurisdiction: "ZA",
	})
	if err != nil || disclosure.Status != PartnerTaxNotRegistered || disclosure.Jurisdiction != "ZA" {
		t.Fatalf("explicit declaration was not decoded: %+v, %v", disclosure, err)
	}
	paths := []string{"/sellers/seller/tax-disclosure", "/partner/tax-disclosure", "/sellers/seller/tax-disclosure"}
	for index, path := range paths {
		request := doer.requests[index]
		method, body := http.MethodGet, ""
		if index == len(paths)-1 {
			method = http.MethodPost
			body = `{"disclosureUuid":"revision","status":"not_registered","jurisdiction":"ZA"}`
		}
		if request.Method != method || request.URL != baseURL+billingPartnersPath+path || string(request.Body) != body {
			t.Fatalf("wrong request for %s", path)
		}
	}
}
