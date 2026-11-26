package custd

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// TestMachineAccessBlockedRefusalIsTyped proves a CI job's log can name the setting that refused it:
// the area, which setting blocked it, the company and the machine, with the server's request ID.
func TestMachineAccessBlockedRefusalIsTyped(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusForbidden, string(readContractFixture(t, "machine-access-blocked-problem.json")))

	_, metadata, err := client.Admin.MachineAccess.GetEffective(context.Background(), "ci-reporting")

	var requestErr *RequestError
	if !errors.As(err, &requestErr) || requestErr.Problem == nil {
		t.Fatalf("error = %#v, want RequestError with problem", err)
	}
	problem := requestErr.Problem
	want := MachineAccessBlock{
		Area: MachineAccessReporting, BlockedBy: MachineAccessDecidedByMachine, CompanySlug: "acme", ClientID: "ci-reporting",
	}
	if problem.Code != ProblemCodeMachineAccessBlocked || problem.MachineAccess == nil || *problem.MachineAccess != want {
		t.Fatalf("problem = %#v, machineAccess = %#v", problem, problem.MachineAccess)
	}
	if problem.TraceID != "req-machine-access-1" || metadata.RequestID != "req-123" {
		t.Fatalf("traceId = %q, requestId = %q", problem.TraceID, metadata.RequestID)
	}
}

// TestMachineAccessUnavailableIsBoundedRetry proves a policy read failure is a retryable refusal
// with no block, so a caller retries instead of changing a setting.
func TestMachineAccessUnavailableIsBoundedRetry(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusServiceUnavailable,
		string(readContractFixture(t, "machine-access-unavailable-problem.json")))

	_, _, err := client.Admin.MachineAccess.Get(context.Background(), "acme")

	var requestErr *RequestError
	if !errors.As(err, &requestErr) || !requestErr.Unavailable() || requestErr.Problem == nil {
		t.Fatalf("error = %#v, want unavailable RequestError", err)
	}
	if requestErr.Problem.Code != ProblemCodeMachineAccessUnavailable || requestErr.Problem.MachineAccess != nil ||
		requestErr.Problem.Retryability != RetryabilityBounded {
		t.Fatalf("problem = %#v", requestErr.Problem)
	}
}
