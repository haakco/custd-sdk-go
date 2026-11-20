package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteMaintenanceRequest schedules a window. Reuse WindowUUID when retrying the same request.
type AdminSiteMaintenanceRequest struct {
	WindowUUID string    `json:"windowUuid"`
	StartsAt   time.Time `json:"startsAt"`
	EndsAt     time.Time `json:"endsAt"`
}

// AdminSiteMaintenanceWindow is a scheduled or historical window.
type AdminSiteMaintenanceWindow struct {
	WindowUUID  string     `json:"windowUuid"`
	StartsAt    time.Time  `json:"startsAt"`
	EndsAt      time.Time  `json:"endsAt"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`
}

// AdminSiteMaintenanceWindows lists the current and historical windows for one domain.
type AdminSiteMaintenanceWindows struct {
	Windows []AdminSiteMaintenanceWindow `json:"windows"`
}

// ListMaintenance returns the domain's maintenance windows.
func (c *SiteDomainAdminClient) ListMaintenance(ctx context.Context, domainUUID string) (*AdminSiteMaintenanceWindows, error) {
	var result AdminSiteMaintenanceWindows
	err := c.request(ctx, http.MethodGet, siteMaintenancePath(domainUUID), nil, &result)
	return &result, err
}

// ScheduleMaintenance schedules a one-off window on a verified domain.
func (c *SiteDomainAdminClient) ScheduleMaintenance(
	ctx context.Context, domainUUID string, request AdminSiteMaintenanceRequest,
) (*AdminSiteMaintenanceWindow, error) {
	var result AdminSiteMaintenanceWindow
	err := c.request(ctx, http.MethodPost, siteMaintenancePath(domainUUID), request, &result)
	return &result, err
}

// CancelMaintenance cancels a scheduled or active window and returns its historical record.
func (c *SiteDomainAdminClient) CancelMaintenance(
	ctx context.Context, domainUUID, windowUUID string,
) (*AdminSiteMaintenanceWindow, error) {
	var result AdminSiteMaintenanceWindow
	path := siteMaintenancePath(domainUUID) + "/" + url.PathEscape(windowUUID)
	err := c.request(ctx, http.MethodDelete, path, nil, &result)
	return &result, err
}

func siteMaintenancePath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/maintenance"
}
