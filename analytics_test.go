package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// CountEvents is tenant-facing, so the assertion that matters is that it reaches the
// analytics route under /api/v1 and carries no company slug: the tenant is the credential.
func TestAnalyticsCountEventsUsesTheTenantAnalyticsRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"count":42}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	count, err := client.Admin.CountEvents(context.Background(), AnalyticsEventQuery{
		Date: "2026-09-13", EventType: "order.created", Source: "auto",
		LabelFilters: []AnalyticsEventLabelFilter{{Key: "environment", Value: "production"}},
	})
	if err != nil {
		t.Fatalf("CountEvents: %v", err)
	}
	if count.Count != 42 {
		t.Fatalf("count = %d, want 42", count.Count)
	}

	request := doer.requests[0]
	if request.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", request.Method)
	}
	if request.URL != "http://localhost:8080/api/v1/analytics/query" {
		t.Fatalf("url = %s, want the analytics route under /api/v1", request.URL)
	}

	var sent AnalyticsEventQuery
	if err := json.Unmarshal(request.Body, &sent); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if sent.Date != "2026-09-13" || sent.EventType != "order.created" || sent.Source != "auto" {
		t.Fatalf("body = %+v", sent)
	}
	if len(sent.LabelFilters) != 1 || sent.LabelFilters[0].Key != "environment" {
		t.Fatalf("filters = %+v", sent.LabelFilters)
	}
}

// A query the server would refuse is refused here, so the failure names the field rather than
// arriving as a service 400.
func TestAnalyticsCountEventsRejectsAnIncompleteQuery(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"count":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, err := client.Admin.CountEvents(context.Background(), AnalyticsEventQuery{Source: "auto"}); err == nil {
		t.Fatal("expected an error for a missing date")
	}
	if _, err := client.Admin.CountEvents(context.Background(), AnalyticsEventQuery{Date: "2026-09-13"}); err == nil {
		t.Fatal("expected an error for a missing source")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("a refused query still made %d requests", len(doer.requests))
	}
}
