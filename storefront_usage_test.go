package custd

import (
	"net/http"
	"testing"
)

func TestStorefrontUsageKeepsCertifiedQuantitiesAndNoAmountDue(t *testing.T) {
	const baseURL = "https://store.example"
	const response = `{"status":"no_amount_due","usage":{"enrollmentUuid":"enrollment","line":` +
		`{"quantity":"9007199254740993","grossAmount":{"currency":"ZAR","amountMinor":"0"}}},"nextBefore":"next","canSchedule":true,"enrollments":[{"tariffDescription":"Metered events"}]}`
	doer := newCaptureDoer(http.StatusOK, response)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "buyer-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	page, err := client.Storefronts.ListUsageEnrollments(t.Context(), StorefrontUsagePageRequest{Storefront: "shop", Before: "previous"})
	if err != nil || page.NextBefore != "next" || !page.CanSchedule || len(page.Enrollments) != 1 || page.Enrollments[0].TariffDescription != "Metered events" {
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

func TestStorefrontUsageScheduleSendsExplicitConsent(t *testing.T) {
	const baseURL = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"enrollmentUuid":"enrollment","state":"scheduled"}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "buyer-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx := t.Context()
	schedule, err := client.Storefronts.ScheduleUsageInvoice(ctx, "shop", "enrollment", UsageInvoiceOptIn{InvoiceOnlyAttested: true})
	if err != nil || schedule.State != UsageScheduled || schedule.EnrollmentUUID != "enrollment" {
		t.Fatalf("decode schedule: %+v, %v", schedule, err)
	}
	if _, err := client.Storefronts.GetUsageSchedule(ctx, "shop", "enrollment"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Storefronts.CancelUsageSchedule(ctx, "shop", "enrollment"); err != nil {
		t.Fatal(err)
	}
	const endpoint = baseURL + "/api/v1/storefronts/shop/usage-enrollments/enrollment/scheduled-issuance"
	for i, method := range []string{http.MethodPost, http.MethodGet, http.MethodDelete} {
		request := doer.requests[i]
		if request.Method != method || request.URL != endpoint {
			t.Fatalf("wrong schedule request for %s: %s %s", method, request.Method, request.URL)
		}
		wantBody := ""
		if method == http.MethodPost {
			wantBody = `{"invoiceOnlyAttested":true}`
		}
		if string(request.Body) != wantBody {
			t.Fatalf("wrong schedule body for %s: %s", method, request.Body)
		}
	}
}

func TestStorefrontSetOffUsesInvoiceOnlyRoutes(t *testing.T) {
	const base = "https://store.example"
	doer := newCaptureDoer(http.StatusOK, `{"eligible":true,"amount":{"currency":"ZAR","amountMinor":"9007199254740993"}}`)
	client := NewClient(&ClientConfig{BaseURL: base, APIKey: "buyer-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx := t.Context()
	if _, err := client.Storefronts.IssueStorefrontInvoice(ctx, "shop", "quote"); err != nil {
		t.Fatal(err)
	}
	eligibility, err := client.Storefronts.GetStorefrontSetOffEligibility(ctx, "shop", "invoice")
	if err != nil || !eligibility.Eligible {
		t.Fatalf("eligibility = %+v, %v", eligibility, err)
	}
	setoff, err := client.Storefronts.ApplyStorefrontSetOff(ctx, "shop", "invoice")
	if err != nil || setoff.Amount.AmountMinor != "9007199254740993" {
		t.Fatalf("set-off = %+v, %v", setoff, err)
	}
	want := []struct{ method, path string }{
		{http.MethodPost, "/quotes/quote/invoice"},
		{http.MethodGet, "/purchases/invoice/set-off"},
		{http.MethodPost, "/purchases/invoice/set-off"},
	}
	for index, expected := range want {
		request := doer.requests[index]
		if request.Method != expected.method || request.URL != base+"/api/v1/storefronts/shop"+expected.path || len(request.Body) != 0 {
			t.Fatalf("unexpected set-off request for %s", expected.path)
		}
	}
}
