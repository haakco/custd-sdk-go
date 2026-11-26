package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteTrafficCoverageState is how many of a day's access-log hours were read.
type AdminSiteTrafficCoverageState string

// The values of AdminSiteTrafficDay.Coverage: how many of the day's access-log hours were read.
const (
	AdminSiteTrafficCoverageComplete AdminSiteTrafficCoverageState = "complete"
	AdminSiteTrafficCoveragePartial  AdminSiteTrafficCoverageState = "partial"
	AdminSiteTrafficCoverageMissing  AdminSiteTrafficCoverageState = "missing"
)

// AdminSiteCrawlerIdentity is whether a crawler's address proved its name.
type AdminSiteCrawlerIdentity string

// The values of AdminSiteTrafficCrawler.Identity: whether the crawler's address proved its name.
const (
	AdminSiteCrawlerVerified   AdminSiteCrawlerIdentity = "verified"
	AdminSiteCrawlerClaimsToBe AdminSiteCrawlerIdentity = "claims_to_be"
)

// AdminSiteReferrerClass is the class of a traffic referrer.
type AdminSiteReferrerClass string

// The referrer classes of AdminSiteTrafficReferrerClass.Class and AdminSiteTrafficReferrer.Class.
const (
	AdminSiteReferrerClassSearch AdminSiteReferrerClass = "search"
	AdminSiteReferrerClassAI     AdminSiteReferrerClass = "ai"
	AdminSiteReferrerClassSocial AdminSiteReferrerClass = "social"
	AdminSiteReferrerClassOther  AdminSiteReferrerClass = "other"
)

// AdminSiteTrafficOtherKey is the key or path that sums the rows beyond a list's bound.
const AdminSiteTrafficOtherKey = "(other)"

// AdminAccessLogHourStatus is how an access-log hour stood before a replay.
type AdminAccessLogHourStatus string

// The values of AdminAccessLogHourReplay.PreviousStatus.
const (
	AdminAccessLogHourDone    AdminAccessLogHourStatus = "done"
	AdminAccessLogHourPartial AdminAccessLogHourStatus = "partial"
	AdminAccessLogHourFailed  AdminAccessLogHourStatus = "failed"
	AdminAccessLogHourMissing AdminAccessLogHourStatus = "missing"
)

// AccessLogHourInvalidCode is the problem code a 400 carries unless hourStart is the start of a
// finished UTC hour within the last 28 days.
const AccessLogHourInvalidCode = "access_log_hour_invalid"

// AdminSiteTrafficCoverage counts the access-log hours a range expected and read.
type AdminSiteTrafficCoverage struct {
	HoursExpected int  `json:"hoursExpected"`
	HoursRead     int  `json:"hoursRead"`
	HoursPartial  int  `json:"hoursPartial"`
	HoursFailed   int  `json:"hoursFailed"`
	HoursMissing  int  `json:"hoursMissing"`
	Complete      bool `json:"complete"`
}

// AdminSiteTrafficTotals are a range's totals; P95Ms is nil when no request was timed.
type AdminSiteTrafficTotals struct {
	Requests int64  `json:"requests"`
	Bytes    int64  `json:"bytes"`
	NotFound int64  `json:"notFound"`
	P95Ms    *int64 `json:"p95Ms,omitempty"`
}

// AdminSiteTrafficDay is one UTC day (YYYY-MM-DD). Requests, Bytes and NotFound are nil when no
// hour of the day was read.
type AdminSiteTrafficDay struct {
	Date         string                        `json:"date"`
	Coverage     AdminSiteTrafficCoverageState `json:"coverage"`
	HoursRead    int                           `json:"hoursRead"`
	HoursPartial int                           `json:"hoursPartial"`
	HoursFailed  int                           `json:"hoursFailed"`
	HoursMissing int                           `json:"hoursMissing"`
	Requests     *int64                        `json:"requests,omitempty"`
	Bytes        *int64                        `json:"bytes,omitempty"`
	NotFound     *int64                        `json:"notFound,omitempty"`
}

// AdminSiteTrafficCount is one key's request count.
type AdminSiteTrafficCount struct {
	Key      string `json:"key"`
	Requests int64  `json:"requests"`
}

// AdminSiteTrafficSlowPath is one slow path's request count and timings.
type AdminSiteTrafficSlowPath struct {
	Path     string `json:"path"`
	Requests int64  `json:"requests"`
	P95Ms    int64  `json:"p95Ms"`
	MaxMs    int64  `json:"maxMs"`
}

// AdminSiteTrafficReferrerClass is one referrer class's request count.
type AdminSiteTrafficReferrerClass struct {
	Class    AdminSiteReferrerClass `json:"class"`
	Requests int64                  `json:"requests"`
}

// AdminSiteTrafficReferrer is one referring domain's request count.
type AdminSiteTrafficReferrer struct {
	Class    AdminSiteReferrerClass `json:"class"`
	Domain   string                 `json:"domain"`
	Requests int64                  `json:"requests"`
}

// AdminSiteTrafficCrawler is one crawler's requests; AI marks AI crawlers.
type AdminSiteTrafficCrawler struct {
	Crawler  string                   `json:"crawler"`
	Identity AdminSiteCrawlerIdentity `json:"identity"`
	AI       bool                     `json:"ai"`
	Requests int64                    `json:"requests"`
	NotFound int64                    `json:"notFound"`
}

// AdminSiteTraffic is a domain's access-log traffic over From..To (YYYY-MM-DD UTC). Totals is nil
// when no hour was read.
type AdminSiteTraffic struct {
	DomainUUID      string                          `json:"domainUuid"`
	From            string                          `json:"from"`
	To              string                          `json:"to"`
	GeneratedAt     time.Time                       `json:"generatedAt"`
	Coverage        AdminSiteTrafficCoverage        `json:"coverage"`
	Totals          *AdminSiteTrafficTotals         `json:"totals,omitempty"`
	Days            []AdminSiteTrafficDay           `json:"days"`
	StatusClasses   []AdminSiteTrafficCount         `json:"statusClasses"`
	NotFoundPaths   []AdminSiteTrafficCount         `json:"notFoundPaths"`
	SlowPaths       []AdminSiteTrafficSlowPath      `json:"slowPaths"`
	Countries       []AdminSiteTrafficCount         `json:"countries"`
	ReferrerClasses []AdminSiteTrafficReferrerClass `json:"referrerClasses"`
	Referrers       []AdminSiteTrafficReferrer      `json:"referrers"`
	Crawlers        []AdminSiteTrafficCrawler       `json:"crawlers"`
}

// AdminSiteAIPage is one path's AI crawler hits and AI assistant referral visits.
type AdminSiteAIPage struct {
	Path                string `json:"path"`
	VerifiedCrawlerHits int64  `json:"verifiedCrawlerHits"`
	ClaimedCrawlerHits  int64  `json:"claimedCrawlerHits"`
	ReferralVisits      int64  `json:"referralVisits"`
}

// AdminSiteAIVisibility is how AI crawlers and assistants reached a domain over From..To.
type AdminSiteAIVisibility struct {
	DomainUUID  string                    `json:"domainUuid"`
	SiteUUID    string                    `json:"siteUuid"`
	From        string                    `json:"from"`
	To          string                    `json:"to"`
	GeneratedAt time.Time                 `json:"generatedAt"`
	Coverage    AdminSiteTrafficCoverage  `json:"coverage"`
	Crawlers    []AdminSiteTrafficCrawler `json:"crawlers"`
	Assistants  []AdminSiteTrafficCount   `json:"assistants"`
	Pages       []AdminSiteAIPage         `json:"pages"`
	// AnswerSample is the latest AI assistant answer sampled for the site's
	// product, nil when none was sampled.
	AnswerSample *AdminSiteAIAnswerSample `json:"answerSample,omitempty"`
}

// AdminSiteAIAnswerSample is one sampled AI assistant answer. An empty Answer
// means no answer was received.
type AdminSiteAIAnswerSample struct {
	Product    SiteAIProduct `json:"product"`
	ObservedAt time.Time     `json:"observedAt"`
	Answer     string        `json:"answer,omitempty"`
}

// AdminAccessLogHourReplay is a queued replay of one access-log hour and that hour's state before.
type AdminAccessLogHourReplay struct {
	HourStart      time.Time                `json:"hourStart"`
	PreviousStatus AdminAccessLogHourStatus `json:"previousStatus"`
	RequestedAt    time.Time                `json:"requestedAt"`
}

type adminAccessLogHourReplayRequest struct {
	HourStart time.Time `json:"hourStart"`
}

// GetTraffic returns the domain's access-log traffic from..to (YYYY-MM-DD UTC days); empty days
// use the server's default range.
func (c *SiteDomainAdminClient) GetTraffic(ctx context.Context, domainUUID, from, to string) (*AdminSiteTraffic, error) {
	var traffic AdminSiteTraffic
	err := c.request(ctx, http.MethodGet, siteRangePath(domainUUID, "/traffic", from, to), nil, &traffic)
	return &traffic, err
}

// GetAIVisibility returns how AI crawlers and assistants reached the domain from..to.
func (c *SiteDomainAdminClient) GetAIVisibility(
	ctx context.Context, domainUUID, from, to string,
) (*AdminSiteAIVisibility, error) {
	var visibility AdminSiteAIVisibility
	err := c.request(ctx, http.MethodGet, siteRangePath(domainUUID, "/ai-visibility", from, to), nil, &visibility)
	return &visibility, err
}

// ReplayAccessLogHour queues the access-log hour starting at hourStart to be read again (202). It
// is platform-wide and needs a system administrator.
func (c *SiteDomainAdminClient) ReplayAccessLogHour(
	ctx context.Context, hourStart time.Time,
) (*AdminAccessLogHourReplay, error) {
	var replay AdminAccessLogHourReplay
	err := c.admin.request(ctx, http.MethodPost, "/access-log-hours/replay",
		adminAccessLogHourReplayRequest{HourStart: hourStart.UTC()}, &replay)
	return &replay, err
}

func siteRangePath(domainUUID, route, from, to string) string {
	return siteQueryPath(siteDomainPath(domainUUID)+route, url.Values{"from": {from}, "to": {to}})
}
