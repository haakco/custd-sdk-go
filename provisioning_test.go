package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestProvisioningDataSpacesUsePublicAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"slug":"agency-store-001","companyName":"Agency Store 001","parentCompanySlug":"agency","enabled":true}`)
	client := newProvisioningTestClient(t, doer)

	space, err := client.Provisioning.DataSpaces.Create(context.Background(), DataSpaceCreate{
		Slug: "agency-store-001", CompanyName: "Agency Store 001",
	}, "")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if space.Slug != "agency-store-001" {
		t.Fatalf("space = %+v", space)
	}
	var body map[string]string
	if err := json.Unmarshal(doer.requests[0].Body, &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/data-spaces" || body["slug"] != "agency-store-001" {
		t.Fatalf("request = %+v body=%+v", doer.requests[0], body)
	}

	doer.status = http.StatusOK
	doer.body = `{"dataSpaces":[],"entitlement":{"enabled":true,"activeDataSpaces":0,"maxActiveDataSpaces":5,"maxActiveProducersPerDataSpace":3}}`
	list, err := client.Provisioning.DataSpaces.List(context.Background(), "agency")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if !list.Entitlement.Enabled || list.Entitlement.MaxActiveDataSpaces != 5 {
		t.Fatalf("list = %+v", list)
	}
	doer.status = http.StatusNoContent
	doer.body = ""
	if err := client.Provisioning.DataSpaces.Revoke(context.Background(), "agency store/001", "agency"); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/data-spaces",
		"GET http://localhost:8080/api/v1/data-spaces?parentCompanySlug=agency",
		"DELETE http://localhost:8080/api/v1/data-spaces/agency%20store%2F001?parentCompanySlug=agency",
	})
}

func TestProvisioningProducersUsePublicAPIAndKeepSecretExplicit(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"baseUrl":"https://api.custd.test","tokenUrl":"https://auth.custd.test/oauth2/token","audience":"custd","companySlug":"agency-store-001","producerSlug":"webhook","clientId":"custd-agency-store-001-webhook","clientSecret":"once","scopes":["events.write"],"environment":"production","allowedEnvironments":["production"]}`)
	client := newProvisioningTestClient(t, doer)

	created, err := client.Provisioning.Producers.Provision(context.Background(), ProducerProvisionRequest{
		CompanySlug: "agency-store-001", ProducerSlug: "webhook", ScopeTemplate: "managed-audit",
	})
	if err != nil {
		t.Fatalf("Provision returned error: %v", err)
	}
	if created.ClientSecret != "once" || created.Environment != "production" ||
		len(created.AllowedEnvironments) != 1 || created.AllowedEnvironments[0] != "production" {
		t.Fatalf("created = %+v", created)
	}
	doer.status = http.StatusOK
	doer.body = `{"producers":[{"clientId":"custd-agency-store-001-webhook","companySlug":"agency-store-001","producerSlug":"webhook","displayName":"Webhook","scopes":["events.write"],"environment":"production","allowedEnvironments":["production","staging"]}]}`
	producers, err := client.Provisioning.Producers.List(context.Background(), "agency-store-001")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(producers) != 1 || producers[0].ProducerSlug != "webhook" || producers[0].DisplayName != "Webhook" ||
		len(producers[0].AllowedEnvironments) != 2 {
		t.Fatalf("producers = %+v", producers)
	}
	doer.body = `{"baseUrl":"https://api.custd.test","tokenUrl":"https://auth.custd.test/oauth2/token","audience":"custd","clientId":"custd-agency-store-001-webhook","clientSecret":"next"}`
	rotated, err := client.Provisioning.Producers.RotateSecret(context.Background(), "custd/agency store")
	if err != nil {
		t.Fatalf("RotateSecret returned error: %v", err)
	}
	if rotated.ClientSecret != "next" || rotated.TokenURL != "https://auth.custd.test/oauth2/token" {
		t.Fatalf("rotated = %+v", rotated)
	}
	doer.status = http.StatusOK
	doer.body = `{"clientId":"custd-agency-store-001-webhook","companySlug":"agency-store-001","producerSlug":"webhook","environment":"staging"}`
	updated, err := client.Provisioning.Producers.UpdateEnvironment(
		context.Background(), "custd/agency store", ProducerEnvironmentUpdateRequest{Environment: "staging"},
	)
	if err != nil {
		t.Fatalf("UpdateEnvironment returned error: %v", err)
	}
	if updated.Environment != "staging" {
		t.Fatalf("updated = %+v", updated)
	}
	doer.status = http.StatusNoContent
	doer.body = ""
	if err := client.Provisioning.Producers.Revoke(context.Background(), "custd/agency store"); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/producer-provisioning",
		"GET http://localhost:8080/api/v1/producer-provisioning?companySlug=agency-store-001",
		"POST http://localhost:8080/api/v1/producer-provisioning/custd%2Fagency%20store/rotate-secret",
		"PATCH http://localhost:8080/api/v1/producer-provisioning/custd%2Fagency%20store/environment",
		"DELETE http://localhost:8080/api/v1/producer-provisioning/custd%2Fagency%20store",
	})
}

func newProvisioningTestClient(t *testing.T, doer *captureDoer) *CustdClient {
	t.Helper()
	client := NewClient(&ClientConfig{
		BaseURL:       "http://localhost:8080/",
		APIKey:        "broker-token",
		HTTPClient:    fakeTransport(doer),
		FlushInterval: time.Hour,
	})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	return client
}
