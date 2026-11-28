package custd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// StorefrontClient browses a seller's offers and checks out using the current buyer identity.
type StorefrontClient struct {
	client *CustdClient
}

// StorefrontMoney is an exact signed minor-unit count encoded as a decimal string.
type StorefrontMoney struct {
	Currency    string `json:"currency"`
	AmountMinor string `json:"amountMinor"`
}

func (money *StorefrontMoney) UnmarshalJSON(data []byte) error {
	type rawStorefrontMoney StorefrontMoney
	var decoded rawStorefrontMoney
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode storefront money: %w", err)
	}
	if len(decoded.Currency) != 3 || strings.IndexFunc(decoded.Currency, func(r rune) bool { return r < 'A' || r > 'Z' }) >= 0 {
		return errors.New("custd: invalid storefront money currency")
	}
	amount, err := strconv.ParseInt(decoded.AmountMinor, 10, 64)
	if err != nil || strconv.FormatInt(amount, 10) != decoded.AmountMinor {
		return errors.New("custd: invalid storefront minor-unit count")
	}
	*money = StorefrontMoney(decoded)
	return nil
}

type StorefrontUnitPrice struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type StorefrontPriceRevision struct {
	UUID          string              `json:"priceRevisionUuid"`
	Revision      int64               `json:"revision"`
	UnitPrice     StorefrontUnitPrice `json:"unitPrice"`
	TaxTreatment  string              `json:"taxTreatment"`
	TaxRate       string              `json:"taxRate"`
	TaxLabel      string              `json:"taxLabel"`
	EffectiveFrom time.Time           `json:"effectiveFrom"`
}

type StorefrontOffer struct {
	UUID            string                  `json:"offerUuid"`
	Name            string                  `json:"name"`
	BillingInterval string                  `json:"billingInterval"`
	PriceRevision   StorefrontPriceRevision `json:"priceRevision"`
}

type StorefrontCatalogue struct {
	CanPublish bool              `json:"canPublish"`
	Offers     []StorefrontOffer `json:"offers"`
}

type StorefrontPublishOffer struct {
	OfferUUID       string              `json:"offerUuid,omitempty"`
	Name            string              `json:"name"`
	BillingInterval string              `json:"billingInterval"`
	UnitPrice       StorefrontUnitPrice `json:"unitPrice"`
	TaxRate         string              `json:"taxRate"`
	TaxLabel        string              `json:"taxLabel"`
	EffectiveFrom   *time.Time          `json:"effectiveFrom,omitempty"`
}

type StorefrontCartLine struct {
	OfferUUID         string `json:"offerUuid"`
	PriceRevisionUUID string `json:"priceRevisionUuid"`
	Quantity          string `json:"quantity"`
}

type StorefrontCart struct {
	UUID      string               `json:"cartUuid"`
	Revision  int64                `json:"revision"`
	Status    string               `json:"status"`
	Currency  string               `json:"currency"`
	CreatedAt time.Time            `json:"createdAt"`
	Lines     []StorefrontCartLine `json:"lines"`
}

// StorefrontLineQuantity sets an exact decimal quantity. Use "0" to remove the line.
type StorefrontLineQuantity struct {
	Quantity string `json:"quantity"`
}

type StorefrontTotals struct {
	NetTotal   StorefrontMoney `json:"netTotal"`
	TaxTotal   StorefrontMoney `json:"taxTotal"`
	GrossTotal StorefrontMoney `json:"grossTotal"`
}

type StorefrontTax struct {
	Label         string          `json:"label"`
	Rate          string          `json:"rate"`
	Rounding      string          `json:"rounding"`
	RoundingPoint string          `json:"roundingPoint"`
	Amount        StorefrontMoney `json:"amount"`
}

type StorefrontServicePeriod struct {
	Start        string `json:"start"`
	EndExclusive string `json:"endExclusive"`
}

type StorefrontQuoteLine struct {
	LineNumber        int                     `json:"lineNumber"`
	OfferUUID         string                  `json:"offerUuid"`
	PriceRevisionUUID string                  `json:"priceRevisionUuid"`
	Description       string                  `json:"description"`
	Quantity          string                  `json:"quantity"`
	UnitPrice         StorefrontUnitPrice     `json:"unitPrice"`
	ExtensionRounding string                  `json:"extensionRounding"`
	ServicePeriod     StorefrontServicePeriod `json:"servicePeriod"`
	NetAmount         StorefrontMoney         `json:"netAmount"`
	Tax               StorefrontTax           `json:"tax"`
	GrossAmount       StorefrontMoney         `json:"grossAmount"`
}

type StorefrontQuote struct {
	UUID         string                `json:"quoteUuid"`
	CartUUID     string                `json:"cartUuid"`
	CartRevision int64                 `json:"cartRevision"`
	IssuedAt     time.Time             `json:"issuedAt"`
	ExpiresAt    time.Time             `json:"expiresAt"`
	Lines        []StorefrontQuoteLine `json:"lines"`
	Totals       StorefrontTotals      `json:"totals"`
}

type StorefrontInvoiceStatus string
type StorefrontPaymentStatus string
type StorefrontCheckoutState string

const (
	StorefrontInvoiceOpen            StorefrontInvoiceStatus = "open"
	StorefrontInvoicePaid            StorefrontInvoiceStatus = "paid"
	StorefrontPaymentPending         StorefrontPaymentStatus = "pending"
	StorefrontPaymentSettled         StorefrontPaymentStatus = "settled"
	StorefrontPaymentFailed          StorefrontPaymentStatus = "failed"
	StorefrontCheckoutCreating       StorefrontCheckoutState = "creating"
	StorefrontCheckoutReady          StorefrontCheckoutState = "ready"
	StorefrontCheckoutReviewRequired StorefrontCheckoutState = "review_required"
)

type StorefrontReceiptDelivery struct {
	Channel               string `json:"channel"`
	NotificationSourceKey string `json:"notificationSourceKey"`
	Status                string `json:"status"`
}

type StorefrontReceipt struct {
	UUID           string                    `json:"receiptUuid"`
	Number         string                    `json:"receiptNumber"`
	InvoiceUUID    string                    `json:"invoiceUuid"`
	PaymentUUID    string                    `json:"paymentUuid"`
	IssuedAt       time.Time                 `json:"issuedAt"`
	AmountReceived StorefrontMoney           `json:"amountReceived"`
	Delivery       StorefrontReceiptDelivery `json:"delivery"`
}

type StorefrontCreditNote struct {
	UUID            string                    `json:"creditNoteUuid"`
	Number          string                    `json:"creditNoteNumber"`
	InvoiceUUID     string                    `json:"invoiceUuid,omitempty"`
	InvoiceNumber   string                    `json:"invoiceNumber,omitempty"`
	InvoiceIssuedAt *time.Time                `json:"invoiceIssuedAt,omitempty"`
	IssuedAt        *time.Time                `json:"issuedAt,omitempty"`
	Status          string                    `json:"status,omitempty"`
	BillTo          *StorefrontBillingContact `json:"billTo,omitempty"`
	SellerName      string                    `json:"sellerName,omitempty"`
	Reason          string                    `json:"reason,omitempty"`
	Lines           []StorefrontQuoteLine     `json:"lines,omitempty"`
	Totals          StorefrontTotals          `json:"totals"`
}

type StorefrontBillingContact struct {
	CompanyUUID string `json:"companyUuid"`
	Name        string `json:"name"`
	Email       string `json:"email"`
}

type StorefrontRefund struct {
	UUID        string          `json:"refundUuid"`
	Status      string          `json:"status"`
	Amount      StorefrontMoney `json:"amount"`
	InvoiceUUID string          `json:"invoiceUuid,omitempty"`
	PaymentUUID string          `json:"paymentUuid,omitempty"`
	RequestedAt *time.Time      `json:"requestedAt,omitempty"`
	ProcessedAt *time.Time      `json:"processedAt,omitempty"`
}

type StorefrontAdjustmentRequest struct {
	Reason string `json:"reason"`
}

type StorefrontSetOff struct {
	UUID              string                    `json:"setOffUuid"`
	InvoiceUUID       string                    `json:"invoiceUuid"`
	Amount            StorefrontMoney           `json:"amount"`
	AppliedAt         time.Time                 `json:"appliedAt"`
	AppliedOn         string                    `json:"appliedOn"`
	PreferenceUUID    string                    `json:"preferenceUuid"`
	TaxDisclosureUUID string                    `json:"taxDisclosureUuid"`
	Reversal          *StorefrontSetOffReversal `json:"reversal,omitempty"`
}

type StorefrontSetOffReversal struct {
	CreditNoteUUID string    `json:"creditNoteUuid"`
	ReversedAt     time.Time `json:"reversedAt"`
}

type StorefrontSetOffEligibility struct {
	Eligible bool `json:"eligible"`
}

type StorefrontPurchase struct {
	InvoiceNotice    *InvoiceNotice          `json:"invoiceNotice,omitempty"`
	SetOff           *StorefrontSetOff       `json:"setOff,omitempty"`
	Usage            *UsageCertificate       `json:"usage,omitempty"`
	ActivationStatus string                  `json:"activationStatus,omitempty"`
	Refund           *StorefrontRefund       `json:"refund,omitempty"`
	InvoiceUUID      string                  `json:"invoiceUuid"`
	InvoiceNumber    string                  `json:"invoiceNumber"`
	Status           StorefrontInvoiceStatus `json:"status"`
	Totals           StorefrontTotals        `json:"totals"`
	CheckoutState    StorefrontCheckoutState `json:"checkoutState,omitempty"`
	PaymentStatus    StorefrontPaymentStatus `json:"paymentStatus,omitempty"`
	PaymentURL       string                  `json:"paymentUrl,omitempty"`
	Receipt          *StorefrontReceipt      `json:"receipt,omitempty"`
	CreditNote       *StorefrontCreditNote   `json:"creditNote,omitempty"`
}

func (s *StorefrontClient) ListOffers(ctx context.Context, storefront string) (*StorefrontCatalogue, error) {
	return storefrontRequest[StorefrontCatalogue](ctx, s, http.MethodGet, nil, storefront, "offers")
}

func (s *StorefrontClient) PublishOffer(ctx context.Context, storefront string, offer StorefrontPublishOffer) (*StorefrontOffer, error) {
	return storefrontRequest[StorefrontOffer](ctx, s, http.MethodPost, offer, storefront, "offers")
}

func (s *StorefrontClient) CreateCart(ctx context.Context, storefront string) (*StorefrontCart, error) {
	return storefrontRequest[StorefrontCart](ctx, s, http.MethodPost, nil, storefront, "carts")
}

func (s *StorefrontClient) GetCart(ctx context.Context, storefront, cartUUID string) (*StorefrontCart, error) {
	return storefrontRequest[StorefrontCart](ctx, s, http.MethodGet, nil, storefront, "carts", cartUUID)
}

func (s *StorefrontClient) UpdateLine(ctx context.Context, storefront, cartUUID, offerUUID string, quantity StorefrontLineQuantity) (*StorefrontCart, error) {
	return storefrontRequest[StorefrontCart](ctx, s, http.MethodPut, quantity, storefront, "carts", cartUUID, "lines", offerUUID)
}

func (s *StorefrontClient) CreateQuote(ctx context.Context, storefront, cartUUID string) (*StorefrontQuote, error) {
	return storefrontRequest[StorefrontQuote](ctx, s, http.MethodPost, nil, storefront, "carts", cartUUID, "quotes")
}

func (s *StorefrontClient) GetQuote(ctx context.Context, storefront, quoteUUID string) (*StorefrontQuote, error) {
	return storefrontRequest[StorefrontQuote](ctx, s, http.MethodGet, nil, storefront, "quotes", quoteUUID)
}

// Checkout starts or resumes the purchase for an accepted quote.
func (s *StorefrontClient) Checkout(ctx context.Context, storefront, quoteUUID string) (*StorefrontPurchase, error) {
	return storefrontRequest[StorefrontPurchase](ctx, s, http.MethodPost, nil, storefront, "quotes", quoteUUID, "checkout")
}

func (s *StorefrontClient) IssueStorefrontInvoice(ctx context.Context, storefront, quoteUUID string) (*StorefrontPurchase, error) {
	return storefrontRequest[StorefrontPurchase](ctx, s, http.MethodPost, nil, storefront, "quotes", quoteUUID, "invoice")
}

func (s *StorefrontClient) GetStorefrontSetOffEligibility(ctx context.Context, storefront, invoiceUUID string) (*StorefrontSetOffEligibility, error) {
	return storefrontRequest[StorefrontSetOffEligibility](ctx, s, http.MethodGet, nil, storefront, "purchases", invoiceUUID, "set-off")
}

func (s *StorefrontClient) ApplyStorefrontSetOff(ctx context.Context, storefront, invoiceUUID string) (*StorefrontSetOff, error) {
	return storefrontRequest[StorefrontSetOff](ctx, s, http.MethodPost, nil, storefront, "purchases", invoiceUUID, "set-off")
}

func (s *StorefrontClient) GetPurchase(ctx context.Context, storefront, invoiceUUID string) (*StorefrontPurchase, error) {
	return storefrontRequest[StorefrontPurchase](ctx, s, http.MethodGet, nil, storefront, "purchases", invoiceUUID)
}

// StartPayment resumes or retries payment for the same invoice.
func (s *StorefrontClient) StartPayment(ctx context.Context, storefront, invoiceUUID string) (*StorefrontPurchase, error) {
	return storefrontRequest[StorefrontPurchase](ctx, s, http.MethodPost, nil, storefront, "purchases", invoiceUUID, "payment")
}

func (s *StorefrontClient) IssueCreditNote(ctx context.Context, storefront, invoiceUUID string, request StorefrontAdjustmentRequest) (*StorefrontCreditNote, error) {
	return storefrontRequest[StorefrontCreditNote](ctx, s, http.MethodPost, request, storefront, "invoices", invoiceUUID, "credit-note")
}

func (s *StorefrontClient) RequestRefund(ctx context.Context, storefront, invoiceUUID string, request StorefrontAdjustmentRequest) (*StorefrontRefund, error) {
	return storefrontRequest[StorefrontRefund](ctx, s, http.MethodPost, request, storefront, "invoices", invoiceUUID, "refund")
}

func (s *StorefrontClient) ReconcileRefund(ctx context.Context, storefront, invoiceUUID string) (*StorefrontRefund, error) {
	return storefrontRequest[StorefrontRefund](ctx, s, http.MethodPost, nil, storefront, "invoices", invoiceUUID, "refund", "reconcile")
}

func storefrontRequest[T any](ctx context.Context, s *StorefrontClient, method string, payload any, parts ...string) (*T, error) {
	path, err := storefrontPath(parts...)
	if err != nil {
		return nil, err
	}
	return billingRequest[T](ctx, s, method, path, payload)
}

func storefrontPath(parts ...string) (string, error) {
	return buildStorefrontPath("/api/v1/storefronts", parts...)
}

func buildStorefrontPath(base string, parts ...string) (string, error) {
	var path strings.Builder
	path.WriteString(base)
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, "/\\") {
			return "", fmt.Errorf("custd: invalid storefront path segment")
		}
		path.WriteByte('/')
		path.WriteString(url.PathEscape(part))
	}
	return path.String(), nil
}

func billingRequest[T any](ctx context.Context, s *StorefrontClient, method, path string, payload any) (*T, error) {
	body, err := encodeRequestBody(payload, "billing")
	if err != nil {
		return nil, err
	}
	response, err := s.client.exchange(ctx, &bufferedRequest{
		Method: method, URL: strings.TrimRight(s.client.config.BaseURL, "/") + path,
		Headers: s.client.headers(false), Body: body,
	})
	if err != nil {
		return nil, &RequestError{Message: "custd: billing request failed: " + err.Error(), Retryable: true, Cause: err}
	}
	recordMetadata(ctx, response.Header.Get(requestIDHeader))
	if err := s.client.checkStatus(response.StatusCode, response.Body); err != nil {
		return nil, err
	}
	var out T
	if err := unmarshalResponse(response.Body, &out); err != nil {
		return nil, fmt.Errorf("custd: decode billing response: %w", err)
	}
	return &out, nil
}
