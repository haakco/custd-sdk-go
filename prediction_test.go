package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestAdminPredictionsUseTenantScopedTypedContract(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"uuid":"definition-1","definition_key":"quota","display_name":"Quota","status":"draft","schedule_kind":"interval","is_paused":false,"created_at":"2026-08-12T12:00:00Z","updated_at":"2026-08-12T12:00:00Z"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	created, err := client.Admin.Predictions.CreateDefinition(context.Background(), "acme", PredictionDefinitionCreateRequest{
		DefinitionKey: "quota", DisplayName: "Quota",
	})
	if err != nil {
		t.Fatalf("CreateDefinition returned error: %v", err)
	}
	if created.UUID != "definition-1" || created.Status != "draft" {
		t.Fatalf("definition = %+v", created)
	}
	if got := doer.requests[0].URL; got != "http://localhost:8080/api/v1/admin/measurement/predictions/definitions?companySlug=acme" {
		t.Fatalf("create URL = %s", got)
	}

	doer.status = http.StatusOK
	doer.body = `[{"uuid":"source-1","source_key":"status","source_mode":"http_json","display_name":"Status","source_status":"active","is_paused":false,"created_at":"2026-08-12T12:00:00Z","updated_at":"2026-08-12T12:00:00Z","consecutive_failed_count":0}]`
	sources, err := client.Admin.Predictions.ListSignalSources(context.Background(), "acme", 20, "next")
	if err != nil {
		t.Fatalf("ListSignalSources returned error: %v", err)
	}
	if len(sources) != 1 || sources[0].SourceMode != "http_json" {
		t.Fatalf("sources = %+v", sources)
	}
	if got := doer.requests[1].URL; got != "http://localhost:8080/api/v1/admin/measurement/predictions/sources?companySlug=acme&pageSize=20&pageToken=next" {
		t.Fatalf("sources URL = %s", got)
	}

	doer.status = http.StatusAccepted
	doer.body = ""
	if err := client.Admin.Predictions.RunNow(context.Background(), "acme", "definition-1", PredictionRunNowRequest{WorkerID: "proof"}); err != nil {
		t.Fatalf("RunNow returned error: %v", err)
	}
	if got := doer.requests[2].URL; got != "http://localhost:8080/api/v1/admin/measurement/predictions/definitions/definition-1/run-now?companySlug=acme" {
		t.Fatalf("run-now URL = %s", got)
	}
}
