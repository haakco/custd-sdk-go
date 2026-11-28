package custd

import (
	"net/http"
	"testing"
)

func TestStorefrontUsageKeepsCertifiedQuantitiesAndNoAmountDue(t *testing.T) {
	const baseURL = "https://store.example"
	const response = `{"status":"no_amount_due","usage":{"enrollmentUuid":"enrollment","line":` +
		`{"quantity":"9007199254740993","grossAmount":{"currency":"ZAR","amountMinor":"0"}}},"nextBefore":"next"}`
	doer := newCaptureDoer(http.StatusOK, response)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "buyer-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	page, err := client.Storefronts.ListUsageEnrollments(t.Context(), StorefrontUsagePageRequest{Storefront: "shop", Before: "previous"})
	if err != nil || page.NextBefore != "next" {
		t.Fatal("did not decode the enrollment page cursor")
	}
	result, err := client.Storefronts.IssueUsageInvoice(t.Context(), "shop", "enrollment")
	if err != nil || result.Status != UsageNoAmountDue || result.Invoice != nil || result.Usage.Line.Quantity != "9007199254740993" {
		t.Fatalf("usage issuance lost the certified quantity or zero-amount outcome: %v", err)
	}
	const endpoint = baseURL + "/api/v1/storefronts/shop/usage-enrollments"
	pageRequest, issueRequest := doer.requests[0], doer.requests[1]
	if pageRequest.Method != http.MethodGet || pageRequest.URL != endpoint+"?before=previous" ||
		issueRequest.Method != http.MethodPost || issueRequest.URL != endpoint+"/enrollment/issue" || len(issueRequest.Body) != 0 {
		t.Fatal("usage calls used the wrong scoped endpoint, cursor or issue body")
	}
}
