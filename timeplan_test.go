package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTimePlanCorrectionRequestUsesTransitionUUID(t *testing.T) {
	request := TimePlanCommandRequest{Type: "append_correction", SupersedesTransitionUUID: "transition-1"}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal correction request: %v", err)
	}
	if string(payload) != `{"commandId":"","idempotencyKey":"","expectedVersion":0,"type":"append_correction","supersedesTransitionUuid":"transition-1"}` {
		t.Fatalf("correction payload = %s", payload)
	}
}

func TestTimePlanAdminLifecycleUsesTenantScopedRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"plans":[{"uuid":"plan-1","planKey":"focus","name":"Focus","status":"ready","draftRevision":1,"definition":{"horizonMs":60000,"blocks":[]}}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")
	plans, err := client.Admin.TimePlans.List(context.Background(), "acme", 25)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(plans.Plans) != 1 || plans.Plans[0].UUID != "plan-1" {
		t.Fatalf("plans = %+v", plans)
	}
	if got := doer.requests[0].URL; got != "http://localhost:8080/api/v1/admin/time-plans?companySlug=acme&limit=25" {
		t.Fatalf("List URL = %s", got)
	}

	doer.status = http.StatusCreated
	doer.body = `{"uuid":"plan-1","planKey":"focus","name":"Focus","status":"draft","draftRevision":1,"definition":{"horizonMs":60000,"blocks":[]}}`
	created, err := client.Admin.TimePlans.Create(context.Background(), "acme", TimePlanDraftRequest{
		PlanKey: "focus", Name: "Focus", Definition: TimePlanDefinition{HorizonMS: 60000, Blocks: []TimePlanBlock{}},
	})
	if err != nil || created.UUID != "plan-1" {
		t.Fatalf("Create result=%+v err=%v", created, err)
	}
	if got := doer.requests[1].URL; got != "http://localhost:8080/api/v1/admin/time-plans?companySlug=acme" {
		t.Fatalf("Create URL = %s", got)
	}
}

func TestTimePlanAdminRunHistoryAndAnnotations(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"transitions":[{"uuid":"transition-1","runUuid":"run-1","streamVersion":1,"commandId":"command-1","type":"start_run","actorKind":"human","actorRef":"user-1","serverReceivedAt":"2026-09-02T12:00:00Z","currentStatus":"running","allocatorVersion":"largest-remainder.v1","schemaVersion":"time-plan.transition.v1","receipt":{"allocatorVersion":"largest-remainder.v1","reason":"","summary":"","source":[],"result":[],"changes":[]}}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	history, err := client.Admin.TimePlans.History(context.Background(), "acme", "run-1", 10)
	if err != nil || len(history.Transitions) != 1 || history.Transitions[0].Type != "start_run" {
		t.Fatalf("History result=%+v err=%v", history, err)
	}
	if got := doer.requests[0].URL; got != "http://localhost:8080/api/v1/admin/time-plans/runs/run-1/history?companySlug=acme&limit=10" {
		t.Fatalf("History URL = %s", got)
	}

	doer.body = `{"uuid":"annotation-1","runUuid":"run-1","type":"note","text":"hello","recordedAt":"2026-09-02T12:00:00Z","actorKind":"human","actorRef":"user-1"}`
	annotation, err := client.Admin.TimePlans.CreateAnnotation(context.Background(), "acme", "run-1", TimePlanAnnotationInput{Type: "note", Text: "hello"})
	if err != nil || annotation.UUID != "annotation-1" {
		t.Fatalf("CreateAnnotation result=%+v err=%v", annotation, err)
	}
	if got := doer.requests[1].URL; got != "http://localhost:8080/api/v1/admin/time-plans/runs/run-1/annotations?companySlug=acme" {
		t.Fatalf("CreateAnnotation URL = %s", got)
	}
	doer.status = http.StatusNoContent
	doer.body = ""
	if err := client.Admin.TimePlans.RedactAnnotation(
		context.Background(), "acme", "run-1", "annotation-1", TimePlanRedactionRequest{Reason: "privacy request"},
	); err != nil {
		t.Fatalf("RedactAnnotation error: %v", err)
	}
	if got := string(doer.requests[2].Body); got != `{"reason":"privacy request"}` {
		t.Fatalf("RedactAnnotation body = %s", got)
	}
}

func TestTimePlanDefinitionValidationPreventsTransport(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	remainingMS := int64(5_000)
	remainingFractionBPS := int64(100)
	_, err := client.Admin.TimePlans.Preview(context.Background(), "acme", TimePlanDefinition{
		HorizonMS: 60_000,
		ThresholdCues: []TimePlanThresholdCue{{
			RemainingMS: &remainingMS, RemainingFractionBPS: &remainingFractionBPS, Severity: TimePlanThresholdWarning,
		}},
	})
	if err == nil {
		t.Fatal("Preview accepted a threshold cue with both triggers")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("invalid definition sent %d requests", len(doer.requests))
	}
}
