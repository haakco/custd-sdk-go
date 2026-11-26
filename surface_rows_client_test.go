package custd

import (
	"context"
	"net/http"
)

var clientSurfaceRows = []surfaceRow{
	{
		id:   "admin-api GET /api/v1/app/profile",
		path: "/api/v1/app/profile",
		call: func(e *surfaceEnv) error {
			_, err := e.client.GetCurrentAppProfile(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/profile/photo",
		path: "/api/v1/app/profile/photo",
		call: func(e *surfaceEnv) error {
			_, err := e.client.GetCurrentAppProfilePhoto(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/profile",
		path: "/api/v1/app/profile",
		call: func(e *surfaceEnv) error {
			_, err := e.client.InitializeCurrentAppProfile(context.Background(), "a1", ApplicationProfileValuesRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/app/profile/photo",
		path: "/api/v1/app/profile/photo",
		call: func(e *surfaceEnv) error {
			_, err := e.client.RemoveCurrentAppProfilePhoto(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/app/profile",
		path: "/api/v1/app/profile",
		call: func(e *surfaceEnv) error {
			_, err := e.client.UpdateCurrentAppProfile(context.Background(), "a1", ApplicationProfileValuesRequest{})
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/projects/{projectUuid}/decisions/{decisionUuid}/simulation-runs/{runUuid}",
		path: "/api/v1/measurement/projects/a1/decisions/a2/simulation-runs/a3",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Projects.GetDecisionSimulationRun(context.Background(), "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/projects/{projectUuid}/forecast-runs/{runUuid}",
		path: "/api/v1/measurement/projects/a1/forecast-runs/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Projects.GetForecastRun(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/measurement/projects/{projectUuid}/decisions/{decisionUuid}/simulation-runs",
		path: "/api/v1/measurement/projects/a1/decisions/a2/simulation-runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Projects.RunDecisionSimulation(context.Background(), "a1", "a2", MeasurementSimulationRunRequest{})
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/measurement/projects/{projectUuid}/forecast-runs",
		path: "/api/v1/measurement/projects/a1/forecast-runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Projects.RunForecast(context.Background(), "a1", MeasurementForecastRunRequest{})
			return err
		},
	},
	{
		id:   "schema-registry POST /api/v1/parsers",
		path: "/api/v1/parsers",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Parsers.Create(context.Background(), ParserSaveRequest{})
			return err
		},
	},
	{
		id:   "schema-registry DELETE /api/v1/parsers/{uuid}",
		path: "/api/v1/parsers/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Parsers.Delete(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "schema-registry GET /api/v1/parsers",
		path: "/api/v1/parsers",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Parsers.List(context.Background())
			return err
		},
	},
	{
		id:   "schema-registry GET /api/v1/parsers/presets",
		path: "/api/v1/parsers/presets",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Parsers.ListPresets(context.Background())
			return err
		},
	},
	{
		id:     "ingest-api ANY /api/v1/parser-preview",
		path:   "/api/v1/parser-preview",
		method: http.MethodPost,
		call: func(e *surfaceEnv) error {
			_, err := e.client.Parsers.Preview(context.Background(), ParserPreviewRequest{})
			return err
		},
	},
	{
		id:   "schema-registry PUT /api/v1/parsers/{uuid}",
		path: "/api/v1/parsers/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Parsers.Update(context.Background(), "a1", ParserSaveRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/data-spaces",
		path: "/api/v1/data-spaces",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.DataSpaces.Create(context.Background(), DataSpaceCreate{}, "p1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/data-spaces",
		path: "/api/v1/data-spaces",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.DataSpaces.List(context.Background(), "p1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/data-spaces/{slug}",
		path: "/api/v1/data-spaces/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Provisioning.DataSpaces.Revoke(context.Background(), "a1", "p1")
		},
	},
	{
		id:   "admin-api POST /api/v1/data-spaces/{slug}/erasure",
		path: "/api/v1/data-spaces/a1/erasure",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.DataSpaces.Erase(context.Background(), "a1", "p1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/data-spaces/{slug}/erasure",
		path: "/api/v1/data-spaces/a1/erasure",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.DataSpaces.GetErasure(context.Background(), "a1", "p1")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/producer-provisioning/{clientId}/signing-keys",
		path: "/api/v1/producer-provisioning/a1/signing-keys",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Producers.SetSigningKeys(context.Background(), "a1", ProducerSigningKeysUpdateRequest{PublicKeys: []string{"AAAA"}})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/producer-provisioning",
		path: "/api/v1/producer-provisioning",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Producers.List(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/producer-provisioning",
		path: "/api/v1/producer-provisioning",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Producers.Provision(context.Background(), ProducerProvisionRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/producer-provisioning/{clientId}",
		path: "/api/v1/producer-provisioning/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Provisioning.Producers.Revoke(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/producer-provisioning/{clientId}/rotate-secret",
		path: "/api/v1/producer-provisioning/a1/rotate-secret",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Producers.RotateSecret(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/producer-provisioning/{clientId}/environment",
		path: "/api/v1/producer-provisioning/a1/environment",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Producers.UpdateEnvironment(context.Background(), "a1", ProducerEnvironmentUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/data-spaces/{slug}/producer-reservations/{producerSlug}/claim",
		path: "/api/v1/data-spaces/a1/producer-reservations/a2/claim",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Reservations.Claim(context.Background(), "a1", "a2", ProducerReservationClaimRequest{}, surfaceReservationScope)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/data-spaces/{slug}/producer-reservations",
		path: "/api/v1/data-spaces/a1/producer-reservations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Reservations.List(context.Background(), "a1", surfaceReservationScope)
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/data-spaces/{slug}/producer-reservations/{producerSlug}",
		path: "/api/v1/data-spaces/a1/producer-reservations/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Provisioning.Reservations.Release(context.Background(), "a1", "a2", surfaceReservationScope)
		},
	},
	{
		id:   "admin-api POST /api/v1/data-spaces/{slug}/producer-reservations",
		path: "/api/v1/data-spaces/a1/producer-reservations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Reservations.Reserve(context.Background(), "a1", ProducerReservationCreateRequest{}, surfaceReservationScope)
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/reporting/dashboards/{key}",
		path: "/api/v1/reporting/dashboards/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.Dashboard(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/reporting/exports",
		path: "/api/v1/reporting/exports",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.ListExports(context.Background(), 1)
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/reporting/outputs",
		path: "/api/v1/reporting/outputs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.Outputs(context.Background())
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/reporting/query",
		path: "/api/v1/reporting/query",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Reporting.Query(context.Background(), ReportingQueryRequest{})
			return err
		},
	},
}
