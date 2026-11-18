package custd

import (
	"net/http"
	"testing"
)

func TestSystemJobsStatusAndRetry(t *testing.T) {
	const baseURL = "https://example.invalid"
	doer := &lifecycleDoer{
		status: http.StatusOK,
		body: `{
			"queueCounts":[{"queue":"notifications","kind":"mail","state":"available","count":2,"oldestWaitingAt":"2026-11-18T09:30:00Z"}],
			"recentFailures":[{"jobId":17,"queue":"notifications","kind":"mail","attempts":3,"maxAttempts":3,"failedAt":"2026-11-18T09:00:00Z"}],
			"outboxes":[{"name":"app_items","pending":1,"failed":0}],"realtimeEvents":12
		}`,
		headers: map[string]string{requestIDHeader: "status-request"},
	}
	client := newLifecycleTestClient(t, doer, baseURL).Admin.System
	status, metadata, err := client.ListJobs(t.Context())
	if err != nil || metadata.RequestID != "status-request" {
		t.Fatalf("list jobs metadata = %+v, error = %v", metadata, err)
	}
	if len(status.QueueCounts) != 1 || status.QueueCounts[0].OldestWaitingAt == nil ||
		len(status.RecentFailures) != 1 || len(status.Outboxes) != 1 || status.RealtimeEvents != 12 {
		t.Fatalf("job status = %+v", status)
	}
	if request := doer.requests[0]; request.Method != http.MethodGet || request.URL != baseURL+"/api/v1/admin/jobs" {
		t.Fatalf("list jobs request = %s %s", request.Method, request.URL)
	}

	doer.status, doer.body = http.StatusAccepted, `{"jobId":17}`
	doer.headers[requestIDHeader] = "retry-request"
	receipt, metadata, err := client.RetryJob(t.Context(), status.RecentFailures[0].JobID)
	if err != nil || receipt.JobID != 17 || metadata.RequestID != "retry-request" {
		t.Fatalf("retry receipt = %+v, metadata = %+v, error = %v", receipt, metadata, err)
	}
	if request := doer.requests[1]; request.Method != http.MethodPost || request.URL != baseURL+"/api/v1/admin/jobs/17/retry" || len(request.Body) != 0 {
		t.Fatalf("retry job request = %s %s, body length %d", request.Method, request.URL, len(request.Body))
	}
}
