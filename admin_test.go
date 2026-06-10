package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type captureDoer struct {
	requests []*HTTPRequest
	status   int
	body     string
}

func (d *captureDoer) Do(req *HTTPRequest) (*HTTPResponse, error) {
	d.requests = append(d.requests, req)
	return &HTTPResponse{StatusCode: d.status, Body: []byte(d.body)}, nil
}

func TestAdminTenantsCreateUsesAdminAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"slug":"acme","companyName":"Acme Inc","enabled":true}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	tenant, err := client.Admin.Tenants.Create(context.Background(), AdminTenantCreate{
		Slug:        "acme",
		CompanyName: "Acme Inc",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if tenant.Slug != "acme" || tenant.CompanyName != "Acme Inc" || !tenant.Enabled {
		t.Fatalf("unexpected tenant: %+v", tenant)
	}
	assertTenantCreateRequest(t, doer.requests[0])
}

func assertTenantCreateRequest(t *testing.T, req *HTTPRequest) {
	t.Helper()
	if req.Method != http.MethodPost {
		t.Fatalf("method = %s", req.Method)
	}
	if req.URL != "http://localhost:8080/api/v1/admin/tenants" {
		t.Fatalf("url = %s", req.URL)
	}
	if req.Headers["Authorization"] != "Bearer admin-token" {
		t.Fatalf("authorization header = %q", req.Headers["Authorization"])
	}
	var body map[string]string
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["slug"] != "acme" || body["companyName"] != "Acme Inc" {
		t.Fatalf("body = %+v", body)
	}
}

func TestAdminOAuthClientsReturnSecretOnlyFromSecretEndpoints(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"clientId":"custd-acme","companySlug":"acme","scopes":["events.write"],"clientSecret":"secret"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	created, err := createAcmeOAuthClient(client)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.ClientSecret != "secret" {
		t.Fatalf("clientSecret = %q", created.ClientSecret)
	}
	req := doer.requests[0]
	if req.URL != "http://localhost:8080/api/v1/admin/oauth-clients" {
		t.Fatalf("url = %s", req.URL)
	}
	assertOAuthClientListDoesNotLeakSecret(t, client, doer)
}

func TestAdminSitesManageBrowserSites(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"siteUuid":"site-123","companySlug":"acme","name":"Docs","identityMode":"cookieless","allowedOrigins":["https://example.com"],"rateLimitPerMinute":600,"retentionDays":365,"enabled":true,"writeKey":"site_pk_test"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	created, err := client.Admin.Sites.Create(context.Background(), AdminSiteCreate{
		CompanySlug:    "acme",
		Name:           "Docs",
		IdentityMode:   "cookieless",
		AllowedOrigins: []string{"https://example.com"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.WriteKey != "site_pk_test" || doer.requests[0].URL != "http://localhost:8080/api/v1/admin/sites" {
		t.Fatalf("created=%+v request=%+v", created, doer.requests[0])
	}
	doer.status = http.StatusOK
	doer.body = `{"writeKey":"site_pk_next"}`
	rotated, err := client.Admin.Sites.RotateWriteKey(context.Background(), "site-123")
	if err != nil {
		t.Fatalf("RotateWriteKey returned error: %v", err)
	}
	if rotated.WriteKey != "site_pk_next" {
		t.Fatalf("writeKey = %q", rotated.WriteKey)
	}
}

func TestAdminSitesListGetDeleteDoNotExposeWriteKeys(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"sites":[{"siteUuid":"site-123","companySlug":"acme","name":"Docs","identityMode":"cookieless","allowedOrigins":["https://example.com"],"rateLimitPerMinute":600,"retentionDays":365,"enabled":true}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, err := client.Admin.Sites.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Sites) != 1 || list.Sites[0].SiteUUID != "site-123" {
		t.Fatalf("listed sites = %+v, want site response without secret fields", list.Sites)
	}
	doer.body = `{"siteUuid":"site-123","companySlug":"acme","name":"Docs","identityMode":"cookieless","allowedOrigins":["https://example.com"],"rateLimitPerMinute":600,"retentionDays":365,"enabled":true}`
	site, err := client.Admin.Sites.Get(context.Background(), "site-123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if site.SiteUUID != "site-123" {
		t.Fatalf("get site = %+v, want site response without secret fields", site)
	}
	doer.status = http.StatusNoContent
	doer.body = ""
	if err := client.Admin.Sites.Delete(context.Background(), "site-123"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	assertSiteRequests(t, doer.requests[0:], []string{
		"GET http://localhost:8080/api/v1/admin/sites",
		"GET http://localhost:8080/api/v1/admin/sites/site-123",
		"DELETE http://localhost:8080/api/v1/admin/sites/site-123",
	})
}

func assertSiteRequests(t *testing.T, requests []*HTTPRequest, want []string) {
	t.Helper()
	if len(requests) != len(want) {
		t.Fatalf("requests = %d, want %d", len(requests), len(want))
	}
	for i := range want {
		got := requests[i].Method + " " + requests[i].URL
		if got != want[i] {
			t.Fatalf("request[%d] = %s, want %s", i, got, want[i])
		}
	}
}

func assertOAuthClientListDoesNotLeakSecret(t *testing.T, client *CustdClient, doer *captureDoer) {
	t.Helper()
	doer.status = http.StatusOK
	doer.body = `{"clients":[{"clientId":"custd-acme","companySlug":"acme","scopes":["events.write"]}]}`
	list, err := client.Admin.OAuthClients.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Clients) != 1 {
		t.Fatalf("clients = %+v", list.Clients)
	}
	if list.Clients[0].ClientSecret != "" {
		t.Fatalf("listed client leaked secret: %+v", list.Clients[0])
	}
}

func newCaptureDoer(status int, body string) *captureDoer {
	return &captureDoer{status: status, body: body}
}

func newAdminTestClient(t *testing.T, doer *captureDoer, baseURL string) *CustdClient {
	t.Helper()
	client := NewClient(&ClientConfig{
		BaseURL:       baseURL,
		APIKey:        "admin-token",
		HTTPClient:    doer,
		FlushInterval: time.Hour,
	})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	return client
}

func createAcmeOAuthClient(client *CustdClient) (*AdminOAuthClientCreateResponse, error) {
	return client.Admin.OAuthClients.Create(context.Background(), AdminOAuthClientCreate{
		ClientID:    "custd-acme",
		CompanySlug: "acme",
		Scopes:      []string{"events.write"},
	})
}
