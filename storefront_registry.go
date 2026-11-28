package custd

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const storefrontRegistryPath = "/api/v1/billing/storefronts"

type StorefrontBinding struct {
	Slug              string    `json:"slug"`
	SellerCompanyUUID string    `json:"sellerCompanyUuid"`
	AppUUID           string    `json:"appUuid"`
	CreatedAt         time.Time `json:"createdAt"`
}

type OwnedStorefronts struct {
	SellerCompanyUUID string              `json:"sellerCompanyUuid"`
	Storefronts       []StorefrontBinding `json:"storefronts"`
}

type RegisterStorefront struct {
	Slug    string `json:"slug"`
	AppUUID string `json:"appUuid"`
}

type SelectStorefrontReceiptTemplate struct {
	Revision string `json:"revision"`
}

type StorefrontReceiptTemplateSelection struct {
	Revision   string     `json:"revision,omitempty"`
	SelectedAt *time.Time `json:"selectedAt,omitempty"`
}

func (s *StorefrontClient) ListOwnedStorefronts(ctx context.Context) (*OwnedStorefronts, error) {
	return billingRequest[OwnedStorefronts](ctx, s, http.MethodGet, storefrontRegistryPath, nil)
}

func (s *StorefrontClient) RegisterStorefront(ctx context.Context, input RegisterStorefront) (*StorefrontBinding, error) {
	return billingRequest[StorefrontBinding](ctx, s, http.MethodPost, storefrontRegistryPath, input)
}

func (s *StorefrontClient) GetReceiptTemplate(ctx context.Context, slug string) (*StorefrontReceiptTemplateSelection, error) {
	return s.receiptTemplate(ctx, http.MethodGet, slug, nil)
}

func (s *StorefrontClient) SelectReceiptTemplate(ctx context.Context, slug string, input SelectStorefrontReceiptTemplate) (*StorefrontReceiptTemplateSelection, error) {
	return s.receiptTemplate(ctx, http.MethodPut, slug, &input)
}

func (s *StorefrontClient) receiptTemplate(ctx context.Context, method, slug string, input *SelectStorefrontReceiptTemplate) (*StorefrontReceiptTemplateSelection, error) {
	if slug == "" || slug == "." || slug == ".." || strings.ContainsAny(slug, "/\\") {
		return nil, errors.New("custd: invalid storefront path segment")
	}
	path := storefrontRegistryPath + "/" + url.PathEscape(slug) + "/receipt-template"
	if input == nil {
		return billingRequest[StorefrontReceiptTemplateSelection](ctx, s, method, path, nil)
	}
	return billingRequest[StorefrontReceiptTemplateSelection](ctx, s, method, path, input)
}
