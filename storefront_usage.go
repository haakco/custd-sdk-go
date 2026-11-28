package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type UsageMeter struct {
	Slug    string `json:"slug"`
	Version int    `json:"version"`
	Unit    string `json:"unit"`
}

type UsageRate struct {
	Currency    string `json:"currency"`
	MajorAmount string `json:"majorAmount"`
	PerUnits    string `json:"perUnits"`
}

type UsageTax struct {
	Label string `json:"label"`
	Rate  string `json:"rate"`
}

type PublishUsageTariffRequest struct {
	UUID          string     `json:"tariffRevisionUuid"`
	ProductUUID   string     `json:"productUuid"`
	Description   string     `json:"description"`
	Meter         UsageMeter `json:"meter"`
	EffectiveFrom time.Time  `json:"effectiveFrom"`
	EffectiveTo   *time.Time `json:"effectiveTo,omitempty"`
	Rate          UsageRate  `json:"rate"`
	Tax           UsageTax   `json:"tax"`
}

type UsageTariffRevision struct {
	PublishUsageTariffRequest
	SellerCompanyUUID string `json:"sellerCompanyUuid"`
	AppUUID           string `json:"appUuid"`
}

type EnrollUsageRequest struct {
	TariffRevisionUUID   string    `json:"tariffRevisionUuid"`
	DataSpaceCompanyUUID string    `json:"dataSpaceCompanyUuid,omitempty"`
	PeriodStart          time.Time `json:"periodStart"`
	CloseNotBefore       time.Time `json:"closeNotBefore"`
}

type UsageEnrollment struct {
	UUID string `json:"enrollmentUuid"`
}

type StorefrontUsagePageRequest struct {
	Storefront string
	Before     string
}

type UsageTariffPage struct {
	Tariffs    []UsageTariffRevision `json:"tariffs"`
	NextBefore string                `json:"nextBefore,omitempty"`
}

type UsageEnrollmentItem struct {
	UUID                 string     `json:"enrollmentUuid"`
	TariffRevisionUUID   string     `json:"tariffRevisionUuid"`
	DataSpaceCompanyUUID string     `json:"dataSpaceCompanyUuid,omitempty"`
	PeriodStart          time.Time  `json:"periodStart"`
	CloseNotBefore       time.Time  `json:"closeNotBefore"`
	ClosedAt             *time.Time `json:"closedAt,omitempty"`
	Status               string     `json:"status"`
	InvoiceUUID          string     `json:"invoiceUuid,omitempty"`
}

type UsageEnrollmentPage struct {
	Enrollments []UsageEnrollmentItem `json:"enrollments"`
	NextBefore  string                `json:"nextBefore,omitempty"`
}

type UsageInvoiceScope struct {
	SellerCompanyUUID string `json:"sellerCompanyUuid"`
	AppUUID           string `json:"appUuid"`
	BuyerCompanyUUID  string `json:"buyerCompanyUuid"`
}

type UsagePeriod struct {
	Start        time.Time `json:"start"`
	EndExclusive time.Time `json:"endExclusive"`
}

type UsageSource struct {
	AccountCompanyUUID   string     `json:"accountCompanyUuid"`
	DataSpaceCompanyUUID string     `json:"dataSpaceCompanyUuid,omitempty"`
	Meter                UsageMeter `json:"meter"`
}

type UsageWindowEvidence struct {
	Period               UsagePeriod `json:"period"`
	Quantity             string      `json:"quantity"`
	State                string      `json:"state"`
	SourceWatermark      string      `json:"sourceWatermark"`
	CorrectionGeneration int         `json:"correctionGeneration"`
	CalculationVersion   int         `json:"calculationVersion"`
}

type UsagePricedLine struct {
	Scope       UsageInvoiceScope     `json:"scope"`
	Period      UsagePeriod           `json:"period"`
	Source      UsageSource           `json:"source"`
	Tariff      UsageTariffRevision   `json:"tariff"`
	Quantity    string                `json:"quantity"`
	Rounding    string                `json:"rounding"`
	Windows     []UsageWindowEvidence `json:"windows"`
	NetAmount   StorefrontMoney       `json:"netAmount"`
	TaxAmount   StorefrontMoney       `json:"taxAmount"`
	GrossAmount StorefrontMoney       `json:"grossAmount"`
}

type UsageCertificate struct {
	EnrollmentUUID string          `json:"enrollmentUuid"`
	ClosedAt       time.Time       `json:"closedAt"`
	Line           UsagePricedLine `json:"line"`
}

type UsageIssuanceStatus string

const (
	UsageInvoiceIssued UsageIssuanceStatus = "invoice_issued"
	UsageNoAmountDue   UsageIssuanceStatus = "no_amount_due"
)

type UsageIssuedInvoice struct {
	UUID     string                  `json:"invoiceUuid"`
	Number   string                  `json:"invoiceNumber"`
	Status   StorefrontInvoiceStatus `json:"status"`
	IssuedAt time.Time               `json:"issuedAt"`
	Totals   StorefrontTotals        `json:"totals"`
	Lines    []StorefrontQuoteLine   `json:"lines"`
}

type IssueUsageResult struct {
	Status  UsageIssuanceStatus `json:"status"`
	Invoice *UsageIssuedInvoice `json:"invoice,omitempty"`
	Usage   UsageCertificate    `json:"usage"`
}

func (s *StorefrontClient) GetUsageTariff(ctx context.Context, storefront, revisionUUID string) (*UsageTariffRevision, error) {
	return storefrontRequest[UsageTariffRevision](ctx, s, http.MethodGet, nil, storefront, "usage-tariffs", revisionUUID)
}

func (s *StorefrontClient) ListUsageTariffs(ctx context.Context, input StorefrontUsagePageRequest) (*UsageTariffPage, error) {
	return listUsage[UsageTariffPage](ctx, s, input, "usage-tariffs")
}

func (s *StorefrontClient) ListUsageEnrollments(ctx context.Context, input StorefrontUsagePageRequest) (*UsageEnrollmentPage, error) {
	return listUsage[UsageEnrollmentPage](ctx, s, input, "usage-enrollments")
}

func listUsage[T any](ctx context.Context, s *StorefrontClient, input StorefrontUsagePageRequest, resource string) (*T, error) {
	path, err := storefrontPath(input.Storefront, resource)
	if err != nil {
		return nil, err
	}
	if input.Before != "" {
		path += "?" + (url.Values{"before": {input.Before}}).Encode()
	}
	return billingRequest[T](ctx, s, http.MethodGet, path, nil)
}

func (s *StorefrontClient) PublishUsageTariff(ctx context.Context, storefront string, input PublishUsageTariffRequest) (*UsageTariffRevision, error) {
	return storefrontRequest[UsageTariffRevision](ctx, s, http.MethodPost, input, storefront, "usage-tariffs")
}

func (s *StorefrontClient) EnrollUsage(ctx context.Context, storefront string, input EnrollUsageRequest) (*UsageEnrollment, error) {
	return storefrontRequest[UsageEnrollment](ctx, s, http.MethodPost, input, storefront, "usage-enrollments")
}

func (s *StorefrontClient) IssueUsageInvoice(ctx context.Context, storefront, enrollmentUUID string) (*IssueUsageResult, error) {
	return storefrontRequest[IssueUsageResult](ctx, s, http.MethodPost, nil, storefront, "usage-enrollments", enrollmentUUID, "issue")
}
