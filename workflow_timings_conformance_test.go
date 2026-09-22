package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// These tests read the shared fixtures under contract-fixtures/. The PHP façade
// reads the same bytes and asserts the same values, so a field rename on either
// side fails here instead of at a consumer.

func TestWorkflowTimingConformanceDeclaresTheWorkflowShape(t *testing.T) {
	client := newAdminTestClient(t, newCaptureDoer(http.StatusOK, string(
		readContractFixture(t, "workflow-timing-definition-reconcile-response.json"),
	)), "http://localhost:8080/")

	result, err := client.Admin.WorkflowTimings.Reconcile(context.Background(), "acme", WorkflowTimingDeclaration{
		WorkflowKey: "hosting.reconcile",
		Name:        "Hosting reconcile",
		Steps: []WorkflowTimingStepDeclaration{
			{StepKey: "foundation", Name: "Foundation", NominalMS: 600000},
			{StepKey: "database", Name: "Database", NominalMS: 900000},
			{StepKey: "deploy", Name: "Deploy", NominalMS: 300000},
		},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if result.Created || !result.RevisionChanged {
		t.Fatalf("reconcile result = %+v", result)
	}
	if result.Definition.RevisionCount != 2 || result.Definition.Current.Number != 2 {
		t.Fatalf("revision = %+v", result.Definition.Current)
	}
	if len(result.Definition.Current.Steps) != 3 {
		t.Fatalf("steps = %+v", result.Definition.Current.Steps)
	}
	deploy := result.Definition.Current.Steps[2]
	if deploy.StepKey != "deploy" || deploy.Sequence != 3 || deploy.NominalMS != 300000 {
		t.Fatalf("deploy step = %+v", deploy)
	}
	if len(result.Definition.Current.Dimensions) != 1 ||
		result.Definition.Current.Dimensions[0].DimensionKey != "cluster" {
		t.Fatalf("dimensions = %+v", result.Definition.Current.Dimensions)
	}
	if result.Definition.Current.PredictionVersionUUID == "" {
		t.Fatal("the selected duration prediction version was dropped")
	}
}

func TestWorkflowTimingConformanceSendsEveryFactShape(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"results":[],"accepted":0,"rejected":0,"duplicates":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	var batch WorkflowTimingObservationBatch
	if err := json.Unmarshal(readContractFixture(t, "workflow-timing-observation-batch-request.json"), &batch); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if len(batch.Observations) != 4 {
		t.Fatalf("observations = %d", len(batch.Observations))
	}

	started := batch.Observations[0]
	if started.Fact.Kind != "run_started" || started.Fact.Run == nil || started.Fact.Run.State != "running" {
		t.Fatalf("run started fact = %+v", started.Fact)
	}
	if started.Fact.Dimensions["cluster"] != "fsn1-a" {
		t.Fatalf("dimensions = %+v", started.Fact.Dimensions)
	}
	if started.Provenance == nil || started.Provenance.ActorKind != "machine" || started.Provenance.ActorRef == "" {
		t.Fatalf("provenance = %+v", started.Provenance)
	}

	finished := batch.Observations[1]
	if finished.Fact.Step == nil || finished.Fact.Step.StepKey != "foundation" || finished.Fact.Step.Attempt != 1 {
		t.Fatalf("step finished fact = %+v", finished.Fact.Step)
	}

	correction := batch.Observations[3]
	if correction.SupersedesFactID == "" || correction.Fact.Step == nil ||
		correction.Fact.Step.ErrorClass != "timeout" {
		t.Fatalf("correction = %+v", correction)
	}

	// The captured bytes must be the same vocabulary the fixture declares, so a
	// renamed field cannot pass locally and fail at the server.
	if _, err := client.Admin.WorkflowTimings.AppendBatch(context.Background(), "acme", batch.Observations); err != nil {
		t.Fatalf("AppendBatch returned error: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(doer.requests[0].Body, &sent); err != nil {
		t.Fatalf("decode sent body: %v", err)
	}
	observations, ok := sent["observations"].([]any)
	if !ok || len(observations) != 4 {
		t.Fatalf("sent body = %s", string(doer.requests[0].Body))
	}
	first := observations[0].(map[string]any)
	fact := first["fact"].(map[string]any)
	if fact["kind"] != "run_started" || fact["occurredAt"] == nil {
		t.Fatalf("sent fact = %+v", fact)
	}
	provenance := first["provenance"].(map[string]any)
	if provenance["actorKind"] != "machine" || provenance["actorRef"] == nil {
		t.Fatalf("sent provenance = %+v", provenance)
	}
}

func TestWorkflowTimingConformanceFailsClosedOnAPartialBatch(t *testing.T) {
	client := newAdminTestClient(t, newCaptureDoer(http.StatusOK, string(
		readContractFixture(t, "workflow-timing-batch-result-partial.json"),
	)), "http://localhost:8080/")

	result, err := client.Admin.WorkflowTimings.Append(context.Background(), "acme", WorkflowTimingObservation{
		WorkflowKey: "hosting.reconcile", ExternalRunID: "op-20260922-42", IdempotencyKey: "k",
		Fact: WorkflowTimingFact{Kind: "run_started", OccurredAt: "2026-09-22T10:00:00Z"},
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if result.Accepted != 2 || result.Rejected != 1 || result.Duplicates != 1 {
		t.Fatalf("batch result = %+v", result)
	}
	if len(result.Results) != 3 || result.Results[0].WorkflowKey != "hosting.reconcile" {
		t.Fatalf("results = %+v", result.Results)
	}
	if result.Results[1].ErrorCode != "workflow_step_unknown" {
		t.Fatalf("rejected item = %+v", result.Results[1])
	}
	if !result.Results[2].Duplicate {
		t.Fatalf("a redelivered fact was not reported as a duplicate: %+v", result.Results[2])
	}
	if requireErr := result.RequireAccepted(); requireErr == nil {
		t.Fatal("a partially applied batch reported success")
	}
	rejected := result.RejectedItems()
	if len(rejected) != 1 || rejected[0].Index != 1 {
		t.Fatalf("rejected items = %+v", rejected)
	}
}

func TestWorkflowTimingConformanceReadsARunWithPendingFacts(t *testing.T) {
	client := newAdminTestClient(t, newCaptureDoer(http.StatusOK, string(
		readContractFixture(t, "workflow-timing-run-read-response.json"),
	)), "http://localhost:8080/")

	run, err := client.Admin.WorkflowTimings.GetRun(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("GetRun returned error: %v", err)
	}
	if run.State != "running" || run.RevisionNumber != 2 || run.Dimensions["cluster"] != "fsn1-a" {
		t.Fatalf("run = %+v", run)
	}
	if len(run.Attempts) != 2 || run.Attempts[0].StepKey != "foundation" || run.Attempts[0].State != "completed" {
		t.Fatalf("attempts = %+v", run.Attempts)
	}
	if run.Attempts[1].FinishedAt != "" {
		t.Fatalf("an open attempt reported a finish time: %+v", run.Attempts[1])
	}
	if !run.ProjectionStatus.Behind() || run.ProjectionStatus.NextAction != "rebuild" {
		t.Fatalf("projection status = %+v", run.ProjectionStatus)
	}
}

func TestWorkflowTimingConformanceKeepsColdStartDistinctFromReady(t *testing.T) {
	cold := newAdminTestClient(t, newCaptureDoer(http.StatusOK, string(
		readContractFixture(t, "workflow-timing-prediction-cold-start.json"),
	)), "http://localhost:8080/")
	coldStart, err := cold.Admin.WorkflowTimings.Prediction(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("Prediction returned error: %v", err)
	}
	if coldStart.Run.State != WorkflowTimingStateColdStart || coldStart.Run.Supported() {
		t.Fatalf("cold start = %+v", coldStart.Run)
	}
	if coldStart.Run.ExpectedMS != nil || coldStart.Run.ConservativeLowMS != nil ||
		coldStart.Run.ConservativeHighMS != nil {
		t.Fatalf("cold start claimed a point or range: %+v", coldStart.Run)
	}
	if coldStart.Run.BaselineMS != 600000 || coldStart.Run.SampleCount != 0 {
		t.Fatalf("cold start baseline = %+v", coldStart.Run)
	}
	if coldStart.GeneratedAt == "" || coldStart.Run.EvidenceWindowFrom == "" {
		t.Fatalf("prediction provenance is incomplete: %+v", coldStart)
	}

	ready := newAdminTestClient(t, newCaptureDoer(http.StatusOK, string(
		readContractFixture(t, "workflow-timing-prediction-ready.json"),
	)), "http://localhost:8080/")
	forecast, err := ready.Admin.WorkflowTimings.Prediction(context.Background(), "acme", "run-2")
	if err != nil {
		t.Fatalf("Prediction returned error: %v", err)
	}
	if !forecast.Run.Supported() || forecast.Run.ExpectedMS == nil {
		t.Fatalf("ready expectation = %+v", forecast.Run)
	}
	if *forecast.Run.ConservativeLowMS >= *forecast.Run.ConservativeHighMS {
		t.Fatalf("ready interval is not ordered: %+v", forecast.Run)
	}
	if forecast.Run.SampleCount != 24 || forecast.Run.PredictionVersion == "" {
		t.Fatalf("ready evidence = %+v", forecast.Run)
	}
	database, ok := forecast.StepExpectation("database")
	if !ok || database.State != WorkflowTimingStateSparse || database.BaselineMS != 900000 {
		t.Fatalf("database step = %+v", database)
	}
	if len(database.Warnings) != 1 || database.Warnings[0] != "sparse_history" {
		t.Fatalf("sparse step warnings = %+v", database.Warnings)
	}
	if _, ok := forecast.StepExpectation("absent"); ok {
		t.Fatal("an undeclared step resolved to an expectation")
	}
}

func TestWorkflowTimingConformanceSeparatesActiveFromWait(t *testing.T) {
	client := newAdminTestClient(t, newCaptureDoer(http.StatusOK, string(
		readContractFixture(t, "workflow-timing-duration-history-response.json"),
	)), "http://localhost:8080/")

	history, err := client.Admin.WorkflowTimings.DurationHistory(context.Background(), "acme", "hosting.reconcile", 50)
	if err != nil {
		t.Fatalf("DurationHistory returned error: %v", err)
	}
	if len(history.Entries) != 2 || !history.Entries[1].Corrected {
		t.Fatalf("entries = %+v", history.Entries)
	}
	if history.Entries[0].BaselineMS != 600000 || history.Entries[0].RevisionNumber != 2 {
		t.Fatalf("entry provenance = %+v", history.Entries[0])
	}
	if history.Outcomes.CompletedRuns != 2 || history.Outcomes.FailedRuns != 1 || history.Outcomes.OpenAttempts != 1 {
		t.Fatalf("outcomes = %+v", history.Outcomes)
	}
	if len(history.Contributions) != 2 || history.Contributions[0].ContributionPM != 662 {
		t.Fatalf("contributions = %+v", history.Contributions)
	}
	if history.LatestCompletedRun == nil {
		t.Fatal("the latest completed run is missing")
	}
	latest := history.LatestCompletedRun
	if latest.WallClockMS != 95000 || latest.ActiveMS != 85400 || latest.WaitMS != 9600 {
		t.Fatalf("latest completed run = %+v", latest)
	}
	if latest.ActiveMS+latest.WaitMS != latest.WallClockMS {
		t.Fatalf("active and wait do not account for the wall clock: %+v", latest)
	}
	if latest.NominalMS != 1800000 {
		t.Fatalf("nominal baseline = %d", latest.NominalMS)
	}
}
