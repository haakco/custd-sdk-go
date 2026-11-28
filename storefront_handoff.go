package custd

import (
	"context"
	"net/http"
	"time"
)

type CreateStorefrontPurchaseHandoff struct {
	OrderReference string          `json:"orderReference"`
	OfferUUID      string          `json:"offerUuid"`
	Amount         StorefrontMoney `json:"amount"`
}

type StorefrontPurchaseHandoff struct {
	UUID           string          `json:"handoffUuid"`
	OrderReference string          `json:"orderReference"`
	OfferUUID      string          `json:"offerUuid"`
	Amount         StorefrontMoney `json:"amount"`
	CheckoutPath   string          `json:"checkoutPath"`
	ExpiresAt      time.Time       `json:"expiresAt"`
	InvoiceUUID    string          `json:"invoiceUuid,omitempty"`
	Settled        bool            `json:"settled"`
}

func (s *StorefrontClient) CreatePurchaseHandoff(ctx context.Context, storefront string, input CreateStorefrontPurchaseHandoff) (*StorefrontPurchaseHandoff, error) {
	return storefrontHandoffRequest(ctx, s, http.MethodPost, &input, storefront, "purchase-handoffs")
}

func (s *StorefrontClient) GetPurchaseHandoff(ctx context.Context, storefront, handoffUUID string) (*StorefrontPurchaseHandoff, error) {
	return storefrontHandoffRequest(ctx, s, http.MethodGet, nil, storefront, "purchase-handoffs", handoffUUID)
}

func storefrontHandoffRequest(ctx context.Context, s *StorefrontClient, method string, payload any, parts ...string) (*StorefrontPurchaseHandoff, error) {
	path, err := buildStorefrontPath(storefrontRegistryPath, parts...)
	if err != nil {
		return nil, err
	}
	return billingRequest[StorefrontPurchaseHandoff](ctx, s, method, path, payload)
}
