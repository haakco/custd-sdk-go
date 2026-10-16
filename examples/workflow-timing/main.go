// Command workflow-timing is a collector example for the Go SDK.
//
// It shows the whole safe path a Go consumer needs: declare a workflow shape
// once, report the run and step facts it observed after its own commit, read the
// run back with its projection status, and ask for a completion expectation.
//
// Custd is not the controller. This program reports work it observed; it does not
// start, retry or cancel that work, and it can finish while Custd is unavailable
// by spooling the facts it could not deliver.
//
//	CUSTD_BASE_URL=https://custd.example \
//	CUSTD_TOKEN=<machine credential with measurement.prediction.read/admin> \
//	CUSTD_COMPANY_SLUG=<tenant slug> \
//	go run ./examples/workflow-timing
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	custd "github.com/haakco/custd-sdk-go/v2"
)

// reconcileWorkflow is the declared shape. Every key is provider-neutral: the
// same vocabulary maps a different provider without changing a primitive.
var reconcileWorkflow = custd.WorkflowTimingDeclaration{
	WorkflowKey: "hosting.reconcile",
	Name:        "Hosting reconcile",
	Description: "Generic hosting provisioning phases as observed by a Go collector.",
	Dimensions:  []string{"cluster"},
	Steps: []custd.WorkflowTimingStepDeclaration{
		{StepKey: "validate-plan", Name: "Validate plan", NominalMS: 60000},
		{StepKey: "provision-host", Name: "Provision host", NominalMS: 600000},
		{StepKey: "configure-database", Name: "Configure database", NominalMS: 900000},
		{StepKey: "publish-site", Name: "Publish site", NominalMS: 300000},
	},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "workflow-timing collector: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL := os.Getenv("CUSTD_BASE_URL")
	token := os.Getenv("CUSTD_TOKEN")
	companySlug := os.Getenv("CUSTD_COMPANY_SLUG")
	if baseURL == "" || token == "" || companySlug == "" {
		return errors.New("set CUSTD_BASE_URL, CUSTD_TOKEN and CUSTD_COMPANY_SLUG")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := custd.NewClient(&custd.ClientConfig{BaseURL: baseURL, APIKey: token})
	timings := client.Admin.WorkflowTimings

	reconciled, err := timings.Reconcile(ctx, companySlug, reconcileWorkflow)
	if err != nil {
		return fmt.Errorf("reconcile: %w", err)
	}
	fmt.Printf(
		"declared %s revision=%d created=%t revisionChanged=%t\n",
		reconciled.Definition.WorkflowKey,
		reconciled.Definition.Current.Number,
		reconciled.Created,
		reconciled.RevisionChanged,
	)

	// The collector observed this run; the timestamps come from the source system,
	// not from Custd.
	observedAt := time.Now().UTC().Add(-30 * time.Minute)
	observations := observedRun("reconcile-"+observedAt.Format("20060102T150405Z"), observedAt)

	result, err := timings.AppendBatch(ctx, companySlug, observations)
	if err != nil {
		return fmt.Errorf("append: %w", err)
	}

	// A transport success is not a durable append. requireAccepted names every
	// rejected item with its machine code, and RejectedItems lets a caller retry
	// exactly those instead of the whole batch.
	if err := result.RequireAccepted(); err != nil {
		return err
	}
	fmt.Printf("appended %d facts (%d duplicates)\n", result.Accepted, result.Duplicates)

	runUUID := result.Results[0].RunUUID
	run, err := timings.GetRun(ctx, companySlug, runUUID)
	if err != nil {
		return fmt.Errorf("read run: %w", err)
	}
	fmt.Printf(
		"run %s state=%s attempts=%d pendingFacts=%d nextAction=%s\n",
		run.RunUUID, run.State, len(run.Attempts), run.ProjectionStatus.PendingFacts, run.ProjectionStatus.NextAction,
	)
	if run.ProjectionStatus.Behind() {
		// The supported repair path, and it needs no SQL.
		if err := timings.Rebuild(ctx, companySlug, runUUID); err != nil {
			return fmt.Errorf("rebuild: %w", err)
		}
		fmt.Println("projection was behind; rebuilt from the ledger")
	}

	prediction, err := timings.Prediction(ctx, companySlug, runUUID)
	if err != nil {
		return fmt.Errorf("prediction: %w", err)
	}
	fmt.Printf(
		"expectation for the run: state=%s baselineMs=%d pointMs=%v interval=[%v, %v] sampleCount=%d method=%s\n",
		prediction.Run.State,
		prediction.Run.BaselineMS,
		prediction.Run.ExpectedMS,
		prediction.Run.ConservativeLowMS,
		prediction.Run.ConservativeHighMS,
		prediction.Run.SampleCount,
		prediction.Run.Method,
	)
	if !prediction.Run.Supported() {
		// A non-ready expectation is not a failure: it is the honest answer when
		// there is not enough comparable history yet, and the point and interval
		// are absent so it cannot be mistaken for a forecast.
		fmt.Printf("expectation is not supported yet: %v\n", prediction.Run.Warnings)
	}

	return nil
}

// observedRun builds the facts for one observed run. Identities come from the
// fact itself, so replaying this function after an outage produces duplicates
// rather than a second duration fact.
func observedRun(externalRunID string, startedAt time.Time) []custd.WorkflowTimingObservation {
	keys := custd.WorkflowTimingIdempotency{}
	at := func(offset time.Duration) string {
		return startedAt.Add(offset).Format(time.RFC3339)
	}
	const collector = "go-collector-example"

	observations := []custd.WorkflowTimingObservation{{
		WorkflowKey:    reconcileWorkflow.WorkflowKey,
		ExternalRunID:  externalRunID,
		IdempotencyKey: keys.RunStarted(reconcileWorkflow.WorkflowKey, externalRunID),
		Provenance:     &custd.WorkflowTimingProvenance{ActorKind: "machine", ActorRef: collector},
		Fact: custd.WorkflowTimingFact{
			Kind:       "run_started",
			OccurredAt: at(0),
			Dimensions: map[string]string{"cluster": "fsn1-a"},
			Run:        &custd.WorkflowTimingRunFact{State: "running"},
		},
	}}

	// validate-plan and configure-database and publish-site complete once;
	// provision-host fails its first attempt and succeeds on the second, which is
	// what a retried attempt looks like in the projection.
	phases := []struct {
		stepKey  string
		attempt  int
		start    time.Duration
		finish   time.Duration
		state    string
		errClass string
	}{
		{"validate-plan", 1, 0, 41 * time.Second, "completed", ""},
		{"provision-host", 1, 41 * time.Second, 341 * time.Second, "failed", "timeout"},
		{"provision-host", 2, 345 * time.Second, 372 * time.Second, "completed", ""},
		{"configure-database", 1, 372 * time.Second, 1203 * time.Second, "completed", ""},
		{"publish-site", 1, 1203 * time.Second, 1427 * time.Second, "completed", ""},
	}

	for _, phase := range phases {
		observations = append(observations, custd.WorkflowTimingObservation{
			WorkflowKey:    reconcileWorkflow.WorkflowKey,
			ExternalRunID:  externalRunID,
			IdempotencyKey: keys.StepStarted(reconcileWorkflow.WorkflowKey, externalRunID, phase.stepKey, phase.attempt),
			Provenance:     &custd.WorkflowTimingProvenance{ActorKind: "machine", ActorRef: collector},
			Fact: custd.WorkflowTimingFact{
				Kind:       "step_started",
				OccurredAt: at(phase.start),
				Step: &custd.WorkflowTimingStepFact{
					StepKey: phase.stepKey, Attempt: phase.attempt, State: "running",
				},
			},
		})
		observations = append(observations, custd.WorkflowTimingObservation{
			WorkflowKey:    reconcileWorkflow.WorkflowKey,
			ExternalRunID:  externalRunID,
			IdempotencyKey: keys.StepFinished(reconcileWorkflow.WorkflowKey, externalRunID, phase.stepKey, phase.attempt),
			Provenance:     &custd.WorkflowTimingProvenance{ActorKind: "machine", ActorRef: collector},
			Fact: custd.WorkflowTimingFact{
				Kind:       "step_finished",
				OccurredAt: at(phase.finish),
				Step: &custd.WorkflowTimingStepFact{
					StepKey: phase.stepKey, Attempt: phase.attempt,
					State: phase.state, ErrorClass: phase.errClass,
				},
			},
		})
	}

	return append(observations, custd.WorkflowTimingObservation{
		WorkflowKey:    reconcileWorkflow.WorkflowKey,
		ExternalRunID:  externalRunID,
		IdempotencyKey: keys.RunFinished(reconcileWorkflow.WorkflowKey, externalRunID),
		Provenance:     &custd.WorkflowTimingProvenance{ActorKind: "machine", ActorRef: collector},
		Fact: custd.WorkflowTimingFact{
			Kind:       "run_finished",
			OccurredAt: at(1427 * time.Second),
			Run:        &custd.WorkflowTimingRunFact{State: "completed"},
		},
	})
}
