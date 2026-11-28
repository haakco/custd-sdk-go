package custd

import (
	"net/http"
	"testing"
)

func TestStorefrontPartnerStatementEncodesCursorAndExactReversal(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"movements":[{"kind":"reversal","partnerDelta":{"currency":"ZAR","amountMinor":"-9007199254740993"}}]}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "seller-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	statement, err := client.Storefronts.GetPartnerStatement(t.Context(), StorefrontPartnerStatementRequest{
		Storefront: "shop", PartnerCompanyUUID: "partner", Month: "2026-12", Cursor: "2026-12-01/entry",
	})
	if err != nil || len(statement.Movements) != 1 || statement.Movements[0].PartnerDelta.AmountMinor != "-9007199254740993" {
		t.Fatal("statement did not preserve the exact reversal")
	}
	request := doer.requests[0]
	if request.Method != http.MethodGet || request.URL != baseURL+"/api/v1/billing/storefronts/shop/partners/partner/statement?cursor=2026-12-01%2Fentry&month=2026-12" {
		t.Fatal("statement used the wrong seller route or cursor")
	}
}

func TestStorefrontAdjustmentsUseSellerInvoiceRoutesAndExactAmounts(t *testing.T) {
	const baseURL = "https://store.example"
	const invoice = baseURL + "/api/v1/storefronts/shop/invoices/invoice"
	const amount = `{"currency":"ZAR","amountMinor":"9007199254740993"}`
	const response = `{"refundUuid":"refund","creditNoteUuid":"credit","reason":"Cancelled","amount":` + amount +
		`,"totals":{"grossTotal":` + amount + `}}`
	doer := newCaptureDoer(http.StatusOK, response)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "seller-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	refund, err := client.Storefronts.RequestRefund(t.Context(), "shop", "invoice", StorefrontAdjustmentRequest{Reason: "Cancelled"})
	if err != nil || refund.Amount.AmountMinor != "9007199254740993" {
		t.Fatal("refund response lost the exact amount")
	}
	if _, err := client.Storefronts.ReconcileRefund(t.Context(), "shop", "invoice"); err != nil {
		t.Fatal(err)
	}
	note, err := client.Storefronts.IssueCreditNote(t.Context(), "shop", "invoice", StorefrontAdjustmentRequest{Reason: "Cancelled"})
	if err != nil || note.Reason != "Cancelled" || note.Totals.GrossTotal.AmountMinor != "9007199254740993" {
		t.Fatal("credit note response lost the reason or exact amount")
	}
	for i, route := range []string{"/refund", "/refund/reconcile", "/credit-note"} {
		request := doer.requests[i]
		wantBody := `{"reason":"Cancelled"}`
		if route == "/refund/reconcile" {
			wantBody = ""
		}
		if request.Method != http.MethodPost || request.URL != invoice+route || string(request.Body) != wantBody {
			t.Fatal("adjustment used the wrong storefront invoice route or body")
		}
	}
}

func TestStorefrontRegistryUsesSellerIdentity(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"sellerCompanyUuid":"seller","slug":"shop","appUuid":"app"}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "seller-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	owned, err := client.Storefronts.ListOwnedStorefronts(t.Context())
	if err != nil || owned.SellerCompanyUUID != "seller" {
		t.Fatal("could not read the authenticated seller")
	}
	binding, err := client.Storefronts.RegisterStorefront(t.Context(), RegisterStorefront{Slug: "shop", AppUUID: "app"})
	if err != nil || binding.Slug != "shop" || binding.AppUUID != "app" {
		t.Fatal("could not decode the registered storefront")
	}
	request := doer.requests[1]
	if doer.requests[0].Method != http.MethodGet || request.Method != http.MethodPost ||
		request.URL != baseURL+storefrontRegistryPath || request.Headers["Authorization"] != "Bearer seller-token" ||
		string(request.Body) != `{"slug":"shop","appUuid":"app"}` {
		t.Fatal("storefront registration did not send the correct seller request")
	}
	if _, err := client.Storefronts.SelectReceiptTemplate(t.Context(), "shop", SelectStorefrontReceiptTemplate{Revision: "fixture"}); err != nil {
		t.Fatal(err)
	}
	request = doer.requests[2]
	if request.Method != http.MethodPut || request.URL != baseURL+storefrontRegistryPath+"/shop/receipt-template" ||
		string(request.Body) != `{"revision":"fixture"}` {
		t.Fatal("receipt template selection did not send the correct seller request")
	}
}

func TestStorefrontAnalyticsDestinationUsesSellerIdentity(t *testing.T) {
	const baseURL = "https://store.example"
	const siteUUID = "0193a000-0000-7000-8000-000000000080"
	doer := newCaptureDoer(http.StatusOK, `{"enabled":false}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "seller-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })

	destination, err := client.Storefronts.GetAnalyticsDestination(t.Context(), "shop")
	if err != nil || *destination != (StorefrontAnalyticsDestination{}) {
		t.Fatalf("read unconfigured destination: %v, %v", destination, err)
	}
	doer.body = `{"siteUuid":"` + siteUUID + `","enabled":true}`
	destination, err = client.Storefronts.SelectAnalyticsDestination(t.Context(), "shop", SelectStorefrontAnalyticsDestination{SiteUUID: siteUUID})
	if err != nil || *destination != (StorefrontAnalyticsDestination{SiteUUID: siteUUID, Enabled: true}) {
		t.Fatalf("select destination: %v, %v", destination, err)
	}
	doer.body = `{"siteUuid":"` + siteUUID + `","enabled":false}`
	destination, err = client.Storefronts.DisableAnalyticsDestination(t.Context(), "shop")
	if err != nil || *destination != (StorefrontAnalyticsDestination{SiteUUID: siteUUID}) {
		t.Fatalf("disable destination: %v, %v", destination, err)
	}
	for i, want := range []struct{ method, body string }{
		{http.MethodGet, ""},
		{http.MethodPut, `{"siteUuid":"` + siteUUID + `"}`},
		{http.MethodDelete, ""},
	} {
		request := doer.requests[i]
		if request.Method != want.method || request.URL != baseURL+storefrontRegistryPath+"/shop/analytics-destination" ||
			request.Headers["Authorization"] != "Bearer seller-token" || string(request.Body) != want.body {
			t.Fatalf("analytics destination request %d used the wrong method, path, identity or body", i)
		}
	}
}

func TestMerchantAccountUsesScopedBillingRoute(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"account":{"merchantId":"fixture","revision":7,"status":"active"}}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "seller-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	result, err := client.Storefronts.ConfigureMerchant(t.Context(), ConfigureMerchant{
		MerchantID: "fixture", ExpectedRevision: 6, BusinessTimeZone: "Africa/Johannesburg",
		CheckoutSecretKey: "fixture-key", PaymentsReadToken: "fixture-read",
	})
	if err != nil || result.Account == nil || result.Account.Revision != 7 {
		t.Fatal("did not decode the configured merchant account")
	}
	request := doer.requests[0]
	if request.Method != http.MethodPut || request.URL != baseURL+merchantPath || request.Headers["Authorization"] != "Bearer seller-token" {
		t.Fatal("configured merchant through the wrong route or identity")
	}
	if _, err := client.Storefronts.RevokeMerchant(t.Context(), RevokeMerchant{ExpectedRevision: 7}); err != nil {
		t.Fatal(err)
	}
	request = doer.requests[1]
	if request.Method != http.MethodDelete || request.URL != baseURL+merchantPath ||
		string(request.Body) != `{"expectedRevision":7}` {
		t.Fatal("revocation did not include the current revision")
	}
}
