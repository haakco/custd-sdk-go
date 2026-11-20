package custd

import (
	"context"
	"net/http"
	"net/url"
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
	IncidentUUID   string     `json:"incidentUuid"`
	StartedAt      time.Time  `json:"startedAt"`
	OpenedAt       time.Time  `json:"openedAt"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	Cause          string     `json:"cause"`
	FailedChecks   int        `json:"failedChecks"`
}

// AdminSiteIncidentList lists a domain's recent outages, newest first.
type AdminSiteIncidentList struct {
	Incidents []AdminSiteIncident `json:"incidents"`
}

// AdminSiteDomainChecks is a domain's uptime, certificate and warming state. The
// problems the checks find are also findings, under the uptime.* and tls.* rules.
type AdminSiteDomainChecks struct {
	Mode             AdminSiteCheckMode         `json:"mode"`
	Uptime           AdminSiteUptime            `json:"uptime"`
	TLS              AdminSiteTLS               `json:"tls"`
	Warming          AdminSiteWarming           `json:"warming"`
	OpenIncident     *AdminSiteIncident         `json:"openIncident,omitempty"`
	ChromeUX         *AdminSiteChromeUX         `json:"chromeUx,omitempty"`
	HTTPExpectations *AdminSiteHTTPExpectations `json:"httpExpectations,omitempty"`
	Registration     *AdminSiteRegistration     `json:"registration,omitempty"`
}

// AdminSiteRegistration records the registry lookup. A missing ExpiresAt means
// the registry did not provide an expiry date.
type AdminSiteRegistration struct {
	Domain    string     `json:"domain"`
	Source    string     `json:"source,omitempty"`
	ReadAt    time.Time  `json:"readAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// AdminSiteChromeUX is the last successful Chrome UX Report lookup. A nil observation
// means the provider returned no field data for the domain's origin.
type AdminSiteChromeUX struct {
	CheckedAt   time.Time                     `json:"checkedAt"`
	Observation *AdminSiteChromeUXObservation `json:"observation,omitempty"`
}

// AdminSiteChromeUXObservation contains the available 75th-percentile field metrics.
type AdminSiteChromeUXObservation struct {
	FirstDate time.Time `json:"firstDate"`
	LastDate  time.Time `json:"lastDate"`
	LCPMs     *float64  `json:"lcpMs,omitempty"`
	INPMs     *float64  `json:"inpMs,omitempty"`
	CLS       *float64  `json:"cls,omitempty"`
}

// AdminSiteHTTPExpectations configures optional checks on the final home page response.
type AdminSiteHTTPExpectations struct {
	StatusCode    int                           `json:"statusCode,omitempty"`
	MaxResponseMs int                           `json:"maxResponseMs,omitempty"`
	BodyContains  string                        `json:"bodyContains,omitempty"`
	JSON          *AdminSiteHTTPJSONExpectation `json:"json,omitempty"`
}

// AdminSiteHTTPJSONExpectation matches one top-level JSON string field.
type AdminSiteHTTPJSONExpectation struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// AdminSiteCheckSettings sets the check mode and optionally updates HTTP expectations.
// Nil expectations preserve existing settings. A pointer to an empty value clears them.
type AdminSiteCheckSettings struct {
	Mode             AdminSiteCheckMode         `json:"mode"`
	HTTPExpectations *AdminSiteHTTPExpectations `json:"httpExpectations,omitempty"`
}

// SiteUptimeCheckInactiveCode identifies a request refused because public checks are inactive.
const SiteUptimeCheckInactiveCode = "site_uptime_check_inactive"

// SiteUptimeCheckMaintenanceCode identifies a request refused during active maintenance.
const SiteUptimeCheckMaintenanceCode = "site_uptime_check_maintenance"

// RequestUptimeCheck queues a public uptime check for a verified domain with active checks.
// The result is the current domain snapshot. Read it again after the scheduler runs.
func (c *SiteDomainAdminClient) RequestUptimeCheck(ctx context.Context, domainUUID string) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/uptime-check", nil, &domain)
	return &domain, err
}

// SetCheckMode chooses how a domain is checked and returns the domain. It works
// before verification too, so a consumer can pause a domain before its first check.
func (c *SiteDomainAdminClient) SetCheckMode(
	ctx context.Context, domainUUID string, mode AdminSiteCheckMode,
) (*AdminSiteDomain, error) {
	return c.SetCheckSettings(ctx, domainUUID, AdminSiteCheckSettings{Mode: mode})
}

// SetCheckSettings updates the check mode and any supplied HTTP expectations.
func (c *SiteDomainAdminClient) SetCheckSettings(
	ctx context.Context, domainUUID string, settings AdminSiteCheckSettings,
) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodPut, siteDomainPath(domainUUID)+"/checks", settings, &domain)
	return &domain, err
}

// ListIncidents returns a domain's recent outages, newest first (at most twenty).
func (c *SiteDomainAdminClient) ListIncidents(ctx context.Context, domainUUID string) (*AdminSiteIncidentList, error) {
	var incidents AdminSiteIncidentList
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/incidents", nil, &incidents)
	return &incidents, err
}

// AcknowledgeIncident records that a person has seen the open incident without resolving it.
func (c *SiteDomainAdminClient) AcknowledgeIncident(
	ctx context.Context, domainUUID, incidentUUID string,
) (*AdminSiteIncident, error) {
	var incident AdminSiteIncident
	path := siteDomainPath(domainUUID) + "/incidents/" + url.PathEscape(incidentUUID) + "/acknowledgment"
	err := c.request(ctx, http.MethodPost, path, nil, &incident)
	return &incident, err
}
