package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// TestManagedAuditProvisionSendsRequestAndDecodesCredential proves the managed
// audit route, which the server handles as producer provisioning, receives the
// provision request and returns the one-time credential.
func TestManagedAuditProvisionSendsRequestAndDecodesCredential(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated,
		`{"clientId":"custd-acme-audit","clientSecret":"s3cret","companySlug":"acme","producerSlug":"audit","scopes":["events.write"]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	producer, err := client.Admin.ManagedAudit.ProvisionProducer(context.Background(),
		ProducerProvisionRequest{CompanySlug: "acme", ProducerSlug: "audit"})
	if err != nil {
		t.Fatalf("ProvisionProducer: %v", err)
	}
	if producer.ClientID != "custd-acme-audit" || producer.ClientSecret != "s3cret" {
		t.Fatalf("producer = %+v, want the returned credential", producer)
	}
	var sent ProducerProvisionRequest
	if err := json.Unmarshal(doer.requests[0].Body, &sent); err != nil {
		t.Fatalf("request body %q: %v", doer.requests[0].Body, err)
	}
	if sent.CompanySlug != "acme" || sent.ProducerSlug != "audit" {
		t.Fatalf("sent = %+v, want the provision request", sent)
	}
}

// TestImpersonationStopAndTerminateDecodeSessionState proves both session
// actions decode the server's {sessionUuid, state} answer.
func TestImpersonationStopAndTerminateDecodeSessionState(t *testing.T) {
	actions := map[string]func(*ImpersonationAdminClient) (*ImpersonationSessionStateResponse, ResponseMetadata, error){
		"stopped": func(c *ImpersonationAdminClient) (*ImpersonationSessionStateResponse, ResponseMetadata, error) {
			return c.Stop(context.Background(), "s1")
		},
		"terminated": func(c *ImpersonationAdminClient) (*ImpersonationSessionStateResponse, ResponseMetadata, error) {
			return c.Terminate(context.Background(), "s1")
		},
	}
	for state, action := range actions {
		doer := newCaptureDoer(http.StatusOK, `{"sessionUuid":"s1","state":"`+state+`"}`)
		client := newAdminTestClient(t, doer, "http://localhost:8080")
		got, _, err := action(client.Admin.Impersonation)
		if err != nil {
			t.Fatalf("%s: %v", state, err)
		}
		if got.SessionUUID != "s1" || got.State != state {
			t.Fatalf("%s: got %+v", state, got)
		}
	}
}

// TestGetScheduledTaskDecodesTaskAndRuns proves the detail read decodes the
// server's {task, runs} envelope, including run proof, rather than a bare task.
func TestGetScheduledTaskDecodesTaskAndRuns(t *testing.T) {
	run := `{"runUuid":"r1","taskName":"compaction","trigger":"schedule","status":"succeeded",` +
		`"scheduledFor":"2026-11-20T06:00:00Z","durationMs":5,"itemsTotal":1,"itemsCompleted":1,"itemsFailed":0,"jobsCompacted":3}`
	body := `{"task":{"taskName":"compaction","displayName":"Compaction","description":"d","ownerService":"scheduler",` +
		`"enabled":true,"scheduleKind":"interval","scheduleExpression":"5m","timezone":"UTC","overlapPolicy":"skip",` +
		`"lockBackend":"postgres","timeoutSeconds":60,"manualRunsAllowed":true,"forceRunsAllowed":false,` +
		`"queuedRuns":0,"runningRuns":0,"failedRuns":0,"latestRun":` + run + `},"runs":[` + run + `]}`
	client := newAdminTestClient(t, newCaptureDoer(http.StatusOK, body), "http://localhost:8080")
	detail, _, err := client.Admin.System.GetScheduledTask(context.Background(), "compaction")
	if err != nil {
		t.Fatalf("GetScheduledTask: %v", err)
	}
	if detail.Task.TaskName != "compaction" || detail.Task.LatestRun == nil || detail.Task.LatestRun.JobsCompacted != 3 ||
		len(detail.Runs) != 1 || detail.Runs[0].RunUUID != "r1" || detail.Runs[0].JobsCompacted != 3 {
		t.Fatalf("detail = %+v", detail)
	}
}

// TestScheduledTaskRunActionsUnwrapTheRun proves run-now and retry return the
// run the server wraps in a {run} envelope.
func TestScheduledTaskRunActionsUnwrapTheRun(t *testing.T) {
	body := `{"run":{"runUuid":"r1","taskName":"compaction","status":"queued","scheduledFor":"2026-11-20T06:00:00Z"}}`
	actions := map[string]func(*SystemAdminClient) (*ScheduledTaskRun, ResponseMetadata, error){
		"run": func(c *SystemAdminClient) (*ScheduledTaskRun, ResponseMetadata, error) {
			return c.RunScheduledTask(context.Background(), "compaction")
		},
		"retry": func(c *SystemAdminClient) (*ScheduledTaskRun, ResponseMetadata, error) {
			return c.RetryScheduledTaskRun(context.Background(), "r0")
		},
	}
	for name, action := range actions {
		client := newAdminTestClient(t, newCaptureDoer(http.StatusAccepted, body), "http://localhost:8080")
		run, _, err := action(client.Admin.System)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if run.RunUUID != "r1" || run.TaskName != "compaction" || run.Status != "queued" {
			t.Fatalf("%s: run = %+v, want the enveloped run", name, run)
		}
	}
}

func TestScheduledTaskRunForceSendsExplicitRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted,
		`{"run":{"runUuid":"r1","taskName":"compaction","status":"queued","scheduledFor":"2026-11-20T06:00:00Z"}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	run, _, err := client.Admin.System.RunScheduledTaskWithOptions(context.Background(), "compaction",
		&ScheduledTaskRunRequest{Force: true})
	if err != nil || run.RunUUID != "r1" {
		t.Fatalf("forced run = %+v, err = %v", run, err)
	}
	var sent ScheduledTaskRunRequest
	if err := json.Unmarshal(doer.requests[0].Body, &sent); err != nil || !sent.Force {
		t.Fatalf("request = %+v, err = %v", sent, err)
	}
}

func TestRLSPlanCreationSendsIdempotencyKey(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"planUuid":"test-plan"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	request := RLSPlanRequest{}
	if _, err := client.Admin.RLS.CreatePlanWithIdempotencyKey(context.Background(), request, ""); err == nil {
		t.Fatal("plan creation accepted an empty idempotency key")
	}
	if _, err := client.Admin.RLS.CreatePlanWithIdempotencyKey(context.Background(), request, "repeat-key"); err != nil {
		t.Fatal(err)
	}
	if len(doer.requests) != 1 || doer.requests[0].Header("Idempotency-Key") != "repeat-key" {
		t.Fatal("plan creation did not send its idempotency key exactly once")
	}
}
