package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestOAuthUpdateScopesUsesPatchEndpoint(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"clientId":"custd-acme","companySlug":"acme","scopes":["events.write","events.read"]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	updated, err := client.Admin.OAuthClients.UpdateScopes(context.Background(), "custd-acme", AdminOAuthClientUpdateScopesRequest{
		Profile: "standard",
		Scopes:  []string{"events.write", "events.read"},
	})
	if err != nil {
		t.Fatalf("UpdateScopes error: %v", err)
	}
	if updated.ClientID != "custd-acme" || updated.ClientSecret != "" {
		t.Fatalf("update leaked secret: %+v", updated)
	}
	if doer.requests[0].Method != http.MethodPatch ||
		doer.requests[0].URL != "http://localhost:8080/api/v1/admin/oauth-clients/custd-acme/scopes" {
		t.Fatalf("UpdateScopes request = %+v", doer.requests[0])
	}
}

func TestSchemaAdminClientValidateEnableDryRunAudit(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"valid":true,"issues":[],"suggestedAction":"register"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	val, err := client.Admin.Schemas.Validate(context.Background(), AdminSchemaValidateRequest{
		TenantSlug:    "acme",
		EventTypeSlug: "courib.delivery.created",
		Dialect:       "jsonschema",
		SchemaJSON:    `{"type":"object"}`,
	})
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if !val.Valid || val.SuggestedAction != "register" {
		t.Fatalf("Validate = %+v", val)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/admin/schema/validate" {
		t.Fatalf("Validate URL = %s", doer.requests[0].URL)
	}

	doer.status = http.StatusNoContent
	if err := client.Admin.Schemas.EnableVersion(context.Background(), "acme", "courib.delivery.created", AdminSchemaEnableRequest{
		TenantSlug:    "acme",
		EventTypeSlug: "courib.delivery.created",
		Version:       "1.0.0",
	}); err != nil {
		t.Fatalf("EnableVersion error: %v", err)
	}
	if doer.requests[1].URL != "http://localhost:8080/api/v1/admin/schema/acme/courib.delivery.created/enable" {
		t.Fatalf("EnableVersion URL = %s", doer.requests[1].URL)
	}

	doer.status = http.StatusOK
	doer.body = `{"passed":3,"failed":0,"issues":[]}`
	dry, err := client.Admin.Schemas.DryRun(context.Background(), "acme", "courib.delivery.created", AdminSchemaDryRunRequest{
		TenantSlug:    "acme",
		EventTypeSlug: "courib.delivery.created",
		Dialect:       "jsonschema",
		SchemaJSON:    `{"type":"object"}`,
		Samples:       []any{map[string]any{"x": 1}},
	})
	if err != nil {
		t.Fatalf("DryRun error: %v", err)
	}
	if dry.Passed != 3 {
		t.Fatalf("DryRun = %+v", dry)
	}

	doer.body = `{"entries":[{"eventId":"ev-1","eventTypeSlug":"courib.delivery.created","version":"1.0.0","action":"registered","registeredAt":"2026-07-23T12:00:00Z","registeredBy":"u-1"}]}`
	audit, err := client.Admin.Schemas.Audit(context.Background())
	if err != nil {
		t.Fatalf("Audit error: %v", err)
	}
	if len(audit.Entries) != 1 || audit.Entries[0].Version != "1.0.0" {
		t.Fatalf("Audit = %+v", audit)
	}
}
