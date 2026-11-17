package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// The Lighthouse profiles of AdminSiteLighthouseRun.Profile and AdminSiteReportSpeed.Profile.
const (
	AdminSiteLighthouseDesktop = "desktop"
	AdminSiteLighthouseMobile  = "mobile"
)

// The outcomes of AdminSiteLighthouseRun.Outcome.
const (
	AdminSiteLighthouseSuccess = "success"
	AdminSiteLighthouseFailed  = "failed"
)

// AdminSiteUptimeWindow counts checks over one window. UpPercent is nil when there were none.
type AdminSiteUptimeWindow struct {
	Checks    int      `json:"checks"`
	UpChecks  int      `json:"upChecks"`
	UpPercent *float64 `json:"upPercent,omitempty"`
}

// AdminSiteUptimeWindows are the last 24 hours, 7 days and 30 days.
type AdminSiteUptimeWindows struct {
	Last24h AdminSiteUptimeWindow `json:"last24h"`
	Last7d  AdminSiteUptimeWindow `json:"last7d"`
	Last30d AdminSiteUptimeWindow `json:"last30d"`
}

// AdminSiteUptimeDay is one UTC day (YYYY-MM-DD). UpPercent is nil when there were no checks.
type AdminSiteUptimeDay struct {
	Date      string   `json:"date"`
	Checks    int      `json:"checks"`
	UpChecks  int      `json:"upChecks"`
	UpPercent *float64 `json:"upPercent,omitempty"`
}

// AdminSiteResponseHour is one hour's response times; AvgMs and MaxMs are nil when no check
// answered.
type AdminSiteResponseHour struct {
	HourStart time.Time `json:"hourStart"`
	Checks    int       `json:"checks"`
	UpChecks  int       `json:"upChecks"`
	AvgMs     *int      `json:"avgMs,omitempty"`
	MaxMs     *int      `json:"maxMs,omitempty"`
}

// AdminSiteUptimeHistory is a domain's uptime windows, ninety daily bars and hourly response
// times over ResponseDays days. MonitoredSince is nil before the first check.
type AdminSiteUptimeHistory struct {
	DomainUUID     string                  `json:"domainUuid"`
	MonitoredSince *time.Time              `json:"monitoredSince,omitempty"`
	GeneratedAt    time.Time               `json:"generatedAt"`
	ResponseDays   int                     `json:"responseDays"`
	Windows        AdminSiteUptimeWindows  `json:"windows"`
	Days           []AdminSiteUptimeDay    `json:"days"`
	Hours          []AdminSiteResponseHour `json:"hours"`
}

// AdminSiteLighthouseScores are 0..100 category scores; a category Lighthouse did not score is nil.
type AdminSiteLighthouseScores struct {
	Performance   *int `json:"performance,omitempty"`
	Accessibility *int `json:"accessibility,omitempty"`
	BestPractices *int `json:"bestPractices,omitempty"`
	SEO           *int `json:"seo,omitempty"`
}

// AdminSiteLighthouseMetrics are lab metrics; a metric Lighthouse did not measure is nil.
type AdminSiteLighthouseMetrics struct {
	FirstContentfulPaintMs   *int     `json:"firstContentfulPaintMs,omitempty"`
	LargestContentfulPaintMs *int     `json:"largestContentfulPaintMs,omitempty"`
	TotalBlockingTimeMs      *int     `json:"totalBlockingTimeMs,omitempty"`
	SpeedIndexMs             *int     `json:"speedIndexMs,omitempty"`
	CumulativeLayoutShift    *float64 `json:"cumulativeLayoutShift,omitempty"`
}

// AdminSiteLighthouseRun is one scheduled lab run; a failed run says why in Failure.
type AdminSiteLighthouseRun struct {
	Profile     string                     `json:"profile"`
	ObservedAt  time.Time                  `json:"observedAt"`
	Outcome     string                     `json:"outcome"`
	Failure     string                     `json:"failure,omitempty"`
	URL         string                     `json:"url"`
	DurationMs  int                        `json:"durationMs"`
	ToolVersion string                     `json:"toolVersion,omitempty"`
	Scores      AdminSiteLighthouseScores  `json:"scores"`
	Metrics     AdminSiteLighthouseMetrics `json:"metrics"`
}

// AdminSiteLighthouseHistory is a domain's lab runs over the last Days days.
type AdminSiteLighthouseHistory struct {
	DomainUUID string                   `json:"domainUuid"`
	Days       int                      `json:"days"`
	Runs       []AdminSiteLighthouseRun `json:"runs"`
}

// AdminSiteReportUptime is the report's uptime: status, windows and ninety daily bars.
// ResponseMs, DownSince and MonitoredSince are nil when unknown.
type AdminSiteReportUptime struct {
	Status         AdminSiteUptimeStatus  `json:"status"`
	ResponseMs     *int                   `json:"responseMs,omitempty"`
	DownSince      *time.Time             `json:"downSince,omitempty"`
	MonitoredSince *time.Time             `json:"monitoredSince,omitempty"`
	Windows        AdminSiteUptimeWindows `json:"windows"`
	Days           []AdminSiteUptimeDay   `json:"days"`
}

// AdminSiteReportCertificate is the report's TLS certificate; Problem means the last check failed.
type AdminSiteReportCertificate struct {
	ValidTo *time.Time `json:"validTo,omitempty"`
	Issuer  string     `json:"issuer,omitempty"`
	Problem bool       `json:"problem"`
}

// AdminSiteReportSpeed is the latest successful Lighthouse run of one profile.
type AdminSiteReportSpeed struct {
	Profile    string                     `json:"profile"`
	ObservedAt time.Time                  `json:"observedAt"`
	Scores     AdminSiteLighthouseScores  `json:"scores"`
	Metrics    AdminSiteLighthouseMetrics `json:"metrics"`
}

// AdminSiteReportSearch is the report's search section; Connected says Google or Bing is connected.
type AdminSiteReportSearch struct {
	Days             int                     `json:"days"`
	Connected        bool                    `json:"connected"`
	Sources          []AdminSiteSearchSource `json:"sources"`
	TopQueries       []AdminSiteSearchRow    `json:"topQueries"`
	TopPages         []AdminSiteSearchRow    `json:"topPages"`
	StrikingDistance []AdminSiteSearchRow    `json:"strikingDistance"`
}

// AdminSiteReportFinding is one open finding as a client report shows it.
type AdminSiteReportFinding struct {
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Fix         string    `json:"fix"`
	FirstSeenAt time.Time `json:"firstSeenAt"`
}

// AdminSiteReportAudit is the report's audit section; AuditedAt and Summary are nil before the
// first audit.
type AdminSiteReportAudit struct {
	AuditedAt *time.Time               `json:"auditedAt,omitempty"`
	Summary   *AdminSiteAuditSummary   `json:"summary,omitempty"`
	Findings  []AdminSiteReportFinding `json:"findings"`
}

// AdminSiteReport is a domain's client report, without verification or search connection details.
type AdminSiteReport struct {
	Host        string                     `json:"host"`
	GeneratedAt time.Time                  `json:"generatedAt"`
	Uptime      AdminSiteReportUptime      `json:"uptime"`
	Certificate AdminSiteReportCertificate `json:"certificate"`
	Speed       []AdminSiteReportSpeed     `json:"speed"`
	Search      AdminSiteReportSearch      `json:"search"`
	Audit       AdminSiteReportAudit       `json:"audit"`
}

// AdminSiteReportShare is one live public share's metadata; its token is shown only on creation.
type AdminSiteReportShare struct {
	ShareUUID string    `json:"shareUuid"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// AdminSiteReportShareList lists a domain's live report shares.
type AdminSiteReportShareList struct {
	Shares []AdminSiteReportShare `json:"shares"`
}

// AdminSiteReportShareCreated is a new share and its token. Open the report at
// /shared-report#<token>; copy the token now, it is never returned again.
type AdminSiteReportShareCreated struct {
	Share AdminSiteReportShare `json:"share"`
	Token string               `json:"token"`
}

// AdminSiteReportShareRequest sets how many days (1..90) a share works; 0 uses the server's 30.
type AdminSiteReportShareRequest struct {
	ExpiresInDays int `json:"expiresInDays,omitempty"`
}

// GetReport returns the domain's complete client report.
func (c *SiteDomainAdminClient) GetReport(ctx context.Context, domainUUID string) (*AdminSiteReport, error) {
	var report AdminSiteReport
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/report", nil, &report)
	return &report, err
}

// GetUptimeHistory returns uptime windows, ninety daily bars and hourly response times over
// responseDays days (1..90); 0 uses the server default.
func (c *SiteDomainAdminClient) GetUptimeHistory(
	ctx context.Context, domainUUID string, responseDays int,
) (*AdminSiteUptimeHistory, error) {
	var history AdminSiteUptimeHistory
	path := siteDomainPath(domainUUID) + "/uptime-history" + siteDaysQuery("responseDays", responseDays)
	err := c.request(ctx, http.MethodGet, path, nil, &history)
	return &history, err
}

// GetLighthouseHistory returns scheduled desktop and mobile lab runs over days days (1..30); 0
// uses the server's seven days.
func (c *SiteDomainAdminClient) GetLighthouseHistory(
	ctx context.Context, domainUUID string, days int,
) (*AdminSiteLighthouseHistory, error) {
	var history AdminSiteLighthouseHistory
	path := siteDomainPath(domainUUID) + "/lighthouse" + siteDaysQuery("days", days)
	err := c.request(ctx, http.MethodGet, path, nil, &history)
	return &history, err
}

// ListReportShares returns the domain's live report shares, without their tokens.
func (c *SiteDomainAdminClient) ListReportShares(ctx context.Context, domainUUID string) (*AdminSiteReportShareList, error) {
	var shares AdminSiteReportShareList
	err := c.request(ctx, http.MethodGet, siteReportSharesPath(domainUUID), nil, &shares)
	return &shares, err
}

// CreateReportShare creates an explicit public share. Copy the token once; creation is never
// retried automatically.
func (c *SiteDomainAdminClient) CreateReportShare(
	ctx context.Context, domainUUID string, req AdminSiteReportShareRequest,
) (*AdminSiteReportShareCreated, error) {
	var created AdminSiteReportShareCreated
	err := c.request(ctx, http.MethodPost, siteReportSharesPath(domainUUID), req, &created)
	return &created, err
}

// RevokeReportShare revokes a share; later public reads of it fail.
func (c *SiteDomainAdminClient) RevokeReportShare(ctx context.Context, domainUUID, shareUUID string) error {
	path := siteReportSharesPath(domainUUID) + "/" + url.PathEscape(shareUUID)
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

func siteReportSharesPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/report-shares"
}

func siteDaysQuery(name string, days int) string {
	if days <= 0 {
		return ""
	}
	return "?" + url.Values{name: {strconv.Itoa(days)}}.Encode()
}
