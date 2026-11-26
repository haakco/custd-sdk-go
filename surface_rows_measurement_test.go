package custd

import "context"

var measurementSurfaceRows = []surfaceRow{
	{
		id:   "admin-api PATCH /api/v1/admin/measurement/projects/{projectUuid}",
		path: "/api/v1/admin/measurement/projects/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.Update(context.Background(), "", "a1", MeasurementProjectUpdate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/cancel",
		path: "/api/v1/admin/measurement/projects/a1/cancel",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.Cancel(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/complete",
		path: "/api/v1/admin/measurement/projects/a1/complete",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.Complete(context.Background(), "", "a1", MeasurementProjectComplete{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/projects/{projectUuid}/estimates",
		path: "/api/v1/admin/measurement/projects/a1/estimates",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.ListEstimates(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/estimates",
		path: "/api/v1/admin/measurement/projects/a1/estimates",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.AddEstimate(context.Background(), "", "a1", MeasurementEstimateCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/runs",
		path: "/api/v1/admin/measurement/projects/a1/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.AddRun(context.Background(), "", "a1", MeasurementRunCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/series",
		path: "/api/v1/admin/measurement/projects/a1/series",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.AddSeries(context.Background(), "", "a1", MeasurementSeriesCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/targets",
		path: "/api/v1/admin/measurement/projects/a1/targets",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.AddTarget(context.Background(), "", "a1", MeasurementTargetCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/targets/{targetUuid}/reopen",
		path: "/api/v1/admin/measurement/projects/a1/targets/a2/reopen",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.ReopenTarget(context.Background(), "", "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/projects/{projectUuid}/decisions",
		path: "/api/v1/admin/measurement/projects/a1/decisions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.ListDecisions(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/projects/{projectUuid}/decisions/{decisionUuid}",
		path: "/api/v1/admin/measurement/projects/a1/decisions/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.GetDecision(context.Background(), "", "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/measurement/projects/{projectUuid}/decisions/{decisionUuid}",
		path: "/api/v1/admin/measurement/projects/a1/decisions/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.UpdateDecision(context.Background(), "", "a1", "a2", MeasurementDecisionModelUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/measurement/projects/{projectUuid}/decisions/{decisionUuid}",
		path: "/api/v1/admin/measurement/projects/a1/decisions/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Measurement.Projects.DeleteDecision(context.Background(), "", "a1", "a2")
		},
	},
}
