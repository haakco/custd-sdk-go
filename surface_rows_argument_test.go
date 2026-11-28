package custd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// argumentSurfaceRows need real arguments: the SDK validates them before it
// sends, or one method selects the operation from an argument.
var argumentSurfaceRows = []surfaceRow{
	{id: "ingest-api POST /api/v1/raw", path: "/api/v1/raw", call: func(e *surfaceEnv) error {
		_, err := e.client.IngestRawLines(context.Background(), RawIngestRequest{
			Content: "line", ParserUUID: "0195f0a0-0000-4000-8000-000000000001"})
		return err
	}},
	{id: "admin-api PUT /api/v1/app/profile/photo", path: "/api/v1/app/profile/photo", call: func(e *surfaceEnv) error {
		_, err := e.client.UploadCurrentAppProfilePhoto(context.Background(), "a1", surfacePhoto())
		return err
	}},
	{id: "admin-api POST /api/v1/app/profile/photo", path: "/api/v1/app/profile/photo", call: func(e *surfaceEnv) error {
		_, err := e.client.InitializeCurrentAppProfilePhoto(context.Background(), "a1", surfacePhoto())
		return err
	}},
	{id: "ingest-api POST /api/v1/events/batch", path: "/api/v1/events/batch", call: func(e *surfaceEnv) error {
		_, err := e.client.SendEvents(context.Background(), SendEventsRequest{Events: []EventEnvelope{*validEvent()}})
		return err
	}},
	{id: "analytics-api POST /api/v1/analytics/query", path: "/api/v1/analytics/query", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.CountEvents(context.Background(), AnalyticsEventQuery{Date: "2026-11-20", Source: "auto"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/oauth-clients", path: "/api/v1/admin/oauth-clients", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.OAuthClients.Create(context.Background(), AdminOAuthClientCreate{ClientID: "a1"},
			OAuthClientRequestOptions{IdempotencyKey: surfaceIdempotencyKey})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/oauth-clients/{clientId}/rotate-secret", path: "/api/v1/admin/oauth-clients/a1/rotate-secret", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.OAuthClients.RotateSecret(context.Background(), "a1",
			OAuthClientRequestOptions{IdempotencyKey: surfaceIdempotencyKey})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/measurement/predictions/definitions/{definitionUuid}/duration-evaluations", path: "/api/v1/admin/measurement/predictions/definitions/a2/duration-evaluations", call: func(e *surfaceEnv) error {
		_, _, err := e.client.Admin.Predictions.EvaluateDurationHistory(context.Background(), "a1", "a2",
			&DurationEvaluationRequest{PlanUUID: "0195f0a0-0000-4000-8000-000000000001", SemanticKey: "a3"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/workflow-timings/observations:batch", path: "/api/v1/admin/workflow-timings/observations:batch",
		skipExample: "both",
		skipReason:  "the example's placeholder result list (null, or one item at index 1) cannot answer a one-input batch; the SDK rejects the mismatch by design", call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.WorkflowTimings.AppendBatch(context.Background(), "a1", []WorkflowTimingObservation{surfaceTimingObservation()})
			return err
		}},
	{id: "admin-api POST /api/v1/admin/workflow-timings/corrections", path: "/api/v1/admin/workflow-timings/corrections", call: func(e *surfaceEnv) error {
		observation := surfaceTimingObservation()
		observation.SupersedesFactID = "0195f0a0-0000-4000-8000-000000000002"
		_, err := e.client.Admin.WorkflowTimings.Correct(context.Background(), "a1", observation)
		return err
	}},
	{id: "admin-api PUT /api/v1/admin/workflow-timings/definitions/{workflowKey}", path: "/api/v1/admin/workflow-timings/definitions/a2", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.WorkflowTimings.Reconcile(context.Background(), "a1", WorkflowTimingDeclaration{WorkflowKey: "a2"})
		return err
	}},
	{id: "admin-api GET /api/v1/admin/audit/events/export", path: "/api/v1/admin/audit/events/export", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.Audit.ExportEvents(context.Background(), surfaceAuditOptions, "csv", io.Discard)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/time-plans/runs/{runUuid}/annotations", path: "/api/v1/admin/time-plans/runs/a2/annotations", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.TimePlans.CreateAnnotation(context.Background(), "a1", "a2", TimePlanAnnotationInput{Type: "note", Text: "a3"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/time-plans/runs/{runUuid}/annotations/{annotationUuid}/corrections", path: "/api/v1/admin/time-plans/runs/a2/annotations/a3/corrections", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.TimePlans.CorrectAnnotation(context.Background(), "a1", "a2", "a3", TimePlanAnnotationInput{Type: "note", Text: "a4"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/time-plans/runs/{runUuid}/annotations/{annotationUuid}/redact", path: "/api/v1/admin/time-plans/runs/a2/annotations/a3/redact", call: func(e *surfaceEnv) error {
		return e.client.Admin.TimePlans.RedactAnnotation(context.Background(), "a1", "a2", "a3", TimePlanRedactionRequest{Reason: "a4"})
	}},
	{id: "admin-api POST /api/v1/admin/time-plans/runs", path: "/api/v1/admin/time-plans/runs", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.TimePlans.CreateRun(context.Background(), "a1", TimePlanRunRequest{PlanUUID: "0195f0a0-0000-4000-8000-000000000001"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/time-plans/runs/{runUuid}/commands", path: "/api/v1/admin/time-plans/runs/a2/commands", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.TimePlans.Execute(context.Background(), "a1", "a2", TimePlanCommandRequest{
			CommandID: "0195f0a0-0000-4000-8000-000000000001", IdempotencyKey: "a3", ExpectedVersion: 1, Type: "start_run"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{action}", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/enable", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ActOnDirectory(context.Background(), "a1", "a2", AuthProjectDirectoryEnable,
			AuthProjectDirectoryRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/apply", path: "/api/v1/admin/auth-projects/a1/environments/a2/apply", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ApplyEnvironmentDesiredState(context.Background(), "a1", "a2",
			AuthProjectDesiredStateRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/apps", path: "/api/v1/admin/apps", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.CreateApp(context.Background(), AppCreateRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects", path: "/api/v1/admin/auth-projects", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.CreateProject(context.Background(), AuthProjectCreateRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/erase", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/erase", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ErasePrincipal(context.Background(), surfacePrincipal,
			ApplicationPrincipalErasureRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/export", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/export", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ExportPrincipal(context.Background(), surfacePrincipal, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/profile", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/profile", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.GetPrincipalProfile(context.Background(), surfacePrincipal, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/profile", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/profile", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.SetPrincipalProfile(context.Background(), surfacePrincipal,
			ApplicationProfileValuesRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/principals/{providerSubject}/audiences", path: "/api/v1/admin/auth-projects/a1/environments/a2/principals/a4/audiences", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ListPrincipalAudiences(context.Background(), surfacePrincipal, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/principals/{providerSubject}/audiences/grant", path: "/api/v1/admin/auth-projects/a1/environments/a2/principals/a4/audiences/grant", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.GrantPrincipalAudience(context.Background(), surfacePrincipal,
			AudienceGrantRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/principals/{providerSubject}/audiences/revoke", path: "/api/v1/admin/auth-projects/a1/environments/a2/principals/a4/audiences/revoke", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.RevokePrincipalAudience(context.Background(), surfacePrincipal,
			AudienceGrantRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/memberships", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/memberships", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ListPrincipalMemberships(context.Background(), surfacePrincipal, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/suspend", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/suspend", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.SuspendPrincipal(context.Background(), surfacePrincipal, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/restore", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/restore", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.RestorePrincipal(context.Background(), surfacePrincipal, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/sessions/revoke", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/sessions/revoke", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.RevokePrincipalSession(context.Background(), "a1", "a2", "a3", "a4",
			ApplicationSessionRevokeRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/sessions/revoke-all", path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/sessions/revoke-all", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.RevokePrincipalSessions(context.Background(), "a1", "a2", "a3", "a4",
			ApplicationSessionsRevokeAllRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/invitations/revoke", path: "/api/v1/admin/auth-projects/a1/environments/a2/invitations/revoke", call: func(e *surfaceEnv) error {
		_, _, err := e.client.Admin.AuthProjects.InvitationByID(context.Background(), AuthProjectInvitationByIDTarget{
			Verb: AuthProjectInvitationVerbRevoke, ProjectID: "a1", EnvironmentID: "a2", InvitationID: "a3"}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/invitations/redeliver", path: "/api/v1/admin/auth-projects/a1/environments/a2/invitations/redeliver", call: func(e *surfaceEnv) error {
		_, _, err := e.client.Admin.AuthProjects.InvitationByID(context.Background(), AuthProjectInvitationByIDTarget{
			Verb: AuthProjectInvitationVerbRedeliver, ProjectID: "a1", EnvironmentID: "a2", InvitationID: "a3"}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/invitations", path: "/api/v1/admin/auth-projects/a1/environments/a2/invitations", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.IssueInvitation(context.Background(), "a1", "a2",
			AuthProjectInvitationIssueRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/invitations/redeliver", path: "/api/v1/admin/auth-projects/a1/environments/a2/invitations/redeliver", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.RedeliverInvitation(context.Background(), "a1", "a2",
			AuthProjectInvitationRedeliveryRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/invitations/revoke", path: "/api/v1/admin/auth-projects/a1/environments/a2/invitations/revoke", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.RevokeInvitation(context.Background(), "a1", "a2",
			AuthProjectInvitationRevocationRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/clients/reconcile", path: "/api/v1/admin/auth-projects/a1/environments/a2/clients/reconcile", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.AuthProjects.ReconcileEnvironmentClients(context.Background(), "a1", "a2",
			AuthProjectClientSyncRequest{}, surfaceAuthOptions)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/data-scopes", path: "/api/v1/admin/data-scopes", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.DataScopes.CreateScope(context.Background(), DataScopeCreateRequest{Code: "a1", Name: "a2"})
		return err
	}},
	{id: "admin-api PATCH /api/v1/admin/data-scopes/{scopeUuid}", path: "/api/v1/admin/data-scopes/a1", call: func(e *surfaceEnv) error {
		name := "a2"
		_, err := e.client.Admin.DataScopes.UpdateScope(context.Background(), "a1", DataScopeUpdateRequest{Name: &name})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/data-scopes/{scopeUuid}/resources", path: "/api/v1/admin/data-scopes/a1/resources", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.DataScopes.CreateResource(context.Background(), "a1", DataScopeResourceCreateRequest{Code: "a2", Name: "a3"})
		return err
	}},
	{id: "admin-api PATCH /api/v1/admin/data-scopes/{scopeUuid}/resources/{resourceUuid}", path: "/api/v1/admin/data-scopes/a1/resources/a2", call: func(e *surfaceEnv) error {
		name := "a3"
		_, err := e.client.Admin.DataScopes.UpdateResource(context.Background(), "a1", "a2", DataScopeResourceUpdateRequest{Name: &name})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/data-scope-grants", path: "/api/v1/admin/data-scope-grants", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.DataScopes.CreateGrant(context.Background(), surfaceGrant)
		return err
	}},
	{id: "admin-api POST /api/v1/admin/data-scope-grants/revoke", path: "/api/v1/admin/data-scope-grants/revoke", call: func(e *surfaceEnv) error {
		return e.client.Admin.DataScopes.RevokeGrant(context.Background(), surfaceGrant)
	}},
	{id: "admin-api GET /api/v1/admin/data-scope-grants", path: "/api/v1/admin/data-scope-grants", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.DataScopes.ListGrants(context.Background(), surfaceGrant.Principal)
		return err
	}},
	{id: "admin-api GET /api/v1/admin/data-scope-grants", path: "/api/v1/admin/data-scope-grants", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.DataScopes.ListGrants(context.Background(),
			DataScopePrincipalRef{Kind: DataScopePrincipalMachine, ClientID: "a1"})
		return err
	}},
	{id: "admin-api POST /api/v1/admin/notification-webhooks/{uuid}/rotate-secret", path: "/api/v1/admin/notification-webhooks/a2/rotate-secret", call: func(e *surfaceEnv) error {
		_, _, err := e.client.Admin.Webhooks.RotateSecret(context.Background(), "a1", "a2")
		return err
	}},
	{id: "admin-api POST /api/v1/admin/notification-deliveries/{uuid}/replay", path: "/api/v1/admin/notification-deliveries/a2/replay", call: func(e *surfaceEnv) error {
		_, err := e.client.Admin.Webhooks.Replay(context.Background(), "a1", "a2", "a3")
		return err
	}},
	{id: "admin-api POST /api/v1/admin/users/{userUuid}/soft-delete", path: "/api/v1/admin/users/a1/soft-delete", call: func(e *surfaceEnv) error {
		_, _, err := e.client.Admin.AdminUsers.Remove(context.Background(), "a1", AdminUserRemovalSoftDelete, "a2", "a3")
		return err
	}},
	{id: "admin-api POST /api/v1/admin/users/{userUuid}/restore", path: "/api/v1/admin/users/a1/restore", call: func(e *surfaceEnv) error {
		_, _, err := e.client.Admin.AdminUsers.Remove(context.Background(), "a1", AdminUserRemovalRestore, "a2", "a3")
		return err
	}},
	{id: "analytics-api POST /api/v1/analytics/query", path: "/api/v1/analytics/query", call: func(e *surfaceEnv) error {
		_, err := e.client.Analytics.Query(context.Background(), AnalyticsEventQueryRequest{Date: "2026-11-20"})
		return err
	}},
	{id: "analytics-api POST /api/v1/analytics/query-range", path: "/api/v1/analytics/query-range", call: func(e *surfaceEnv) error {
		_, err := e.client.Analytics.QueryRange(context.Background(), AnalyticsEventRangeQueryRequest{From: "2026-11-01", To: "2026-11-20"})
		return err
	}},
	{id: "analytics-api POST /api/v1/reporting/exports", path: "/api/v1/reporting/exports", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.CreateExport(context.Background(), ReportExportCreateRequest{DashboardKey: "a1", Formats: []ExportFormat{FormatCSV}})
		return err
	}},
	{id: "analytics-api GET /api/v1/reporting/exports/{exportID}", path: "/api/v1/reporting/exports/0195f0a0-0000-4000-8000-000000000001", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.GetExport(context.Background(), surfaceUUID)
		return err
	}},
	{id: "analytics-api POST /api/v1/reporting/exports/{exportID}/cancel", path: "/api/v1/reporting/exports/0195f0a0-0000-4000-8000-000000000001/cancel", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.CancelExport(context.Background(), surfaceUUID, "a1")
		return err
	}},
	{id: "analytics-api GET /api/v1/reporting/exports/{exportID}/artifacts/{artifactID}", path: "/api/v1/reporting/exports/0195f0a0-0000-4000-8000-000000000001/artifacts/0195f0a0-0000-4000-8000-000000000002", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.DownloadExport(context.Background(), surfaceUUID, surfaceUUID2)
		return err
	}},
	{id: "analytics-api GET /api/v1/reporting/exports/{exportID}/artifacts/{artifactID}", path: "/api/v1/reporting/exports/0195f0a0-0000-4000-8000-000000000001/artifacts/0195f0a0-0000-4000-8000-000000000002", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.DownloadExportArtifact(context.Background(), surfaceUUID, surfaceUUID2)
		return err
	}},
	{id: "analytics-api GET /api/v1/reporting/outputs/{outputUuid}", path: "/api/v1/reporting/outputs/0195f0a0-0000-4000-8000-000000000001", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.Output(context.Background(), surfaceUUID)
		return err
	}},
	{id: "analytics-api POST /api/v1/reporting/outputs/{outputUuid}/query", path: "/api/v1/reporting/outputs/0195f0a0-0000-4000-8000-000000000001/query", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.QueryOutput(context.Background(), surfaceUUID, ReportingQueryRequest{})
		return err
	}},
	{id: "analytics-api GET /api/v1/processing/{receiptUuid}", path: "/api/v1/processing/0195f0a0-0000-4000-8000-000000000001", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.Receipt(context.Background(), surfaceUUID)
		return err
	}},
	{id: "analytics-api POST /api/v1/reporting/insights/subject", path: "/api/v1/reporting/insights/subject", call: func(e *surfaceEnv) error {
		_, err := e.client.Reporting.SubjectInsight(context.Background(), SubjectInsightRequest{Template: "a1", Subject: "a2", RangeDays: 7})
		return err
	}},
	{id: "ingest-api ANY /api/v1/errors/api/{project}/envelope/", path: "/api/v1/errors/api/1/envelope/", method: http.MethodPost, call: func(e *surfaceEnv) error {
		client, err := NewErrorCaptureClient(ErrorCaptureConfig{BaseURL: e.serverURL(), WriteKey: "site_pk_" + strings.Repeat("b", 48), Environment: "test"})
		if err != nil {
			return err
		}
		if result := client.CaptureException(context.Background(), CaptureExceptionRequest{Error: errors.New("a1")}); result.Status == 0 {
			return fmt.Errorf("error capture outcome %s", result.Outcome)
		}
		return nil
	}},
}

const (
	surfaceUUID           = "0195f0a0-0000-4000-8000-000000000001"
	surfaceUUID2          = "0195f0a0-0000-4000-8000-000000000002"
	surfaceIdempotencyKey = "0195f0a0-0000-4000-8000-00000000000a"
)

var (
	surfaceAuthOptions  = AuthProjectRequestOptions{IdempotencyKey: surfaceIdempotencyKey}
	surfacePrincipal    = ApplicationPrincipalTarget{ProjectID: "a1", EnvironmentID: "a2", DirectoryID: "a3", ProviderSubject: "a4"}
	surfaceAuditOptions = AuditListOptions{
		Scope: AuditScopeGlobal, AffectedTenantSlug: "a1", Since: "2026-11-01T00:00:00Z", Until: "2026-11-02T00:00:00Z",
		ActorKind: "a2", ActorReference: "a3", Action: "a4", ResourceType: "a5", ResourceID: "a6",
		Outcome: AuditOutcomeSuccess, CorrelationID: "a7", EnvironmentID: surfaceUUID, Limit: 5, Cursor: "a8",
	}
	surfaceReservationScope = ProducerReservationScope{ParentCompanySlug: "p1"}
	surfaceUsageQuery       = UsageQuery{
		MeterSlug: "a1", Start: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), Limit: 5,
	}
	surfaceGrant = DataScopeGrant{
		Principal: DataScopePrincipalRef{Kind: DataScopePrincipalHuman, UserUUID: surfaceUUID},
		Target:    DataScopeGrantTarget{Kind: DataScopeGrantTargetParent},
		Action:    DataScopeActionRead,
	}
)

func surfacePhoto() CurrentAppProfilePhoto {
	return CurrentAppProfilePhoto{ContentType: "image/png", Bytes: []byte{1, 2}}
}

func surfaceTimingObservation() WorkflowTimingObservation {
	return WorkflowTimingObservation{WorkflowKey: "a2", ExternalRunID: "a3", IdempotencyKey: "a4",
		Fact: WorkflowTimingFact{}}
}
