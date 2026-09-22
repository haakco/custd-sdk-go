package custd

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestWorkflowTimingsReconcileSendsTheDeclarativeShape(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"definition":{"uuid":"wf-1","workflowKey":"hosting.reconcile","name":"Hosting reconcile","status":"active","revisionCount":1,"currentRevision":{"uuid":"rev-1","number":1,"hash":"`+strings.Repeat("a", 64)+`","steps":[{"stepKey":"foundation","name":"Foundation","nominalMs":600000}]}},"created":true,"revisionChanged":true}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	result, err := client.Admin.WorkflowTimings.Reconcile(context.Background(), "acme", WorkflowTimingDeclaration{
		WorkflowKey: "hosting.reconcile",
		Name:        "Hosting reconcile",
		Steps: []WorkflowTimingStepDeclaration{
			{StepKey: "foundation", Name: "Foundation", NominalMS: 600000},
		},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if !result.Created || !result.RevisionChanged || result.Definition.Current.Number != 1 {
		t.Fatalf("reconcile result = %+v", result)
	}
	if len(result.Definition.Current.Steps) != 1 || result.Definition.Current.Steps[0].NominalMS != 600000 {
		t.Fatalf("steps = %+v", result.Definition.Current.Steps)
	}
	assertSiteRequests(t, doer.requests, []string{
		"PUT http://localhost:8080/api/v1/admin/workflow-timings/definitions/hosting.reconcile?companySlug=acme",
	})
}

func TestWorkflowTimingsBatchResultFailsClosed(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"results":[{"index":0,"accepted":true,"duplicate":false,"idempotencyKey":"k0","runUuid":"run-1"},{"index":1,"accepted":false,"duplicate":false,"idempotencyKey":"k1","errorCode":"workflow_step_unknown"}],"accepted":1,"rejected":1,"duplicates":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	result, err := client.Admin.WorkflowTimings.AppendBatch(context.Background(), "acme", []WorkflowTimingObservation{
		{
			WorkflowKey: "hosting.reconcile", ExternalRunID: "op-1", IdempotencyKey: "k0",
			Fact: WorkflowTimingFact{Kind: "run_started", OccurredAt: "2026-09-22T10:00:00Z", Run: &WorkflowTimingRunFact{State: "running"}},
		},
		{
			WorkflowKey: "hosting.reconcile", ExternalRunID: "op-1", IdempotencyKey: "k1",
			Fact: WorkflowTimingFact{
				Kind: "step_started", OccurredAt: "2026-09-22T10:00:01Z",
				Step: &WorkflowTimingStepFact{StepKey: "nope", Attempt: 1, State: "running"},
			},
		},
	})
	if err != nil {
		t.Fatalf("AppendBatch returned error: %v", err)
	}
	if result.Accepted != 1 || result.Rejected != 1 {
		t.Fatalf("batch result = %+v", result)
	}
	requireErr := result.RequireAccepted()
	if requireErr == nil {
		t.Fatal("a partially applied batch reported success")
	}
	for _, want := range []string{"1 of 2", "k1", "workflow_step_unknown"} {
		if !strings.Contains(requireErr.Error(), want) {
			t.Fatalf("error %q does not name %q", requireErr.Error(), want)
		}
	}
	if _, err := client.Admin.WorkflowTimings.AppendBatch(context.Background(), "acme", nil); err == nil {
		t.Fatal("an empty batch was accepted")
	}
	if _, err := client.Admin.WorkflowTimings.Correct(context.Background(), "acme", WorkflowTimingObservation{}); err == nil {
		t.Fatal("a correction without a superseded fact was accepted")
	}
}

func TestWorkflowTimingsPredictionKeepsColdStartDistinct(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"runUuid":"run-1","workflowKey":"hosting.reconcile","externalRunId":"op-1","state":"running","allowOverlap":false,"run":{"seriesKey":"workflow_run","state":"cold_start","baselineMs":600000,"sampleCount":0,"method":"nominal","methodVersion":"nominal.v1","inputHash":"`+strings.Repeat("b", 64)+`","warnings":["cold_start"]},"steps":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	prediction, err := client.Admin.WorkflowTimings.Prediction(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("Prediction returned error: %v", err)
	}
	if prediction.Run.State != "cold_start" {
		t.Fatalf("run state = %q", prediction.Run.State)
	}
	if prediction.Run.ConservativeLowMS != nil || prediction.Run.ConservativeHighMS != nil {
		t.Fatalf("cold start claimed a range: %+v", prediction.Run)
	}
	if prediction.Run.BaselineMS != 600000 || prediction.Run.SampleCount != 0 {
		t.Fatalf("cold start = %+v", prediction.Run)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET http://localhost:8080/api/v1/admin/workflow-timings/runs/run-1/prediction?companySlug=acme",
	})
}

func TestWorkflowTimingIdempotencyIsDeterministicPerFact(t *testing.T) {
	keys := WorkflowTimingIdempotency{}
	if got := keys.RunStarted("hosting.reconcile", "op-1"); got != "workflow-timing:hosting.reconcile:op-1:run_started" {
		t.Fatalf("run started key = %q", got)
	}
	if got := keys.StepStarted("hosting.reconcile", "op-1", "foundation", 2); got !=
		"workflow-timing:hosting.reconcile:op-1:step_started:foundation:2" {
		t.Fatalf("step started key = %q", got)
	}
	if keys.StepStarted("hosting.reconcile", "op-1", "foundation", 1) ==
		keys.StepFinished("hosting.reconcile", "op-1", "foundation", 1) {
		t.Fatal("start and finish share an idempotency identity")
	}
	if keys.RunStarted("hosting.reconcile", "op-1") == keys.RunStarted("hosting.reconcile", "op-2") {
		t.Fatal("two runs share an idempotency identity")
	}
}
