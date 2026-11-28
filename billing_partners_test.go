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
