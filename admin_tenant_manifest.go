package custd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TenantManifestAdminClient owns a tenant's declarative desired state.
//
// Apply reconciles; Readiness reads. Neither returns secret material: the
// manifest is the tenant's own declaration and the readiness projection is
// secret-free by contract.
type TenantManifestAdminClient struct {
	admin *AdminClient
}

// TenantManifestResourceStatus is one declared resource's reconciliation state.
// SafeNextActionCode is the stable identifier; SafeNextAction is for a person.
type TenantManifestResourceStatus struct {
	Kind               string `json:"kind"`
	Key                string `json:"key"`
	State              string `json:"state"`
	Ready              bool   `json:"ready"`
	SafeNextAction     string `json:"safeNextAction"`
	SafeNextActionCode string `json:"safeNextActionCode"`
}

// TenantManifestOperationStatus is the durable operation Apply enqueued.
type TenantManifestOperationStatus struct {
	UUID           string     `json:"uuid"`
	IdempotencyKey string     `json:"idempotencyKey"`
	State          string     `json:"state"`
	ErrorCode      string     `json:"errorCode,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
}

// TenantManifestApplyResponse is the reconciliation result.
type TenantManifestApplyResponse struct {
	TenantSlug         string                         `json:"tenantSlug"`
	ManifestDigest     string                         `json:"manifestDigest"`
	Ready              bool                           `json:"ready"`
	State              string                         `json:"state"`
	ResourceStatuses   []TenantManifestResourceStatus `json:"resources"`
	SafeNextAction     string                         `json:"safeNextAction"`
	SafeNextActionCode string                         `json:"safeNextActionCode"`
	ObservedAt         time.Time                      `json:"observedAt"`
	Operation          *TenantManifestOperationStatus `json:"operation,omitempty"`
}

// TenantManifestReadinessResponse is the secret-free desired-state projection.
type TenantManifestReadinessResponse struct {
	TenantSlug         string                         `json:"tenantSlug"`
	ManifestDigest     string                         `json:"manifestDigest"`
	Ready              bool                           `json:"ready"`
	State              string                         `json:"state"`
	ResourceStatuses   []TenantManifestResourceStatus `json:"resources"`
	SafeNextAction     string                         `json:"safeNextAction"`
	SafeNextActionCode string                         `json:"safeNextActionCode"`
	ObservedAt         time.Time                      `json:"observedAt"`
	Operation          *TenantManifestOperationStatus `json:"operation,omitempty"`
}

const tenantManifestPath = "/tenant-manifest/"

// Apply reconciles one tenant's declarative desired state.
//
// IdempotencyKey is required: the server refuses a reconciling request that
// arrives without one, so an empty value is a caller error rather than "no
// opinion".
func (c *TenantManifestAdminClient) Apply(
	ctx context.Context,
	tenantSlug string,
	idempotencyKey string,
	manifest json.RawMessage,
) (*TenantManifestApplyResponse, ResponseMetadata, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ResponseMetadata{}, errors.New("custd: tenant manifest idempotency key is required")
	}
	var out TenantManifestApplyResponse
	headers := map[string]string{"Idempotency-Key": idempotencyKey}
	metadata, err := c.admin.requestWithMetadataHeaders(
		ctx, http.MethodPut, tenantManifestPath+url.PathEscape(tenantSlug), manifest, &out, headers,
	)
	return &out, metadata, err
}

// Readiness reads one tenant's secret-free desired-state projection.
func (c *TenantManifestAdminClient) Readiness(
	ctx context.Context,
	tenantSlug string,
) (*TenantManifestReadinessResponse, ResponseMetadata, error) {
	var out TenantManifestReadinessResponse
	path := tenantManifestPath + url.PathEscape(tenantSlug) + "/readiness"
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, path, nil, &out)
	return &out, metadata, err
}
