package custd

import "context"

var operationsSurfaceRows = []surfaceRow{
	{
		id:   "admin-api PUT /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/image-override",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/image-override",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.SetDomainImageOverride(context.Background(), "a1", "a2", AuthDomainImageOverrideSetRequest{}, AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/image-override",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/image-override",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ClearDomainImageOverride(context.Background(), "a1", "a2", AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/mint",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/mint",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.MintDomain(context.Background(), "a1", "a2", MintAuthDomainRequest{}, AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/app-links",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/app-links",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListAppLinks(context.Background(), "a1", "a2", AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/apps/{slug}/deploy-credentials",
		path: "/api/v1/admin/apps/a1/deploy-credentials",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListAppDeployCredentials(context.Background(), "a1", AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/apps/{slug}/pause",
		path: "/api/v1/admin/apps/a1/pause",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.PauseApp(context.Background(), "a1", AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/apps/{slug}/resume",
		path: "/api/v1/admin/apps/a1/resume",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ResumeApp(context.Background(), "a1", AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/apps/{slug}/repair",
		path: "/api/v1/admin/apps/a1/repair",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.RepairApp(context.Background(), "a1", AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/apps/{slug}/sign-in",
		path: "/api/v1/admin/apps/a1/sign-in",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.SetAppSignIn(context.Background(), "a1", AppSignInRequest{}, AuthProjectRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/tenants/{slug}/storage-limit",
		path: "/api/v1/admin/tenants/a1/storage-limit",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.SetStorageLimit(context.Background(), "a1", StorageLimitUpdateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/oauth-clients/{clientId}/replace-unknown-secret",
		path: "/api/v1/admin/oauth-clients/a1/replace-unknown-secret",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.OAuthClients.ReplaceUnknownSecret(context.Background(), "a1", OAuthClientRequestOptions{IdempotencyKey: "k1"})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/privacy/consent",
		path: "/api/v1/admin/privacy/consent",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.ListConsent(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/privacy/consent",
		path: "/api/v1/admin/privacy/consent",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.RecordConsent(context.Background(), PrivacyConsentRecordRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/privacy/consent/{subjectId}",
		path: "/api/v1/admin/privacy/consent/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Privacy.DeleteConsent(context.Background(), "", "a1", "a2", "a3")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/storage/alerts/{tenantSlug}/events",
		path: "/api/v1/admin/storage/alerts/a1/events",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.StorageAlerts.ListEvents(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/time-plans/runs/{runUuid}/predictions",
		path: "/api/v1/admin/time-plans/runs/a1/predictions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.TimePlans.ListRunPredictions(context.Background(), "", "a1", 5)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/storage-usage",
		path: "/api/v1/admin/storage-usage",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Usage.ListStorageUsage(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-domains/{domainUuid}/approval",
		path: "/api/v1/site-domains/a1/approval",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.Approve(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/site-domains/{domainUuid}/approval",
		path: "/api/v1/site-domains/a1/approval",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SiteDomains.RevokeApproval(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/producer-provisioning/{clientId}",
		path: "/api/v1/producer-provisioning/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.Producers.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/data-spaces/usage",
		path: "/api/v1/data-spaces/usage",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Provisioning.DataSpaces.Usage(context.Background(), "p1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/event-categories",
		path: "/api/v1/admin/event-categories",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.EventTypes.ListCategories(context.Background())
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/event-categories/{slug}",
		path: "/api/v1/admin/event-categories/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.EventTypes.UpdateCategory(context.Background(), "a1", EventCategoryConfigUpdateRequest{})
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/event-types",
		path: "/api/v1/admin/event-types",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.EventTypes.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/admin/event-types/{slug}",
		path: "/api/v1/admin/event-types/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.EventTypes.Update(context.Background(), "a1", EventTypeConfigUpdateRequest{})
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/ingest/claims/{tenantSlug}",
		path: "/api/v1/admin/ingest/claims/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.IngestClaims.List(context.Background(), "a1", 5)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/ingest/claims/{tenantSlug}/{objectUuid}/retry",
		path: "/api/v1/admin/ingest/claims/a1/a2/retry",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.IngestClaims.Retry(context.Background(), "a1", "a2")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/ingest/claims/{tenantSlug}/{objectUuid}/terminal",
		path: "/api/v1/admin/ingest/claims/a1/a2/terminal",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.IngestClaims.MarkTerminal(context.Background(), "a1", "a2", IngestClaimTerminalRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/rls/status",
		path: "/api/v1/admin/rls/status",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.RLS.Status(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/rls/reconciliation-plans",
		path: "/api/v1/admin/rls/reconciliation-plans",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.RLS.CreatePlan(context.Background(), RLSPlanRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/rls/reconciliation-plans/{planUuid}/execute",
		path: "/api/v1/admin/rls/reconciliation-plans/a1/execute",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.RLS.ExecutePlan(context.Background(), "a1", RLSExecuteRequest{}, "k1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/rls/reconciliation-plans/{planUuid}/proof",
		path: "/api/v1/admin/rls/reconciliation-plans/a1/proof",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.RLS.GetPlanProof(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/error-issues",
		path: "/api/v1/error-issues",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ErrorIssues.List(context.Background(), ErrorIssueListOptions{SiteUUID: "a1", State: "a2", Sort: "a3", Limit: 5, Cursor: "a4"})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/error-issues/{issueUuid}",
		path: "/api/v1/error-issues/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ErrorIssues.Get(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/error-issues/{issueUuid}/state",
		path: "/api/v1/error-issues/a1/state",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ErrorIssues.SetState(context.Background(), "", "a1", ErrorIssueStateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/sites/{siteUuid}/config",
		path: "/api/v1/sites/a1/config",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Sites.GetPublicConfig(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/managed-audit/provision",
		path: "/api/v1/managed-audit/provision",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.ManagedAudit.ProvisionProducer(context.Background(), ProducerProvisionRequest{})
			return err
		},
	},

	{
		id:   "admin-api POST /api/v1/admin/tenants/{slug}/historical-cleanup",
		path: "/api/v1/admin/tenants/a1/historical-cleanup",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.Tenants.HistoricalCleanup(context.Background(), "a1", 7, "k1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/bootstrap",
		path: "/api/v1/admin/bootstrap",
		call: func(e *surfaceEnv) error {
			_, _, err := NewPreparedDataClient(e.client.Admin).GetClientBootstrap(context.Background())
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/bootstrap",
		path: "/api/v1/admin/bootstrap",
		call: func(e *surfaceEnv) error {
			_, _, err := NewPreparedDataClient(e.client.Admin).SubmitClientBootstrap(context.Background(), PreparedDataBootstrapSubmitRequest{})
			return err
		},
	},
}
