package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const apiEndpoint = "/api/v1"

type ProvisioningClient struct {
	DataSpaces   *DataSpaceProvisioningClient
	Producers    *ProducerProvisioningClient
	Reservations *ProducerReservationsClient
	client       *CustdClient
}

type DataSpaceProvisioningClient struct {
	provisioning *ProvisioningClient
}

type ProducerProvisioningClient struct {
	provisioning *ProvisioningClient
}

// ProducerReservationsClient owns the producer-reservation lifecycle on a data
// space. Reservations are scoped to a data space; callers carry the resolved
// tenant via the bearer token.
type ProducerReservationsClient struct {
	provisioning *ProvisioningClient
}

// ProducerReservationStatus is where a producer reservation stands.
type ProducerReservationStatus string

const (
	ProducerReservationStatusPending  ProducerReservationStatus = "pending"
	ProducerReservationStatusClaimed  ProducerReservationStatus = "claimed"
	ProducerReservationStatusReleased ProducerReservationStatus = "released"
	ProducerReservationStatusExpired  ProducerReservationStatus = "expired"
)

type ProducerReservation struct {
	ProducerSlug      string                    `json:"producerSlug"`
	ParentCompanySlug string                    `json:"parentCompanySlug"`
	ChildCompanySlug  string                    `json:"childCompanySlug"`
	ClaimedByClientID string                    `json:"claimedByClientId,omitempty"`
	Status            ProducerReservationStatus `json:"status"`
	ReservedAt        string                    `json:"reservedAt,omitempty"`
	ExpiresAt         string                    `json:"expiresAt,omitempty"`
	MaxTTLSeconds     int64                     `json:"maxTtlSeconds"`
}

type ProducerReservationListResponse struct {
	Reservations []ProducerReservation `json:"reservations"`
}

type ProducerReservationCreateRequest struct {
	ProducerSlug string `json:"producerSlug"`
	TTLSeconds   int64  `json:"ttlSeconds,omitempty"`
}

type ProducerReservationClaimRequest struct {
	ClaimedByClientID string `json:"claimedByClientId"`
}

func newProvisioningClient(client *CustdClient) *ProvisioningClient {
	provisioning := &ProvisioningClient{client: client}
	provisioning.DataSpaces = &DataSpaceProvisioningClient{provisioning: provisioning}
	provisioning.Producers = &ProducerProvisioningClient{provisioning: provisioning}
	provisioning.Reservations = &ProducerReservationsClient{provisioning: provisioning}
	return provisioning
}

// Create adds a data space under a parent. An empty parentCompanySlug uses the
// credential's own company; a platform administrator may name another parent.
func (c *DataSpaceProvisioningClient) Create(
	ctx context.Context,
	req DataSpaceCreate,
	parentCompanySlug string,
) (*DataSpace, error) {
	var space DataSpace
	err := c.provisioning.request(ctx, http.MethodPost, dataSpacePath("/data-spaces", parentCompanySlug), req, &space)
	return &space, err
}

// List returns a parent's data spaces and entitlement. An empty parentCompanySlug
// uses the credential's own company.
func (c *DataSpaceProvisioningClient) List(ctx context.Context, parentCompanySlug string) (*DataSpaceList, error) {
	var spaces DataSpaceList
	err := c.provisioning.request(ctx, http.MethodGet, dataSpacePath("/data-spaces", parentCompanySlug), nil, &spaces)
	return &spaces, err
}

// Revoke removes one data space from a parent. An empty parentCompanySlug uses
// the credential's own company.
func (c *DataSpaceProvisioningClient) Revoke(ctx context.Context, slug, parentCompanySlug string) error {
	path := dataSpacePath("/data-spaces/"+url.PathEscape(slug), parentCompanySlug)
	return c.provisioning.request(ctx, http.MethodDelete, path, nil, nil)
}

// DataSpaceErasureStatus is where a data space erasure is in its life.
type DataSpaceErasureStatus string

const (
	DataSpaceErasureRequested DataSpaceErasureStatus = "requested"
	DataSpaceErasureErasing   DataSpaceErasureStatus = "erasing"
	DataSpaceErasureCompleted DataSpaceErasureStatus = "completed"
	DataSpaceErasureFailed    DataSpaceErasureStatus = "failed"
)

// DataSpaceErasure is the erasure request for one data space and how far it has got.
type DataSpaceErasure struct {
	RequestUUID       string                 `json:"requestUuid"`
	DataSpaceSlug     string                 `json:"dataSpaceSlug"`
	ParentCompanySlug string                 `json:"parentCompanySlug"`
	Status            DataSpaceErasureStatus `json:"status"`
	RequestedAt       string                 `json:"requestedAt"`
}

// Erase asks Custd to erase everything one data space holds. The answer is the
// accepted request, not the finished erasure; read its progress with GetErasure.
func (c *DataSpaceProvisioningClient) Erase(ctx context.Context, slug, parentCompanySlug string) (*DataSpaceErasure, error) {
	var erasure DataSpaceErasure
	path := dataSpacePath("/data-spaces/"+url.PathEscape(slug)+"/erasure", parentCompanySlug)
	err := c.provisioning.request(ctx, http.MethodPost, path, nil, &erasure)
	return &erasure, err
}

// GetErasure reads the erasure request of one data space.
func (c *DataSpaceProvisioningClient) GetErasure(ctx context.Context, slug, parentCompanySlug string) (*DataSpaceErasure, error) {
	var erasure DataSpaceErasure
	path := dataSpacePath("/data-spaces/"+url.PathEscape(slug)+"/erasure", parentCompanySlug)
	err := c.provisioning.request(ctx, http.MethodGet, path, nil, &erasure)
	return &erasure, err
}

// dataSpacePath adds the parentCompanySlug a platform administrator names.
func dataSpacePath(path, parentCompanySlug string) string {
	return siteQueryPath(path, url.Values{"parentCompanySlug": {parentCompanySlug}})
}

func (c *ProducerProvisioningClient) Provision(
	ctx context.Context,
	req ProducerProvisionRequest,
) (*ProducerProvisionResponse, error) {
	var producer ProducerProvisionResponse
	err := c.provisioning.request(ctx, http.MethodPost, "/producer-provisioning", req, &producer)
	return &producer, err
}

func (c *ProducerProvisioningClient) List(
	ctx context.Context,
	companySlug string,
) ([]ProducerProvisionPublicClient, error) {
	path := "/producer-provisioning"
	if companySlug != "" {
		path += "?companySlug=" + url.QueryEscape(companySlug)
	}
	var response struct {
		Producers []ProducerProvisionPublicClient `json:"producers"`
	}
	err := c.provisioning.request(ctx, http.MethodGet, path, nil, &response)
	return response.Producers, err
}

func (c *ProducerProvisioningClient) RotateSecret(
	ctx context.Context,
	clientID string,
) (*ProducerSecretRotationResponse, error) {
	var producer ProducerSecretRotationResponse
	err := c.provisioning.request(
		ctx,
		http.MethodPost,
		"/producer-provisioning/"+url.PathEscape(clientID)+"/rotate-secret",
		nil,
		&producer,
	)
	return &producer, err
}

// ProducerEnvironmentUpdateRequest sets the environment recorded for a producer.
// It is the authenticated default for events that do not declare their own
// environment, not a requirement: one producer credential is expected to carry
// every environment unless an operator deliberately restricts it.
type ProducerEnvironmentUpdateRequest struct {
	Environment string `json:"environment"`
	// AllowedEnvironments replaces the producer's optional restriction. An empty
	// list allows every environment; a nil leaves it unchanged.
	AllowedEnvironments *[]string `json:"allowedEnvironments,omitempty"`
}

// UpdateEnvironment changes the source environment recorded for a producer.
func (c *ProducerProvisioningClient) UpdateEnvironment(
	ctx context.Context,
	clientID string,
	req ProducerEnvironmentUpdateRequest,
) (*ProducerProvisionPublicClient, error) {
	var producer ProducerProvisionPublicClient
	err := c.provisioning.request(
		ctx,
		http.MethodPatch,
		"/producer-provisioning/"+url.PathEscape(clientID)+"/environment",
		req,
		&producer,
	)
	return &producer, err
}

// ProducerSigningKeysUpdateRequest replaces the Ed25519 public keys a signing
// producer's events are verified with: one or two base64 keys, two while rotating.
type ProducerSigningKeysUpdateRequest struct {
	PublicKeys []string `json:"publicKeys"`
}

// ProducerSigningKeys is the set of public keys Custd now verifies a producer's events with.
type ProducerSigningKeys struct {
	ClientID    string   `json:"clientId"`
	CompanySlug string   `json:"companySlug"`
	PublicKeys  []string `json:"publicKeys"`
}

// SetSigningKeys replaces the public keys of a producer minted with
// ScopeTemplateHostingEmailDelivery.
func (c *ProducerProvisioningClient) SetSigningKeys(
	ctx context.Context,
	clientID string,
	req ProducerSigningKeysUpdateRequest,
) (*ProducerSigningKeys, error) {
	var keys ProducerSigningKeys
	err := c.provisioning.request(
		ctx,
		http.MethodPut,
		"/producer-provisioning/"+url.PathEscape(clientID)+"/signing-keys",
		req,
		&keys,
	)
	return &keys, err
}

func (c *ProducerProvisioningClient) Revoke(ctx context.Context, clientID string) error {
	return c.provisioning.request(ctx, http.MethodDelete, "/producer-provisioning/"+url.PathEscape(clientID), nil, nil)
}

// ProducerReservationScope names the parent company a platform administrator
// acts for. Omit it to act for the caller's own company.
type ProducerReservationScope struct {
	ParentCompanySlug string
}

// producerReservationPath builds a reservation route with the optional parent.
func producerReservationPath(dataSpaceSlug, suffix string, scope []ProducerReservationScope) string {
	parent := ""
	if len(scope) > 0 {
		parent = scope[0].ParentCompanySlug
	}
	return dataSpacePath("/data-spaces/"+url.PathEscape(dataSpaceSlug)+"/producer-reservations"+suffix, parent)
}

func (c *ProducerReservationsClient) Reserve(
	ctx context.Context,
	dataSpaceSlug string,
	req ProducerReservationCreateRequest,
	scope ...ProducerReservationScope,
) (*ProducerReservation, error) {
	var out ProducerReservation
	err := c.provisioning.request(ctx, http.MethodPost, producerReservationPath(dataSpaceSlug, "", scope), req, &out)
	return &out, err
}

func (c *ProducerReservationsClient) List(
	ctx context.Context,
	dataSpaceSlug string,
	scope ...ProducerReservationScope,
) (*ProducerReservationListResponse, error) {
	var out ProducerReservationListResponse
	err := c.provisioning.request(ctx, http.MethodGet, producerReservationPath(dataSpaceSlug, "", scope), nil, &out)
	return &out, err
}

func (c *ProducerReservationsClient) Claim(
	ctx context.Context,
	dataSpaceSlug string,
	producerSlug string,
	req ProducerReservationClaimRequest,
	scope ...ProducerReservationScope,
) (*ProducerReservation, error) {
	var out ProducerReservation
	path := producerReservationPath(dataSpaceSlug, "/"+url.PathEscape(producerSlug)+"/claim", scope)
	err := c.provisioning.request(ctx, http.MethodPost, path, req, &out)
	return &out, err
}

func (c *ProducerReservationsClient) Release(
	ctx context.Context, dataSpaceSlug string, producerSlug string, scope ...ProducerReservationScope,
) error {
	path := producerReservationPath(dataSpaceSlug, "/"+url.PathEscape(producerSlug), scope)
	return c.provisioning.request(ctx, http.MethodDelete, path, nil, nil)
}

func (c *ProvisioningClient) request(ctx context.Context, method string, path string, payload any, out any) error {
	body, err := encodeRequestBody(payload, "provisioning")
	if err != nil {
		return err
	}
	response, err := c.client.exchange(ctx, &bufferedRequest{
		Method: method, URL: c.endpoint(path), Headers: c.client.headers(false), Body: body,
	})
	if err != nil {
		return fmt.Errorf("custd: provisioning request failed: %w", err)
	}
	recordMetadata(ctx, response.Header.Get(requestIDHeader))
	if err := c.client.checkStatus(response.StatusCode, response.Body); err != nil {
		return err
	}
	return decodeAdminResponse(response.Body, out)
}

func (c *ProvisioningClient) endpoint(path string) string {
	return strings.TrimRight(c.client.config.BaseURL, "/") + apiEndpoint + path
}
