package custd

import (
	"context"
	"strings"
	"time"
)

var publicReadSurfaceRows = []surfaceRow{
	{
		id:   "admin-api GET /api/v1/app-hosts/{host}/discovery",
		path: "/api/v1/app-hosts/a1/discovery",
		call: func(e *surfaceEnv) error {
			_, err := e.client.GetAppDiscovery(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/shared-site-reports/{token}",
		path: "/api/v1/shared-site-reports/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.GetSharedSiteReport(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/auth-domain-monitoring/inventory",
		path: "/api/v1/auth-domain-monitoring/inventory",
		call: func(e *surfaceEnv) error {
			_, err := e.client.GetAuthDomainMonitoringInventory(context.Background(), "k1")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/public-status-pages/{key}",
		path: "/api/v1/public-status-pages/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.GetPublicStatusPage(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/public-status-pages/{key}/subscribe",
		path: "/api/v1/public-status-pages/a1/subscribe",
		call: func(e *surfaceEnv) error {
			return e.client.SubscribeToStatusPage(context.Background(), "a1", SiteStatusSubscriptionRequest{Email: "owner@example.com"})
		},
	},
	{
		id:   "admin-api POST /api/v1/public-status-pages/{key}/confirm-subscription",
		path: "/api/v1/public-status-pages/a1/confirm-subscription",
		call: func(e *surfaceEnv) error {
			return e.client.ConfirmStatusPageSubscription(context.Background(), "a1", SiteStatusSubscriptionToken{Token: "t1"})
		},
	},
	{
		id:   "admin-api POST /api/v1/public-status-pages/{key}/unsubscribe",
		path: "/api/v1/public-status-pages/a1/unsubscribe",
		call: func(e *surfaceEnv) error {
			return e.client.UnsubscribeFromStatusPage(context.Background(), "a1", SiteStatusSubscriptionToken{Token: "t1"})
		},
	},
	{
		id:   "admin-api POST /api/v1/site-heartbeat-reports/{companySlug}/{monitorUuid}",
		path: "/api/v1/site-heartbeat-reports/a1/a2",
		call: func(e *surfaceEnv) error {
			_, err := SubmitSiteHeartbeat(context.Background(), surfaceHeartbeatDestination(e, heartbeatReportPrefix),
				SiteHeartbeatReport{ScheduledFor: time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC), Outcome: SiteHeartbeatSucceeded})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-private-uptime-reports/{companySlug}/{monitorUuid}",
		path: "/api/v1/site-private-uptime-reports/a1/a2",
		call: func(e *surfaceEnv) error {
			return SubmitPrivateUptime(context.Background(),
				PrivateUptimeDestination(surfaceHeartbeatDestination(e, privateUptimeReportPrefix)), PrivateUptimeReport{})
		},
	},
	{
		id:   "admin-api GET /api/v1/site-private-uptime-reports/{companySlug}/{monitorUuid}",
		path: "/api/v1/site-private-uptime-reports/a1/a2",
		call: func(e *surfaceEnv) error {
			_, err := GetPrivateUptimeWork(context.Background(),
				PrivateUptimeDestination(surfaceHeartbeatDestination(e, privateUptimeReportPrefix)))
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/site-ai-owner/redeem",
		path: "/api/v1/site-ai-owner/redeem",
		call: func(e *surfaceEnv) error {
			_, err := e.client.RedeemSiteAIOwnerHandoff(context.Background(),
				SiteAIOwnerHandoff{Code: "c1", RedeemURL: "http://localhost:8080/api/v1/site-ai-owner/redeem"}, "https://hub.example")
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/site-ai-owner/settings",
		path: "/api/v1/site-ai-owner/settings",
		call: func(e *surfaceEnv) error {
			_, err := surfaceAIOwnerSession(e).GetSettings(context.Background())
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-ai-owner/settings",
		path: "/api/v1/site-ai-owner/settings",
		call: func(e *surfaceEnv) error {
			_, err := surfaceAIOwnerSession(e).SetSettings(context.Background(), SiteAISettingsRequest{Mode: SiteAIOff})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/site-ai-owner/consent",
		path: "/api/v1/site-ai-owner/consent",
		call: func(e *surfaceEnv) error {
			_, err := surfaceAIOwnerSession(e).SetConsent(context.Background(), SiteAIConsentRequest{Action: SiteAIRevokeConsent})
			return err
		},
	},
}

// surfaceHeartbeatDestination points a monitor call at the env's local server.
func surfaceHeartbeatDestination(e *surfaceEnv, prefix string) SiteHeartbeatDestination {
	return SiteHeartbeatDestination{BaseURL: e.serverURL(), ReportPath: prefix + "a1/a2", WriteToken: strings.Repeat("a", 64)}
}

// surfaceAIOwnerSession is a redeemed session that already holds its write binding.
func surfaceAIOwnerSession(e *surfaceEnv) *SiteAIOwnerSession {
	return &SiteAIOwnerSession{client: e.client, origin: "http://localhost:8080", handle: "h1", binding: "b1"}
}
