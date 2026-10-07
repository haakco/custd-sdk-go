package custd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"
)

const usageReportBody = `{
  "schemaVersion": "usage.v1",
  "companySlug": "acme",
  "start": "2026-09-01T00:00:00Z",
  "end": "2026-10-01T00:00:00Z",
  "rows": [{
    "accountCompanySlug": "acme",
    "dataSpaceCompanySlug": "acme-web",
    "meterSlug": "events.ingested",
    "meterVersion": 1,
    "unit": "event",
    "windowStart": "2026-09-01T00:00:00Z",
    "windowEnd": "2026-09-02T00:00:00Z",
    "quantity": 42,
    "sourceWatermark": 100,
    "completenessState": "final",
    "correctionGeneration": 0,
    "calculationVersion": 1
  }],
  "totals": [{
    "accountCompanySlug": "acme",
    "dataSpaceCompanySlug": "acme-web",
    "meterSlug": "events.ingested",
    "unit": "event",
    "quantity": 42
  }],
  "sourceWatermark": 100,
  "containsProvisional": false,
  "containsIncomplete": false
}`

func TestAdminUsageGetReadsCurrentTenantUsage(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, usageReportBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	report, err := client.Admin.Usage.Get(context.Background(), UsageQuery{})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	req := doer.requests[0]
	if req.Method != http.MethodGet {
		t.Fatalf("method = %s", req.Method)
	}
	if req.URL != "http://localhost:8080/api/v1/admin/usage/me" {
		t.Fatalf("url = %s", req.URL)
	}
	if req.Headers["Authorization"] != "Bearer admin-token" {
		t.Fatalf("authorization header = %q", req.Headers["Authorization"])
	}
	if report.SchemaVersion != "usage.v1" || report.CompanySlug != "acme" {
		t.Fatalf("unexpected report: %+v", report)
	}
	if len(report.Rows) != 1 || report.Rows[0].Quantity != 42 {
		t.Fatalf("unexpected rows: %+v", report.Rows)
	}
	if report.Rows[0].CompletenessState != UsageCompletenessFinal {
		t.Fatalf("completeness = %q", report.Rows[0].CompletenessState)
	}
	if len(report.Totals) != 1 || report.Totals[0].MeterSlug != "events.ingested" {
		t.Fatalf("unexpected totals: %+v", report.Totals)
	}
	if report.ContainsProvisional || report.ContainsIncomplete {
		t.Fatalf("completeness flags = %+v", report)
	}
}

func TestAdminUsageGetEncodesWindow(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, usageReportBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.Usage.Get(context.Background(), UsageQuery{
		MeterSlug: "events.ingested",
		Start:     time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		End:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Limit:     100,
	})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	want := "http://localhost:8080/api/v1/admin/usage/me" +
		"?end=2026-10-01T00%3A00%3A00Z&limit=100&meter=events.ingested&start=2026-09-01T00%3A00%3A00Z"
	if doer.requests[0].URL != want {
		t.Fatalf("url = %s, want %s", doer.requests[0].URL, want)
	}
}

// usageReportBodyWith applies a mutation to the canonical usage report so a
// malformed variant can be built without restating every field.
func usageReportBodyWith(t *testing.T, mutate func(report map[string]any)) string {
	t.Helper()
	var report map[string]any
	if err := json.Unmarshal([]byte(usageReportBody), &report); err != nil {
		t.Fatalf("unmarshal usage fixture: %v", err)
	}
	mutate(report)
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal usage fixture: %v", err)
	}
	return string(encoded)
}

func TestAdminUsageGetRejectsMalformedSuccessBody(t *testing.T) {
	cases := map[string]func() string{
		"empty body": func() string { return "" },
		"empty object": func() string {
			return usageReportBodyWith(t, func(report map[string]any) {
				for key := range report {
					delete(report, key)
				}
			})
		},
		"missing completeness flag": func() string {
			return usageReportBodyWith(t, func(report map[string]any) { delete(report, "containsIncomplete") })
		},
		"missing collection": func() string {
			return usageReportBodyWith(t, func(report map[string]any) { delete(report, "rows") })
		},
		"collection is an object": func() string {
			return usageReportBodyWith(t, func(report map[string]any) {
				report["rows"] = map[string]any{"unexpected": report["rows"]}
			})
		},
		"collection is null": func() string {
			return usageReportBodyWith(t, func(report map[string]any) { report["totals"] = nil })
		},
		"wrong field type": func() string {
			return usageReportBodyWith(t, func(report map[string]any) { report["containsProvisional"] = "false" })
		},
		"row missing field": func() string {
			return usageReportBodyWith(t, func(report map[string]any) {
				rows := report["rows"].([]any)
				delete(rows[0].(map[string]any), "quantity")
			})
		},
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, body())
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			_, err := client.Admin.Usage.Get(context.Background(), UsageQuery{})
			var validationErr *ResponseValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Get error = %v, want *ResponseValidationError", err)
			}
		})
	}
}

func TestAdminUsageGetPreservesFractionalSecondsInWindow(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, usageReportBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.Usage.Get(context.Background(), UsageQuery{
		Start: time.Date(2026, 9, 1, 0, 0, 0, 100_000_000, time.UTC),
		End:   time.Date(2026, 9, 1, 0, 0, 0, 200_000_000, time.UTC),
	})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	want := "http://localhost:8080/api/v1/admin/usage/me" +
		"?end=2026-09-01T00%3A00%3A00.2Z&start=2026-09-01T00%3A00%3A00.1Z"
	if doer.requests[0].URL != want {
		t.Fatalf("url = %s, want %s", doer.requests[0].URL, want)
	}
}

func TestAdminUsageGetRejectsInvalidQueryBeforeSending(t *testing.T) {
	cases := map[string]UsageQuery{
		"start not before end": {
			Start: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		},
		"limit above maximum": {Limit: UsageMaxLimit + 1},
		"limit below minimum": {Limit: -1},
	}

	for name, query := range cases {
		t.Run(name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, usageReportBody)
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			if _, err := client.Admin.Usage.Get(context.Background(), query); err == nil {
				t.Fatal("Get returned nil error for an invalid query")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("invalid query sent %d request(s)", len(doer.requests))
			}
		})
	}
}
