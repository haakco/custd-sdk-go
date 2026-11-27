package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestStorefrontQuantityAndPurchase(t *testing.T) {
	const baseURL = "https://store.example"
	const storePath = baseURL + "/api/v1/storefronts/custd"
	const money = `{"currency":"ZAR","amountMinor":"9007199254740993"}`
	const totals = `{"netTotal":` + money + `,"taxTotal":` + money + `,"grossTotal":` + money + `}`
	doer := newCaptureDoer(http.StatusOK, `{"invoiceUuid":"invoice","status":"open","paymentStatus":"pending","totals":`+totals+`}`)
	client := NewClient(&ClientConfig{BaseURL: baseURL, APIKey: "buyer-token", HTTPClient: fakeTransport(doer)})
	ctx := context.Background()
	t.Cleanup(func() { _ = client.Close(ctx) })

	_, err := client.Storefronts.UpdateLine(ctx, "custd", "cart", "offer", StorefrontLineQuantity{Quantity: "1.25"})
	if err != nil {
		t.Fatal(err)
	}
	purchase, err := client.Storefronts.StartPayment(ctx, "custd", "invoice")
	if err != nil || purchase.InvoiceUUID != "invoice" || purchase.PaymentStatus != StorefrontPaymentPending ||
		purchase.Totals.GrossTotal.AmountMinor != "9007199254740993" {
		t.Fatalf("decoded purchase = %+v, %v", purchase, err)
	}
	if len(doer.requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(doer.requests))
	}
	line, retry := doer.requests[0], doer.requests[1]
	if line.Method != http.MethodPut || line.URL != storePath+"/carts/cart/lines/offer" ||
		string(line.Body) != `{"quantity":"1.25"}` || line.Headers["Authorization"] != "Bearer buyer-token" {
		t.Fatalf("line request = %+v", line)
	}
	if retry.Method != http.MethodPost || retry.URL != storePath+"/purchases/invoice/payment" ||
		len(retry.Body) != 0 || retry.Headers["Authorization"] != "Bearer buyer-token" {
		t.Fatalf("retry request = %+v", retry)
	}
	doer.body = `{"totals":{"grossTotal":{"currency":"ZAR","amountMinor":9007199254740993}}}`
	if _, err := client.Storefronts.GetPurchase(ctx, "custd", "invoice"); err == nil {
		t.Fatal("accepted a numeric minor-unit amount")
	}
}
