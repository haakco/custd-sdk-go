package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// siteDomainsPath is the site ownership collection. It lives on the top-level
// /api/v1 surface, not under /api/v1/admin.
const siteDomainsPath = "/site-domains"

// SiteDomainNotVerifiedCode is the problem code a 409 carries when an audit is
// requested for a domain whose tenant has not yet proven control of it. Branch on
// it with errors.As(err, &requestErr) and requestErr.Problem.Code.
const SiteDomainNotVerifiedCode = "site_domain_not_verified"

// AdminSiteFindingState selects which findings a list returns.
type AdminSiteFindingState string

const (
	AdminSiteFindingStateOpen  AdminSiteFindingState = "open"
	AdminSiteFindingStateFixed AdminSiteFindingState = "fixed"
	AdminSiteFindingStateAll   AdminSiteFindingState = "all"
)

// AdminSiteDomainCreate claims a host for one of the tenant's sites.
type AdminSiteDomainCreate struct {
	SiteUUID string `json:"siteUuid"`
	Host     string `json:"host"`
}

// AdminSiteDomainVerification tells the tenant how to prove control: publish either
// the DNS TXT record or the well-known file. Either one is enough.
type AdminSiteDomainVerification struct {
	DNSRecordName  string `json:"dnsRecordName"`
	DNSRecordValue string `json:"dnsRecordValue"`
	FileURL        string `json:"fileUrl"`
	FileContent    string `json:"fileContent"`
}

// AdminSiteDomainOwnership is the proof of control and what the last check saw.
type AdminSiteDomainOwnership struct {
	Verified         bool                        `json:"verified"`
	VerifiedAt       *time.Time                  `json:"verifiedAt,omitempty"`
	VerifiedBy       string                      `json:"verifiedBy,omitempty"`
	LastCheckedAt    *time.Time                  `json:"lastCheckedAt,omitempty"`
	LastCheckFailure string                      `json:"lastCheckFailure,omitempty"`
	Instructions     AdminSiteDomainVerification `json:"instructions"`
}

// AdminSiteAuditSummary counts the findings of one audit run by severity.
type AdminSiteAuditSummary struct {
	High         int  `json:"high"`
	Medium       int  `json:"medium"`
	Low          int  `json:"low"`
	Info         int  `json:"info"`
	PagesChecked int  `json:"pagesChecked"`
	Requests     int  `json:"requests"`
	Truncated    bool `json:"truncated"`
}

// AdminSiteDomainAudit is the scheduled audit's state for one domain.
type AdminSiteDomainAudit struct {
	RequestedAt   *time.Time             `json:"requestedAt,omitempty"`
	LastAuditedAt *time.Time             `json:"lastAuditedAt,omitempty"`
	LastSummary   *AdminSiteAuditSummary `json:"lastSummary,omitempty"`
	LastFailure   string                 `json:"lastFailure,omitempty"`
}

// AdminSiteDomain is a host a site has claimed, whether its tenant has proven
// control, its latest audit and its uptime, certificate and warming checks. Only a
// verified domain is audited or checked.
type AdminSiteDomain struct {
	DomainUUID string                   `json:"domainUuid"`
	SiteUUID   string                   `json:"siteUuid"`
	Host       string                   `json:"host"`
	Ownership  AdminSiteDomainOwnership `json:"ownership"`
	Audit      AdminSiteDomainAudit     `json:"audit"`
	Checks     AdminSiteDomainChecks    `json:"checks"`
	CreatedAt  time.Time                `json:"createdAt"`
}

// AdminSiteDomainList lists one site's domains.
type AdminSiteDomainList struct {
	Domains []AdminSiteDomain `json:"domains"`
}

// AdminSiteFindingHistory says when a finding was seen and whether it is fixed.
type AdminSiteFindingHistory struct {
	FirstSeenAt time.Time  `json:"firstSeenAt"`
	LastSeenAt  time.Time  `json:"lastSeenAt"`
	Fixed       bool       `json:"fixed"`
	FixedAt     *time.Time `json:"fixedAt,omitempty"`
}

// AdminSiteFinding is one problem an audit found on a verified domain, with the
// advice to fix it.
type AdminSiteFinding struct {
	FindingUUID string                  `json:"findingUuid"`
	Rule        string                  `json:"rule"`
	Severity    string                  `json:"severity"`
	Title       string                  `json:"title"`
	URL         string                  `json:"url"`
	Evidence    string                  `json:"evidence"`
	Fix         string                  `json:"fix"`
	Source      string                  `json:"source"`
	History     AdminSiteFindingHistory `json:"history"`
}

// AdminSiteFindingList lists a domain's findings, most severe first.
type AdminSiteFindingList struct {
	Findings []AdminSiteFinding `json:"findings"`
}

// SiteDomainAdminClient claims, verifies and audits the domains a site serves. A
// system administrator names the tenant with ClientConfig.ActingCompanySlug.
type SiteDomainAdminClient struct {
	admin *AdminClient
}

// Create claims a host for a site and returns the verification instructions.
func (c *SiteDomainAdminClient) Create(ctx context.Context, req AdminSiteDomainCreate) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodPost, siteDomainsPath, req, &domain)
	return &domain, err
}

// List returns one site's domains.
func (c *SiteDomainAdminClient) List(ctx context.Context, siteUUID string) (*AdminSiteDomainList, error) {
	var domains AdminSiteDomainList
	path := siteDomainsPath + "?" + url.Values{"siteUuid": {siteUUID}}.Encode()
	err := c.request(ctx, http.MethodGet, path, nil, &domains)
	return &domains, err
}

// Get returns one domain.
func (c *SiteDomainAdminClient) Get(ctx context.Context, domainUUID string) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID), nil, &domain)
	return &domain, err
}

// Verify checks the domain's DNS record or well-known file now and returns the
// result. An unproven domain comes back with Ownership.Verified false and the
// reason in Ownership.LastCheckFailure.
func (c *SiteDomainAdminClient) Verify(ctx context.Context, domainUUID string) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/verify", nil, &domain)
	return &domain, err
}

// RequestAudit asks for the domain to be audited on the scheduler's next tick. An
// unverified domain is refused with a 409 carrying SiteDomainNotVerifiedCode.
func (c *SiteDomainAdminClient) RequestAudit(ctx context.Context, domainUUID string) (*AdminSiteDomain, error) {
	var domain AdminSiteDomain
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/audit", nil, &domain)
	return &domain, err
}

// ListFindings returns a domain's findings in the given state. An empty state lets
// the server default to open findings.
func (c *SiteDomainAdminClient) ListFindings(
	ctx context.Context, domainUUID string, state AdminSiteFindingState,
) (*AdminSiteFindingList, error) {
	var findings AdminSiteFindingList
	path := siteDomainPath(domainUUID) + "/findings"
	if state != "" {
		path += "?" + url.Values{"state": {string(state)}}.Encode()
	}
	err := c.request(ctx, http.MethodGet, path, nil, &findings)
	return &findings, err
}

// Delete removes a domain and its findings.
func (c *SiteDomainAdminClient) Delete(ctx context.Context, domainUUID string) error {
	return c.request(ctx, http.MethodDelete, siteDomainPath(domainUUID), nil, nil)
}

func (c *SiteDomainAdminClient) request(ctx context.Context, method, path string, payload, out any) error {
	return c.admin.requestNonAdmin(ctx, method, c.admin.actingCompanyPath(path), payload, out)
}

func siteDomainPath(domainUUID string) string {
	return siteDomainsPath + "/" + url.PathEscape(domainUUID)
}
