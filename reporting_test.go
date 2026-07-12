package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestReportingDashboardReadsGenericPackDashboard(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-dashboard-security.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	dashboard, err := client.Reporting.Dashboard(context.Background(), "security_operations")
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}
	if dashboard.Key != "security_operations" || len(dashboard.Widgets) != 1 {
		t.Fatalf("dashboard = %#v", dashboard)
	}
	if dashboard.DefaultRange != "14d" || dashboard.RefreshSeconds != 300 || !reflect.DeepEqual(dashboard.RequiredScopes, []string{"reporting:read"}) {
		t.Fatalf("dashboard metadata = %#v", dashboard)
	}
	widget := dashboard.Widgets[0]
	if widget.Template != "security_events" || !reflect.DeepEqual(widget.Metrics, []string{"event_count"}) || !reflect.DeepEqual(widget.Dimensions, []string{"severity"}) {
		t.Fatalf("dashboard widget = %#v", widget)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/reporting/dashboards/security_operations" {
		t.Fatalf("url = %s", doer.requests[0].URL)
	}
}

func TestReportingQueryReturnsTrustDiagnostics(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-query-security-trust.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	requestFixtureBytes := readContractFixture(t, "reporting-query-security.json")
	var request ReportingQueryRequest
	if err := json.Unmarshal(requestFixtureBytes, &request); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}
	widget, err := client.Reporting.Query(context.Background(), request)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if widget.Trust == nil || widget.Trust.Status != "healthy" || widget.Trust.RollupState != "healthy" {
		t.Fatalf("trust = %#v", widget.Trust)
	}
	if widget.Count != 2 || !widget.Complete || widget.Truncated || widget.QueryDurationMs != 42 || widget.ParquetURICount != 1 || widget.SnapshotAgeMs != 120000 || widget.EventLagP95Ms != 8000 || widget.DeltaCount != 1 || widget.DeltaPercent != 100 || widget.DeltaLabel != "vs previous period" || widget.SecondaryLabel != "reviewed events" {
		t.Fatalf("widget metadata = %#v", widget)
	}
	if len(widget.Buckets) != 1 || widget.Buckets[0].Source != "auto" || !widget.Buckets[0].Complete || widget.Buckets[0].QueryDurationMs != 42 || widget.Buckets[0].ParquetURICount != 1 {
		t.Fatalf("bucket metadata = %#v", widget.Buckets)
	}
	if widget.Trust.SchemaVersion != "security-event/1.0.0" || widget.Trust.Coverage != "complete" || widget.Trust.PermissionClass != "reporting.read" || len(widget.Trust.QueryWarnings) != 0 {
		t.Fatalf("trust metadata = %#v", widget.Trust)
	}
	var requestBody map[string]any
	if err := json.Unmarshal(doer.requests[0].Body, &requestBody); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	var requestFixture map[string]any
	if err := json.Unmarshal(requestFixtureBytes, &requestFixture); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}
	if !reflect.DeepEqual(requestBody, requestFixture) {
		t.Fatalf("request body = %#v, want %#v", requestBody, requestFixture)
	}
}

func TestReportingQuerySerializesMaxRowsWithoutRowLimit(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-query-security-trust.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	fixtureBytes := readContractFixture(t, "reporting-query-max-rows.json")
	var request ReportingQueryRequest
	if err := json.Unmarshal(fixtureBytes, &request); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}
	if _, err := client.Reporting.Query(context.Background(), request); err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(doer.requests[0].Body, &body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if body["maxRows"] != float64(50) {
		t.Fatalf("maxRows = %#v, want 50", body["maxRows"])
	}
	if _, ok := body["rowLimit"]; ok {
		t.Fatalf("request body contains rowLimit: %#v", body)
	}
}

func TestReportingQueryRejectsUnsafeTrustDiagnostics(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-query-unsafe-trust.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Reporting.Query(context.Background(), ReportingQueryRequest{
		Template:  "awthy_secure_checkout_flow",
		Metrics:   []string{"flow_completion_rate"},
		RangeDays: 1,
	})
	if err == nil {
		t.Fatal("Query returned no error for unsafe trust diagnostics")
	}
	message := err.Error()
	if message != "custd: decode reporting response: custd: unsafe reporting trust diagnostics" {
		t.Fatalf("error = %q, want exact generic message", message)
	}
	for _, unsafeValue := range []string{
		"customer@example.test",
		"unknown",
		"failed",
		"none",
		"not_enough_data",
		"enabled",
		"present",
	} {
		if strings.Contains(message, unsafeValue) {
			t.Fatalf("error contains unsafe diagnostic value %q: %q", unsafeValue, message)
		}
	}
}
