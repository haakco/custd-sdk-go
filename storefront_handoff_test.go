package custd

import (
	"net/http"
	"testing"
)

func TestStorefrontMachinePurchaseHandoff(t *testing.T) {
	const origin = "https://store.example"
	const path = origin + "/api/v1/billing/storefronts/shop/purchase-handoffs"
	const amount = `{"currency":"ZAR","amountMinor":"9007199254740993"}`
	doer := newCaptureDoer(http.StatusOK, `{"handoffUuid":"handoff","orderReference":"order-123","offerUuid":"offer","amount":`+amount+`,"invoiceUuid":"invoice","settled":true}`)
	client := NewClient(&ClientConfig{BaseURL: origin, APIKey: "seller-token", HTTPClient: fakeTransport(doer)})
	ctx := t.Context()
	t.Cleanup(func() { _ = client.Close(ctx) })

	result, err := client.Storefronts.CreatePurchaseHandoff(ctx, "shop", CreateStorefrontPurchaseHandoff{
		OrderReference: "order-123", OfferUUID: "offer", Amount: StorefrontMoney{Currency: "ZAR", AmountMinor: "9007199254740993"},
	})
	if err != nil || result.Amount.AmountMinor != "9007199254740993" {
		t.Fatalf("create handoff = %+v, %v", result, err)
	}
	result, err = client.Storefronts.GetPurchaseHandoff(ctx, "shop", "handoff")
	if err != nil || !result.Settled || result.InvoiceUUID != "invoice" {
		t.Fatalf("get handoff = %+v, %v", result, err)
	}
	doer.body = `{"handoffUuid":"handoff","amount":` + amount + `,"cancelled":true}`
	cancelled, err := client.Storefronts.CancelPurchaseHandoff(ctx, "shop", "handoff")
	if err != nil || !cancelled.Cancelled || cancelled.Settled {
		t.Fatalf("cancel handoff = %+v, %v", cancelled, err)
	}
	request := doer.requests[2]
	if request.Method != http.MethodPost || request.URL != path+"/handoff/cancel" || len(request.Body) != 0 {
		t.Fatal("unexpected handoff cancellation request")
	}
	created, fetched := doer.requests[0], doer.requests[1]
	if created.Method != http.MethodPost || created.URL != path ||
		string(created.Body) != `{"orderReference":"order-123","offerUuid":"offer","amount":`+amount+`}` ||
		fetched.Method != http.MethodGet || fetched.URL != path+"/handoff" {
		t.Fatal("unexpected handoff request")
	}
}
