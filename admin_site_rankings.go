package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// The states of AdminSiteRankObservation.State.
const (
	AdminSiteRankStateRanking    = "ranking"
	AdminSiteRankStateNotRanking = "not_ranking"
	AdminSiteRankStateNoData     = "no_data"
)

// The values of AdminSiteTrackedKeyword.Status.
const (
	AdminSiteTrackedKeywordActive = "active"
	AdminSiteTrackedKeywordPaused = "paused"
)

// AdminSiteRankAll is the country or device that tracks every country or device.
const AdminSiteRankAll = "all"

// The devices a keyword can be tracked on besides AdminSiteRankAll.
const (
	AdminSiteRankDeviceDesktop = "desktop"
	AdminSiteRankDeviceMobile  = "mobile"
	AdminSiteRankDeviceTablet  = "tablet"
)

// AdminSiteRankSourceNotConnected is AdminSiteRankings.Source when no Search Console connection
// exists; otherwise Source is an AdminSiteSearchConnectionStatus.
const AdminSiteRankSourceNotConnected = "not_connected"

// Rank tracking limits and defaults.
const (
	AdminSiteMaxTrackedKeywords        = 50
	AdminSiteDefaultRankNotifyPosition = 10
	AdminSiteRankingsDefaultDays       = 28
	AdminSiteRankingsMaxDays           = 90
)

// SiteTrackedKeywordInvalidCode is the problem code a 400 carries for a bad keyword, country,
// device or notify position.
const SiteTrackedKeywordInvalidCode = "site_tracked_keyword_invalid"

// SiteTrackedKeywordLimitCode is the problem code a 409 carries once the Site tracks
// AdminSiteMaxTrackedKeywords keywords.
const SiteTrackedKeywordLimitCode = "site_tracked_keyword_limit"

// SiteTrackedKeywordDuplicateCode is the problem code a 409 carries when the domain already
// tracks the keyword in that country and device.
const SiteTrackedKeywordDuplicateCode = "site_tracked_keyword_duplicate"

// SiteTrackedKeywordNotFoundCode is the problem code a 404 carries when the domain does not track
// the keyword.
const SiteTrackedKeywordNotFoundCode = "site_tracked_keyword_not_found"

// SiteRankingsDaysInvalidCode is the problem code a 400 carries unless days is 1 to 90.
const SiteRankingsDaysInvalidCode = "site_rankings_days_invalid"

// AdminSiteRankObservation is one day's rank. Position and RankingURL are absent unless State is
// AdminSiteRankStateRanking; a nil Position is not position 0.
type AdminSiteRankObservation struct {
	Day         string   `json:"day"`
	State       string   `json:"state"`
	RankingURL  string   `json:"rankingUrl,omitempty"`
	Position    *float64 `json:"position,omitempty"`
	Clicks      int64    `json:"clicks"`
	Impressions int64    `json:"impressions"`
}

// AdminSiteTrackedKeyword is one keyword a domain tracks, its latest rank and its daily ranks.
type AdminSiteTrackedKeyword struct {
	KeywordUUID    string                     `json:"keywordUuid"`
	Keyword        string                     `json:"keyword"`
	Country        string                     `json:"country"`
	Device         string                     `json:"device"`
	Status         string                     `json:"status"`
	NotifyPosition int                        `json:"notifyPosition"`
	CreatedAt      time.Time                  `json:"createdAt"`
	Latest         AdminSiteRankObservation   `json:"latest"`
	Observations   []AdminSiteRankObservation `json:"observations"`
	// Pool is the shared search pool's bands for the keyword; nil when the tenant has not joined
	// the pool, the keyword is below its threshold, or the server predates the pool.
	Pool *AdminSiteSearchPoolBand `json:"pool,omitempty"`
}

// AdminSiteRankings are a domain's tracked keywords over From..To; SourceFailure says why a
// failing Search Console connection failed. Pool is AdminSiteSearchPoolJoined or
// AdminSiteSearchPoolNotJoined, and empty from a server that predates the shared search pool.
type AdminSiteRankings struct {
	Source        string                    `json:"source"`
	Pool          string                    `json:"pool,omitempty"`
	SourceFailure string                    `json:"sourceFailure,omitempty"`
	From          string                    `json:"from"`
	To            string                    `json:"to"`
	MaxKeywords   int                       `json:"maxKeywords"`
	Keywords      []AdminSiteTrackedKeyword `json:"keywords"`
}

// AdminSiteTrackKeywordRequest tracks a keyword. Country and Device are omitted when empty (the
// server tracks AdminSiteRankAll); NotifyPosition 0 uses the server's 10.
type AdminSiteTrackKeywordRequest struct {
	Keyword        string `json:"keyword"`
	Country        string `json:"country,omitempty"`
	Device         string `json:"device,omitempty"`
	NotifyPosition int    `json:"notifyPosition,omitempty"`
}

// AdminSiteTrackedKeywordUpdate pauses or resumes a keyword. Paused is always sent; a
// NotifyPosition of 0 is omitted and keeps the stored alert position.
type AdminSiteTrackedKeywordUpdate struct {
	Paused         bool `json:"paused"`
	NotifyPosition int  `json:"notifyPosition,omitempty"`
}

// GetRankings returns the domain's tracked keywords over days days (1..90); 0 uses the server's 28.
func (c *SiteDomainAdminClient) GetRankings(ctx context.Context, domainUUID string, days int) (*AdminSiteRankings, error) {
	var rankings AdminSiteRankings
	path := siteQueryPath(siteDomainPath(domainUUID)+"/rankings", url.Values{"days": {sitePositiveInt(days)}})
	err := c.request(ctx, http.MethodGet, path, nil, &rankings)
	return &rankings, err
}

// TrackKeyword starts tracking a keyword on the domain (201).
func (c *SiteDomainAdminClient) TrackKeyword(
	ctx context.Context, domainUUID string, req AdminSiteTrackKeywordRequest,
) (*AdminSiteTrackedKeyword, error) {
	return c.trackedKeyword(ctx, http.MethodPost, siteTrackedKeywordsPath(domainUUID), req)
}

// UpdateTrackedKeyword pauses or resumes a tracked keyword and sets its notify position.
func (c *SiteDomainAdminClient) UpdateTrackedKeyword(
	ctx context.Context, domainUUID, keywordUUID string, update AdminSiteTrackedKeywordUpdate,
) (*AdminSiteTrackedKeyword, error) {
	return c.trackedKeyword(ctx, http.MethodPut, siteTrackedKeywordPath(domainUUID, keywordUUID), update)
}

// UntrackKeyword stops tracking a keyword and deletes its ranks.
func (c *SiteDomainAdminClient) UntrackKeyword(ctx context.Context, domainUUID, keywordUUID string) error {
	return c.request(ctx, http.MethodDelete, siteTrackedKeywordPath(domainUUID, keywordUUID), nil, nil)
}

func (c *SiteDomainAdminClient) trackedKeyword(
	ctx context.Context, method, path string, payload any,
) (*AdminSiteTrackedKeyword, error) {
	var keyword AdminSiteTrackedKeyword
	err := c.request(ctx, method, path, payload, &keyword)
	return &keyword, err
}

func siteTrackedKeywordsPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/rankings/keywords"
}

func siteTrackedKeywordPath(domainUUID, keywordUUID string) string {
	return siteTrackedKeywordsPath(domainUUID) + "/" + url.PathEscape(keywordUUID)
}
