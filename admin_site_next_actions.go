package custd

import (
	"context"
	"net/http"
	"time"
)

// The kinds of AdminSiteNextAction, in the order Custd ranks them: proving ownership, an open
// outage, the certificate, search access, serious findings, first-run setup, then findings worth
// fixing. Custd may add kinds; show one this SDK does not name by its Title and Why.
const (
	AdminSiteNextActionVerify            = "verify"
	AdminSiteNextActionOutage            = "outage"
	AdminSiteNextActionCertificate       = "certificate"
	AdminSiteNextActionReconnectSearch   = "reconnect-search"
	AdminSiteNextActionFinding           = "finding"
	AdminSiteNextActionFirstAuditPending = "first-audit-pending"
	AdminSiteNextActionConnectSearch     = "connect-search"
)

// The places of AdminSiteNextActionTarget.Place: the ownership proof steps, the uptime and
// certificate section, search settings, the latest audit, a finding group (FindingUUID) or one
// page (PageURL). Custd may add places.
const (
	AdminSiteNextActionPlaceOwnership      = "ownership"
	AdminSiteNextActionPlaceUptime         = "uptime"
	AdminSiteNextActionPlaceSearchSettings = "search-settings"
	AdminSiteNextActionPlaceAudit          = "audit"
	AdminSiteNextActionPlaceFinding        = "finding"
	AdminSiteNextActionPlacePage           = "page"
)

// The values of AdminSiteNextActions.NotChecked: what Custd cannot check yet.
const (
	AdminSiteNotCheckedUptime = "uptime"
	AdminSiteNotCheckedAudit  = "audit"
	AdminSiteNotCheckedSearch = "search"
)

// AdminSiteNextActionsMax is how many actions a next-actions read returns.
const AdminSiteNextActionsMax = 3

// AdminSiteNextActions is one domain's status, one plain headline, what is not checked yet and
// its most important actions, most important first. ActionCount counts every action; at most
// AdminSiteNextActionsMax are returned.
type AdminSiteNextActions struct {
	AIMode      SiteAIMode            `json:"aiMode,omitempty"`
	DomainUUID  string                `json:"domainUuid"`
	Host        string                `json:"host"`
	Status      AdminSiteUptimeStatus `json:"status"`
	Headline    string                `json:"headline"`
	CheckedAt   *time.Time            `json:"checkedAt,omitempty"`
	NotChecked  []string              `json:"notChecked"`
	ActionCount int                   `json:"actionCount"`
	Actions     []AdminSiteNextAction `json:"actions"`
}

// AdminSiteNextAction is one thing to do. AffectedCount counts a finding's pages and is 0 for an
// action about the whole site; Rule is set for a finding; Recheckable says Check again (Verify
// now, or a one-page recheck) can confirm the fix at once.
type AdminSiteNextAction struct {
	Kind          string                    `json:"kind"`
	Title         string                    `json:"title"`
	Why           string                    `json:"why"`
	Severity      AdminSiteFindingSeverity  `json:"severity"`
	AffectedCount int                       `json:"affectedCount"`
	Rule          string                    `json:"rule,omitempty"`
	Target        AdminSiteNextActionTarget `json:"target"`
	Recheckable   bool                      `json:"recheckable"`
}

// AdminSiteNextActionTarget is where an action is done, and for a finding its first finding and page.
type AdminSiteNextActionTarget struct {
	Place       string `json:"place"`
	FindingUUID string `json:"findingUuid,omitempty"`
	PageURL     string `json:"pageUrl,omitempty"`
}

// AdminSitePortfolioSummary is every domain the caller may see, worst first.
type AdminSitePortfolioSummary struct {
	Domains []AdminSitePortfolioDomain `json:"domains"`
}

// AdminSitePortfolioDomain is one domain's status and first action, as its next-actions read
// gives them, and its open findings by severity. FirstAction is nil when there is nothing to do.
type AdminSitePortfolioDomain struct {
	DomainUUID   string                 `json:"domainUuid"`
	SiteUUID     string                 `json:"siteUuid"`
	Host         string                 `json:"host"`
	Status       AdminSiteUptimeStatus  `json:"status"`
	Headline     string                 `json:"headline"`
	CheckedAt    *time.Time             `json:"checkedAt,omitempty"`
	ActionCount  int                    `json:"actionCount"`
	OpenFindings AdminSiteFindingCounts `json:"openFindings"`
	FirstAction  *AdminSiteNextAction   `json:"firstAction,omitempty"`
}

// AdminSiteFindingCounts counts a domain's open findings by severity.
type AdminSiteFindingCounts struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Info   int `json:"info"`
}

// NextActions returns the domain's status and its most important next actions. A site-bound
// credential reads only its own Site's domains; another tenant's domain is 404.
func (c *SiteDomainAdminClient) NextActions(ctx context.Context, domainUUID string) (*AdminSiteNextActions, error) {
	var next AdminSiteNextActions
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/next-actions", nil, &next)
	return &next, err
}

// PortfolioSummary returns every domain of the tenant (only its own Site's for a site-bound
// credential), worst first. A site_insights.read credential (ScopeTemplateSiteInsightsRead) may
// call it and NextActions, and nothing else.
func (c *SiteDomainAdminClient) PortfolioSummary(ctx context.Context) (*AdminSitePortfolioSummary, error) {
	var summary AdminSitePortfolioSummary
	err := c.request(ctx, http.MethodGet, siteDomainsPath+"/portfolio-summary", nil, &summary)
	return &summary, err
}
