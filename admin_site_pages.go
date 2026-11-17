package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// What a finding's fix changes, in AdminSiteRemedy.Kind. A rule Custd has no remedy for carries
// an empty kind, so compare against these rather than assuming the set is closed.
const (
	AdminSiteRemedyKindMarkup  = "markup"
	AdminSiteRemedyKindSetting = "setting"
	AdminSiteRemedyKindContent = "content"
	AdminSiteRemedyKindServer  = "server"
)

// The formats of AdminSiteGeneratedFix.Format.
const (
	AdminSiteGeneratedFormatHTML      = "html"
	AdminSiteGeneratedFormatRobotsTxt = "robots.txt"
)

// The states of AdminSitePageRecheck.State.
const (
	AdminSitePageRecheckQueued  = "queued"
	AdminSitePageRecheckChecked = "checked"
	AdminSitePageRecheckFailed  = "failed"
)

// The sources of AdminSiteRobotsTestResult.Source: the robots.txt the caller pasted, or the
// domain's live file.
const (
	AdminSiteRobotsSourcePasted = "pasted"
	AdminSiteRobotsSourceLive   = "live"
)

// SitePageNotOnDomainCode is the problem code a 400 carries when a page tool is given a URL on
// another host than the domain's own.
const SitePageNotOnDomainCode = "site_page_not_on_domain"

// SitePageRecheckLimitCode is the problem code a 429 carries once the domain has used its hourly
// page rechecks.
const SitePageRecheckLimitCode = "site_page_recheck_limit"

// AdminSiteRemedy says what a finding's fix changes; the recommendation itself is the finding's
// Fix text. Generated is set only for rules with a generator: code to copy, never a published
// change.
type AdminSiteRemedy struct {
	Kind      string                 `json:"kind"`
	Element   string                 `json:"element"`
	Generated *AdminSiteGeneratedFix `json:"generated,omitempty"`
}

// AdminSiteGeneratedFix is ready-to-copy code for one finding.
type AdminSiteGeneratedFix struct {
	Format string `json:"format"`
	Code   string `json:"code"`
}

// AdminSitePageFacts is what the latest audit, recheck or draft check read from one page.
type AdminSitePageFacts struct {
	URL         string                   `json:"url"`
	ObservedAt  time.Time                `json:"observedAt"`
	Status      int                      `json:"status"`
	Lang        string                   `json:"lang"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Canonical   string                   `json:"canonical"`
	Robots      AdminSitePageRobots      `json:"robots"`
	OpenGraph   AdminSitePageOpenGraph   `json:"openGraph"`
	Twitter     AdminSitePageTwitter     `json:"twitter"`
	Hreflang    []AdminSiteHreflangLink  `json:"hreflang"`
	Headings    []AdminSitePageHeading   `json:"headings"`
	WordCount   int                      `json:"wordCount"`
	Readability AdminSitePageReadability `json:"readability"`
}

// AdminSitePageRobots is the page's robots meta and X-Robots-Tag header.
type AdminSitePageRobots struct {
	Noindex  bool   `json:"noindex"`
	Nofollow bool   `json:"nofollow"`
	Header   string `json:"header"`
}

// AdminSitePageOpenGraph is the page's og: tags, as written.
type AdminSitePageOpenGraph struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	SiteName    string `json:"siteName"`
}

// AdminSitePageTwitter is the page's twitter: tags, as written.
type AdminSitePageTwitter struct {
	Card        string `json:"card"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Site        string `json:"site"`
}

// AdminSiteHreflangLink is one alternate-language link.
type AdminSiteHreflangLink struct {
	Code string `json:"code"`
	Href string `json:"href"`
}

// AdminSitePageHeading is one heading, level 1 to 6.
type AdminSitePageHeading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// AdminSitePageReadability measures the page's visible text. FleschReadingEase is nil when the
// page is not English or has no words.
type AdminSitePageReadability struct {
	Sentences          int      `json:"sentences"`
	WordsPerSentence   float64  `json:"wordsPerSentence"`
	ReadingTimeSeconds int      `json:"readingTimeSeconds"`
	FleschReadingEase  *float64 `json:"fleschReadingEase,omitempty"`
}

// AdminSitePageRecheck is one queued check of one page. CheckedAt is nil until the scheduler
// checks it; OpenRules and FixedRules are the one-page rules it opened and fixed.
type AdminSitePageRecheck struct {
	RecheckUUID string     `json:"recheckUuid"`
	URL         string     `json:"url"`
	State       string     `json:"state"`
	RequestedAt time.Time  `json:"requestedAt"`
	CheckedAt   *time.Time `json:"checkedAt,omitempty"`
	Failure     string     `json:"failure,omitempty"`
	OpenRules   []string   `json:"openRules"`
	FixedRules  []string   `json:"fixedRules"`
}

// AdminSitePageFinding is one problem a draft check found on a page. Unlike AdminSiteFinding it is
// not stored, so it has no UUID or history.
type AdminSitePageFinding struct {
	Rule     string          `json:"rule"`
	Severity string          `json:"severity"`
	Title    string          `json:"title"`
	URL      string          `json:"url"`
	Evidence string          `json:"evidence"`
	Fix      string          `json:"fix"`
	Source   string          `json:"source"`
	Remedy   AdminSiteRemedy `json:"remedy"`
}

// AdminSitePageCheck is what the page checks found on draft HTML, and the page's facts.
type AdminSitePageCheck struct {
	Page     AdminSitePageFacts     `json:"page"`
	Findings []AdminSitePageFinding `json:"findings"`
}

// AdminSitePageCheckRequest checks HTML as if URL, on the domain's host, served it. HTML is the
// full page, at most 512 KiB; it is never stored.
type AdminSitePageCheckRequest struct {
	URL  string `json:"url"`
	HTML string `json:"html"`
}

// AdminSiteRobotsTestRequest asks whether UserAgent may crawl URL (a path or an http(s) URL).
// An empty RobotsTxt tests the domain's live robots.txt.
type AdminSiteRobotsTestRequest struct {
	URL       string `json:"url"`
	UserAgent string `json:"userAgent"`
	RobotsTxt string `json:"robotsTxt,omitempty"`
}

// AdminSiteRobotsRule is the robots.txt rule that decided a test.
type AdminSiteRobotsRule struct {
	Line      int    `json:"line"`
	Directive string `json:"directive"`
	Pattern   string `json:"pattern"`
}

// AdminSiteRobotsTestResult is the RFC 9309 answer. RobotsStatus is nil for a pasted file or an
// unreachable live one; Rule is nil when no rule matched; Note says why a live file decided
// without a rule.
type AdminSiteRobotsTestResult struct {
	Allowed      bool                 `json:"allowed"`
	UserAgent    string               `json:"userAgent"`
	Path         string               `json:"path"`
	Source       string               `json:"source"`
	RobotsStatus *int                 `json:"robotsStatus,omitempty"`
	Groups       []string             `json:"groups"`
	Rule         *AdminSiteRobotsRule `json:"rule,omitempty"`
	Note         string               `json:"note,omitempty"`
}

type adminSitePageRecheckRequest struct {
	URL string `json:"url"`
}

// GetPageFacts returns what the latest audit or recheck read from one page of the domain.
func (c *SiteDomainAdminClient) GetPageFacts(ctx context.Context, domainUUID, pageURL string) (*AdminSitePageFacts, error) {
	var facts AdminSitePageFacts
	path := siteDomainPath(domainUUID) + "/page-facts?" + url.Values{"url": {pageURL}}.Encode()
	err := c.request(ctx, http.MethodGet, path, nil, &facts)
	return &facts, err
}

// RequestPageRecheck queues one page of a verified domain for the scheduler's next tick (202). A
// page already waiting comes back as it is; past the hourly allowance the server answers 429
// with SitePageRecheckLimitCode.
func (c *SiteDomainAdminClient) RequestPageRecheck(
	ctx context.Context, domainUUID, pageURL string,
) (*AdminSitePageRecheck, error) {
	var recheck AdminSitePageRecheck
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/page-rechecks",
		adminSitePageRecheckRequest{URL: pageURL}, &recheck)
	return &recheck, err
}

// GetPageRecheck returns one recheck, to poll until its State is checked or failed.
func (c *SiteDomainAdminClient) GetPageRecheck(
	ctx context.Context, domainUUID, recheckUUID string,
) (*AdminSitePageRecheck, error) {
	var recheck AdminSitePageRecheck
	path := siteDomainPath(domainUUID) + "/page-rechecks/" + url.PathEscape(recheckUUID)
	err := c.request(ctx, http.MethodGet, path, nil, &recheck)
	return &recheck, err
}

// CheckPage runs the one-page checks on draft HTML without fetching or storing anything.
func (c *SiteDomainAdminClient) CheckPage(
	ctx context.Context, domainUUID string, req AdminSitePageCheckRequest,
) (*AdminSitePageCheck, error) {
	var check AdminSitePageCheck
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/page-check", req, &check)
	return &check, err
}

// TestRobots answers whether a crawler may fetch a URL under a pasted or the live robots.txt.
func (c *SiteDomainAdminClient) TestRobots(
	ctx context.Context, domainUUID string, req AdminSiteRobotsTestRequest,
) (*AdminSiteRobotsTestResult, error) {
	var result AdminSiteRobotsTestResult
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/robots-test", req, &result)
	return &result, err
}
