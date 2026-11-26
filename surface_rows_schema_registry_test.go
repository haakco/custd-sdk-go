package custd

import "context"

var schemaRegistrySurfaceRows = []surfaceRow{
	{
		id:   "schema-registry GET /api/v1/schemas",
		path: "/api/v1/schemas",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.List(context.Background())
			return err
		},
	},
	{
		id:   "schema-registry GET /api/v1/schemas/{eventType}/versions",
		path: "/api/v1/schemas/a1/versions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.ListVersions(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "schema-registry GET /api/v1/schemas/{eventType}/{version}",
		path: "/api/v1/schemas/a1/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Get(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "schema-registry POST /api/v1/schemas",
		path: "/api/v1/schemas",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Register(context.Background(), EventSchemaRegisterRequest{})
			return err
		},
	},
	{
		id:   "schema-registry POST /api/v1/schemas/validate",
		path: "/api/v1/schemas/validate",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Validate(context.Background(), EventSchemaCheckRequest{})
			return err
		},
	},
	{
		id:   "schema-registry POST /api/v1/schemas/dry-run",
		path: "/api/v1/schemas/dry-run",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.DryRun(context.Background(), EventSchemaCheckRequest{})
			return err
		},
	},
	{
		id:   "schema-registry POST /api/v1/schemas/infer",
		path: "/api/v1/schemas/infer",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Infer(context.Background(), EventSchemaInferRequest{})
			return err
		},
	},
	{
		id:   "schema-registry POST /api/v1/schemas/{eventType}/{version}/active",
		path: "/api/v1/schemas/a1/a2/active",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Activate(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "schema-registry PUT /api/v1/schemas/{eventType}/{version}/enabled",
		path: "/api/v1/schemas/a1/a2/enabled",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Enable(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "schema-registry DELETE /api/v1/schemas/{eventType}/{version}/enabled",
		path: "/api/v1/schemas/a1/a2/enabled",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.Disable(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "schema-registry DELETE /api/v1/schemas/{eventType}",
		path: "/api/v1/schemas/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Schemas.DisableEventType(context.Background(), "a1")
			return err
		},
	},
}
