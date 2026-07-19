package custd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
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

func TestReportingSubjectInsightUsesClosedRequestAndTypedResponse(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-subject-insight-response.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	requestFixture := readContractFixture(t, "reporting-subject-insight-request.json")
	var request SubjectInsightRequest
	if err := json.Unmarshal(requestFixture, &request); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}
	response, err := client.Reporting.SubjectInsight(context.Background(), request)
	if err != nil {
		t.Fatalf("SubjectInsight returned error: %v", err)
	}
	if response.Data.Value.Unit != "count" || response.Data.Value.Value != 2 || len(response.Data.Buckets) != 1 {
		t.Fatalf("response = %#v", response)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/reporting/insights/subject" {
		t.Fatalf("url = %s", doer.requests[0].URL)
	}
	var body map[string]any
	if err := json.Unmarshal(doer.requests[0].Body, &body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	var want map[string]any
	if err := json.Unmarshal(requestFixture, &want); err != nil {
		t.Fatalf("decode expected request: %v", err)
	}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("request body = %#v, want %#v", body, want)
	}
}

func TestReportingSubjectInsightSerializesDateRange(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-subject-insight-response.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	fixture := readContractFixture(t, "reporting-subject-insight-date-range-request.json")
	var request SubjectInsightRequest
	if err := json.Unmarshal(fixture, &request); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}
	if _, err := client.Reporting.SubjectInsight(context.Background(), request); err != nil {
		t.Fatalf("SubjectInsight returned error: %v", err)
	}
	var got, want map[string]any
	if err := json.Unmarshal(doer.requests[0].Body, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatalf("decode expected request: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("request body = %#v, want %#v", got, want)
	}
}

func TestReportingSubjectInsightRejectsInvalidRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	fixture := readContractFixture(t, "invalid-reporting-subject-insight-missing-subject.json")
	var request SubjectInsightRequest
	if err := json.Unmarshal(fixture, &request); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}
	_, err := client.Reporting.SubjectInsight(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), "subject is required") {
		t.Fatalf("error = %v", err)
	}
	if len(doer.requests) != 0 {
		t.Fatalf("requests = %d, want 0", len(doer.requests))
	}
}

func TestReportingSubjectInsightRejectsMalformedRenderedData(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "invalid-reporting-subject-insight-response.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err == nil || !strings.Contains(err.Error(), "rendered widget data missing value") {
		t.Fatalf("error = %v", err)
	}
}

func TestReportingSubjectInsightUsesRenderedTrustContract(t *testing.T) {
	responseBody := `{"data":{"buckets":[],"value":{"value":2,"unit":"count","sampleCount":2,"dataSufficiency":"sufficient","complete":true},"queryDurationMs":12,"snapshotAgeMs":1500,"eventLagP95Ms":320,"trust":{"status":"healthy","dataFreshness":"fresh","lastExport":"2026-07-18T00:00:00Z","rollupState":"healthy","coverage":"complete","captureState":"enabled","consentState":"present","exportState":"complete"}}}`
	doer := newCaptureDoer(http.StatusOK, responseBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	response, err := client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err != nil {
		t.Fatalf("SubjectInsight returned error: %v", err)
	}
	if response.Data.Trust == nil || response.Data.Trust.LastExport != "2026-07-18T00:00:00Z" {
		t.Fatalf("trust = %#v", response.Data.Trust)
	}

	unsafeDoer := newCaptureDoer(http.StatusOK, strings.Replace(responseBody, `"lastExport":"2026-07-18T00:00:00Z"`, `"token":"unsafe"`, 1))
	unsafeClient := newAdminTestClient(t, unsafeDoer, "http://localhost:8080")
	_, err = unsafeClient.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err == nil || err.Error() != "custd: decode reporting response: custd: unsafe reporting trust diagnostics" {
		t.Fatalf("unsafe trust error = %v", err)
	}
}

func TestReportingSubjectInsightNativeHTTPPreservesCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseHandler
	}))
	defer server.Close()
	defer close(releaseHandler)
	client := NewClient(&ClientConfig{BaseURL: server.URL, APIKey: "reporting-token"})
	defer func() { _ = client.Close(context.Background()) }()

	ctx, cancel := context.WithCancel(context.Background())
	errorResult := make(chan error, 1)
	go func() {
		_, err := client.Reporting.SubjectInsight(ctx, SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
		errorResult <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("SubjectInsight request did not reach the server")
	}
	cancel()

	select {
	case err := <-errorResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("SubjectInsight error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("SubjectInsight did not return after context cancellation")
	}
}

func TestReportingSubjectInsightNativeHTTPPreservesAuthAndErrors(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		http.Error(response, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()
	client := NewClient(&ClientConfig{BaseURL: server.URL, APIKey: "reporting-token"})
	defer func() { _ = client.Close(context.Background()) }()

	_, err := client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("SubjectInsight error = %v, want propagated 403", err)
	}
	if authorization != "Bearer reporting-token" {
		t.Fatalf("Authorization = %q, want bearer token", authorization)
	}
}

func TestReportingSubjectInsightRejectsMalformedOptionalMetadata(t *testing.T) {
	responseBody := `{"data":{"buckets":[],"value":{"value":2,"unit":"count","sampleCount":2,"dataSufficiency":"sufficient","complete":true},"queryDurationMs":12,"snapshotAgeMs":1500,"eventLagP95Ms":320,"metadata":{}}}`
	doer := newCaptureDoer(http.StatusOK, responseBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err == nil {
		t.Fatal("SubjectInsight accepted malformed empty metadata; want rejection")
	}
	if !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("error = %v, want mention of metadata", err)
	}
}

func TestReportingSubjectInsightRejectsMalformedOptionalSources(t *testing.T) {
	responseBody := `{"data":{"buckets":[],"value":{"value":2,"unit":"count","sampleCount":2,"dataSufficiency":"sufficient","complete":true},"queryDurationMs":12,"snapshotAgeMs":1500,"eventLagP95Ms":320,"sources":[{}]}}`
	doer := newCaptureDoer(http.StatusOK, responseBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err == nil {
		t.Fatal("SubjectInsight accepted malformed empty source entry; want rejection")
	}
	if !strings.Contains(err.Error(), "source") {
		t.Fatalf("error = %v, want mention of source", err)
	}
}

func TestReportingSubjectInsightRejectsMalformedOptionalTrust(t *testing.T) {
	responseBody := `{"data":{"buckets":[],"value":{"value":2,"unit":"count","sampleCount":2,"dataSufficiency":"sufficient","complete":true},"queryDurationMs":12,"snapshotAgeMs":1500,"eventLagP95Ms":320,"trust":{}}}`
	doer := newCaptureDoer(http.StatusOK, responseBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "security_events", Subject: "subject-123", RangeDays: 7})
	if err == nil {
		t.Fatal("SubjectInsight accepted structurally incomplete trust; want rejection")
	}
	if !strings.Contains(err.Error(), "trust") {
		t.Fatalf("error = %v, want mention of trust", err)
	}
}
