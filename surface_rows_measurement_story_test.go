package custd

import (
	"context"
	"time"
)

var measurementStorySurfaceRows = []surfaceRow{
	{
		id:   "analytics-api GET /api/v1/measurement/overview",
		path: "/api/v1/measurement/overview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Overview(context.Background())
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/projects/{projectUuid}/forecast-story",
		path: "/api/v1/measurement/projects/a1/forecast-story",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Projects.GetForecastStory(context.Background(), "a1")
			return err
		},
	},
}

var predictionReadSurfaceRows = []surfaceRow{
	{
		id:   "analytics-api GET /api/v1/measurement/predictions/definitions",
		path: "/api/v1/measurement/predictions/definitions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Predictions.ListDefinitions(context.Background())
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/predictions/definitions/{definitionUuid}",
		path: "/api/v1/measurement/predictions/definitions/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Predictions.GetDefinition(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/predictions/definitions/{definitionUuid}/runs",
		path: "/api/v1/measurement/predictions/definitions/a1/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Predictions.ListRuns(context.Background(), "a1", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/predictions/definitions/{definitionUuid}/outcomes",
		path: "/api/v1/measurement/predictions/definitions/a1/outcomes",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Predictions.ListOutcomes(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/measurement/predictions/definitions/{definitionUuid}/evaluations",
		path: "/api/v1/measurement/predictions/definitions/a1/evaluations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Measurement.Predictions.GetLatestEvaluation(context.Background(), "a1")
			return err
		},
	},
}
