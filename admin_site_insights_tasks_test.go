package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteInsightsTasksFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-insights-tasks.json", key)
}

// The operator task status reads on the admin namespace; an unregistered task decodes with no run.
func TestSiteInsightsTasksStatus(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteInsightsTasksFixture(t, "tasks"))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, _, err := client.Admin.System.ListSiteInsightsTasks(context.Background())
	if err != nil || len(list.Tasks) != 3 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	checks, imports, graph := list.Tasks[0], list.Tasks[1], list.Tasks[2]
	if checks.TaskName != SiteInsightsTaskChecks || !checks.Registered || checks.Backlog[0].Count != 3 ||
		imports.TaskName != SiteInsightsTaskSearchImports || imports.LastFailureAt == nil || imports.LastFailure == "" ||
		graph.TaskName != SiteInsightsTaskWebGraph || graph.Registered || graph.LastRunAt != nil || len(graph.Backlog) != 0 {
		t.Fatalf("tasks = %+v", list.Tasks)
	}
	assertSiteRequests(t, doer.requests, []string{"GET http://localhost:8080/api/v1/admin/site-insights/tasks"})
}

// A failing import is retried on the provider's retry route; one that is not failing is a 409.
func TestSiteSearchImportRetry(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteInsightsTasksFixture(t, "searchConnectionRetried"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	connection, err := domains.RetrySearchImport(ctx, "d-1", AdminSiteSearchProviderGoogle)
	if err != nil || connection.Status != AdminSiteSearchConnectionFailing || connection.ImportedThrough != "2026-11-14" {
		t.Fatalf("connection = %+v, %v", connection, err)
	}
	assertSiteRequests(t, doer.requests, []string{"POST " + siteTestBase + "/search/google/retry"})
	doer.status, doer.body = http.StatusConflict, siteInsightsTasksFixture(t, "retryNotFailingProblem")
	_, err = domains.RetrySearchImport(ctx, "d-1", AdminSiteSearchProviderBing)
	assertSiteProblemCode(t, err, http.StatusConflict, SiteSearchRetryNotFailingCode)
}

// The schema audit's unknown-tenant refusal carries a code a client can act on.
func TestSchemaAuditTenantNotFoundCode(t *testing.T) {
	doer := newCaptureDoer(http.StatusNotFound, recordedFixture(t, "schema-audit.json", "tenantNotFoundProblem"))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.Admin.Schemas.Audit(context.Background(), "nobody")
	assertSiteProblemCode(t, err, http.StatusNotFound, SchemaAuditTenantNotFoundCode)
}
