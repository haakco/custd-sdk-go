package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

const partnerReferralReviewPath = billingPartnersPath + "/referral-reviews"

// PartnerReferralPageRequest addresses a page before the last returned NextBefore value.
type PartnerReferralPageRequest struct {
	Before string
}

type PublishReferralTerms struct {
	Code                    string `json:"code"`
	EffectiveFrom           string `json:"effectiveFrom"`
	AgreementDate           string `json:"agreementDate"`
	AgreementReference      string `json:"agreementReference"`
	SignedAgreementAttested bool   `json:"signedAgreementAttested"`
	Share                   string `json:"share"`
	Months                  int    `json:"months"`
	HoldDays                int    `json:"holdDays"`
}

type ReferralTermsVersion struct {
	TermsUUID          string `json:"termsUuid"`
	EffectiveFrom      string `json:"effectiveFrom"`
	AgreementDate      string `json:"agreementDate"`
	AgreementReference string `json:"agreementReference"`
	RecordedByUserUUID string `json:"recordedByUserUuid"`
	Share              string `json:"share"`
	Months             int    `json:"months"`
	HoldDays           int    `json:"holdDays"`
}

type ReferralCodeTerms struct {
	Code             string                 `json:"code,omitempty"`
	BusinessTimeZone string                 `json:"businessTimeZone,omitempty"`
	Versions         []ReferralTermsVersion `json:"versions"`
	NextBefore       string                 `json:"nextBefore,omitempty"`
}

type ReferralReviewEventType string

const (
	ReferralReviewPaymentSettled  ReferralReviewEventType = "billing.payment.settled"
	ReferralReviewPaymentRefunded ReferralReviewEventType = "billing.payment.refunded"
)

type ReferralReviewItem struct {
	AttributionUUID    string                  `json:"attributionUuid"`
	EventUUID          string                  `json:"eventUuid"`
	BuyerCompanyUUID   string                  `json:"buyerCompanyUuid"`
	PartnerCompanyUUID string                  `json:"partnerCompanyUuid"`
	InvoiceUUID        string                  `json:"invoiceUuid"`
	EventType          ReferralReviewEventType `json:"eventType"`
	FailureCode        string                  `json:"failureCode"`
	ReviewedAt         time.Time               `json:"reviewedAt"`
	FirstReceivedOn    string                  `json:"firstReceivedOn,omitempty"`
	RetryEligible      bool                    `json:"retryEligible"`
}

type ReferralReviewPage struct {
	Reviews    []ReferralReviewItem `json:"reviews"`
	NextBefore string               `json:"nextBefore,omitempty"`
}

type ReferralRetryStatus string

const (
	ReferralRetryBlocked ReferralRetryStatus = "blocked"
	ReferralRequeued     ReferralRetryStatus = "requeued"
	ReferralNoReview     ReferralRetryStatus = "no_review_to_retry"
)

type ReferralReviewRetry struct {
	AttributionUUID string              `json:"attributionUuid"`
	Status          ReferralRetryStatus `json:"status"`
	FailureCode     string              `json:"failureCode,omitempty"`
	RequeuedCount   int                 `json:"requeuedCount"`
	HasMore         bool                `json:"hasMore"`
}

// ListPartnerReferralTerms lists the seller’s referral terms for one partner.
func (s *StorefrontClient) ListPartnerReferralTerms(ctx context.Context, partnerUUID string, input PartnerReferralPageRequest) (*ReferralCodeTerms, error) {
	path, err := buildStorefrontPath(billingPartnersPath, partnerUUID, "referral-terms")
	if err != nil {
		return nil, err
	}
	if input.Before != "" {
		path += "?" + (url.Values{"before": {input.Before}}).Encode()
	}
	return billingRequest[ReferralCodeTerms](ctx, s, http.MethodGet, path, nil)
}

// PublishPartnerReferralTerms records one effective-dated version of a partner’s referral terms.
func (s *StorefrontClient) PublishPartnerReferralTerms(ctx context.Context, partnerUUID string, input PublishReferralTerms) (*ReferralCodeTerms, error) {
	path, err := buildStorefrontPath(billingPartnersPath, partnerUUID, "referral-terms")
	if err != nil {
		return nil, err
	}
	return billingRequest[ReferralCodeTerms](ctx, s, http.MethodPost, path, input)
}

// ListReferralReviews lists the seller’s referral fee events requiring review.
func (s *StorefrontClient) ListReferralReviews(ctx context.Context, input PartnerReferralPageRequest) (*ReferralReviewPage, error) {
	path := partnerReferralReviewPath
	if input.Before != "" {
		path += "?" + (url.Values{"before": {input.Before}}).Encode()
	}
	return billingRequest[ReferralReviewPage](ctx, s, http.MethodGet, path, nil)
}

// RetryReferralReview requeues eligible events using Billing’s recorded payment evidence.
func (s *StorefrontClient) RetryReferralReview(ctx context.Context, attributionUUID string) (*ReferralReviewRetry, error) {
	path, err := buildStorefrontPath(partnerReferralReviewPath, attributionUUID, "retry")
	if err != nil {
		return nil, err
	}
	return billingRequest[ReferralReviewRetry](ctx, s, http.MethodPost, path, nil)
}
