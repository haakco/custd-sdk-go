package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

const billingPartnersPath = "/api/v1/billing/partners"

type BillingPartner struct {
	PartnerCompanyUUID string `json:"partnerCompanyUuid"`
	Name               string `json:"name"`
	Active             bool   `json:"active"`
}

type BillingPartnerPageRequest struct {
	After string
}

type BillingPartnerPage struct {
	Partners  []BillingPartner `json:"partners"`
	NextAfter string           `json:"nextAfter,omitempty"`
}

type BillingPartnerInvitation struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type AcceptBillingPartnerInvitationRequest struct {
	Code string `json:"code"`
}

type AcceptedBillingPartnerInvitation struct {
	SellerCompanyName string `json:"sellerCompanyName"`
}

type PublishPartnerProductTermsRequest struct {
	PartnerCompanyUUID      string `json:"partnerCompanyUuid"`
	EffectiveFrom           string `json:"effectiveFrom"`
	AgreementDate           string `json:"agreementDate"`
	AgreementReference      string `json:"agreementReference"`
	SignedAgreementAttested bool   `json:"signedAgreementAttested"`
	UnitCost                string `json:"unitCost"`
	SellerCut               string `json:"sellerCut"`
}

type PartnerProductTermsVersion struct {
	PublishPartnerProductTermsRequest
	TermsUUID          string `json:"termsUuid"`
	RecordedByUserUUID string `json:"recordedByUserUuid"`
	Currency           string `json:"currency"`
	HoldDays           int    `json:"holdDays"`
}

type PartnerProductTermsPageRequest struct {
	Storefront string
	OfferUUID  string
	Before     string
}

type PartnerProductTermsPage struct {
	Versions   []PartnerProductTermsVersion `json:"versions"`
	NextBefore string                       `json:"nextBefore,omitempty"`
}

func (s *StorefrontClient) ListBillingPartners(ctx context.Context, input BillingPartnerPageRequest) (*BillingPartnerPage, error) {
	path := billingPartnersPath
	if input.After != "" {
		path += "?" + (url.Values{"after": {input.After}}).Encode()
	}
	return billingRequest[BillingPartnerPage](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) CreateBillingPartnerInvitation(ctx context.Context) (*BillingPartnerInvitation, error) {
	return billingRequest[BillingPartnerInvitation](ctx, s, http.MethodPost, billingPartnersPath+"/invitations", nil)
}

func (s *StorefrontClient) AcceptBillingPartnerInvitation(ctx context.Context, input AcceptBillingPartnerInvitationRequest) (*AcceptedBillingPartnerInvitation, error) {
	return billingRequest[AcceptedBillingPartnerInvitation](ctx, s, http.MethodPost, billingPartnersPath+"/invitations/accept", input)
}

func (s *StorefrontClient) ListPartnerProductTerms(ctx context.Context, input PartnerProductTermsPageRequest) (*PartnerProductTermsPage, error) {
	path, err := buildStorefrontPath(storefrontRegistryPath, input.Storefront, "offers", input.OfferUUID, "partner-terms")
	if err != nil {
		return nil, err
	}
	if input.Before != "" {
		path += "?" + (url.Values{"before": {input.Before}}).Encode()
	}
	return billingRequest[PartnerProductTermsPage](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) PublishPartnerProductTerms(ctx context.Context, storefront, offerUUID string, input PublishPartnerProductTermsRequest) (*PartnerProductTermsVersion, error) {
	path, err := buildStorefrontPath(storefrontRegistryPath, storefront, "offers", offerUUID, "partner-terms")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerProductTermsVersion](ctx, s, http.MethodPost, path, input)
}
