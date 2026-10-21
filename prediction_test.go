package custd

import (
	"context"
	"encoding/json"
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

// predictionWireCase is one row of the prediction wire-contract table: the
// server response body, the call that produces it, and the request fields
// the Admin API must receive on the wire.
type predictionWireCase struct {
	name        string
	response    string
	invoke      func(*testing.T, *CustdClient)
	wantRequest map[string]string
}

// TestPredictionDefinitionFieldsReachTheWire proves the prediction fields a
// caller needs survive as real wire bytes: the request body carries the
// values the Admin API expects, and the response body decodes into the
// typed definition.
func TestPredictionDefinitionFieldsReachTheWire(t *testing.T) {
	for _, tc := range []predictionWireCase{
		predictionCreateDefinitionWireCase(),
		predictionActivateWireCase(),
	} {
		t.Run(tc.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, tc.response)
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			tc.invoke(t, client)

			if len(doer.requests) != 1 {
				t.Fatalf("requests = %d", len(doer.requests))
			}
			assertPredictionWireFields(t, doer.requests[0].Body, tc.wantRequest)
		})
	}
}

func predictionCreateDefinitionWireCase() predictionWireCase {
	return predictionWireCase{
		name:     "create definition carries project kind",
		response: `{"uuid":"definition-1","project_kind":"duration","definition_key":"quota","display_name":"Quota","status":"draft","schedule_kind":"interval","is_paused":false,"active_version_uuid":"version-active","candidate_version_uuid":"version-candidate","candidate_version_count":2,"created_at":"2026-08-12T12:00:00Z","updated_at":"2026-08-12T12:00:00Z"}`,
		invoke: func(t *testing.T, c *CustdClient) {
			t.Helper()
			definition, err := c.Admin.Predictions.CreateDefinition(context.Background(), "acme", PredictionDefinitionCreateRequest{
				ProjectKind:   ProjectKindDuration,
				DefinitionKey: "quota",
				DisplayName:   "Quota",
			})
			if err != nil {
				t.Fatalf("CreateDefinition returned error: %v", err)
			}
			if definition.ProjectKind != ProjectKindDuration {
				t.Fatalf("project kind = %q", definition.ProjectKind)
			}
			if definition.ActiveVersionUUID != "version-active" ||
				definition.CandidateVersionUUID != "version-candidate" ||
				definition.CandidateVersionCount != 2 {
				t.Fatalf("version state = %+v", definition)
			}
		},
		wantRequest: map[string]string{"project_kind": `"duration"`},
	}
}

func predictionActivateWireCase() predictionWireCase {
	return predictionWireCase{
		name:     "activate version carries reason",
		response: `{"uuid":"version-1","version_uuid":"version-1","version_number":1,"version_status":"active","definition_hash":"hash","definition":{},"feature_count":0,"source_count":0,"created_by":"ops","created_at":"2026-08-12T12:00:00Z"}`,
		invoke: func(t *testing.T, c *CustdClient) {
			t.Helper()
			if _, err := c.Admin.Predictions.ActivateVersion(context.Background(), "acme", "definition-1", PredictionActivateRequest{
				VersionUUID: "version-1",
				Reason:      "scheduled activation",
			}); err != nil {
				t.Fatalf("ActivateVersion returned error: %v", err)
			}
		},
		wantRequest: map[string]string{"reason": `"scheduled activation"`},
	}
}

// assertPredictionWireFields checks the exact JSON bytes the SDK put on the
// wire, so a renamed or mistyped tag fails the test instead of silently
// changing the request body.
func assertPredictionWireFields(t *testing.T, body []byte, want map[string]string) {
	t.Helper()
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode request body %s: %v", body, err)
	}
	for key, wantValue := range want {
		got, ok := decoded[key]
		if !ok {
			t.Fatalf("request body %s is missing %q", body, key)
		}
		if string(got) != wantValue {
			t.Fatalf("request body %s: %s = %s, want %s", body, key, got, wantValue)
		}
	}
}
