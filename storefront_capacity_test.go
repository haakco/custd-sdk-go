package custd

import (
	"net/http"
	"testing"
)

func TestPublishCapacityOfferPreservesStableOfferAndTypedGrant(t *testing.T) {
	const origin = "https://billing.example"
	const offerUUID = "019b0000-0000-7000-8000-000000000001"
	const requestBody = `{"offerUuid":"` + offerUUID + `","name":"Capacity bundle","billingInterval":"month",` +
		`"unitPrice":{"currency":"ZAR","amount":"30.00"},"taxRate":"0.15","taxLabel":"VAT",` +
		`"dataSpaceCapacity":{"extraSpacesPerUnit":2,"extraProducersPerSpacePerUnit":5}}`
	doer := newCaptureDoer(http.StatusCreated, `{"offerUuid":"`+offerUUID+`","name":"Capacity bundle","billingInterval":"month",`+
		`"priceRevision":{"priceRevisionUuid":"revision","revision":1,"unitPrice":{"currency":"ZAR","amount":"30.00"},`+
		`"taxTreatment":"exclusive","taxRate":"0.15","taxLabel":"VAT","effectiveFrom":"2026-11-28T00:00:00Z"},`+
		`"dataSpaceCapacity":{"extraSpacesPerUnit":2,"extraProducersPerSpacePerUnit":5}}`)
	client := NewClient(&ClientConfig{BaseURL: origin, APIKey: "fixture-platform-admin", HTTPClient: fakeTransport(doer)})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	input := PublishStorefrontCapacityOffer{
		OfferUUID: offerUUID, Name: "Capacity bundle", BillingInterval: "month",
		UnitPrice: StorefrontUnitPrice{Currency: "ZAR", Amount: "30.00"}, TaxRate: "0.15", TaxLabel: "VAT",
		DataSpaceCapacity: StorefrontDataSpaceCapacity{ExtraSpacesPerUnit: 2, ExtraProducersPerSpacePerUnit: 5},
	}

	for i := range 2 {
		result, err := client.Storefronts.PublishCapacityOffer(t.Context(), "shop", input)
		if err != nil {
			t.Fatal(err)
		}
		if result.UUID != offerUUID || result.PriceRevision.Revision != 1 || result.PriceRevision.UnitPrice != input.UnitPrice ||
			result.DataSpaceCapacity == nil || *result.DataSpaceCapacity != input.DataSpaceCapacity {
			t.Fatalf("wrong published capacity: %+v", result)
		}
		request := doer.requests[i]
		if request.Method != http.MethodPost || request.URL != origin+"/api/v1/billing/storefronts/shop/capacity-offers" ||
			request.Headers["Authorization"] != "Bearer fixture-platform-admin" || string(request.Body) != requestBody {
			t.Fatalf("publication attempt %d changed the scope, route or exact payload", i+1)
		}
	}
}
