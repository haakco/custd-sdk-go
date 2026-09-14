package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const analyticsQueryResponseFixture = `{
  "results": [{"eventTypeSlug": "page_view", "payload": {"path": "/"}}],
  "count": 1,
  "sources": [{"name": "postgres", "count": 1, "complete": true, "fresh": true, "queryDurationMs": 4, "freshnessLagMs": 12}],
  "timing": {"eventLagP50Ms": 10, "eventLagP95Ms": 20, "eventLagMaxMs": 30, "queryDurationMs": 4, "snapshotAgeMs": 100}
}`

func TestAnalyticsQueryUsesTheTenantRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, analyticsQueryResponseFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	response, err := client.Analytics.Query(context.Background(), AnalyticsEventQueryRequest{
		Date:      "2026-09-14",
		EventType: "page_view",
		Limit:     50,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The tenant route, not the admin one: the credential already names the tenant.
	if len(doer.requests) != 1 || !strings.HasSuffix(doer.requests[0].URL, "/api/v1/analytics/query") {
		t.Fatalf("request URL = %s", doer.requests[0].URL)
	}
	// The server's own completeness and freshness assessment is surfaced verbatim; the
	// SDK must not substitute a confidence judgement of its own.
	if len(response.Sources) != 1 || !response.Sources[0].Complete || !response.Sources[0].Fresh {
		t.Fatalf("sources = %+v", response.Sources)
	}
	if response.Timing.EventLagP95Ms != 20 || response.Count != 1 {
		t.Fatalf("timing/count = %+v/%d", response.Timing, response.Count)
	}
}

func TestAnalyticsQueryRejectsInvalidRequestsBeforeSending(t *testing.T) {
	tooMany := make([]AnalyticsLabelFilter, MaxAnalyticsLabelFilters+1)
	for i := range tooMany {
		tooMany[i] = AnalyticsLabelFilter{Key: "k", Value: "v"}
	}

	cases := []struct {
		name string
		req  AnalyticsEventQueryRequest
	}{
		{"missing date", AnalyticsEventQueryRequest{}},
		{"too many label filters", AnalyticsEventQueryRequest{Date: "2026-09-14", LabelFilters: tooMany}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, analyticsQueryResponseFixture)
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			if _, err := client.Analytics.Query(context.Background(), testCase.req); err == nil {
				t.Fatal("expected a validation error")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("validation should not cost a request, sent %d", len(doer.requests))
			}
		})
	}
}

// The service accepts an internal anonymousId predicate and countOnly flag, but both are
// absent from its public JSON contract. The request struct must not be able to carry them.
func TestAnalyticsQueryRequestCarriesOnlyPublicFields(t *testing.T) {
	body, err := json.Marshal(AnalyticsEventQueryRequest{Date: "2026-09-14"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != `{"date":"2026-09-14"}` {
		t.Fatalf("body = %s", got)
	}
}
