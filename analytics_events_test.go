package custd

import (
	"context"
	"encoding/json"
	"errors"
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

const analyticsRangeQueryResponseFixture = `{
  "rows": [{"eventTypeSlug": "page_view", "payload": {"path": "/"}}],
  "count": 3,
  "buckets": [
    {"date": "2026-09-14", "count": 1, "source": "duckdb", "complete": true, "queryDurationMs": 4, "parquetUriCount": 2},
    {"date": "2026-09-15", "count": 2, "source": "duckdb", "complete": true, "queryDurationMs": 5}
  ],
  "sources": [{"name": "duckdb", "count": 3, "complete": true, "fresh": false, "queryDurationMs": 9, "freshnessLagMs": 0}],
  "timing": {"eventLagP50Ms": 10, "eventLagP95Ms": 20, "eventLagMaxMs": 30, "queryDurationMs": 9, "snapshotAgeMs": 100}
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

func TestAnalyticsQueryRangeUsesTheRangeRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, analyticsRangeQueryResponseFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	response, err := client.Analytics.QueryRange(context.Background(), AnalyticsEventRangeQueryRequest{
		From:      "2026-09-14",
		To:        "2026-09-15",
		EventType: "page_view",
		Limit:     10000,
		Source:    AnalyticsRangeSourceAuto,
		GroupBy:   AnalyticsRangeGroupByDay,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(doer.requests) != 1 || !strings.HasSuffix(doer.requests[0].URL, "/api/v1/analytics/query-range") {
		t.Fatalf("request URL = %s", doer.requests[0].URL)
	}
	// The per-day completeness assessment belongs to the server; it is surfaced unchanged.
	if len(response.Buckets) != 2 || response.Buckets[0].Date != "2026-09-14" || !response.Buckets[0].Complete {
		t.Fatalf("buckets = %+v", response.Buckets)
	}
	if response.Buckets[0].Source != AnalyticsSourceDuckDB || response.Buckets[0].ParquetURICount != 2 {
		t.Fatalf("bucket source = %+v", response.Buckets[0])
	}
	if response.Count != 3 || response.Timing.QueryDurationMs != 9 {
		t.Fatalf("count/timing = %d/%+v", response.Count, response.Timing)
	}
}

func TestAnalyticsQueryRangeRejectsInvalidRequestsBeforeSending(t *testing.T) {
	tooMany := make([]AnalyticsLabelFilter, MaxAnalyticsLabelFilters+1)
	for i := range tooMany {
		tooMany[i] = AnalyticsLabelFilter{Key: "k", Value: "v"}
	}

	cases := []struct {
		name string
		req  AnalyticsEventRangeQueryRequest
	}{
		{"missing from", AnalyticsEventRangeQueryRequest{To: "2026-09-15"}},
		{"missing to", AnalyticsEventRangeQueryRequest{From: "2026-09-14"}},
		{"malformed from", AnalyticsEventRangeQueryRequest{From: "14-09-2026", To: "2026-09-15"}},
		{"to before from", AnalyticsEventRangeQueryRequest{From: "2026-09-15", To: "2026-09-14"}},
		{"over the day cap", AnalyticsEventRangeQueryRequest{From: "2026-01-01", To: "2026-05-02"}},
		{"unsupported groupBy", AnalyticsEventRangeQueryRequest{From: "2026-09-14", To: "2026-09-15", GroupBy: "week"}},
		{"too many label filters", AnalyticsEventRangeQueryRequest{From: "2026-09-14", To: "2026-09-15", LabelFilters: tooMany}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, analyticsRangeQueryResponseFixture)
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			if _, err := client.Analytics.QueryRange(context.Background(), testCase.req); err == nil {
				t.Fatal("expected a validation error")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("validation should not cost a request, sent %d", len(doer.requests))
			}
		})
	}
}

func TestAnalyticsQueryRangeAcceptsExactlyTheDayCap(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, analyticsRangeQueryResponseFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	// 2026-01-01..2026-04-30 is 120 inclusive days, the documented maximum.
	if _, err := client.Analytics.QueryRange(context.Background(), AnalyticsEventRangeQueryRequest{
		From: "2026-01-01",
		To:   "2026-04-30",
	}); err != nil {
		t.Fatalf("120-day range should be accepted: %v", err)
	}
	if len(doer.requests) != 1 {
		t.Fatalf("requests = %d", len(doer.requests))
	}
}

// rangeResponseBodyWith applies a mutation to the canonical range response so a
// malformed variant can be built without restating every field.
func rangeResponseBodyWith(t *testing.T, mutate func(response map[string]any)) string {
	t.Helper()
	var response map[string]any
	if err := json.Unmarshal([]byte(analyticsRangeQueryResponseFixture), &response); err != nil {
		t.Fatalf("unmarshal range fixture: %v", err)
	}
	mutate(response)
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal range fixture: %v", err)
	}
	return string(encoded)
}

func TestAnalyticsQueryRangeRejectsMalformedSuccessBody(t *testing.T) {
	cases := map[string]func(t *testing.T) string{
		"empty body": func(t *testing.T) string { t.Helper(); return "" },
		"empty object": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) {
				for key := range response {
					delete(response, key)
				}
			})
		},
		"missing rows collection": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) { delete(response, "rows") })
		},
		"rows is an object": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) {
				response["rows"] = map[string]any{"unexpected": response["rows"]}
			})
		},
		"sources is an object": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) {
				response["sources"] = map[string]any{"unexpected": response["sources"]}
			})
		},
		"missing timing": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) { delete(response, "timing") })
		},
		"timing missing completeness field": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) {
				timing := response["timing"].(map[string]any)
				delete(timing, "snapshotAgeMs")
			})
		},
		"bucket missing field": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) {
				buckets := response["buckets"].([]any)
				delete(buckets[0].(map[string]any), "complete")
			})
		},
		"source missing completeness flag": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) {
				sources := response["sources"].([]any)
				delete(sources[0].(map[string]any), "fresh")
			})
		},
		"wrong field type": func(t *testing.T) string {
			t.Helper()
			return rangeResponseBodyWith(t, func(response map[string]any) { response["count"] = "3" })
		},
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, body(t))
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			_, err := client.Analytics.QueryRange(context.Background(), AnalyticsEventRangeQueryRequest{
				From: "2026-09-14",
				To:   "2026-09-15",
			})
			var validationErr *ResponseValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("QueryRange error = %v, want *ResponseValidationError", err)
			}
		})
	}
}

func TestAnalyticsQueryRangeAcceptsEmptySource(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, analyticsRangeQueryResponseFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	// The range contract accepts an omitted, empty, auto, or duckdb source; an empty
	// source is normalized to omitted by the omitempty wire tag.
	if _, err := client.Analytics.QueryRange(context.Background(), AnalyticsEventRangeQueryRequest{
		From:   "2026-09-14",
		To:     "2026-09-15",
		Source: AnalyticsRangeQuerySource(""),
	}); err != nil {
		t.Fatalf("empty source should be accepted: %v", err)
	}
	if len(doer.requests) != 1 {
		t.Fatalf("requests = %d", len(doer.requests))
	}
}

func TestAnalyticsQueryRangeRejectsRetiredSourcesBeforeSending(t *testing.T) {
	for _, source := range []AnalyticsRangeQuerySource{"postgres", "rollup", "materialized"} {
		t.Run(string(source), func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, analyticsRangeQueryResponseFixture)
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			_, err := client.Analytics.QueryRange(context.Background(), AnalyticsEventRangeQueryRequest{
				From:   "2026-09-14",
				To:     "2026-09-15",
				Source: source,
			})
			if err == nil {
				t.Fatalf("QueryRange accepted retired source %q", source)
			}
			if len(doer.requests) != 0 {
				t.Fatalf("retired source %q sent %d request(s)", source, len(doer.requests))
			}
		})
	}
}
