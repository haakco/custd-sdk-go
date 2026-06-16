package custd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestSetupProducerRegistersSchemasFromDirectory(t *testing.T) {
	dir := t.TempDir()
	schema := `{"eventTypeSlug":"courib.delivery.created","version":"1.0.0","jsonSchema":{"type":"object"}}`
	if err := os.WriteFile(filepath.Join(dir, "delivery-created.json"), []byte(schema), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	doer := &capturingSetupDoer{}
	admin := NewClient(&ClientConfig{
		BaseURL:    "http://localhost:8090",
		APIKey:     "admin-token",
		HTTPClient: doer,
	})
	defer func() { _ = admin.Close(context.Background()) }()

	_, err := SetupProducer(context.Background(), admin, ProducerSetupRequest{
		BaseURL:      "http://localhost:8087",
		TokenURL:     "http://localhost:4444/oauth2/token",
		TenantSlug:   "acme",
		ClientID:     "acme-producer",
		Environment:  "development",
		EnsureTenant: false,
		SchemaDir:    dir,
	})
	if err != nil {
		t.Fatalf("SetupProducer returned error: %v", err)
	}

	if len(doer.requests) != 2 {
		t.Fatalf("expected schema and oauth requests, got %d", len(doer.requests))
	}
	if doer.requests[0].path != "/api/v1/admin/schemas" {
		t.Fatalf("first path = %q", doer.requests[0].path)
	}
	if doer.requests[1].path != "/api/v1/admin/oauth-clients" {
		t.Fatalf("second path = %q", doer.requests[1].path)
	}
	if doer.requests[0].body["eventTypeSlug"] != "courib.delivery.created" {
		t.Fatalf("schema body = %+v", doer.requests[0].body)
	}
}

func TestSetupProducerDoesNotCreateOAuthClientWhenSchemaRegistrationFails(t *testing.T) {
	dir := t.TempDir()
	schema := `{"eventTypeSlug":"courib.delivery.created","version":"1.0.0","jsonSchema":{"type":"object"}}`
	if err := os.WriteFile(filepath.Join(dir, "delivery-created.json"), []byte(schema), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	doer := &capturingSetupDoer{schemaStatus: 503}
	admin := NewClient(&ClientConfig{
		BaseURL:    "http://localhost:8090",
		APIKey:     "admin-token",
		HTTPClient: doer,
	})
	defer func() { _ = admin.Close(context.Background()) }()

	_, err := SetupProducer(context.Background(), admin, ProducerSetupRequest{
		BaseURL:      "http://localhost:8087",
		TokenURL:     "http://localhost:4444/oauth2/token",
		TenantSlug:   "acme",
		ClientID:     "acme-producer",
		Environment:  "development",
		EnsureTenant: false,
		SchemaDir:    dir,
	})

	if err == nil || !strings.Contains(err.Error(), "register schema") {
		t.Fatalf("expected register schema error, got %v", err)
	}
	if len(doer.requests) != 1 || doer.requests[0].path != "/api/v1/admin/schemas" {
		t.Fatalf("expected only schema request before failure, got %+v", doer.requests)
	}
}

func TestSetupProducerPreflightsSchemaDirectoryBeforeCreatingOAuthClient(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{"eventTypeSlug":`), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	doer := &capturingSetupDoer{}
	admin := NewClient(&ClientConfig{
		BaseURL:    "http://localhost:8090",
		APIKey:     "admin-token",
		HTTPClient: doer,
	})
	defer func() { _ = admin.Close(context.Background()) }()

	_, err := SetupProducer(context.Background(), admin, ProducerSetupRequest{
		BaseURL:      "http://localhost:8087",
		TokenURL:     "http://localhost:4444/oauth2/token",
		TenantSlug:   "acme",
		ClientID:     "acme-producer",
		Environment:  "development",
		EnsureTenant: true,
		SchemaDir:    dir,
	})

	if err == nil || !strings.Contains(err.Error(), "decode schema") {
		t.Fatalf("expected decode schema error, got %v", err)
	}
	if len(doer.requests) != 0 {
		t.Fatalf("expected no mutating requests before schema preflight, got %+v", doer.requests)
	}
}

type setupRequest struct {
	path string
	body map[string]any
}

type capturingSetupDoer struct {
	requests     []setupRequest
	schemaStatus int
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
	case "/api/v1/admin/schemas":
		if d.schemaStatus != 0 {
			return jsonResponse(d.schemaStatus, map[string]any{"error": "schema unavailable"})
		}
		return jsonResponse(201, body)
	default:
		return jsonResponse(404, map[string]any{"error": "not found"})
	}
}

func TestNewClientFromProvisionedProducerConsumesBundleDirectly(t *testing.T) {
	creds := loadProvisionedProducerFixture(t, "valid-provisioned-producer.json")

	client, err := NewClientFromProvisionedProducer(creds)
	if err != nil {
		t.Fatalf("NewClientFromProvisionedProducer returned error: %v", err)
	}
	defer func() { _ = client.Close(context.Background()) }()

	if client == nil {
		t.Fatal("expected client, got nil")
	}
}

func TestNewClientFromProvisionedProducerRejectsMissingSecret(t *testing.T) {
	creds := loadProvisionedProducerFixture(t, "invalid-provisioned-producer-missing-secret.json")

	_, err := NewClientFromProvisionedProducer(creds)
	if err == nil || !strings.Contains(err.Error(), "client secret") {
		t.Fatalf("expected missing client secret error, got %v", err)
	}
}

func TestRedactedProvisionedProducerOmitsSecret(t *testing.T) {
	creds := loadProvisionedProducerFixture(t, "valid-provisioned-producer.json")

	redacted := RedactedProvisionedProducer(creds)

	if redacted.ClientID != creds.ClientID {
		t.Fatalf("expected client ID %q, got %q", creds.ClientID, redacted.ClientID)
	}
	encoded, err := json.Marshal(redacted)
	if err != nil {
		t.Fatalf("marshal redacted: %v", err)
	}
	if strings.Contains(string(encoded), creds.ClientSecret) {
		t.Fatalf("redacted view leaked secret: %s", encoded)
	}
	if strings.Contains(string(encoded), "clientSecret") {
		t.Fatalf("redacted view exposed clientSecret field: %s", encoded)
	}
}

func loadProvisionedProducerFixture(t *testing.T, name string) ProvisionedProducerCredentials {
	t.Helper()
	path := filepath.Join("..", "contract-fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var creds ProvisionedProducerCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return creds
}

func jsonResponse(status int, body map[string]any) (*HTTPResponse, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{StatusCode: status, Body: encoded}, nil
}
