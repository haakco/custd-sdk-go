package custd

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestReportingDashboardReadsAwthyDashboard(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-dashboard-awthy.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	dashboard, err := client.Reporting.Dashboard(context.Background(), "awthy_managed_audit_reporting")
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}
	if dashboard.Key != "awthy_managed_audit_reporting" || len(dashboard.Widgets) != 1 {
		t.Fatalf("dashboard = %#v", dashboard)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/reporting/dashboards/awthy_managed_audit_reporting" {
		t.Fatalf("url = %s", doer.requests[0].URL)
	}
}

func TestReportingQueryReturnsTrustDiagnostics(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, string(readContractFixture(t, "reporting-query-awthy-trust.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	widget, err := client.Reporting.Query(context.Background(), ReportingQueryRequest{
		Template: "awthy_secure_checkout_flow",
		Metrics:  []string{"flow_completion_rate"},
		From:     "2026-07-06",
		To:       "2026-07-06",
		MaxRows:  50,
	})
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if widget.Trust == nil || widget.Trust.Status != "healthy" || widget.Trust.RollupState != "healthy" {
		t.Fatalf("trust = %#v", widget.Trust)
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
	if err == nil || !strings.Contains(err.Error(), "unsafe reporting trust diagnostics") {
		t.Fatalf("error = %v, want unsafe reporting trust diagnostics", err)
	}
}
