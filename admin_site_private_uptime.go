package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PrivateUptimeMonitor gives the network label, latest observation, and pending check.
type PrivateUptimeMonitor struct {
	MonitorUUID  string               `json:"monitorUuid"`
	NetworkLabel string               `json:"networkLabel"`
	Latest       *PrivateUptimeReport `json:"latest,omitempty"`
	RequestedAt  *time.Time           `json:"requestedAt,omitempty"`
	Status       string               `json:"status"`
}

// PrivateUptimeCredential is returned when a monitor is created or its token is rotated.
type PrivateUptimeCredential struct {
	Monitor    PrivateUptimeMonitor `json:"monitor"`
	ReportPath string               `json:"reportPath"`
	Token      string               `json:"token"`
}

type createPrivateUptimeRequest struct {
	NetworkLabel string `json:"networkLabel"`
}

// GetPrivateUptime reads a domain's private uptime monitor and its latest result.
func (c *SiteDomainAdminClient) GetPrivateUptime(ctx context.Context, domainUUID string) (*PrivateUptimeMonitor, error) {
	var monitor PrivateUptimeMonitor
	err := c.request(ctx, http.MethodGet, privateUptimePath(domainUUID), nil, &monitor)
	return &monitor, err
}

// CreatePrivateUptime creates a monitor and returns its one-time write credential.
func (c *SiteDomainAdminClient) CreatePrivateUptime(
	ctx context.Context, domainUUID, networkLabel string,
) (*PrivateUptimeCredential, error) {
	var credential PrivateUptimeCredential
	err := c.request(ctx, http.MethodPost, privateUptimePath(domainUUID),
		createPrivateUptimeRequest{NetworkLabel: networkLabel}, &credential)
	return &credential, err
}

// RequestPrivateUptime queues a check for the assigned private probe.
func (c *SiteDomainAdminClient) RequestPrivateUptime(ctx context.Context, domainUUID string) (*PrivateUptimeWork, error) {
	var work PrivateUptimeWork
	err := c.request(ctx, http.MethodPost, privateUptimePath(domainUUID)+"/check", nil, &work)
	return &work, err
}

// RotatePrivateUptime revokes the current write token and returns its replacement.
func (c *SiteDomainAdminClient) RotatePrivateUptime(
	ctx context.Context, domainUUID, monitorUUID string,
) (*PrivateUptimeCredential, error) {
	var credential PrivateUptimeCredential
	path := privateUptimePath(domainUUID) + "/" + url.PathEscape(monitorUUID) + "/credential"
	err := c.request(ctx, http.MethodPost, path, nil, &credential)
	return &credential, err
}

// DeletePrivateUptime removes the monitor and revokes its write token.
func (c *SiteDomainAdminClient) DeletePrivateUptime(ctx context.Context, domainUUID, monitorUUID string) error {
	path := privateUptimePath(domainUUID) + "/" + url.PathEscape(monitorUUID)
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

func privateUptimePath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/private-uptime"
}
