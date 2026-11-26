package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteSearchImpactStatus is whether a fix's search impact could be measured.
type AdminSiteSearchImpactStatus string

// The values of AdminSiteFindingImpact.Status and of a source's or period's Status (measured or
// insufficient).
const (
	AdminSiteSearchImpactOpen         AdminSiteSearchImpactStatus = "open"
	AdminSiteSearchImpactMeasured     AdminSiteSearchImpactStatus = "measured"
	AdminSiteSearchImpactInsufficient AdminSiteSearchImpactStatus = "insufficient"
)

// The bounds of GetFindingSearchImpact's days.
const (
	AdminSiteSearchImpactDefaultDays = 28
	AdminSiteSearchImpactMinDays     = 7
	AdminSiteSearchImpactMaxDays     = 90
)

// AdminSiteSearchValueStatus is whether a domain's search goal value could be measured.
type AdminSiteSearchValueStatus string

// The values of AdminSiteSearchValue.GoalStatus.
const (
	AdminSiteSearchValueMeasured    AdminSiteSearchValueStatus = "measured"
	AdminSiteSearchValuePartial     AdminSiteSearchValueStatus = "partial"
	AdminSiteSearchValueUnavailable AdminSiteSearchValueStatus = "unavailable"
)

// The bounds of GetSearchValue: days, goals per request and pages returned.
const (
	AdminSiteSearchValueDefaultDays = 28
	AdminSiteSearchValueMaxDays     = 90
	AdminSiteSearchValueMaxGoals    = 5
	AdminSiteSearchValuePages       = 50
)

// AdminSiteSearchImpactWindow is one period's UTC days (YYYY-MM-DD).
type AdminSiteSearchImpactWindow struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// AdminSiteSearchImpactPeriod is one provider's numbers for one period. An insufficient period
// says why in Reason and has nil Clicks, Impressions and CTR, never 0.
type AdminSiteSearchImpactPeriod struct {
	Status       AdminSiteSearchImpactStatus `json:"status"`
	CoveredDays  int                         `json:"coveredDays"`
	ExpectedDays int                         `json:"expectedDays"`
	Reason       string                      `json:"reason,omitempty"`
	Clicks       *int64                      `json:"clicks,omitempty"`
	Impressions  *int64                      `json:"impressions,omitempty"`
	CTR          *float64                    `json:"ctr,omitempty"`
}

// AdminSiteFindingImpactSource is one provider's numbers before and after the fix.
type AdminSiteFindingImpactSource struct {
	Provider    AdminSiteSearchProvider     `json:"provider"`
	Source      string                      `json:"source"`
	Granularity AdminSiteSearchGranularity  `json:"granularity"`
	Status      AdminSiteSearchImpactStatus `json:"status"`
	Before      AdminSiteSearchImpactPeriod `json:"before"`
	After       AdminSiteSearchImpactPeriod `json:"after"`
}

// AdminSiteFindingImpact is a finding's page's search clicks before and after an audit first saw
// it fixed. An open finding has nil FixedAt, Before and After; Note says what the numbers mean.
type AdminSiteFindingImpact struct {
	FindingUUID string                         `json:"findingUuid"`
	Rule        string                         `json:"rule"`
	URL         string                         `json:"url"`
	PageURL     string                         `json:"pageUrl"`
	Status      AdminSiteSearchImpactStatus    `json:"status"`
	FixedAt     *time.Time                     `json:"fixedAt,omitempty"`
	WindowDays  int                            `json:"windowDays"`
	Before      *AdminSiteSearchImpactWindow   `json:"before,omitempty"`
	After       *AdminSiteSearchImpactWindow   `json:"after,omitempty"`
	Sources     []AdminSiteFindingImpactSource `json:"sources"`
	Note        string                         `json:"note"`
}

// AdminSiteSearchValuePage is one search page's totals and the goals its visitors reached.
// Landings and GoalConversions are nil when goals were not measured for it.
type AdminSiteSearchValuePage struct {
	Provider        AdminSiteSearchProvider `json:"provider"`
	Source          string                  `json:"source"`
	PageURL         string                  `json:"pageUrl"`
	Path            string                  `json:"path"`
	Search          AdminSiteSearchTotals   `json:"search"`
	Landings        *int64                  `json:"landings,omitempty"`
	GoalConversions *int64                  `json:"goalConversions,omitempty"`
}

// AdminSiteSearchValue is a domain's top search pages with the analytics goals credited to them;
// Attribution explains how goals are credited.
type AdminSiteSearchValue struct {
	DomainUUID  string                     `json:"domainUuid"`
	SiteUUID    string                     `json:"siteUuid"`
	Days        int                        `json:"days"`
	From        string                     `json:"from"`
	To          string                     `json:"to"`
	Goals       []string                   `json:"goals"`
	GoalStatus  AdminSiteSearchValueStatus `json:"goalStatus"`
	GoalReason  string                     `json:"goalReason,omitempty"`
	Attribution string                     `json:"attribution"`
	Pages       []AdminSiteSearchValuePage `json:"pages"`
}

// GetFindingSearchImpact compares the finding's page's search numbers over days days (7..90)
// before and after it was fixed; 0 uses the server's 28.
func (c *SiteDomainAdminClient) GetFindingSearchImpact(
	ctx context.Context, domainUUID, findingUUID string, days int,
) (*AdminSiteFindingImpact, error) {
	var impact AdminSiteFindingImpact
	path := siteQueryPath(siteDomainPath(domainUUID)+"/findings/"+url.PathEscape(findingUUID)+"/search-impact",
		url.Values{"days": {sitePositiveInt(days)}})
	err := c.request(ctx, http.MethodGet, path, nil, &impact)
	return &impact, err
}

// GetSearchValue returns the domain's top search pages with the goals (1 to 5 analytics event
// types) credited to them over days days (1..90); 0 uses the server's 28.
func (c *SiteDomainAdminClient) GetSearchValue(
	ctx context.Context, domainUUID string, goals []string, days int,
) (*AdminSiteSearchValue, error) {
	var value AdminSiteSearchValue
	path := siteQueryPath(siteDomainPath(domainUUID)+"/search-value", url.Values{
		"goal": goals, "days": {sitePositiveInt(days)},
	})
	err := c.request(ctx, http.MethodGet, path, nil, &value)
	return &value, err
}
