package custd

import (
	"net/http"
	"net/url"
	"testing"
)

func TestReadAccountingExportPreservesScopePagingAndExactFacts(t *testing.T) {
	const origin = "https://billing.example"
	const cursor = "2026-11-02/payment_allocation/019b0000-0000-7000-8000-000000000001"
	const payload = `{
		"sellerCompanyUuid":"seller","month":"2026-11","businessTimeZone":"Africa/Johannesburg",
		"nextAfter":"` + cursor + `","sourceGaps":["provider_fee_evidence_missing","provider_net_deposit_evidence_missing"],
		"records":[
			{"kind":"payment_allocation","recordUuid":"payment","appUuid":"shop","invoiceUuid":"invoice",
			 "postedOn":"2026-11-02","paymentAllocation":{"amount":{"currency":"ZAR","amountMinor":"9007199254740993"},
			 "verifiedAt":"2026-11-03T00:00:00Z","providerReceivedAt":"2026-11-02T20:00:00Z",
			 "provider":"yoco","providerReference":"verified-reference","providerAccountUuid":"merchant",
			 "postingBasis":"provider_received_at"}},
			{"kind":"partner_movement","recordUuid":"movement","postedOn":"2026-11-05","partnerMovement":{
			 "feeKind":"referral_fee","originalMovementUuid":"earning","partnerDelta":{"currency":"ZAR","amountMinor":"-730"},
			 "sellerCutDelta":{"currency":"ZAR","amountMinor":"0"}}},
			{"kind":"partner_payout","recordUuid":"payout","postedOn":"2026-11-06","partnerPayout":{
			 "amount":{"currency":"ZAR","amountMinor":"9007199254740993"},"externalReference":"bank-proof",
			 "allocations":[{"earningMovementUuid":"earning","amountMinor":"9007199254740993"}]}}
		]
	}`
	doer := newCaptureDoer(http.StatusOK, payload)
	client := NewClient(&ClientConfig{BaseURL: origin, APIKey: "fixture-seller-token", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	request := AccountingExportRequest{Month: "2026-11", BusinessTimeZone: "Africa/Johannesburg"}
	page, err := client.Storefronts.ReadAccountingExport(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	if page.SellerCompanyUUID != "seller" || page.Month != request.Month ||
		page.BusinessTimeZone != request.BusinessTimeZone || page.NextAfter != cursor ||
		len(page.SourceGaps) != 2 || page.SourceGaps[0] != AccountingProcessorFees ||
		page.SourceGaps[1] != AccountingNetDeposits || len(page.Records) != 3 {
		t.Fatalf("unexpected accounting page: %+v", page)
	}
	receipt := page.Records[0]
	if receipt.Kind != AccountingKindPayment || receipt.AppUUID != "shop" || receipt.Payment == nil ||
		receipt.Payment.Amount.AmountMinor != "9007199254740993" || receipt.Payment.ProviderAccountUUID != "merchant" ||
		receipt.Payment.ProviderReceivedAt == nil || receipt.Payment.PostingBasis != AccountingProviderReceiptTime {
		t.Fatalf("wrong payment allocation: %+v", receipt)
	}
	movement, payout := page.Records[1].Partner, page.Records[2].Payout
	if movement == nil || movement.FeeKind != AccountingReferralFee || movement.PartnerDelta.AmountMinor != "-730" ||
		payout == nil || len(payout.Allocations) != 1 || payout.Allocations[0].AmountMinor != "9007199254740993" {
		t.Fatal("export changed exact partner movement or payout")
	}
	request.After = page.NextAfter
	if _, err := client.Storefronts.ReadAccountingExport(t.Context(), request); err != nil {
		t.Fatal(err)
	}
	for i, after := range []string{"", cursor} {
		captured := doer.requests[i]
		target, err := url.Parse(captured.URL)
		if err != nil {
			t.Fatal(err)
		}
		query := target.Query()
		if captured.Method != http.MethodGet || target.Path != "/api/v1/billing/accounting-export" ||
			len(captured.Body) != 0 || captured.Headers["Authorization"] != "Bearer fixture-seller-token" ||
			query.Get("month") != request.Month || query.Get("businessTimeZone") != request.BusinessTimeZone ||
			query.Get("after") != after || query.Has("after") != (after != "") {
			t.Fatalf("unexpected seller accounting request %s", target)
		}
	}
}
