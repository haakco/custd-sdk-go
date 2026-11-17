package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// The values of AdminSiteAuditPage.Indexability. Compare against these rather than assuming the
// set is closed.
const (
	AdminSiteAuditPageIndexable     = "indexable"
	AdminSiteAuditPageNoindex       = "noindex"
	AdminSiteAuditPageCanonicalized = "canonicalized"
	AdminSiteAuditPageRedirected    = "redirected"
	AdminSiteAuditPageError         = "error"
	AdminSiteAuditPageNotHTML       = "not_html"
)

// The paging bounds of ListAuditPages.
const (
	AdminSiteAuditPagesDefaultLimit = 50
	AdminSiteAuditPagesMaxLimit     = 200
)

// The values of AdminSiteAuditDelta.Run.
const (
	AdminSiteAuditRunNotAudited = "not_audited"
	AdminSiteAuditRunFailed     = "failed"
	AdminSiteAuditRunTruncated  = "truncated"
	AdminSiteAuditRunComplete   = "complete"
)

// AdminSiteLighthouseMaxExtraPages is how many pages beyond the home page Lighthouse measures.
const AdminSiteLighthouseMaxExtraPages = 5

// SiteLighthousePageNotExtraCode is the problem code a 400 carries when a Lighthouse page is not
// on the domain's host or is its home page.
const SiteLighthousePageNotExtraCode = "site_lighthouse_page_not_extra"

// AdminSiteAuditPage is one URL the last audit crawled. FinalURL is empty unless it redirected;
// Status is 0 when the request failed. Depth and InboundLinks are nil when not known.
type AdminSiteAuditPage struct {
	URL          string `json:"url"`
	FinalURL     string `json:"finalUrl,omitempty"`
	Status       int    `json:"status"`
	Indexability string `json:"indexability"`
	Depth        *int   `json:"depth,omitempty"`
	InboundLinks *int   `json:"inboundLinks,omitempty"`
	OpenFindings int    `json:"openFindings"`
}

// AdminSiteAuditPageList is one page of the last audit's crawled URLs. AuditedAt is nil before
// the first audit; Truncated means the crawl stopped at its bound.
type AdminSiteAuditPageList struct {
	DomainUUID string               `json:"domainUuid"`
	AuditedAt  *time.Time           `json:"auditedAt,omitempty"`
	Truncated  bool                 `json:"truncated"`
	Offset     int                  `json:"offset"`
	Limit      int                  `json:"limit"`
	Total      int                  `json:"total"`
	Pages      []AdminSiteAuditPage `json:"pages"`
}

// AdminSiteAuditDelta is what the last audit changed: findings it opened, reopened and fixed.
// A failed run says why in Failure.
type AdminSiteAuditDelta struct {
	DomainUUID string             `json:"domainUuid"`
	Run        string             `json:"run"`
	AuditedAt  *time.Time         `json:"auditedAt,omitempty"`
	Failure    string             `json:"failure,omitempty"`
	New        []AdminSiteFinding `json:"new"`
	Reopened   []AdminSiteFinding `json:"reopened"`
	Fixed      []AdminSiteFinding `json:"fixed"`
}

// AdminSiteLighthousePageComparison is one measured page and profile's latest score and the one
// before it; PreviousAt and PreviousScore are nil after a single run.
type AdminSiteLighthousePageComparison struct {
	URL            string     `json:"url"`
	Profile        string     `json:"profile"`
	LatestAt       time.Time  `json:"latestAt"`
	LatestScore    int        `json:"latestScore"`
	PreviousAt     *time.Time `json:"previousAt,omitempty"`
	PreviousScore  *int       `json:"previousScore,omitempty"`
	RegressionOpen bool       `json:"regressionOpen"`
}

// AdminSiteLighthousePages are the extra pages Lighthouse measures beside HomePage.
type AdminSiteLighthousePages struct {
	DomainUUID string   `json:"domainUuid"`
	HomePage   string   `json:"homePage"`
	URLs       []string `json:"urls"`
	MaxPages   int      `json:"maxPages"`
}

type adminSiteLighthousePagesRequest struct {
	URLs []string `json:"urls"`
}

// ListAuditPages returns one page of the last audit's crawled URLs. A limit of 0 uses the
// server's 50 (at most AdminSiteAuditPagesMaxLimit); offset 0 starts at the first URL.
func (c *SiteDomainAdminClient) ListAuditPages(
	ctx context.Context, domainUUID string, offset, limit int,
) (*AdminSiteAuditPageList, error) {
	var pages AdminSiteAuditPageList
	path := siteQueryPath(siteDomainPath(domainUUID)+"/audit-pages", url.Values{
		"offset": {sitePositiveInt(offset)}, "limit": {sitePositiveInt(limit)},
	})
	err := c.request(ctx, http.MethodGet, path, nil, &pages)
	return &pages, err
}

// GetAuditDelta returns the findings the last audit opened, reopened and fixed.
func (c *SiteDomainAdminClient) GetAuditDelta(ctx context.Context, domainUUID string) (*AdminSiteAuditDelta, error) {
	var delta AdminSiteAuditDelta
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/audit-delta", nil, &delta)
	return &delta, err
}

// GetLighthousePages returns the extra pages Lighthouse measures.
func (c *SiteDomainAdminClient) GetLighthousePages(ctx context.Context, domainUUID string) (*AdminSiteLighthousePages, error) {
	var pages AdminSiteLighthousePages
	err := c.request(ctx, http.MethodGet, siteLighthousePagesPath(domainUUID), nil, &pages)
	return &pages, err
}

// SetLighthousePages replaces the extra pages (at most AdminSiteLighthouseMaxExtraPages, on the
// domain's host, not its home page); an empty list clears them.
func (c *SiteDomainAdminClient) SetLighthousePages(
	ctx context.Context, domainUUID string, urls []string,
) (*AdminSiteLighthousePages, error) {
	var pages AdminSiteLighthousePages
	err := c.request(ctx, http.MethodPut, siteLighthousePagesPath(domainUUID),
		adminSiteLighthousePagesRequest{URLs: nonNilSlice(urls)}, &pages)
	return &pages, err
}

func siteLighthousePagesPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/lighthouse/pages"
}
