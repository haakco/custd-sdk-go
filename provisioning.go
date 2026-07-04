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

const apiEndpoint = "/api/v1"

type ProvisioningClient struct {
	DataSpaces *DataSpaceProvisioningClient
	Producers  *ProducerProvisioningClient
	client     *CustdClient
}

type DataSpaceProvisioningClient struct {
	provisioning *ProvisioningClient
}

type ProducerProvisioningClient struct {
	provisioning *ProvisioningClient
}

func newProvisioningClient(client *CustdClient) *ProvisioningClient {
	provisioning := &ProvisioningClient{client: client}
	provisioning.DataSpaces = &DataSpaceProvisioningClient{provisioning: provisioning}
	provisioning.Producers = &ProducerProvisioningClient{provisioning: provisioning}
	return provisioning
}

func (c *DataSpaceProvisioningClient) Create(
	ctx context.Context,
	req DataSpaceCreate,
) (*DataSpace, error) {
	var space DataSpace
	err := c.provisioning.request(ctx, http.MethodPost, "/data-spaces", req, &space)
	return &space, err
}

func (c *DataSpaceProvisioningClient) List(ctx context.Context) (*DataSpaceList, error) {
	var spaces DataSpaceList
	err := c.provisioning.request(ctx, http.MethodGet, "/data-spaces", nil, &spaces)
	return &spaces, err
}

func (c *DataSpaceProvisioningClient) Revoke(ctx context.Context, slug string) error {
	return c.provisioning.request(ctx, http.MethodDelete, "/data-spaces/"+url.PathEscape(slug), nil, nil)
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
	var producers []ProducerProvisionPublicClient
	err := c.provisioning.request(ctx, http.MethodGet, path, nil, &producers)
	return producers, err
}

func (c *ProducerProvisioningClient) RotateSecret(
	ctx context.Context,
	clientID string,
) (*ProducerProvisionResponse, error) {
	var producer ProducerProvisionResponse
	err := c.provisioning.request(
		ctx,
		http.MethodPost,
		"/producer-provisioning/"+url.PathEscape(clientID)+"/rotate-secret",
		nil,
		&producer,
	)
	return &producer, err
}

func (c *ProducerProvisioningClient) Revoke(ctx context.Context, clientID string) error {
	return c.provisioning.request(ctx, http.MethodDelete, "/producer-provisioning/"+url.PathEscape(clientID), nil, nil)
}

func (c *ProvisioningClient) request(ctx context.Context, method string, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal provisioning request: %w", err)
		}
	}
	if c.client.config.HTTPClient != nil {
		return c.requestViaDoer(method, path, body, out)
	}
	return c.requestViaHTTP(ctx, method, path, body, out)
}

func (c *ProvisioningClient) requestViaDoer(method string, path string, body []byte, out any) error {
	resp, err := c.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     c.endpoint(path),
		Headers: c.client.headers(false),
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("custd: provisioning request failed: %w", err)
	}
	if err := c.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return err
	}
	return decodeAdminResponse(resp.Body, out)
}

func (c *ProvisioningClient) requestViaHTTP(ctx context.Context, method string, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create provisioning request: %w", err)
	}
	for k, v := range c.client.headers(false) {
		req.Header.Set(k, v)
	}
	resp, err := c.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("custd: provisioning request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if err := c.client.checkStatus(resp.StatusCode, respBody); err != nil {
		return err
	}
	return decodeAdminResponse(respBody, out)
}

func (c *ProvisioningClient) endpoint(path string) string {
	return strings.TrimRight(c.client.config.BaseURL, "/") + apiEndpoint + path
}
