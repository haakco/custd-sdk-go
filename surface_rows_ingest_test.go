package custd

import (
	"context"
	"net/http"
)

var ingestSurfaceRows = []surfaceRow{
	{
		id:     "ingest-api ANY /api/v1/events",
		path:   "/api/v1/events",
		method: http.MethodPost,
		call: func(e *surfaceEnv) error {
			_, err := e.client.SendEvent(context.Background(), *validEvent())
			return err
		},
	},
	{
		id:   "ingest-api POST /api/v1/collect/events",
		path: "/api/v1/collect/events",
		call: func(e *surfaceEnv) error {
			_, err := e.client.CollectEvent(context.Background(), SiteCollectorConfig{WriteKey: "site_pk_test", Origin: "https://app.example"}, *validEvent())
			return err
		},
	},
	{
		id:   "ingest-api POST /api/v1/collect/events/batch",
		path: "/api/v1/collect/events/batch",
		call: func(e *surfaceEnv) error {
			_, err := e.client.CollectEvents(context.Background(), SiteCollectorConfig{WriteKey: "site_pk_test", Origin: "https://app.example"}, SendEventsRequest{Events: []EventEnvelope{*validEvent()}})
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/reprocess",
		path:   "/api/v1/reprocess",
		method: http.MethodPost,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Reprocess(context.Background())
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/managed-audit/redactions",
		path:   "/api/v1/managed-audit/redactions",
		method: http.MethodPost,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.RedactManagedAudit(context.Background(), ManagedAuditRedactionRequest{})
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/monitoring/summary",
		path:   "/api/v1/monitoring/summary",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.MonitoringSummary(context.Background())
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/monitoring/alerts",
		path:   "/api/v1/monitoring/alerts",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.MonitoringAlerts(context.Background())
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/system/monitoring/summary",
		path:   "/api/v1/system/monitoring/summary",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.SystemMonitoringSummary(context.Background())
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/diagnostics/ingest",
		path:   "/api/v1/diagnostics/ingest",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Diagnostics(context.Background(), IngestDiagnosticsIngest)
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/diagnostics/buffers",
		path:   "/api/v1/diagnostics/buffers",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Diagnostics(context.Background(), IngestDiagnosticsBuffers)
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/diagnostics/database",
		path:   "/api/v1/diagnostics/database",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Diagnostics(context.Background(), IngestDiagnosticsDatabase)
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/diagnostics/segments",
		path:   "/api/v1/diagnostics/segments",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Diagnostics(context.Background(), IngestDiagnosticsSegments)
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/diagnostics/reporting",
		path:   "/api/v1/diagnostics/reporting",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Diagnostics(context.Background(), IngestDiagnosticsReporting)
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/admin/cleanup",
		path:   "/api/v1/admin/cleanup",
		method: http.MethodPost,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.Cleanup(context.Background())
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/admin/integrity",
		path:   "/api/v1/admin/integrity",
		method: http.MethodPost,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.CheckIntegrity(context.Background())
			return err
		},
	},
	{
		id:   "ingest-api GET /api/v1/admin/cleanup/plans",
		path: "/api/v1/admin/cleanup/plans",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.ListCleanupPlans(context.Background())
			return err
		},
	},
	{
		id:   "ingest-api POST /api/v1/admin/cleanup/plans",
		path: "/api/v1/admin/cleanup/plans",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.CreateCleanupPlan(context.Background(), CleanupPlanCreateRequest{})
			return err
		},
	},
	{
		id:   "ingest-api GET /api/v1/cleanup/jobs",
		path: "/api/v1/cleanup/jobs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.ListCleanupJobs(context.Background())
			return err
		},
	},
	{
		id:   "ingest-api POST /api/v1/cleanup/jobs",
		path: "/api/v1/cleanup/jobs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.CreateCleanupJob(context.Background(), CleanupJobCreateRequest{})
			return err
		},
	},
	{
		id:   "ingest-api GET /api/v1/cleanup/jobs/{uuid}",
		path: "/api/v1/cleanup/jobs/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.GetCleanupJob(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "ingest-api POST /api/v1/cleanup/jobs/{uuid}",
		path: "/api/v1/cleanup/jobs/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.CancelCleanupJob(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "ingest-api GET /api/v1/cleanup/jobs/{uuid}/proof",
		path: "/api/v1/cleanup/jobs/a1/proof",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Ingest.GetCleanupJobProof(context.Background(), "a1")
			return err
		},
	},
}
