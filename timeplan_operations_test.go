package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// timePlanRequestCase is one table-driven expectation about the request an SDK
// method sends: its method, its URL, and its body.
type timePlanRequestCase struct {
	name          string
	call          func() error
	wantMethod    string
	wantURL       string
	wantBody      string
	wantJSONKeys  []string
	wantEmptyBody bool
}

// assertTimePlanRequestCase runs one case and checks the captured request.
func assertTimePlanRequestCase(t *testing.T, doer *captureDoer, test timePlanRequestCase) {
	t.Helper()
	before := len(doer.requests)
	if err := test.call(); err != nil {
		t.Fatalf("call error: %v", err)
	}
	if len(doer.requests) != before+1 {
		t.Fatalf("requests = %d, want one request", len(doer.requests)-before)
	}
	request := doer.requests[len(doer.requests)-1]
	if request.Method != test.wantMethod {
		t.Fatalf("method = %s, want %s", request.Method, test.wantMethod)
	}
	if request.URL != test.wantURL {
		t.Fatalf("url = %s, want %s", request.URL, test.wantURL)
	}
	assertTimePlanRequestBody(t, request, test)
}

// assertTimePlanRequestBody checks the captured body against the case.
func assertTimePlanRequestBody(t *testing.T, request *HTTPRequest, test timePlanRequestCase) {
	t.Helper()
	if test.wantEmptyBody {
		if len(request.Body) != 0 {
			t.Fatalf("body = %s, want no body", request.Body)
		}
		return
	}
	if test.wantBody != "" && string(request.Body) != test.wantBody {
		t.Fatalf("body = %s, want %s", request.Body, test.wantBody)
	}
	for _, key := range test.wantJSONKeys {
		var payload map[string]any
		if err := json.Unmarshal(request.Body, &payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if _, exists := payload[key]; !exists {
			t.Fatalf("request body is missing %q: %s", key, request.Body)
		}
	}
}

func TestTimePlanOperationReadsUseTenantScopedRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	tests := []timePlanRequestCase{
		{
			name: "operational status",
			call: func() error {
				_, _, err := client.Admin.TimePlans.GetTimePlanOperationalStatus(context.Background(), "acme")
				return err
			},
			wantMethod:    http.MethodGet,
			wantURL:       "http://localhost:8080/api/v1/admin/time-plans/operations/status?companySlug=acme",
			wantEmptyBody: true,
		},
		{
			name: "duration observations",
			call: func() error {
				_, err := client.Admin.TimePlans.ListTimePlanDurationObservations(context.Background(), "acme", "run-1")
				return err
			},
			wantMethod:    http.MethodGet,
			wantURL:       "http://localhost:8080/api/v1/admin/time-plans/runs/run-1/duration-observations?companySlug=acme",
			wantEmptyBody: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertTimePlanRequestCase(t, doer, test)
		})
	}
}

func TestTimePlanOperationMutationsUseTenantScopedRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	tests := []timePlanRequestCase{
		{
			name: "repair run",
			call: func() error {
				_, _, err := client.Admin.TimePlans.RepairTimePlanRun(context.Background(), "acme", "run-1")
				return err
			},
			wantMethod:    http.MethodPost,
			wantURL:       "http://localhost:8080/api/v1/admin/time-plans/runs/run-1/repair?companySlug=acme",
			wantEmptyBody: true,
		},
		{
			name: "delete run projection",
			call: func() error {
				_, err := client.Admin.TimePlans.DeleteTimePlanRunProjection(context.Background(), "acme", "run-1")
				return err
			},
			wantMethod:    http.MethodPost,
			wantURL:       "http://localhost:8080/api/v1/admin/time-plans/runs/run-1/projection-loss?companySlug=acme",
			wantEmptyBody: true,
		},
		{
			name: "replay outbox",
			call: func() error {
				_, _, err := client.Admin.TimePlans.ReplayTimePlanOutbox(
					context.Background(), "acme", TimePlanOutboxReplayRequest{Limit: 25},
				)
				return err
			},
			wantMethod: http.MethodPost,
			wantURL:    "http://localhost:8080/api/v1/admin/time-plans/operations/outbox/replay?companySlug=acme",
			wantBody:   `{"limit":25}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertTimePlanRequestCase(t, doer, test)
		})
	}
}

func TestTimePlanOperationalStatusDecodesRedactedCounts(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"pendingOutbox":2,"failedOutbox":1,"publishedOutbox":9,"activeRuns":3,"staleRuns":0,"checkedAt":"2026-09-02T12:00:00Z"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	status, metadata, err := client.Admin.TimePlans.GetTimePlanOperationalStatus(context.Background(), "acme")
	if err != nil {
		t.Fatalf("GetTimePlanOperationalStatus error: %v", err)
	}
	if status.PendingOutbox != 2 || status.FailedOutbox != 1 || status.PublishedOutbox != 9 || status.ActiveRuns != 3 {
		t.Fatalf("status = %+v", status)
	}
	if status.CheckedAt.IsZero() {
		t.Fatalf("checkedAt was not decoded: %+v", status)
	}
	if metadata.RequestID != "" {
		t.Fatalf("request id = %q, want empty when the server sends none", metadata.RequestID)
	}
}

func TestTimePlanDurationObservationsDecodeNamedIdentities(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"runUuid":"run-1","nonDeletedCount":1,"observations":[{"observationUuid":"observation-1","idempotencyKey":"key-1","runUuid":"run-1","blockUuid":"block-1","planUuid":"plan-1","semanticKey":"focus"}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	observations, err := client.Admin.TimePlans.ListTimePlanDurationObservations(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("ListTimePlanDurationObservations error: %v", err)
	}
	if observations.RunUUID != "run-1" || observations.NonDeletedCount != 1 {
		t.Fatalf("observations = %+v", observations)
	}
	if len(observations.Observations) != 1 || observations.Observations[0].ObservationUUID != "observation-1" {
		t.Fatalf("observations = %+v", observations.Observations)
	}
}

func TestTimePlanRepairReturnsRebuiltProjection(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"run":{"uuid":"run-1","planUuid":"plan-1","status":"completed","streamVersion":4,"baselineHorizonMs":60000,"lostMs":0,"unusedMs":0,"overrunMs":0,"blocks":[]},"repairedAt":"2026-09-02T12:00:00Z"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	repair, _, err := client.Admin.TimePlans.RepairTimePlanRun(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("RepairTimePlanRun error: %v", err)
	}
	if repair.Run.UUID != "run-1" || repair.Run.StreamVersion != 4 || repair.Run.BaselineHorizonMS != 60000 {
		t.Fatalf("repair = %+v", repair)
	}
	if repair.RepairedAt.IsZero() {
		t.Fatalf("repairedAt was not decoded: %+v", repair)
	}
}

func TestTimePlanProjectionLossAndOutboxReplayDecodeCounters(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"runUuid":"run-1","deletedRows":3,"lostAt":"2026-09-02T12:00:00Z"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	loss, err := client.Admin.TimePlans.DeleteTimePlanRunProjection(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("DeleteTimePlanRunProjection error: %v", err)
	}
	if loss.RunUUID != "run-1" || loss.DeletedRows != 3 {
		t.Fatalf("loss = %+v", loss)
	}

	doer.body = `{"requeued":4,"requested":5,"replayedAt":"2026-09-02T12:05:00Z"}`
	replay, _, err := client.Admin.TimePlans.ReplayTimePlanOutbox(
		context.Background(), "acme", TimePlanOutboxReplayRequest{Limit: 5},
	)
	if err != nil {
		t.Fatalf("ReplayTimePlanOutbox error: %v", err)
	}
	if replay.Requeued != 4 || replay.Requested != 5 {
		t.Fatalf("replay = %+v", replay)
	}
}

func TestTimePlanDefinitionMathProofDecodesVectors(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"schemaVersion":"time-plan.definition-math-proof.v1","companySlug":"acme","source":"admin-api.timeplans",`+
			`"observedAt":"2026-09-02T12:00:00Z","conformanceSchemaVersion":"time-plan.conformance.v1",`+
			`"allocatorVersion":"largest-remainder.v1","definitions":[{"vectorId":"v1","index":0,"horizonMs":1000,`+
			`"blockCount":1,"allocatedMs":1000,"exactConservation":true,"runtimeVocabularyClean":true,"passed":true}],`+
			`"allocations":[{"vectorId":"v1","index":0,"horizonMs":1000,"allocations":[{"blockId":"b1","sequence":0,"durationMs":1000}],`+
			`"exactConservation":true,"deterministicTieBreak":true,"negativeAllocations":0,"passed":true}],`+
			`"adjustments":[{"vectorId":"v2","index":1,"reason":"mark_future_not_run","horizonMs":1000,"elapsedMs":100,`+
			`"actualMs":100,"remainingMs":900,"unusedMs":0,"overrunMs":0,"markFutureNotRun":true,"allocations":[],`+
			`"exactConservation":true,"negativeAllocations":0,"passed":true}],`+
			`"summary":{"initialVectorCount":1,"adjustmentVectorCount":1,"vectorCount":2,"passed":true,`+
			`"exactConservation":true,"deterministicTieBreak":true,"negativeAllocations":0}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	proof, err := client.Admin.TimePlans.GetTimePlanDefinitionMathProof(context.Background(), "acme")
	if err != nil {
		t.Fatalf("GetTimePlanDefinitionMathProof error: %v", err)
	}
	if proof.SchemaVersion != TimePlanDefinitionMathProofSchemaVersion || proof.Source != TimePlanDefinitionMathProofSource {
		t.Fatalf("proof = %+v", proof)
	}
	if len(proof.Definitions) != 1 || proof.Definitions[0].VectorID != "v1" || !proof.Definitions[0].Passed {
		t.Fatalf("definitions = %+v", proof.Definitions)
	}
	if len(proof.Allocations) != 1 || len(proof.Allocations[0].Allocations) != 1 {
		t.Fatalf("allocations = %+v", proof.Allocations)
	}
	if len(proof.Adjustments) != 1 || proof.Adjustments[0].Reason != "mark_future_not_run" {
		t.Fatalf("adjustments = %+v", proof.Adjustments)
	}
	if proof.Adjustments[0].HorizonMS != 1000 || proof.Adjustments[0].ElapsedMS != 100 {
		t.Fatalf("adjustment timing = %+v", proof.Adjustments[0])
	}
	if !proof.Summary.Passed || proof.Summary.VectorCount != 2 {
		t.Fatalf("summary = %+v", proof.Summary)
	}
}

func TestTimePlanDefinitionMathProofUsesTheServerOwnedRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, err := client.Admin.TimePlans.GetTimePlanDefinitionMathProof(context.Background(), "acme"); err != nil {
		t.Fatalf("GetTimePlanDefinitionMathProof error: %v", err)
	}
	want := "http://localhost:8080/api/v1/admin/time-plans/operations/definition-math?companySlug=acme"
	if got := doer.requests[0].URL; got != want {
		t.Fatalf("url = %s, want %s", got, want)
	}
	if len(doer.requests[0].Body) != 0 {
		t.Fatalf("body = %s, want no body", doer.requests[0].Body)
	}
}
