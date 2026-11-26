package custd

import "context"

var uncertaintySurfaceRows = []surfaceRow{
	{
		id:   "admin-api GET /api/v1/admin/uncertainty-models",
		path: "/api/v1/admin/uncertainty-models",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.List(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/uncertainty-models",
		path: "/api/v1/admin/uncertainty-models",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.Create(context.Background(), "", UncertaintyCreateModelRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/uncertainty-models/{modelUuid}",
		path: "/api/v1/admin/uncertainty-models/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.Get(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/uncertainty-models/{modelUuid}",
		path: "/api/v1/admin/uncertainty-models/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.UncertaintyModels.Delete(context.Background(), "", "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/uncertainty-models/{modelUuid}/copy",
		path: "/api/v1/admin/uncertainty-models/a1/copy",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.Copy(context.Background(), "", "a1", UncertaintyCopyModelRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/uncertainty-models/{modelUuid}/versions",
		path: "/api/v1/admin/uncertainty-models/a1/versions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.ListVersions(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/uncertainty-models/{modelUuid}/versions",
		path: "/api/v1/admin/uncertainty-models/a1/versions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.CreateVersion(context.Background(), "", "a1", UncertaintyCreateVersionRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/uncertainty-models/{modelUuid}/versions/{versionNumber}",
		path: "/api/v1/admin/uncertainty-models/a1/versions/8",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.UncertaintyModels.GetVersion(context.Background(), "", "a1", 8)
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/uncertainty-models/{modelUuid}/runs",
		path: "/api/v1/measurement/uncertainty-models/a1/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.UncertaintyModels.ListRuns(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/measurement/uncertainty-models/{modelUuid}/runs",
		path: "/api/v1/measurement/uncertainty-models/a1/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.UncertaintyModels.CreateRun(context.Background(), "a1", UncertaintyRunRequest{})
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/uncertainty-models/{modelUuid}/runs/{runUuid}",
		path: "/api/v1/measurement/uncertainty-models/a1/runs/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.UncertaintyModels.GetRun(context.Background(), "a1", "a2")
			return err
		},
	},
}
