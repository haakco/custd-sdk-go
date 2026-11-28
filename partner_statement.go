package custd

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type StorefrontPartnerStatementRequest struct {
	Storefront         string
	PartnerCompanyUUID string
	Month              string
	Cursor             string
}

type StorefrontPartnerStatement struct {
	SellerCompanyUUID  string                      `json:"sellerCompanyUuid"`
	AppUUID            string                      `json:"appUuid"`
	PartnerCompanyUUID string                      `json:"partnerCompanyUuid"`
	Month              string                      `json:"month"`
	Movements          []StorefrontPartnerMovement `json:"movements"`
	NextCursor         string                      `json:"nextCursor,omitempty"`
}

type StorefrontPartnerMovement struct {
	MovementUUID         string                    `json:"movementUuid"`
	Kind                 string                    `json:"kind"`
	OriginalMovementUUID string                    `json:"originalMovementUuid,omitempty"`
	TermsUUID            string                    `json:"termsUuid"`
	BuyerCompanyUUID     string                    `json:"buyerCompanyUuid"`
	InvoiceUUID          string                    `json:"invoiceUuid"`
	PaymentUUID          string                    `json:"paymentUuid"`
	LineNumber           int                       `json:"lineNumber"`
	PostedOn             string                    `json:"postedOn"`
	AvailableOn          string                    `json:"availableOn"`
	PartnerDelta         StorefrontMoney           `json:"partnerDelta"`
	SellerCutDelta       StorefrontMoney           `json:"sellerCutDelta"`
	Earning              *StorefrontPartnerEarning `json:"earning,omitempty"`
}

type StorefrontPartnerEarning struct {
	Gross         StorefrontMoney `json:"gross"`
	Tax           StorefrontMoney `json:"tax"`
	Net           StorefrontMoney `json:"net"`
	CostToProvide StorefrontMoney `json:"costToProvide"`
	Remainder     StorefrontMoney `json:"remainder"`
}

func (s *StorefrontClient) GetPartnerStatement(ctx context.Context, input StorefrontPartnerStatementRequest) (*StorefrontPartnerStatement, error) {
	for _, segment := range []string{input.Storefront, input.PartnerCompanyUUID} {
		if segment == "" || segment == "." || segment == ".." || strings.ContainsAny(segment, "/\\") {
			return nil, errors.New("custd: invalid partner statement path segment")
		}
	}
	query := url.Values{"month": {input.Month}}
	if input.Cursor != "" {
		query.Set("cursor", input.Cursor)
	}
	path := storefrontRegistryPath + "/" + url.PathEscape(input.Storefront) + "/partners/" +
		url.PathEscape(input.PartnerCompanyUUID) + "/statement?" + query.Encode()
	return billingRequest[StorefrontPartnerStatement](ctx, s, http.MethodGet, path, nil)
}
