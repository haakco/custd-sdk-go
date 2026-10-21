package custd

import (
	"context"
	"net/http"
	"net/url"
)

// NotificationWebhookAdminClient owns tenant notification webhook
// subscriptions and their deliveries.
//
// Signing secrets are returned exactly once, on create and on rotate. The SDK
// hands them back to the caller and never stores, logs or echoes them.
type NotificationWebhookAdminClient struct {
	admin *AdminClient
}

// NotificationWebhookSubscription is one tenant's webhook endpoint.
type NotificationWebhookSubscription struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	Enabled   bool   `json:"enabled"`
	Company   string `json:"companySlug,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// NotificationWebhookListResponse lists a tenant's subscriptions.
type NotificationWebhookListResponse struct {
	Subscriptions []NotificationWebhookSubscription `json:"subscriptions"`
}

// NotificationWebhookSecret carries a signing secret. The value is sensitive:
// hand it to the caller's secret store and do not log it.
type NotificationWebhookSecret struct {
	SigningSecret string `json:"signingSecret"`
}

// NotificationWebhookCreateRequest registers a webhook endpoint. The tenant is
// server-derived from the authenticated session.
type NotificationWebhookCreateRequest struct {
	CompanySlug string `json:"companySlug,omitempty"`
	Name        string `json:"name"`
	Endpoint    string `json:"endpoint"`
	Enabled     bool   `json:"enabled"`
}

// NotificationDeliveryStatus is one delivery attempt's state.
//
// AttemptCount, LastErrorClass, FailureReason and NextAction are part of the
// contract, not decoration: a caller polling a delivery needs the failure class
// and the server's next action to decide whether to wait, retry or report.
type NotificationDeliveryStatus struct {
	MessageUUID    string `json:"messageUuid"`
	DeliveryUUID   string `json:"deliveryUuid"`
	TemplateID     string `json:"templateId"`
	Channel        string `json:"channel"`
	State          string `json:"state"`
	AttemptCount   int    `json:"attemptCount"`
	LastErrorClass string `json:"lastErrorClass,omitempty"`
	FailureReason  string `json:"failureReason,omitempty"`
	NextAction     string `json:"nextAction,omitempty"`
}

const (
	notificationWebhooksPath    = "/notification-webhooks"
	notificationDeliveriesPath  = "/notification-deliveries"
	companySlugQueryParameter   = "companySlug"
	rotateSecretPathSegment     = "/rotate-secret"
	notificationReplayPathPart  = "/replay"
	notificationEnabledFieldKey = "enabled"
)

// Create registers a webhook and returns its signing secret once. The secret is
// sensitive; the caller owns storing it.
func (c *NotificationWebhookAdminClient) Create(
	ctx context.Context,
	req NotificationWebhookCreateRequest,
) (*NotificationWebhookSecret, ResponseMetadata, error) {
	var out NotificationWebhookSecret
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, notificationWebhooksPath, req, &out,
	)
	return &out, metadata, err
}

// List returns a tenant's subscriptions. Company is required because the
// route is tenant-scoped even for an operator.
func (c *NotificationWebhookAdminClient) List(
	ctx context.Context,
	company string,
) (*NotificationWebhookListResponse, ResponseMetadata, error) {
	var out NotificationWebhookListResponse
	path := tenantScopedPath(notificationWebhooksPath, "", company)
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, path, nil, &out)
	return &out, metadata, err
}

// SetEnabled turns a subscription on or off without replacing its secret.
func (c *NotificationWebhookAdminClient) SetEnabled(
	ctx context.Context,
	company string,
	uuid string,
	enabled bool,
) (*NotificationWebhookSubscription, ResponseMetadata, error) {
	var out NotificationWebhookSubscription
	path := tenantScopedPath(notificationWebhooksPath, uuid, company)
	body := map[string]bool{notificationEnabledFieldKey: enabled}
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPatch, path, body, &out)
	return &out, metadata, err
}

// RotateSecret issues a new signing secret and invalidates the previous one.
// The response is the only place the new secret exists.
func (c *NotificationWebhookAdminClient) RotateSecret(
	ctx context.Context,
	company string,
	uuid string,
) (*NotificationWebhookSecret, ResponseMetadata, error) {
	var out NotificationWebhookSecret
	path := tenantScopedPath(notificationWebhooksPath, uuid, company) + rotateSecretPathSegment
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, &out)
	return &out, metadata, err
}

// Delete removes a subscription. There is no body to carry the request
// identifier, so the metadata is the whole result.
func (c *NotificationWebhookAdminClient) Delete(
	ctx context.Context,
	company string,
	uuid string,
) (ResponseMetadata, error) {
	path := tenantScopedPath(notificationWebhooksPath, uuid, company)
	return c.admin.requestWithMetadata(ctx, http.MethodDelete, path, nil, nil)
}

// GetDelivery reads one delivery attempt's state.
func (c *NotificationWebhookAdminClient) GetDelivery(
	ctx context.Context,
	company string,
	uuid string,
) (*NotificationDeliveryStatus, ResponseMetadata, error) {
	var out NotificationDeliveryStatus
	path := tenantScopedPath(notificationDeliveriesPath, uuid, company)
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, path, nil, &out)
	return &out, metadata, err
}

// Replay re-sends one delivery. The reason is carried to the server so the
// replay is attributable; a bodyless answer returns the metadata alone.
func (c *NotificationWebhookAdminClient) Replay(
	ctx context.Context,
	company string,
	uuid string,
	reason string,
) (ResponseMetadata, error) {
	path := tenantScopedPath(notificationDeliveriesPath, uuid, company) + notificationReplayPathPart
	return c.admin.requestWithMetadata(
		ctx, http.MethodPost, path, map[string]string{"reason": reason}, nil,
	)
}

// tenantScopedPath builds "/base[/uuid]?companySlug=...". The company slug is a
// query parameter on these routes, not a path segment.
func tenantScopedPath(base, uuid, company string) string {
	path := base
	if uuid != "" {
		path += "/" + url.PathEscape(uuid)
	}
	return path + "?" + companySlugQueryParameter + "=" + url.QueryEscape(company)
}
