package custd

import "context"

var emailSurfaceRows = []surfaceRow{
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-permissions/check",
		path: "/api/v1/admin/campaign-email-permissions/check",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.CheckPermission(context.Background(), "", CampaignEmailRecipient{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-permissions/confirm",
		path: "/api/v1/admin/campaign-email-permissions/confirm",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.ConfirmPermission(context.Background(), "", CampaignEmailConfirmation{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-permissions/withdraw",
		path: "/api/v1/admin/campaign-email-permissions/withdraw",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.WithdrawPermission(context.Background(), "", CampaignEmailRecipient{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-revisions",
		path: "/api/v1/admin/campaign-email-revisions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.CreateRevision(context.Background(), "", PublishCampaignEmailRevisionRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/campaign-email-revisions/{revision}",
		path: "/api/v1/admin/campaign-email-revisions/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.GetRevision(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-revisions/{revision}/render",
		path: "/api/v1/admin/campaign-email-revisions/a1/render",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.RenderRevision(context.Background(), "a1", CampaignRenderRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-runs",
		path: "/api/v1/admin/campaign-email-runs",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.StartRun(context.Background(), "", CampaignRunScheduleRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/campaign-email-runs/{runId}",
		path: "/api/v1/admin/campaign-email-runs/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.GetRun(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-runs/{runId}/cancel",
		path: "/api/v1/admin/campaign-email-runs/a1/cancel",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.CancelRun(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/campaign-email-welcome/definition",
		path: "/api/v1/admin/campaign-email-welcome/definition",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.PutWelcomeDefinition(context.Background(), "", CampaignWelcomeDefinition{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/campaign-email-welcome/pause",
		path: "/api/v1/admin/campaign-email-welcome/pause",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.SetWelcomePaused(context.Background(), "", CampaignWelcomePause{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-welcome/activation",
		path: "/api/v1/admin/campaign-email-welcome/activation",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.RecordWelcomeActivation(context.Background(), "", CampaignWelcomeActivation{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/campaign-email-welcome/confirmed-signup",
		path: "/api/v1/admin/campaign-email-welcome/confirmed-signup",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.RecordWelcomeConfirmedSignup(context.Background(), "", CampaignWelcomeSignup{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/campaign-email-welcome/contacts/{contactId}",
		path: "/api/v1/admin/campaign-email-welcome/contacts/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.CampaignEmail.GetWelcomeContact(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/email-relay/{tenantSlug}",
		path: "/api/v1/admin/email-relay/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.EmailRelay.Get(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/email-relay/{tenantSlug}",
		path: "/api/v1/admin/email-relay/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.EmailRelay.Put(context.Background(), "a1", EmailRelayWriteRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/email-relay/{tenantSlug}",
		path: "/api/v1/admin/email-relay/a1",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.EmailRelay.Delete(context.Background(), "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/email-relay/{tenantSlug}/test-send",
		path: "/api/v1/admin/email-relay/a1/test-send",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.EmailRelay.TestSend(context.Background(), "a1", EmailRelayTestSendRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/sdk-usage",
		path: "/api/v1/admin/sdk-usage",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SDKUsage.List(context.Background())
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/sdk-usage/deprecations/{sdk}/{versionScope}",
		path: "/api/v1/admin/sdk-usage/deprecations/a1/a2",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.SDKUsage.PutDeprecation(context.Background(), "a1", "a2", SDKVersionDeprecationRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/sdk-usage/deprecations/{sdk}/{versionScope}",
		path: "/api/v1/admin/sdk-usage/deprecations/a1/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.SDKUsage.DeleteDeprecation(context.Background(), "a1", "a2")
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/accounting/evidence",
		path: "/api/v1/admin/accounting/evidence",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AccountingEvidence.List(context.Background(), "2026-11-26", 5)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/accounting/evidence",
		path: "/api/v1/admin/accounting/evidence",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AccountingEvidence.Create(context.Background(), AccountingEvidenceCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/notification-destinations/{eventSlug}",
		path: "/api/v1/notification-destinations/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.NotificationDestinations.Get(context.Background(), "", "a1")
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/notification-destinations/{eventSlug}",
		path: "/api/v1/notification-destinations/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.NotificationDestinations.Put(context.Background(), "", "a1", NotificationDestinationRequest{})
			return err
		},
	},
}
