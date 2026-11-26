package custd

import "context"

var analyticsReadSurfaceRows = []surfaceRow{
	{
		id:   "analytics-api GET /api/v1/reporting/dashboards",
		path: "/api/v1/reporting/dashboards",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.ListDashboards(context.Background())
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/reporting/dashboard-data-generations",
		path: "/api/v1/reporting/dashboard-data-generations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.GenerateDashboardData(context.Background(), ReportingDashboardGenerationRequest{})
			return err
		},
	},

	{
		id:   "analytics-api POST /api/v1/reporting/draft-preview",
		path: "/api/v1/reporting/draft-preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.PreviewDraft(context.Background(), ReportingDraftPreviewRequest{})
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/analytics/active-now",
		path: "/api/v1/analytics/active-now",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Analytics.ActiveNow(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/analytics/dashboard-summary",
		path: "/api/v1/analytics/dashboard-summary",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Analytics.DashboardSummary(context.Background())
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/telemetry/event-count",
		path: "/api/v1/telemetry/event-count",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Analytics.TelemetryEventCount(context.Background(), "2026-11-26", "a1")
			return err
		},
	},
}
