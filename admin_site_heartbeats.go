package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteHeartbeatSchedule describes when a scheduled job should report.
type AdminSiteHeartbeatSchedule struct {
	Cron         string `json:"cron"`
	Timezone     string `json:"timezone"`
	GraceMinutes int    `json:"graceMinutes"`
}

// AdminSiteHeartbeatCreate names a monitor and defines its schedule.
type AdminSiteHeartbeatCreate struct {
	Name     string                     `json:"name"`
	Schedule AdminSiteHeartbeatSchedule `json:"schedule"`
}

// SiteHeartbeatReport identifies the scheduled occurrence and its final outcome.
type SiteHeartbeatReport struct {
	ScheduledFor time.Time            `json:"scheduledFor"`
	Outcome      SiteHeartbeatOutcome `json:"outcome"`
}

// AdminSiteHeartbeatReceipt is the first accepted report for an occurrence.
type AdminSiteHeartbeatReceipt struct {
	Report     SiteHeartbeatReport `json:"report"`
	ReceivedAt time.Time           `json:"receivedAt"`
}

// AdminSiteHeartbeatMonitor contains a monitor and its latest observed state.
type AdminSiteHeartbeatMonitor struct {
	MonitorUUID    string                     `json:"monitorUuid"`
	Name           string                     `json:"name"`
	Schedule       AdminSiteHeartbeatSchedule `json:"schedule"`
	StartedAt      time.Time                  `json:"startedAt"`
	Latest         *AdminSiteHeartbeatReceipt `json:"latest,omitempty"`
	NextExpectedAt time.Time                  `json:"nextExpectedAt"`
	Status         string                     `json:"status"`
}

// AdminSiteHeartbeatCredential holds the one-time credential returned by create or rotate.
type AdminSiteHeartbeatCredential struct {
	Monitor    AdminSiteHeartbeatMonitor `json:"monitor"`
	Token      string                    `json:"token"`
	ReportPath string                    `json:"reportPath"`
}

// AdminSiteHeartbeatList lists the domain's heartbeat monitors without their tokens.
type AdminSiteHeartbeatList struct {
	Monitors []AdminSiteHeartbeatMonitor `json:"monitors"`
}

// ListHeartbeats lists scheduled-job monitors for a domain.
func (c *SiteDomainAdminClient) ListHeartbeats(ctx context.Context, domainUUID string) (*AdminSiteHeartbeatList, error) {
	var result AdminSiteHeartbeatList
	err := c.request(ctx, http.MethodGet, siteHeartbeatPath(domainUUID), nil, &result)
	return &result, err
}

// CreateHeartbeat schedules a monitor and returns its one-time write credential.
func (c *SiteDomainAdminClient) CreateHeartbeat(
	ctx context.Context, domainUUID string, request AdminSiteHeartbeatCreate,
) (*AdminSiteHeartbeatCredential, error) {
	var result AdminSiteHeartbeatCredential
	err := c.request(ctx, http.MethodPost, siteHeartbeatPath(domainUUID), request, &result)
	return &result, err
}

// RotateHeartbeat revokes the current write token and returns its replacement.
func (c *SiteDomainAdminClient) RotateHeartbeat(
	ctx context.Context, domainUUID, monitorUUID string,
) (*AdminSiteHeartbeatCredential, error) {
	var result AdminSiteHeartbeatCredential
	path := siteHeartbeatPath(domainUUID) + "/" + url.PathEscape(monitorUUID) + "/credential"
	err := c.request(ctx, http.MethodPost, path, nil, &result)
	return &result, err
}

// DeleteHeartbeat removes the monitor and revokes its write token.
func (c *SiteDomainAdminClient) DeleteHeartbeat(ctx context.Context, domainUUID, monitorUUID string) error {
	path := siteHeartbeatPath(domainUUID) + "/" + url.PathEscape(monitorUUID)
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

func siteHeartbeatPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/heartbeats"
}
