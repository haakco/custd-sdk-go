package custd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnvSnippetsIncludesAllConsumerFamilies(t *testing.T) {
	snippets := EnvSnippets("vorrent media-cache", ProducerCredentials{
		BaseURL:      "https://custd.example",
		TokenURL:     "https://auth.example/oauth2/token",
		Audience:     "custd",
		TenantSlug:   "vorrent",
		ClientID:     "vorrent-media-cache",
		ClientSecret: "secret",
		Scopes:       []string{"events.write"},
		Environment:  "production",
	})

	for _, want := range []string{
		"# Generic",
		"# Go / TypeScript / Python / PHP",
		"# Laravel",
		"# WordPress",
		"VORRENT_MEDIA_CACHE_OAUTH_CLIENT_ID=\"vorrent-media-cache\"",
		"CUSTD_OAUTH_CLIENT_SECRET=\"secret\"",
		"CUSTD_WP_TENANT_SLUG=\"vorrent\"",
	} {
		if !strings.Contains(snippets, want) {
			t.Fatalf("expected snippet to contain %q:\n%s", want, snippets)
		}
	}
}

func TestValidateProducerSetupRequestRejectsPlainRemoteHTTP(t *testing.T) {
	err := ValidateProducerSetupRequest(ProducerSetupRequest{
		BaseURL:    "http://custd.example",
		TokenURL:   "https://auth.example/oauth2/token",
		TenantSlug: "acme",
		ClientID:   "acme-producer",
	})

	if err == nil || !strings.Contains(err.Error(), "base URL must use https") {
		t.Fatalf("expected secure URL error, got %v", err)
	}
}

func TestSetupProducerCreatesTenantAndOAuthClient(t *testing.T) {
	doer := &capturingSetupDoer{}
	admin := NewClient(&ClientConfig{
		BaseURL:    "http://localhost:8090",
		APIKey:     "admin-token",
		HTTPClient: doer,
	})
	defer func() { _ = admin.Close(context.Background()) }()

	creds, err := SetupProducer(context.Background(), admin, ProducerSetupRequest{
		BaseURL:      "http://localhost:8087",
		TokenURL:     "http://localhost:4444/oauth2/token",
		Audience:     "custd",
		TenantSlug:   "acme",
		CompanyName:  "Acme Inc",
		ClientID:     "acme-producer",
		Scopes:       []string{"events.write"},
		Environment:  "development",
		EnsureTenant: true,
	})
	if err != nil {
		t.Fatalf("SetupProducer returned error: %v", err)
	}
	if creds.ClientSecret != "created-secret" {
		t.Fatalf("expected created secret, got %q", creds.ClientSecret)
	}
	if len(doer.requests) != 2 {
		t.Fatalf("expected tenant and oauth requests, got %d", len(doer.requests))
	}
	if doer.requests[0].path != "/api/v1/admin/tenants" {
		t.Fatalf("unexpected first path %q", doer.requests[0].path)
	}
	if doer.requests[1].path != "/api/v1/admin/oauth-clients" {
		t.Fatalf("unexpected second path %q", doer.requests[1].path)
	}
}

type setupRequest struct {
	path string
	body map[string]any
}

type capturingSetupDoer struct {
	requests []setupRequest
}

func (d *capturingSetupDoer) Do(req *HTTPRequest) (*HTTPResponse, error) {
	path := strings.TrimPrefix(req.URL, "http://localhost:8090")
	var body map[string]any
	if len(req.Body) > 0 {
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return nil, err
		}
	}
	d.requests = append(d.requests, setupRequest{path: path, body: body})
	switch path {
	case "/api/v1/admin/tenants":
		return jsonResponse(201, map[string]any{
			"slug":        body["slug"],
			"companyName": body["companyName"],
			"enabled":     true,
		})
	case "/api/v1/admin/oauth-clients":
		return jsonResponse(201, map[string]any{
			"clientId":     body["clientId"],
			"companySlug":  body["companySlug"],
			"scopes":       body["scopes"],
			"clientSecret": "created-secret",
		})
	default:
		return jsonResponse(404, map[string]any{"error": "not found"})
	}
}

func jsonResponse(status int, body map[string]any) (*HTTPResponse, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{StatusCode: status, Body: encoded}, nil
}
