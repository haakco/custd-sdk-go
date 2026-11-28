package custd

import (
	"net/http"
	"testing"
)

func TestPartnerReferralTermsPassSignedVersionAndDateCursor(t *testing.T) {
	const origin = "https://billing.example"
	const path = origin + billingPartnersPath + "/partner/referral-terms"
	doer := newCaptureDoer(http.StatusOK, `{"code":"PARTNER-1","businessTimeZone":"Africa/Johannesburg",
		"versions":[{"termsUuid":"terms","effectiveFrom":"2026-12-01","agreementDate":"2026-11-10",
		"agreementReference":"agreement-1","recordedByUserUuid":"owner","share":"0.125","months":12,"holdDays":30}],
		"nextBefore":"2026-12-01"}`)
	client := NewClient(&ClientConfig{BaseURL: origin, APIKey: "fixture-seller", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx := t.Context()

	terms, err := client.Storefronts.ListPartnerReferralTerms(ctx, "partner", PartnerReferralPageRequest{Before: "2027-01-01"})
	if err != nil || terms.Code != "PARTNER-1" || terms.NextBefore != "2026-12-01" || len(terms.Versions) != 1 ||
		terms.Versions[0].RecordedByUserUUID != "owner" || terms.Versions[0].Share != "0.125" {
		t.Fatalf("unexpected terms: %+v, %v", terms, err)
	}
	request := PublishReferralTerms{
		Code: "PARTNER-1", EffectiveFrom: "2026-12-01", AgreementDate: "2026-11-10", AgreementReference: "agreement-1",
		SignedAgreementAttested: true, Share: "0.125", Months: 12, HoldDays: 30,
	}
	published, err := client.Storefronts.PublishPartnerReferralTerms(ctx, "partner", request)
	if err != nil || len(published.Versions) != 1 || published.Versions[0].EffectiveFrom != request.EffectiveFrom {
		t.Fatalf("unexpected published terms: %+v, %v", published, err)
	}
	read, publish := doer.requests[0], doer.requests[1]
	if read.Method != http.MethodGet || read.URL != path+"?before=2027-01-01" ||
		len(read.Body) != 0 || read.Headers["Authorization"] != "Bearer fixture-seller" {
		t.Fatal("terms history used an incorrect identity, route or query")
	}
	if publish.Method != http.MethodPost || publish.URL != path || publish.Headers["Authorization"] != "Bearer fixture-seller" ||
		string(publish.Body) != `{"code":"PARTNER-1","effectiveFrom":"2026-12-01","agreementDate":"2026-11-10","agreementReference":"agreement-1","signedAgreementAttested":true,"share":"0.125","months":12,"holdDays":30}` {
		t.Fatal("terms publication did not preserve the seller's signed agreement intent")
	}
}

func TestReferralReviewPagesAndRetriesRecordedEvidence(t *testing.T) {
	const origin = "https://billing.example"
	const path = origin + partnerReferralReviewPath
	const event = "019b0000-0000-7000-8000-000000000001"
	doer := newCaptureDoer(http.StatusOK, `{"reviews":[{"attributionUuid":"attribution","eventUuid":"`+event+`",
		"invoiceUuid":"invoice","eventType":"billing.payment.settled","failureCode":"earlier_payment_unresolved",
		"reviewedAt":"2026-11-28T07:10:00Z","retryEligible":true}],"nextBefore":"`+event+`"}`)
	client := NewClient(&ClientConfig{BaseURL: origin, APIKey: "fixture-seller", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx := t.Context()

	page, err := client.Storefronts.ListReferralReviews(ctx, PartnerReferralPageRequest{})
	if err != nil || page.NextBefore != event || len(page.Reviews) != 1 || !page.Reviews[0].RetryEligible ||
		page.Reviews[0].EventType != ReferralReviewPaymentSettled || page.Reviews[0].ReviewedAt.IsZero() {
		t.Fatalf("unexpected review page: %+v, %v", page, err)
	}
	if _, err := client.Storefronts.ListReferralReviews(ctx, PartnerReferralPageRequest{Before: page.NextBefore}); err != nil {
		t.Fatal(err)
	}
	doer.body = `{"attributionUuid":"attribution","status":"requeued","requeuedCount":1,"hasMore":true}`
	result, err := client.Storefronts.RetryReferralReview(ctx, page.Reviews[0].AttributionUUID)
	if err != nil || result.Status != ReferralRequeued || result.RequeuedCount != 1 || !result.HasMore {
		t.Fatalf("unexpected review retry: %+v, %v", result, err)
	}
	first, next, retry := doer.requests[0], doer.requests[1], doer.requests[2]
	if first.Method != http.MethodGet || first.URL != path || next.Method != http.MethodGet ||
		next.URL != path+"?before="+event || retry.Method != http.MethodPost || retry.URL != path+"/attribution/retry" ||
		retry.Headers["Authorization"] != "Bearer fixture-seller" || len(retry.Body) != 0 {
		t.Fatal("referral review requests changed the event cursor or attached retry evidence")
	}
}
