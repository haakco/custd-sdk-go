package custd

import (
	"context"
	"net/http"
	"time"
)

// AdminSiteCheckMode is how a verified domain is checked. Custd cannot tell that a
// site is stopped, so whoever stops a site sets its domain to paused; nothing then
// requests it until the mode is set back.
type AdminSiteCheckMode string

const (
	// AdminSiteCheckModePaused sends no requests: no uptime, certificate or warming check.
	AdminSiteCheckModePaused AdminSiteCheckMode = "paused"
	// AdminSiteCheckModeUptime checks the home page every five minutes and the
	// certificate every six hours. It is the default.
	AdminSiteCheckModeUptime AdminSiteCheckMode = "uptime"
	// AdminSiteCheckModeWarm adds one anonymous home page GET every minute.
	AdminSiteCheckModeWarm AdminSiteCheckMode = "warm"
)

// AdminSiteUptimeStatus is the uptime state a domain shows.
type AdminSiteUptimeStatus string

const (
	// AdminSiteUptimeUnknown means no check in the last fifteen minutes, or none yet.
	AdminSiteUptimeUnknown AdminSiteUptimeStatus = "unknown"
	AdminSiteUptimeUp      AdminSiteUptimeStatus = "up"
	// AdminSiteUptimeFailing is one failed check, not yet an outage.
	AdminSiteUptimeFailing AdminSiteUptimeStatus = "failing"
	// AdminSiteUptimeDown is an open incident: two or more failed checks in a row.
	AdminSiteUptimeDown   AdminSiteUptimeStatus = "down"
	AdminSiteUptimePaused AdminSiteUptimeStatus = "paused"
)

// AdminSiteUptime is the latest home page check.
type AdminSiteUptime struct {
	Status              AdminSiteUptimeStatus `json:"status"`
	LastCheckedAt       *time.Time            `json:"lastCheckedAt,omitempty"`
	StatusCode          int                   `json:"statusCode,omitempty"`
	ResponseMs          int                   `json:"responseMs,omitempty"`
	LastFailure         string                `json:"lastFailure,omitempty"`
	ConsecutiveFailures int                   `json:"consecutiveFailures"`
}

// AdminSiteTLS is the latest certificate check. LastFailure says why the
// certificate is not valid, or why the check could not reach the host.
type AdminSiteTLS struct {
	LastCheckedAt *time.Time `json:"lastCheckedAt,omitempty"`
	ValidTo       *time.Time `json:"validTo,omitempty"`
	Issuer        string     `json:"issuer,omitempty"`
	LastFailure   string     `json:"lastFailure,omitempty"`
}

// AdminSiteWarming is the average home page response time measured while warming
// was off and while it was on. Either is nil until measured.
type AdminSiteWarming struct {
	LastWarmedAt       *time.Time `json:"lastWarmedAt,omitempty"`
	UnwarmedResponseMs *int       `json:"unwarmedResponseMs,omitempty"`
	WarmedResponseMs   *int       `json:"warmedResponseMs,omitempty"`
}

// AdminSiteIncident is one outage, from the first failed check until the first check
// that passed. ResolvedAt is nil while it is open.
type AdminSiteIncident struct {
	IncidentUUID string     `json:"incidentUuid"`
	StartedAt    time.Time  `json:"startedAt"`
	OpenedAt     time.Time  `json:"openedAt"`
	ResolvedAt   *time.Time `json:"resolvedAt,omitempty"`
	Cause        string     `json:"cause"`
	FailedChecks int        `json:"failedChecks"`
}

// AdminSiteIncidentList lists a domain's recent outages, newest first.
type AdminSiteIncidentList struct {
	Incidents []AdminSiteIncident `json:"incidents"`
}

// AdminSiteDomainChecks is a domain's uptime, certificate and warming state. The
// problems the checks find are also findings, under the uptime.* and tls.* rules.
type AdminSiteDomainChecks struct {
	Mode         AdminSiteCheckMode `json:"mode"`
	Uptime       AdminSiteUptime    `json:"uptime"`
	TLS          AdminSiteTLS       `json:"tls"`
	Warming      AdminSiteWarming   `json:"warming"`
	OpenIncident *AdminSiteIncident `json:"openIncident,omitempty"`
}

type adminSiteCheckSettings struct {
	Mode AdminSiteCheckMode `json:"mode"`
}

// SetCheckMode chooses how a domain is checked and returns the domain. It works
// before verification too, so a consumer can pause a domain before its first check.
func (c *SiteDomainAdminClient) SetCheckMode(
	ctx context.Context, domainUUID string, mode AdminSiteCheckMode,
) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodPut, siteDomainPath(domainUUID)+"/checks", adminSiteCheckSettings{Mode: mode}, &domain)
	return &domain, err
}

// ListIncidents returns a domain's recent outages, newest first (at most twenty).
func (c *SiteDomainAdminClient) ListIncidents(ctx context.Context, domainUUID string) (*AdminSiteIncidentList, error) {
	var incidents AdminSiteIncidentList
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/incidents", nil, &incidents)
	return &incidents, err
}
