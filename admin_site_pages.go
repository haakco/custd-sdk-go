package custd

import (
	"context"
	"encoding/json"
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

// The formats of AdminSiteGeneratedFix.Format. A redirect is one Apache line and one nginx line.
const (
	AdminSiteGeneratedFormatHTML      = "html"
	AdminSiteGeneratedFormatRobotsTxt = "robots.txt"
	AdminSiteGeneratedFormatRedirect  = "redirect"
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
// ShareImage is the og:image a fix would name (the largest declared image, else the first unsized
// one; never fetched), empty when the page has none or the server predates it. FAQSchema is
// FAQPage JSON-LD from the page's own question headings, nil when it asks none; Google shows no
// FAQ rich results since 7 May 2026.
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
	ShareImage  string                   `json:"shareImage,omitempty"`
	FAQSchema   *AdminSiteGeneratedFix   `json:"faqSchema,omitempty"`
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

// AdminSitePageCheck is what the page checks found on draft HTML, and the page's facts. The draft
// tools are empty when not asked for or from an older server: Slug is made from the draft's first
// h1, else its title; Redirect is the permanent redirect from PreviousURL (format redirect);
// Phrase says where the requested phrase appears.
type AdminSitePageCheck struct {
	Page     AdminSitePageFacts     `json:"page"`
	Findings []AdminSitePageFinding `json:"findings"`
	Slug     string                 `json:"slug,omitempty"`
	Redirect *AdminSiteGeneratedFix `json:"redirect,omitempty"`
	Phrase   *AdminSitePhraseUse    `json:"phrase,omitempty"`
}

// AdminSitePhraseUse is where one phrase appears in a draft and how often, with no target or
// verdict. DensityPercent is Count over WordCount as a percentage, nil when the draft has no words.
type AdminSitePhraseUse struct {
	Phrase         string   `json:"phrase"`
	InTitle        bool     `json:"inTitle"`
	InDescription  bool     `json:"inDescription"`
	InH1           bool     `json:"inH1"`
	Count          int      `json:"count"`
	WordCount      int      `json:"wordCount"`
	DensityPercent *float64 `json:"densityPercent,omitempty"`
}

// UnmarshalJSON drops the density the server sends as 0 for a draft without words.
func (p *AdminSitePhraseUse) UnmarshalJSON(data []byte) error {
	type wire AdminSitePhraseUse
	if err := json.Unmarshal(data, (*wire)(p)); err != nil {
		return err
	}
	if p.WordCount == 0 {
		p.DensityPercent = nil
	}
	return nil
}

// AdminSitePageCheckRequest checks HTML as if URL, on the domain's host, served it. HTML is the
// full page, at most 512 KiB; it is never stored. PreviousURL, the page's current address on the
// same host, asks for the redirect a move needs; Phrase (at most 100 characters) asks where it
// appears. Both are left out when empty.
type AdminSitePageCheckRequest struct {
	URL         string `json:"url"`
	HTML        string `json:"html"`
	PreviousURL string `json:"previousUrl,omitempty"`
	Phrase      string `json:"phrase,omitempty"`
}

// AdminSiteWordingCompareRequest is two drafts, as HTML or plain text (each at most 512 KiB), to
// compare. Nothing is stored or fetched.
type AdminSiteWordingCompareRequest struct {
	First  string `json:"first"`
	Second string `json:"second"`
}

// AdminSiteWordingCompare is how much wording two drafts share, measured as the crawl's
// content.near-duplicate rule measures it; wording overlap, not plagiarism. Texts under MinWords
// are not compared: Compared is false and SharedPercent nil.
type AdminSiteWordingCompare struct {
	FirstWords       int  `json:"firstWords"`
	SecondWords      int  `json:"secondWords"`
	MinWords         int  `json:"minWords"`
	Compared         bool `json:"compared"`
	SharedPercent    *int `json:"sharedPercent,omitempty"`
	ThresholdPercent int  `json:"thresholdPercent"`
	NearDuplicate    bool `json:"nearDuplicate"`
}

// UnmarshalJSON drops the share the server sends as 0 when it did not compare the texts.
func (w *AdminSiteWordingCompare) UnmarshalJSON(data []byte) error {
	type wire AdminSiteWordingCompare
	if err := json.Unmarshal(data, (*wire)(w)); err != nil {
		return err
	}
	if !w.Compared {
		w.SharedPercent = nil
	}
	return nil
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

// CheckPage runs the one-page checks on draft HTML without fetching or storing anything, and the
// draft tools the request asks for.
func (c *SiteDomainAdminClient) CheckPage(
	ctx context.Context, domainUUID string, req AdminSitePageCheckRequest,
) (*AdminSitePageCheck, error) {
	var check AdminSitePageCheck
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/page-check", req, &check)
	return &check, err
}

// CompareWording measures how much wording two drafts share, as the near-duplicate rule does.
func (c *SiteDomainAdminClient) CompareWording(
	ctx context.Context, domainUUID string, req AdminSiteWordingCompareRequest,
) (*AdminSiteWordingCompare, error) {
	var result AdminSiteWordingCompare
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/wording-compare", req, &result)
	return &result, err
}

// TestRobots answers whether a crawler may fetch a URL under a pasted or the live robots.txt.
func (c *SiteDomainAdminClient) TestRobots(
	ctx context.Context, domainUUID string, req AdminSiteRobotsTestRequest,
) (*AdminSiteRobotsTestResult, error) {
	var result AdminSiteRobotsTestResult
	err := c.request(ctx, http.MethodPost, siteDomainPath(domainUUID)+"/robots-test", req, &result)
	return &result, err
}
