package custd

import (
	"context"
	"io"
	"net/http"
)

var adminSurfaceRows = []surfaceRow{
	{
		id:     "ingest-api ANY /api/v1/account/me",
		path:   "/api/v1/account/me",
		method: http.MethodGet,
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.GetAccount(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/tenants",
		path: "/api/v1/admin/tenants",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.Create(context.Background(), AdminTenantCreate{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/tenants/{slug}/data-space-entitlement",
		path: "/api/v1/admin/tenants/a1/data-space-entitlement",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.DataSpaceEntitlement(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/tenants/{slug}",
		path: "/api/v1/admin/tenants/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Tenants.Delete(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/tenants/{slug}/finalize",
		path: "/api/v1/admin/tenants/a1/finalize",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Tenants.Finalize(context.Background(), "a1", 2)
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/tenants/{slug}",
		path: "/api/v1/admin/tenants/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/tenants",
		path: "/api/v1/admin/tenants",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/tenants/{slug}/data-space-entitlement",
		path: "/api/v1/admin/tenants/a1/data-space-entitlement",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.UpdateDataSpaceEntitlement(context.Background(), "a1", DataSpaceEntitlementUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/oauth-clients/{clientId}",
		path: "/api/v1/admin/oauth-clients/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.OAuthClients.Delete(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/oauth-clients/{clientId}",
		path: "/api/v1/admin/oauth-clients/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClients.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/oauth-clients",
		path: "/api/v1/admin/oauth-clients",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClients.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/oauth-clients/{clientId}/scopes",
		path: "/api/v1/admin/oauth-clients/a1/scopes",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClients.UpdateScopes(context.Background(), "a1", AdminOAuthClientUpdateScopesRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/oauth-client-admissions/{admissionUuid}/approve",
		path: "/api/v1/admin/oauth-client-admissions/a1/approve",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClientAdmissions.Approve(context.Background(), "a1", AdminOAuthClientAdmissionApproveRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/oauth-client-admissions",
		path: "/api/v1/admin/oauth-client-admissions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClientAdmissions.List(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/oauth-client-admissions/{admissionUuid}/review",
		path: "/api/v1/admin/oauth-client-admissions/a1/review",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClientAdmissions.Review(context.Background(), "a1", AdminOAuthClientAdmissionReviewRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/oauth-client-admissions/{admissionUuid}/revoke",
		path: "/api/v1/admin/oauth-client-admissions/a1/revoke",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClientAdmissions.Revoke(context.Background(), "a1", AdminOAuthClientAdmissionRevokeRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/sites",
		path: "/api/v1/admin/sites",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Sites.Create(context.Background(), AdminSiteCreate{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/sites/{siteUuid}",
		path: "/api/v1/admin/sites/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Sites.Delete(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/sites/{siteUuid}",
		path: "/api/v1/admin/sites/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Sites.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/sites",
		path: "/api/v1/admin/sites",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Sites.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/sites/{siteUuid}/rotate-write-key",
		path: "/api/v1/admin/sites/a1/rotate-write-key",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Sites.RotateWriteKey(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/sites/{siteUuid}",
		path: "/api/v1/admin/sites/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Sites.Update(context.Background(), "a1", AdminSiteUpdate{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/schema/audit",
		path: "/api/v1/admin/schema/audit",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Schemas.Audit(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/schema/{tenantSlug}/{eventType}/dry-run",
		path: "/api/v1/admin/schema/a1/a2/dry-run",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Schemas.DryRun(context.Background(), "a1", "a2", AdminSchemaDryRunRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/schema/{tenantSlug}/{eventType}/enable",
		path: "/api/v1/admin/schema/a1/a2/enable",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Schemas.EnableVersion(context.Background(), "a1", "a2", AdminSchemaEnableRequest{})
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/schema/{tenantSlug}/{eventType}/register",
		path: "/api/v1/admin/schema/a1/a2/register",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Schemas.Register(context.Background(), "a1", "a2", AdminSchemaRegister{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/schema/validate",
		path: "/api/v1/admin/schema/validate",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Schemas.Validate(context.Background(), AdminSchemaValidateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects",
		path: "/api/v1/admin/measurement/projects",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.Create(context.Background(), MeasurementProjectCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/calibration-exercises",
		path: "/api/v1/admin/measurement/projects/a1/calibration-exercises",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.CreateCalibrationExercise(context.Background(), "a1", MeasurementCalibrationExerciseCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/decisions",
		path: "/api/v1/admin/measurement/projects/a1/decisions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.CreateDecisionModel(context.Background(), "a1", MeasurementDecisionModelCreate{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/projects/{projectUuid}",
		path: "/api/v1/admin/measurement/projects/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/projects/{projectUuid}/imports/{importUuid}",
		path: "/api/v1/admin/measurement/projects/a1/imports/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.GetImportRun(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:          "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/observations:csv",
		path:        "/api/v1/admin/measurement/projects/a1/observations:csv",
		skipExample: "minimal",
		skipReason:  "an empty result list cannot answer a one-row submission; the SDK rejects the count mismatch by design",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.ImportCSVString(context.Background(), "a1", "a2", 1)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/projects",
		path: "/api/v1/admin/measurement/projects",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.List(context.Background())
			return err
		},
	},
	{
		id:          "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/observations:bulk",
		path:        "/api/v1/admin/measurement/projects/a1/observations:bulk",
		skipExample: "minimal",
		skipReason:  "an empty result list cannot answer a one-row submission; the SDK rejects the count mismatch by design",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.SubmitObservation(context.Background(), "a1", MeasurementObservationInput{})
			return err
		},
	},
	{
		id:          "admin-api POST /api/v1/admin/measurement/projects/{projectUuid}/observations:bulk",
		path:        "/api/v1/admin/measurement/projects/a1/observations:bulk",
		skipExample: "minimal",
		skipReason:  "an empty result list cannot answer a one-row submission; the SDK rejects the count mismatch by design",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Measurement.Projects.SubmitObservations(context.Background(), "a1", measurementBulkRequest(1))
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/sources/{sourceUuid}/activate",
		path: "/api/v1/admin/measurement/predictions/sources/a2/activate",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Predictions.ActivateSignalSource(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/activate",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/activate",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.ActivateVersion(context.Background(), "a1", "a2", PredictionActivateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/archive",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/archive",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Predictions.ArchiveDefinition(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/sources/{sourceUuid}/archive",
		path: "/api/v1/admin/measurement/predictions/sources/a2/archive",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Predictions.ArchiveSignalSource(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions",
		path: "/api/v1/admin/measurement/predictions/definitions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.CreateDefinition(context.Background(), "a1", PredictionDefinitionCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/sources",
		path: "/api/v1/admin/measurement/predictions/sources",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.CreateSignalSource(context.Background(), "a1", PredictionSignalSourceCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions/{definitionUuid}",
		path: "/api/v1/admin/measurement/predictions/definitions/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.GetDefinition(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/evaluations",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/evaluations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.GetEvaluation(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/sources/{sourceUuid}",
		path: "/api/v1/admin/measurement/predictions/sources/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.GetSignalSource(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/versions/{versionUuid}",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/versions/a3",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.GetVersion(context.Background(), "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions",
		path: "/api/v1/admin/measurement/predictions/definitions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.ListDefinitions(context.Background(), "a1", 2, "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/outcomes",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/outcomes",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.ListOutcomes(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/runs",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.ListRuns(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/sources",
		path: "/api/v1/admin/measurement/predictions/sources",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.ListSignalSources(context.Background(), "a1", 2, "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/threshold-events",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/threshold-events",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.ListThresholdEvents(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/pause",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/pause",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Predictions.PauseDefinition(context.Background(), "a1", "a2", PredictionPauseRequest{})
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/publish",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/publish",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.PublishVersion(context.Background(), "a1", "a2", PredictionVersionPublishRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/resume",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/resume",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Predictions.ResumeDefinition(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/run-now",
		path: "/api/v1/admin/measurement/predictions/definitions/a2/run-now",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Predictions.RunNow(context.Background(), "a1", "a2", PredictionRunNowRequest{})
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/measurement/predictions/definitions/{definitionUuid}",
		path: "/api/v1/admin/measurement/predictions/definitions/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Predictions.UpdateDefinition(context.Background(), "a1", "a2", PredictionDefinitionUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/workflow-timings/observations",
		path: "/api/v1/admin/workflow-timings/observations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.Append(context.Background(), "a1", WorkflowTimingObservation{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/definitions/{workflowKey}/duration-history",
		path: "/api/v1/admin/workflow-timings/definitions/a2/duration-history",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.DurationHistory(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/definitions/{workflowKey}/evaluation",
		path: "/api/v1/admin/workflow-timings/definitions/a2/evaluation",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.Evaluation(context.Background(), "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/definitions/{workflowKey}",
		path: "/api/v1/admin/workflow-timings/definitions/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.GetDefinition(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/runs/{runUuid}",
		path: "/api/v1/admin/workflow-timings/runs/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.GetRun(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/definitions",
		path: "/api/v1/admin/workflow-timings/definitions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.ListDefinitions(context.Background(), "a1", 2)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/runs",
		path: "/api/v1/admin/workflow-timings/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.ListRuns(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/workflow-timings/runs/{runUuid}/prediction",
		path: "/api/v1/admin/workflow-timings/runs/a2/prediction",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.Prediction(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/workflow-timings/runs/{runUuid}/rebuild",
		path: "/api/v1/admin/workflow-timings/runs/a2/rebuild",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.Rebuild(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/privacy/rules",
		path: "/api/v1/admin/privacy/rules",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.GetRules(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/privacy/identifiers/{companySlug}",
		path: "/api/v1/admin/privacy/identifiers/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.ListIdentifierMappings(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/privacy/identifiers/{companySlug}/map",
		path: "/api/v1/admin/privacy/identifiers/a1/map",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.MapIdentifier(context.Background(), "a1", PrivacyIdentifierMapRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/privacy/rules",
		path: "/api/v1/admin/privacy/rules",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.SetRules(context.Background(), PrivacyRules{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/retention/policies/{tenantSlug}/apply",
		path: "/api/v1/admin/retention/policies/a1/apply",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Retention.Apply(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/retention/policies/{tenantSlug}",
		path: "/api/v1/admin/retention/policies/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.Retention.Delete(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/retention/policies/{tenantSlug}",
		path: "/api/v1/admin/retention/policies/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Retention.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/retention/policies",
		path: "/api/v1/admin/retention/policies",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Retention.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/retention/policies/{tenantSlug}/runs",
		path: "/api/v1/admin/retention/policies/a1/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Retention.ListRuns(context.Background(), "a1", RetentionRunsPage{Limit: 5, Offset: 2})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/retention/policies/{tenantSlug}/preview",
		path: "/api/v1/admin/retention/policies/a1/preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Retention.Preview(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/retention/policies/{tenantSlug}",
		path: "/api/v1/admin/retention/policies/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Retention.Upsert(context.Background(), "a1", RetentionPolicyUpsertRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/storage/alerts/{tenantSlug}",
		path: "/api/v1/admin/storage/alerts/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.StorageAlerts.CreateRule(context.Background(), "a1", StorageAlertRuleCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/storage/alerts/{tenantSlug}/{ruleId}",
		path: "/api/v1/admin/storage/alerts/a1/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.StorageAlerts.DeleteRule(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/storage/alerts/{tenantSlug}",
		path: "/api/v1/admin/storage/alerts/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.StorageAlerts.ListRules(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/audit/events/{eventId}",
		path: "/api/v1/admin/audit/events/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Audit.GetEvent(context.Background(), "a1", AuditListOptions{Scope: AuditScopeGlobal})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/audit/events",
		path: "/api/v1/admin/audit/events",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Audit.ListEvents(context.Background(), surfaceAuditOptions)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/audit-events",
		path: "/api/v1/admin/reporting-packs/audit-events",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Audit.ListReportingPackEvents(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/requests/{requestUuid}/acknowledge",
		path: "/api/v1/admin/offboarding/requests/a1/acknowledge",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Acknowledge(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/{requestUuid}/cancel",
		path: "/api/v1/admin/offboarding/a1/cancel",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.CancelRequest(context.Background(), "a1", OffboardingCancelRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/schedules/{tenantSlug}/cancel",
		path: "/api/v1/admin/offboarding/schedules/a1/cancel",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.CancelSchedule(context.Background(), "a1", OffboardingCancelRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/{requestUuid}/confirm",
		path: "/api/v1/admin/offboarding/a1/confirm",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.ConfirmRequest(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/offboarding/requests/{requestUuid}/download",
		path: "/api/v1/admin/offboarding/requests/a1/download",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Download(context.Background(), "a1", io.Discard)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/requests/{requestUuid}/execute",
		path: "/api/v1/admin/offboarding/requests/a1/execute",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Execute(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/requests/{requestUuid}/export",
		path: "/api/v1/admin/offboarding/requests/a1/export",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Export(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/offboarding/{requestUuid}",
		path: "/api/v1/admin/offboarding/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.GetRequest(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/offboarding/schedules/{tenantSlug}",
		path: "/api/v1/admin/offboarding/schedules/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.GetSchedule(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/offboarding/schedules",
		path: "/api/v1/admin/offboarding/schedules",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.ListSchedules(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/requests/{requestUuid}/preview",
		path: "/api/v1/admin/offboarding/requests/a1/preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Preview(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/offboarding/requests/{requestUuid}/receipt",
		path: "/api/v1/admin/offboarding/requests/a1/receipt",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Receipt(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding",
		path: "/api/v1/admin/offboarding",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.RequestOffboarding(context.Background(), OffboardingRequestCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/requests/{requestUuid}/retry",
		path: "/api/v1/admin/offboarding/requests/a1/retry",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Retry(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/offboarding/schedules",
		path: "/api/v1/admin/offboarding/schedules",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Offboarding.Schedule(context.Background(), OffboardingScheduleRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/setup-templates/{kind}/bind",
		path: "/api/v1/admin/reporting-packs/setup-templates/a1/bind",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.BindSetupTemplate(context.Background(), "a1", ReportingPackSetupTemplateBindRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/drafts",
		path: "/api/v1/admin/reporting-packs/drafts",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.CreateDraft(context.Background(), ReportingPackDraftCreate{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/drafts/{draftId}",
		path: "/api/v1/admin/reporting-packs/drafts/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.GetDraft(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/generations/{generationId}",
		path: "/api/v1/admin/reporting-packs/generations/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.GetGeneration(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/generations/{generationId}/status",
		path: "/api/v1/admin/reporting-packs/generations/a1/status",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.GetGenerationStatus(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/generations/{generationId}/rollup-provenance",
		path: "/api/v1/admin/reporting-packs/generations/a1/rollup-provenance",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.GetRollupProvenance(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/drafts",
		path: "/api/v1/admin/reporting-packs/drafts",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.ListDrafts(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/generation-validity",
		path: "/api/v1/admin/reporting-packs/generation-validity",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.ListGenerationValidity(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/reporting-packs/setup-templates",
		path: "/api/v1/admin/reporting-packs/setup-templates",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.ListSetupTemplates(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/preview",
		path: "/api/v1/admin/reporting-packs/preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.Preview(context.Background(), ReportingPackPreviewRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/drafts/{draftId}/publish",
		path: "/api/v1/admin/reporting-packs/drafts/a1/publish",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.Publish(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/drafts/{draftId}/restart",
		path: "/api/v1/admin/reporting-packs/drafts/a1/restart",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.Restart(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/generations/{generationId}/rollback",
		path: "/api/v1/admin/reporting-packs/generations/a1/rollback",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.RollbackGeneration(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/reporting-packs/drafts/{draftId}",
		path: "/api/v1/admin/reporting-packs/drafts/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.UpdateDraft(context.Background(), "a1", ReportingPackDraftUpdate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/reporting-packs/validate",
		path: "/api/v1/admin/reporting-packs/validate",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ReportingPacks.Validate(context.Background(), ReportingPackDraftCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/tenant-storage-locations",
		path: "/api/v1/tenant-storage-locations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TenantStorage.Create(context.Background(), TenantStorageCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/tenant-storage-locations/{id}",
		path: "/api/v1/tenant-storage-locations/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TenantStorage.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/tenant-storage-locations",
		path: "/api/v1/tenant-storage-locations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TenantStorage.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/tenant-storage-locations/{id}",
		path: "/api/v1/tenant-storage-locations/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.TenantStorage.Revoke(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/subject-exports/{requestId}/cancel",
		path: "/api/v1/admin/subject-exports/a1/cancel",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SubjectExports.Cancel(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/subject-exports",
		path: "/api/v1/admin/subject-exports",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SubjectExports.Create(context.Background(), SubjectExportCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/subject-exports/{requestId}/download",
		path: "/api/v1/admin/subject-exports/a1/download",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SubjectExports.Download(context.Background(), "a1", SubjectExportScopeAccess)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/subject-exports/{requestId}/force",
		path: "/api/v1/admin/subject-exports/a1/force",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SubjectExports.Force(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/subject-exports/{requestId}",
		path: "/api/v1/admin/subject-exports/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SubjectExports.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/subject-exports",
		path: "/api/v1/admin/subject-exports",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SubjectExports.List(context.Background(), 5)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/privacy/erasures",
		path: "/api/v1/admin/privacy/erasures",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Erasures.Create(context.Background(), PrivacyErasureCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/privacy/erasures/{requestUuid}/force",
		path: "/api/v1/admin/privacy/erasures/a1/force",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Erasures.Force(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/privacy/erasures/{requestUuid}",
		path: "/api/v1/admin/privacy/erasures/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Erasures.Get(context.Background(), "a1", "")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/privacy/erasures",
		path: "/api/v1/admin/privacy/erasures",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Erasures.List(context.Background(), "")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-labels/catalogue",
		path: "/api/v1/admin/data-labels/catalogue",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.Catalogue(context.Background(), true)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/data-labels",
		path: "/api/v1/admin/data-labels",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.Create(context.Background(), DataLabelDefinitionCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/data-labels/{definitionUuid}/values",
		path: "/api/v1/admin/data-labels/a1/values",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.CreateValue(context.Background(), "a1", DataLabelValueCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/data-labels/{definitionUuid}/disable",
		path: "/api/v1/admin/data-labels/a1/disable",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataLabels.Disable(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/data-label-values/{valueUuid}/disable",
		path: "/api/v1/admin/data-label-values/a1/disable",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataLabels.DisableValue(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-labels/{definitionUuid}",
		path: "/api/v1/admin/data-labels/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.Get(context.Background(), "a1", true)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-labels",
		path: "/api/v1/admin/data-labels",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.List(context.Background(), true)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-label-assignments",
		path: "/api/v1/admin/data-label-assignments",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.ListAssignments(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-labels/usage",
		path: "/api/v1/admin/data-labels/usage",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.ListUsage(context.Background())
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/event-types/{slug}/data-label-defaults/{definitionUuid}",
		path: "/api/v1/admin/event-types/a1/data-label-defaults/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataLabels.RemoveEventTypeDefault(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/data-label-assignments/schema-fields/{assignmentUuid}",
		path: "/api/v1/admin/data-label-assignments/schema-fields/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataLabels.RemoveSchemaFieldAssignment(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/event-types/{slug}/data-label-defaults/{definitionUuid}",
		path: "/api/v1/admin/event-types/a1/data-label-defaults/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataLabels.SetEventTypeDefault(context.Background(), "a1", "a2", EventTypeDataLabelDefaultRequest{})
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/event-schemas/{schemaUuid}/field-data-labels",
		path: "/api/v1/admin/event-schemas/a1/field-data-labels",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataLabels.SetSchemaFieldAssignment(context.Background(), "a1", SchemaFieldDataLabelAssignmentRequest{})
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/data-labels/{definitionUuid}",
		path: "/api/v1/admin/data-labels/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.Update(context.Background(), "a1", DataLabelDefinitionUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/data-label-values/{valueUuid}",
		path: "/api/v1/admin/data-label-values/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataLabels.UpdateValue(context.Background(), "a1", DataLabelValueUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans",
		path: "/api/v1/admin/time-plans",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.Create(context.Background(), "a1", TimePlanDraftRequest{})
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/reporting/exports/completed-run",
		path: "/api/v1/reporting/exports/completed-run",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.CreateCompletedRunExport(context.Background(), CompletedRunExportRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans/runs/{runUuid}/projection-loss",
		path: "/api/v1/admin/time-plans/runs/a2/projection-loss",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.DeleteTimePlanRunProjection(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/{planUuid}",
		path: "/api/v1/admin/time-plans/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.Get(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/runs/{runUuid}/report",
		path: "/api/v1/admin/time-plans/runs/a2/report",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.GetCompletedDurationReport(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/plans/{planUuid}/versions/{versionUuid}",
		path: "/api/v1/admin/time-plans/plans/a2/versions/a3",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.GetPublishedTimePlanVersion(context.Background(), "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/runs/{runUuid}",
		path: "/api/v1/admin/time-plans/runs/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.GetRun(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/operations/definition-math",
		path: "/api/v1/admin/time-plans/operations/definition-math",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.GetTimePlanDefinitionMathProof(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/operations/status",
		path: "/api/v1/admin/time-plans/operations/status",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.TimePlans.GetTimePlanOperationalStatus(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/runs/{runUuid}/history",
		path: "/api/v1/admin/time-plans/runs/a2/history",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.History(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans",
		path: "/api/v1/admin/time-plans",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.List(context.Background(), "a1", 2)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/runs/{runUuid}/annotations",
		path: "/api/v1/admin/time-plans/runs/a2/annotations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.ListAnnotations(context.Background(), "a1", "a2", 3)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/runs/{runUuid}/duration-observations",
		path: "/api/v1/admin/time-plans/runs/a2/duration-observations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.ListTimePlanDurationObservations(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans/preview",
		path: "/api/v1/admin/time-plans/preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.Preview(context.Background(), "a1", TimePlanDefinition{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans/{planUuid}/publish",
		path: "/api/v1/admin/time-plans/a2/publish",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.Publish(context.Background(), "a1", "a2", TimePlanRevisionRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans/runs/{runUuid}/repair",
		path: "/api/v1/admin/time-plans/runs/a2/repair",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.TimePlans.RepairTimePlanRun(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans/operations/outbox/replay",
		path: "/api/v1/admin/time-plans/operations/outbox/replay",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.TimePlans.ReplayTimePlanOutbox(context.Background(), "a1", TimePlanOutboxReplayRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/time-plans/{planUuid}/retire",
		path: "/api/v1/admin/time-plans/a2/retire",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.Retire(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/time-plans/{planUuid}",
		path: "/api/v1/admin/time-plans/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.Revise(context.Background(), "a1", "a2", TimePlanDraftRevisionRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/usage/export",
		path: "/api/v1/admin/usage/export",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Usage.ExportUsage(context.Background(), "a1", surfaceUsageQuery)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/usage/me",
		path: "/api/v1/admin/usage/me",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Usage.Get(context.Background(), surfaceUsageQuery)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/storage-usage/me",
		path: "/api/v1/admin/storage-usage/me",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Usage.GetRetainedUsage(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/usage/me",
		path: "/api/v1/admin/usage/me",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Usage.ListCurrentUsage(context.Background(), surfaceUsageQuery)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/usage",
		path: "/api/v1/admin/usage",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Usage.ListUsage(context.Background(), "a1", surfaceUsageQuery)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/data-scopes/{scopeUuid}/resources/{resourceUuid}/disable",
		path: "/api/v1/admin/data-scopes/a1/resources/a2/disable",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataScopes.DisableResource(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/data-scopes/{scopeUuid}/disable",
		path: "/api/v1/admin/data-scopes/a1/disable",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.DataScopes.DisableScope(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-scopes/{scopeUuid}/resources/{resourceUuid}",
		path: "/api/v1/admin/data-scopes/a1/resources/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataScopes.GetResource(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-scopes/{scopeUuid}",
		path: "/api/v1/admin/data-scopes/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataScopes.GetScope(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-scopes/{scopeUuid}/resources",
		path: "/api/v1/admin/data-scopes/a1/resources",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataScopes.ListResources(context.Background(), "a1", true)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/data-scopes",
		path: "/api/v1/admin/data-scopes",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.DataScopes.ListScopes(context.Background(), true)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/support-access",
		path: "/api/v1/admin/support-access",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.SupportAccess.Create(context.Background(), SupportAccessGrantCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/support-access",
		path: "/api/v1/admin/support-access",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.SupportAccess.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/support-access/{supportAccessGrantUuid}/revoke",
		path: "/api/v1/admin/support-access/a1/revoke",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.SupportAccess.Revoke(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/impersonation-sessions",
		path: "/api/v1/admin/impersonation-sessions",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Impersonation.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/impersonation-sessions",
		path: "/api/v1/admin/impersonation-sessions",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Impersonation.Start(context.Background(), ImpersonationSessionCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/impersonation-sessions/{impersonationSessionUuid}/stop",
		path: "/api/v1/admin/impersonation-sessions/a1/stop",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Impersonation.Stop(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/impersonation-sessions/{impersonationSessionUuid}/terminate",
		path: "/api/v1/admin/impersonation-sessions/a1/terminate",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Impersonation.Terminate(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/notification-webhooks",
		path: "/api/v1/admin/notification-webhooks",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Webhooks.Create(context.Background(), NotificationWebhookCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/notification-webhooks/{uuid}",
		path: "/api/v1/admin/notification-webhooks/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Webhooks.Delete(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/notification-deliveries/{uuid}",
		path: "/api/v1/admin/notification-deliveries/a2",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Webhooks.GetDelivery(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/notification-webhooks",
		path: "/api/v1/admin/notification-webhooks",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Webhooks.List(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/notification-webhooks/{uuid}",
		path: "/api/v1/admin/notification-webhooks/a2",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.Webhooks.SetEnabled(context.Background(), "a1", "a2", false)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/environment-policy",
		path: "/api/v1/admin/environment-policy",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.EnvironmentPolicy.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/environment-policy",
		path: "/api/v1/admin/environment-policy",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.EnvironmentPolicy.Update(context.Background(), "a1", EnvironmentPolicyUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/tenant-manifest/{tenantSlug}",
		path: "/api/v1/admin/tenant-manifest/a1",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.TenantManifest.Apply(context.Background(), "a1", "a2", nil)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/tenant-manifest/{tenantSlug}/readiness",
		path: "/api/v1/admin/tenant-manifest/a1/readiness",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.TenantManifest.Readiness(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/users/{userUuid}/memberships",
		path: "/api/v1/admin/users/a1/memberships",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AdminUsers.AddMembership(context.Background(), "a1", AdminUserMembershipCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/users",
		path: "/api/v1/admin/users",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AdminUsers.Create(context.Background(), AdminUserCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/users/{userUuid}/disable",
		path: "/api/v1/admin/users/a1/disable",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AdminUsers.Disable(context.Background(), "a1", "a2", true)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/users/{userUuid}/enable",
		path: "/api/v1/admin/users/a1/enable",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AdminUsers.Enable(context.Background(), "a1", nil, true)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/users/{userUuid}",
		path: "/api/v1/admin/users/a1",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AdminUsers.Get(context.Background(), "a1", true)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/users",
		path: "/api/v1/admin/users",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AdminUsers.List(context.Background(), AdminUserListOptions{CompanySlug: "a1", Role: "a2", GlobalState: "a3",
				MembershipState: "a4", WelcomeStatus: "a5", IncludeEnded: true, IncludeDeleted: true})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/user-roles",
		path: "/api/v1/admin/user-roles",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AdminUsers.ListRoles(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/users/{userUuid}/reset-password",
		path: "/api/v1/admin/users/a1/reset-password",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AdminUsers.ResetPassword(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/users/{userUuid}/retry-welcome",
		path: "/api/v1/admin/users/a1/retry-welcome",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AdminUsers.RetryWelcome(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/users/{userUuid}",
		path: "/api/v1/admin/users/a1",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AdminUsers.Update(context.Background(), "a1", nil)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/scheduled-tasks/{taskName}",
		path: "/api/v1/admin/scheduled-tasks/a1",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.GetScheduledTask(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/site-insights/ai-usage",
		path: "/api/v1/admin/site-insights/ai-usage",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.GetSiteAIUsage(context.Background(), SiteAIUsagePeriod("a1"))
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/jobs",
		path: "/api/v1/admin/jobs",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.ListJobs(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/scheduled-tasks",
		path: "/api/v1/admin/scheduled-tasks",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.ListScheduledTasks(context.Background())
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/site-insights/tasks",
		path: "/api/v1/admin/site-insights/tasks",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.ListSiteInsightsTasks(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/scheduled-tasks/{taskName}/pause",
		path: "/api/v1/admin/scheduled-tasks/a1/pause",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.System.PauseScheduledTask(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/scheduled-tasks/{taskName}/resume",
		path: "/api/v1/admin/scheduled-tasks/a1/resume",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.System.ResumeScheduledTask(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/jobs/{jobId}/retry",
		path: "/api/v1/admin/jobs/1/retry",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.RetryJob(context.Background(), 1)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/scheduled-task-runs/{runUuid}/retry",
		path: "/api/v1/admin/scheduled-task-runs/a1/retry",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.RetryScheduledTaskRun(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/scheduled-tasks/{taskName}/run",
		path: "/api/v1/admin/scheduled-tasks/a1/run",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.RunScheduledTask(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/system/status",
		path: "/api/v1/admin/system/status",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.System.Status(context.Background())
			return err
		},
	},
}
