package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type workflowBatchDoer struct {
	calls int
	seen  int
}

func (d *workflowBatchDoer) Do(request *HTTPRequest) (*HTTPResponse, error) {
	if len(request.Body) > 1<<20 {
		return nil, fmt.Errorf("workflow batch exceeds the server byte limit: %d", len(request.Body))
	}
	var batch WorkflowTimingObservationBatch
	if err := json.Unmarshal(request.Body, &batch); err != nil {
		return nil, err
	}
	result := WorkflowTimingBatchResult{}
	for index, observation := range batch.Observations {
		accepted := observation.IdempotencyKey != "reject"
		result.Results = append(result.Results, WorkflowTimingObservationResult{
			Index: index, Accepted: accepted, IdempotencyKey: observation.IdempotencyKey,
		})
		if accepted {
			result.Accepted++
		} else {
			result.Rejected++
		}
	}
	d.calls++
	d.seen += len(batch.Observations)
	body, err := json.Marshal(result)
	return &HTTPResponse{StatusCode: 200, Body: body}, err
}

func TestWorkflowTimingBatchSplitsByBytesAndKeepsGlobalResults(t *testing.T) {
	doer := &workflowBatchDoer{}
	client := NewClient(&ClientConfig{BaseURL: "http://localhost:8080", APIKey: "test", HTTPClient: doer})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	observations := make([]WorkflowTimingObservation, 10001)
	for index := range observations {
		observations[index] = WorkflowTimingObservation{
			WorkflowKey: "workflow", ExternalRunID: strings.Repeat("r", 200), IdempotencyKey: fmt.Sprint(index),
		}
	}
	observations[len(observations)-1].IdempotencyKey = "reject"
	result, err := client.Admin.WorkflowTimings.AppendBatch(context.Background(), "acme", observations)
	if err != nil {
		t.Fatal(err)
	}
	if doer.calls < 2 || doer.seen != len(observations) || result.Accepted != 10000 || result.Rejected != 1 {
		t.Fatalf("calls=%d seen=%d accepted=%d rejected=%d", doer.calls, doer.seen, result.Accepted, result.Rejected)
	}
	for index, item := range result.Results {
		if item.Index != index || item.IdempotencyKey != observations[index].IdempotencyKey {
			t.Fatalf("result %d lost its input identity: %+v", index, item)
		}
	}
	if result.RequireAccepted() == nil {
		t.Fatal("a rejection in a later chunk reported success")
	}
}

func TestWorkflowTimingBatchRejectsMissingResults(t *testing.T) {
	doer := newCaptureDoer(200, `{"results":[],"accepted":0,"rejected":0,"duplicates":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.Admin.WorkflowTimings.AppendBatch(context.Background(), "acme", []WorkflowTimingObservation{{IdempotencyKey: "key"}})
	if err == nil || !strings.Contains(err.Error(), "does not account for every input") {
		t.Fatalf("missing results reported success: %v", err)
	}
}

func TestWorkflowTimingBatchRejectsOversizedFactBeforeSending(t *testing.T) {
	doer := newCaptureDoer(200, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.Admin.WorkflowTimings.AppendBatch(context.Background(), "acme", []WorkflowTimingObservation{
		{IdempotencyKey: "small"}, {IdempotencyKey: strings.Repeat("x", 1<<20)},
	})
	if err == nil || len(doer.requests) != 0 {
		t.Fatalf("oversized item: err=%v requests=%d", err, len(doer.requests))
	}
}
