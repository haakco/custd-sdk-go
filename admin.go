package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const adminEndpoint = "/api/v1/admin"

type AdminClient struct {
	Tenants      *TenantAdminClient
	OAuthClients *OAuthClientAdminClient
	Sites        *SiteAdminClient
	Schemas      *SchemaAdminClient
	client       *CustdClient
}

type TenantAdminClient struct {
	admin *AdminClient
}

type OAuthClientAdminClient struct {
	admin *AdminClient
}

type SiteAdminClient struct {
	admin *AdminClient
}

type SchemaAdminClient struct {
	admin *AdminClient
}

func newAdminClient(client *CustdClient) *AdminClient {
	admin := &AdminClient{client: client}
	admin.Tenants = &TenantAdminClient{admin: admin}
	admin.OAuthClients = &OAuthClientAdminClient{admin: admin}
	admin.Sites = &SiteAdminClient{admin: admin}
	admin.Schemas = &SchemaAdminClient{admin: admin}
	return admin
}

func (c *TenantAdminClient) Create(ctx context.Context, req AdminTenantCreate) (*AdminTenant, error) {
	var tenant AdminTenant
	err := c.admin.request(ctx, http.MethodPost, "/tenants", req, &tenant)
	return &tenant, err
}

func (c *TenantAdminClient) List(ctx context.Context) (*AdminTenantList, error) {
	var tenants AdminTenantList
	err := c.admin.request(ctx, http.MethodGet, "/tenants", nil, &tenants)
	return &tenants, err
}

func (c *TenantAdminClient) Get(ctx context.Context, slug string) (*AdminTenant, error) {
	return adminGetByID[AdminTenant](ctx, c.admin, "/tenants/", slug)
}

func (c *TenantAdminClient) Delete(ctx context.Context, slug string) error {
	return c.admin.request(ctx, http.MethodDelete, "/tenants/"+url.PathEscape(slug), nil, nil)
}

func (c *OAuthClientAdminClient) Create(
	ctx context.Context,
	req AdminOAuthClientCreate,
) (*AdminOAuthClientCreateResponse, error) {
	var client AdminOAuthClientCreateResponse
	err := c.admin.request(ctx, http.MethodPost, "/oauth-clients", req, &client)
	return &client, err
}

func (c *OAuthClientAdminClient) List(ctx context.Context) (*AdminOAuthClientList, error) {
	var clients AdminOAuthClientList
	err := c.admin.request(ctx, http.MethodGet, "/oauth-clients", nil, &clients)
	return &clients, err
}

func (c *OAuthClientAdminClient) Get(ctx context.Context, clientID string) (*AdminOAuthClient, error) {
	return adminGetByID[AdminOAuthClient](ctx, c.admin, "/oauth-clients/", clientID)
}

// adminGetByID fetches a single resource of type T from an admin collection,
// escaping id into the path. It is the shared body for the per-resource Get
// helpers so they don't duplicate the marshal/return boilerplate.
func adminGetByID[T any](ctx context.Context, admin *AdminClient, prefix, id string) (*T, error) {
	var out T
	err := admin.request(ctx, http.MethodGet, prefix+url.PathEscape(id), nil, &out)
	return &out, err
}

func (c *OAuthClientAdminClient) Delete(ctx context.Context, clientID string) error {
	return c.admin.request(ctx, http.MethodDelete, "/oauth-clients/"+url.PathEscape(clientID), nil, nil)
}

func (c *OAuthClientAdminClient) RotateSecret(
	ctx context.Context,
	clientID string,
) (*AdminOAuthClientSecretResponse, error) {
	var secret AdminOAuthClientSecretResponse
	err := c.admin.request(ctx, http.MethodPost, "/oauth-clients/"+url.PathEscape(clientID)+"/rotate-secret", nil, &secret)
	return &secret, err
}

func (c *SiteAdminClient) Create(ctx context.Context, req AdminSiteCreate) (*AdminSiteCreateResponse, error) {
	var site AdminSiteCreateResponse
	err := c.admin.request(ctx, http.MethodPost, "/sites", req, &site)
	return &site, err
}

func (c *SiteAdminClient) List(ctx context.Context) (*AdminSiteList, error) {
	var sites AdminSiteList
	err := c.admin.request(ctx, http.MethodGet, "/sites", nil, &sites)
	return &sites, err
}

func (c *SiteAdminClient) Get(ctx context.Context, siteUUID string) (*AdminSite, error) {
	return adminGetByID[AdminSite](ctx, c.admin, "/sites/", siteUUID)
}

func (c *SiteAdminClient) Delete(ctx context.Context, siteUUID string) error {
	return c.admin.request(ctx, http.MethodDelete, "/sites/"+url.PathEscape(siteUUID), nil, nil)
}

func (c *SiteAdminClient) RotateWriteKey(
	ctx context.Context,
	siteUUID string,
) (*AdminSiteWriteKeyResponse, error) {
	var key AdminSiteWriteKeyResponse
	err := c.admin.request(ctx, http.MethodPost, "/sites/"+url.PathEscape(siteUUID)+"/rotate-write-key", nil, &key)
	return &key, err
}

func (c *SchemaAdminClient) Register(ctx context.Context, req AdminSchemaRegister) (*AdminSchema, error) {
	var schema AdminSchema
	err := c.admin.request(ctx, http.MethodPost, "/schemas", req, &schema)
	return &schema, err
}

func (c *SchemaAdminClient) List(ctx context.Context) (*AdminSchemaList, error) {
	var schemas AdminSchemaList
	err := c.admin.request(ctx, http.MethodGet, "/schemas", nil, &schemas)
	return &schemas, err
}

func (c *SchemaAdminClient) Get(ctx context.Context, eventTypeSlug string) (*AdminSchema, error) {
	return adminGetByID[AdminSchema](ctx, c.admin, "/schemas/", eventTypeSlug)
}

func (c *SchemaAdminClient) CreateVersion(
	ctx context.Context,
	eventTypeSlug string,
	req AdminSchemaVersionCreate,
) (*AdminSchema, error) {
	var schema AdminSchema
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/schemas/"+url.PathEscape(eventTypeSlug)+"/versions",
		req,
		&schema,
	)
	return &schema, err
}

func (c *AdminClient) request(ctx context.Context, method string, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal admin request: %w", err)
		}
	}
	if c.client.config.HTTPClient != nil {
		return c.requestViaDoer(method, path, body, out)
	}
	return c.requestViaHTTP(ctx, method, path, body, out)
}

func (c *AdminClient) requestViaDoer(method string, path string, body []byte, out any) error {
	resp, err := c.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     c.endpoint(path),
		Headers: c.client.headers(),
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("custd: admin request failed: %w", err)
	}
	if err := c.client.checkStatus(resp.StatusCode); err != nil {
		return err
	}
	return decodeAdminResponse(resp.Body, out)
}

func (c *AdminClient) requestViaHTTP(ctx context.Context, method string, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create admin request: %w", err)
	}
	for k, v := range c.client.headers() {
		req.Header.Set(k, v)
	}
	resp, err := c.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("custd: admin request failed: %w", err)
	}
	// nolint:errcheck // response body fully read below; a close error cannot affect the already-read admin response
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if err := c.client.checkStatus(resp.StatusCode); err != nil {
		return err
	}
	return decodeAdminResponse(respBody, out)
}

func (c *AdminClient) endpoint(path string) string {
	return strings.TrimRight(c.client.config.BaseURL, "/") + adminEndpoint + path
}

func decodeAdminResponse(body []byte, out any) error {
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("custd: decode admin response: %w", err)
	}
	return nil
}
