package custd

import (
	"net/http"
	"testing"
)

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
