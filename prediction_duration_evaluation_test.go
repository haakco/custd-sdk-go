package custd

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// durationEvaluationResponseBody is one server answer: the artifact, the
// outcome counts and the version evidence the route returns.
const durationEvaluationResponseBody = `{
	"definitionUuid": "definition-1",
	"planUuid": "plan-1",
	"semanticKey": "build_step",
	"artifact": {
		"schema_version": "duration-evaluation-artifact.v1",
		"source": "libs/go/measurement",
		"prediction_version_uuid": "version-1",
		"content_sha256": "abc123",
		"evaluation": {
			"tenant_id": "acme co",
			"project_id": "project-1",
			"series_id": "series-1",
			"minimum_training_history": 5,
			"fold_count": 1,
			"primary_metric": "mean_absolute_error_ms",
			"definition": {
				"schema_version": "duration-forecast.v1",
				"method": "shrinkage",
				"method_version": "duration-shrinkage.v3",
				"prior_weight": 8,
				"minimum_history": 5
			},
			"metrics": {
				"nominal": {
					"method": "nominal",
					"evaluated_fold_count": 1,
					"mean_absolute_error_ms": 12.5,
					"interval": {
						"available_fold_count": 1,
						"covered_fold_count": 1,
						"coverage": 1,
						"mean_width_ms": 40
					}
				},
				"shrinkage": {"method": "shrinkage", "evaluated_fold_count": 1}
			},
			"folds": [{
				"fold_number": 1,
				"held_out_run_id": "run-1",
				"observed_at": "2026-08-12T12:00:00Z",
				"training_sample_count": 5,
				"actual_ms": 100,
				"nominal_point_ms": 90,
				"shrinkage_point_ms": 95,
				"nominal_absolute_error_ms": 10,
				"shrinkage_absolute_error_ms": 5,
				"shrinkage_interval": {"p10_ms": 80, "p90_ms": 120}
			}],
			"input_hash": "input-1",
			"as_of": "2026-08-12T12:00:00Z"
		}
	},
	"outcomes": {"eligible": 5, "skipped": 1, "cancelled": 1, "notRun": 1, "censored": 1},
	"versionEvidence": {
		"selected": {
			"uuid": "version-1",
			"versionNumber": 2,
			"status": "candidate",
			"createdAt": "2026-08-12T12:00:00Z"
		},
		"activeVersionUuid": "version-0",
		"history": {"items": [{
			"versionUuid": "version-1",
			"transitionKind": "activate",
			"fromStatus": "draft",
			"toStatus": "candidate",
			"actor": "operator",
			"reason": "evaluated",
			"transitionedAt": "2026-08-12T12:00:00Z"
		}]}
	}
}`

func TestPredictionEvaluateDurationHistoryUsesTypedScopedRoute(t *testing.T) {
	tests := []struct {
		name           string
		definitionUUID string
		wantPath       string
	}{
		{
			name:           "plain definition",
			definitionUUID: "definition-1",
			wantPath: "/api/v1/admin/measurement/predictions/definitions/definition-1/" +
				"duration-evaluations?companySlug=acme+co",
		},
		{
			name:           "definition with a slash",
			definitionUUID: "definition/1",
			wantPath: "/api/v1/admin/measurement/predictions/definitions/definition%2F1/" +
				"duration-evaluations?companySlug=acme+co",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, durationEvaluationResponseBody)
			doer.headers = map[string]string{requestIDHeader: "req-1"}
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			response, metadata, err := client.Admin.Predictions.EvaluateDurationHistory(
				context.Background(), "acme co", test.definitionUUID, &DurationEvaluationRequest{
					PlanUUID:              "plan-1",
					SemanticKey:           "build_step",
					PredictionVersionUUID: "version-1",
					WindowStart:           time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
					AsOf:                  time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC),
					Limit:                 128,
				},
			)
			if err != nil {
				t.Fatalf("EvaluateDurationHistory returned error: %v", err)
			}
			if metadata.RequestID != "req-1" {
				t.Fatalf("request id = %q", metadata.RequestID)
			}
			assertDurationEvaluationResponse(t, response)
			assertDurationEvaluationRequest(t, doer.requests, test.wantPath)
		})
	}
}

// assertDurationEvaluationResponse proves the answer carries the artifact, the
// outcome counts and the version evidence the route returned.
func assertDurationEvaluationResponse(t *testing.T, response *DurationEvaluationResponse) {
	t.Helper()
	if response.DefinitionUUID != "definition-1" || response.PlanUUID != "plan-1" {
		t.Fatalf("response = %+v", response)
	}
	if response.Artifact.Evaluation.TenantID != "acme co" ||
		response.Artifact.Evaluation.Definition.MinimumHistory != 5 ||
		len(response.Artifact.Evaluation.Folds) != 1 ||
		response.Artifact.Evaluation.Folds[0].ShrinkageInterval.P10MS == nil {
		t.Fatalf("artifact = %+v", response.Artifact)
	}
	if response.Outcomes.Censored != 1 || response.Outcomes.NotRun != 1 {
		t.Fatalf("outcomes = %+v", response.Outcomes)
	}
	if response.VersionEvidence.Selected.Status != "candidate" ||
		response.VersionEvidence.ActiveVersionUUID != "version-0" ||
		len(response.VersionEvidence.History.Items) != 1 {
		t.Fatalf("version evidence = %+v", response.VersionEvidence)
	}
}

// assertDurationEvaluationRequest proves the route, the query and the body the
// evaluation was sent with.
func assertDurationEvaluationRequest(t *testing.T, requests []*HTTPRequest, wantPath string) {
	t.Helper()
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	req := requests[0]
	if req.Method != http.MethodPost {
		t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
	}
	if req.URL != "http://localhost:8080"+wantPath {
		t.Fatalf("url = %s, want %s", req.URL, "http://localhost:8080"+wantPath)
	}
	body := decodeAuthProjectBody(t, req)
	if body["planUuid"] != "plan-1" || body["semanticKey"] != "build_step" ||
		body["predictionVersionUuid"] != "version-1" || body["limit"] != float64(128) {
		t.Fatalf("body = %+v", body)
	}
	if body["windowStart"] != "2026-06-01T00:00:00Z" || body["asOf"] != "2026-08-12T12:00:00Z" {
		t.Fatalf("body times = %+v", body)
	}
}
