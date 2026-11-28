package custd

import (
	"context"
	"net/http"
	"time"
)

const merchantPath = "/api/v1/billing/merchant"

type MerchantAccount struct {
	UUID              string    `json:"accountUuid"`
	SellerCompanyUUID string    `json:"sellerCompanyUuid"`
	MerchantID        string    `json:"merchantId"`
	Mode              string    `json:"mode"`
	Revision          int64     `json:"revision"`
	Status            string    `json:"status"`
	RefundsEnabled    bool      `json:"refundsEnabled"`
	BusinessTimeZone  string    `json:"businessTimeZone"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type MerchantAccountResponse struct {
	Account *MerchantAccount `json:"account,omitempty"`
}

type ConfigureMerchant struct {
	MerchantID        string `json:"merchantId"`
	ExpectedRevision  int64  `json:"expectedRevision"`
	BusinessTimeZone  string `json:"businessTimeZone"`
	RefundsEnabled    bool   `json:"refundsEnabled"`
	CheckoutSecretKey string `json:"checkoutSecretKey"`
	PaymentsReadToken string `json:"paymentsReadToken"`
}

type RevokeMerchant struct {
	ExpectedRevision int64 `json:"expectedRevision"`
}

type VerifyMerchant struct {
	AppUUID     string `json:"appUuid"`
	InvoiceUUID string `json:"invoiceUuid"`
}

type VerifyMerchantResult struct {
	Verified bool `json:"verified"`
}

func (s *StorefrontClient) GetMerchant(ctx context.Context) (*MerchantAccountResponse, error) {
	return billingRequest[MerchantAccountResponse](ctx, s, http.MethodGet, merchantPath, nil)
}

func (s *StorefrontClient) ConfigureMerchant(ctx context.Context, input ConfigureMerchant) (*MerchantAccountResponse, error) {
	return billingRequest[MerchantAccountResponse](ctx, s, http.MethodPut, merchantPath, input)
}

func (s *StorefrontClient) RevokeMerchant(ctx context.Context, input RevokeMerchant) (*MerchantAccountResponse, error) {
	return billingRequest[MerchantAccountResponse](ctx, s, http.MethodDelete, merchantPath, input)
}

func (s *StorefrontClient) VerifyMerchant(ctx context.Context, input VerifyMerchant) (*VerifyMerchantResult, error) {
	return billingRequest[VerifyMerchantResult](ctx, s, http.MethodPost, merchantPath+"/test", input)
}
