package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AccountingExportRequest selects a seller month, business time zone and optional page cursor.
type AccountingExportRequest struct {
	Month            string
	BusinessTimeZone string
	After            string
}

// AccountingExport is one page of the authenticated seller’s source facts and evidence gaps.
type AccountingExport struct {
	SellerCompanyUUID string             `json:"sellerCompanyUuid"`
	Month             string             `json:"month"`
	BusinessTimeZone  string             `json:"businessTimeZone"`
	Records           []AccountingRecord `json:"records"`
	SourceGaps        []AccountingGap    `json:"sourceGaps"`
	NextAfter         string             `json:"nextAfter,omitempty"`
}

type AccountingGap string

const (
	AccountingProcessorFees     AccountingGap = "provider_fee_evidence_missing"
	AccountingNetDeposits       AccountingGap = "provider_net_deposit_evidence_missing"
	AccountingSettlementBatches AccountingGap = "provider_settlement_batch_reference_missing"
)

type AccountingRecordKind string

const (
	AccountingKindInvoice        AccountingRecordKind = "invoice"
	AccountingKindCreditNote     AccountingRecordKind = "credit_note"
	AccountingKindPayment        AccountingRecordKind = "payment_allocation"
	AccountingKindRefund         AccountingRecordKind = "refund"
	AccountingKindPartner        AccountingRecordKind = "partner_movement"
	AccountingKindPayout         AccountingRecordKind = "partner_payout"
	AccountingKindSetOff         AccountingRecordKind = "partner_setoff"
	AccountingKindSetOffReversal AccountingRecordKind = "partner_setoff_reversal"
	AccountingKindProductGap     AccountingRecordKind = "partner_product_projection_gap"
	AccountingKindReferralGap    AccountingRecordKind = "partner_referral_projection_gap"
)

type AccountingRecord struct {
	Kind             AccountingRecordKind     `json:"kind"`
	UUID             string                   `json:"recordUuid"`
	AppUUID          string                   `json:"appUuid,omitempty"`
	InvoiceUUID      string                   `json:"invoiceUuid,omitempty"`
	InvoiceNumber    string                   `json:"invoiceNumber,omitempty"`
	BuyerCompanyUUID string                   `json:"buyerCompanyUuid,omitempty"`
	PostedOn         string                   `json:"postedOn"`
	Document         *AccountingDocument      `json:"document,omitempty"`
	Payment          *AccountingPayment       `json:"paymentAllocation,omitempty"`
	Refund           *AccountingRefund        `json:"refund,omitempty"`
	Partner          *AccountingMovement      `json:"partnerMovement,omitempty"`
	Payout           *AccountingPayout        `json:"partnerPayout,omitempty"`
	SetOff           *AccountingSetOff        `json:"partnerSetOff,omitempty"`
	Projection       *AccountingProjectionGap `json:"partnerProjectionGap,omitempty"`
}

type AccountingDocument struct {
	Number   string                `json:"number"`
	IssuedAt time.Time             `json:"issuedAt"`
	Lines    []StorefrontQuoteLine `json:"lines"`
	Totals   StorefrontTotals      `json:"totals"`
}

type AccountingPostingBasis string

const (
	AccountingProviderReceiptTime AccountingPostingBasis = "provider_received_at"
	AccountingRecordedReceiptDate AccountingPostingBasis = "recorded_received_date"
	AccountingVerificationDate    AccountingPostingBasis = "verification_date"
)

type AccountingPayment struct {
	Amount              StorefrontMoney        `json:"amount"`
	VerifiedAt          time.Time              `json:"verifiedAt"`
	ProviderReceivedAt  *time.Time             `json:"providerReceivedAt,omitempty"`
	Provider            string                 `json:"provider"`
	ProviderReference   string                 `json:"providerReference"`
	ProviderAccountUUID string                 `json:"providerAccountUuid,omitempty"`
	PostingBasis        AccountingPostingBasis `json:"postingBasis"`
}

type AccountingRefund struct {
	Amount              StorefrontMoney `json:"amount"`
	PaymentUUID         string          `json:"paymentUuid"`
	ProcessedAt         time.Time       `json:"processedAt"`
	Provider            string          `json:"provider"`
	ProviderReference   string          `json:"providerReference"`
	ProviderAccountUUID string          `json:"providerAccountUuid,omitempty"`
}

type AccountingFeeKind string

const (
	AccountingProductFee  AccountingFeeKind = "product_fee"
	AccountingReferralFee AccountingFeeKind = "referral_fee"
)

type AccountingMovement struct {
	FeeKind              AccountingFeeKind `json:"feeKind"`
	PartnerCompanyUUID   string            `json:"partnerCompanyUuid"`
	OriginalMovementUUID string            `json:"originalMovementUuid,omitempty"`
	PaymentUUID          string            `json:"paymentUuid"`
	TermsUUID            string            `json:"termsUuid"`
	LineNumber           int               `json:"lineNumber"`
	AvailableOn          string            `json:"availableOn"`
	PartnerDelta         StorefrontMoney   `json:"partnerDelta"`
	SellerCutDelta       StorefrontMoney   `json:"sellerCutDelta"`
}

type AccountingPayout struct {
	PartnerCompanyUUID string                        `json:"partnerCompanyUuid"`
	Amount             StorefrontMoney               `json:"amount"`
	ExternalReference  string                        `json:"externalReference"`
	PaidAt             time.Time                     `json:"paidAt"`
	PreferenceUUID     string                        `json:"preferenceUuid"`
	TaxDisclosureUUID  string                        `json:"taxDisclosureUuid"`
	Allocations        []AccountingEarningAllocation `json:"allocations"`
}

type AccountingEarningAllocation struct {
	EarningMovementUUID string `json:"earningMovementUuid"`
	AmountMinor         string `json:"amountMinor"`
}

type AccountingSetOff struct {
	SetOffUUID         string                        `json:"setOffUuid"`
	PartnerCompanyUUID string                        `json:"partnerCompanyUuid"`
	Amount             StorefrontMoney               `json:"amount"`
	OccurredAt         time.Time                     `json:"occurredAt"`
	CreditNoteUUID     string                        `json:"creditNoteUuid,omitempty"`
	PreferenceUUID     string                        `json:"preferenceUuid"`
	TaxDisclosureUUID  string                        `json:"taxDisclosureUuid"`
	Allocations        []AccountingEarningAllocation `json:"allocations"`
}

type AccountingProjectionGap struct {
	Outcome     string `json:"outcome"`
	FailureCode string `json:"failureCode,omitempty"`
}

// ReadAccountingExport reads one page across the authenticated seller’s storefronts.
// Use the returned NextAfter with the same Month and BusinessTimeZone to continue.
func (s *StorefrontClient) ReadAccountingExport(ctx context.Context, input AccountingExportRequest) (*AccountingExport, error) {
	query := url.Values{
		"month": {input.Month}, "businessTimeZone": {input.BusinessTimeZone},
	}
	if input.After != "" {
		query.Set("after", input.After)
	}
	path := "/api/v1/billing/accounting-export?" + query.Encode()
	return billingRequest[AccountingExport](ctx, s, http.MethodGet, path, nil)
}
