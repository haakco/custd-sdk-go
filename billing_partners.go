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

type BillingPartnerSeller struct {
	SellerCompanyUUID string `json:"sellerCompanyUuid"`
	Name              string `json:"name"`
	Active            bool   `json:"active"`
}

type BillingPartnerSellerPage struct {
	Sellers   []BillingPartnerSeller `json:"sellers"`
	NextAfter string                 `json:"nextAfter,omitempty"`
}

type PartnerSettlementMode string

const (
	PartnerSettlementPayout PartnerSettlementMode = "payout"
	PartnerSettlementSetOff PartnerSettlementMode = "set_off"
)

type PartnerSettlementPreference struct {
	PreferenceUUID string                `json:"preferenceUuid"`
	Mode           PartnerSettlementMode `json:"mode"`
	EffectiveFrom  time.Time             `json:"effectiveFrom"`
	RecordedAt     time.Time             `json:"recordedAt"`
}

type PartnerSettlementSelection struct {
	SellerCompanyUUID  string                       `json:"sellerCompanyUuid"`
	PartnerCompanyUUID string                       `json:"partnerCompanyUuid"`
	Current            *PartnerSettlementPreference `json:"current,omitempty"`
	Next               *PartnerSettlementPreference `json:"next,omitempty"`
}

type RecordPartnerSettlementPreferenceRequest struct {
	PreferenceUUID string                `json:"preferenceUuid"`
	Mode           PartnerSettlementMode `json:"mode"`
	EffectiveFrom  *time.Time            `json:"effectiveFrom,omitempty"`
}

func (s *StorefrontClient) ListBillingPartnerSellers(ctx context.Context, input BillingPartnerPageRequest) (*BillingPartnerSellerPage, error) {
	path := billingPartnersPath + "/sellers"
	if input.After != "" {
		path += "?" + (url.Values{"after": {input.After}}).Encode()
	}
	return billingRequest[BillingPartnerSellerPage](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) GetSellerSettlementPreference(ctx context.Context, sellerUUID string) (*PartnerSettlementSelection, error) {
	path, err := buildStorefrontPath(billingPartnersPath+"/sellers", sellerUUID, "settlement-preference")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerSettlementSelection](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) GetPartnerSettlementPreference(ctx context.Context, partnerUUID string) (*PartnerSettlementSelection, error) {
	path, err := buildStorefrontPath(billingPartnersPath, partnerUUID, "settlement-preference")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerSettlementSelection](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) RecordPartnerSettlementPreference(ctx context.Context, sellerUUID string, input RecordPartnerSettlementPreferenceRequest) (*PartnerSettlementPreference, error) {
	path, err := buildStorefrontPath(billingPartnersPath+"/sellers", sellerUUID, "settlement-preference")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerSettlementPreference](ctx, s, http.MethodPost, path, input)
}

type PartnerTaxStatus string

const (
	PartnerTaxRegistered    PartnerTaxStatus = "registered"
	PartnerTaxNotRegistered PartnerTaxStatus = "not_registered"
)

type RecordPartnerTaxDisclosureRequest struct {
	DisclosureUUID     string           `json:"disclosureUuid"`
	Status             PartnerTaxStatus `json:"status"`
	Jurisdiction       string           `json:"jurisdiction"`
	RegistrationNumber string           `json:"registrationNumber,omitempty"`
	EffectiveFrom      *time.Time       `json:"effectiveFrom,omitempty"`
}

type PartnerTaxDisclosure struct {
	DisclosureUUID     string           `json:"disclosureUuid"`
	Status             PartnerTaxStatus `json:"status"`
	Jurisdiction       string           `json:"jurisdiction"`
	RegistrationNumber string           `json:"registrationNumber,omitempty"`
	EffectiveFrom      time.Time        `json:"effectiveFrom"`
	RecordedAt         time.Time        `json:"recordedAt"`
}

type PartnerTaxDisclosureSelection struct {
	SellerCompanyUUID  string                `json:"sellerCompanyUuid"`
	PartnerCompanyUUID string                `json:"partnerCompanyUuid"`
	Current            *PartnerTaxDisclosure `json:"current,omitempty"`
	Next               *PartnerTaxDisclosure `json:"next,omitempty"`
}

func (s *StorefrontClient) GetSellerTaxDisclosure(ctx context.Context, sellerUUID string) (*PartnerTaxDisclosureSelection, error) {
	path, err := buildStorefrontPath(billingPartnersPath+"/sellers", sellerUUID, "tax-disclosure")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerTaxDisclosureSelection](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) GetPartnerTaxDisclosure(ctx context.Context, partnerUUID string) (*PartnerTaxDisclosureSelection, error) {
	path, err := buildStorefrontPath(billingPartnersPath, partnerUUID, "tax-disclosure")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerTaxDisclosureSelection](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) RecordPartnerTaxDisclosure(ctx context.Context, sellerUUID string, input RecordPartnerTaxDisclosureRequest) (*PartnerTaxDisclosure, error) {
	path, err := buildStorefrontPath(billingPartnersPath+"/sellers", sellerUUID, "tax-disclosure")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerTaxDisclosure](ctx, s, http.MethodPost, path, input)
}

type RecordPartnerPayoutRequest struct {
	ExternalReference string          `json:"externalReference"`
	Amount            StorefrontMoney `json:"amount"`
	PaidAt            time.Time       `json:"paidAt"`
}

type PartnerPayout struct {
	RecordPartnerPayoutRequest
	PayoutUUID         string    `json:"payoutUuid"`
	PartnerCompanyUUID string    `json:"partnerCompanyUuid"`
	PaidOn             string    `json:"paidOn"`
	RecordedAt         time.Time `json:"recordedAt"`
	PreferenceUUID     string    `json:"preferenceUuid"`
	TaxDisclosureUUID  string    `json:"taxDisclosureUuid"`
}

type PartnerPayoutPageRequest struct {
	PartnerCompanyUUID string
	Before             string
}

type PartnerPayoutPage struct {
	Payouts    []PartnerPayout `json:"payouts"`
	NextBefore string          `json:"nextBefore,omitempty"`
}

func (s *StorefrontClient) ListPartnerPayouts(ctx context.Context, input PartnerPayoutPageRequest) (*PartnerPayoutPage, error) {
	path, err := buildStorefrontPath(billingPartnersPath, input.PartnerCompanyUUID, "payouts")
	if err != nil {
		return nil, err
	}
	if input.Before != "" {
		path += "?" + (url.Values{"before": {input.Before}}).Encode()
	}
	return billingRequest[PartnerPayoutPage](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) RecordPartnerPayout(ctx context.Context, partnerUUID string, input RecordPartnerPayoutRequest) (*PartnerPayout, error) {
	path, err := buildStorefrontPath(billingPartnersPath, partnerUUID, "payouts")
	if err != nil {
		return nil, err
	}
	return billingRequest[PartnerPayout](ctx, s, http.MethodPost, path, input)
}
