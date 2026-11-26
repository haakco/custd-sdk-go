package custd

import (
	"context"
	"net/http"
	"time"
)

// AdminSiteSearchInspectionStatus is where a URL inspection stands.
type AdminSiteSearchInspectionStatus string

// The states of AdminSiteSearchInspection.Status.
const (
	AdminSiteSearchInspectionQueued AdminSiteSearchInspectionStatus = "queued"
	AdminSiteSearchInspectionDone   AdminSiteSearchInspectionStatus = "done"
	AdminSiteSearchInspectionFailed AdminSiteSearchInspectionStatus = "failed"
)

// SiteSearchInspectionURLInvalidCode is the problem code a 400 carries when the URL to inspect is
// not an http(s) URL of at most 2048 bytes on the domain's own host.
const SiteSearchInspectionURLInvalidCode = "site_search_inspection_url_invalid"

// SiteSearchInspectionNotConnectedCode is the problem code a 409 carries when the domain is not
// verified and connected to Google Search Console.
const SiteSearchInspectionNotConnectedCode = "site_search_inspection_not_connected"

// SiteSearchInspectionLimitCode is the problem code a 429 carries once the tenant has used its
// daily URL inspections.
const SiteSearchInspectionLimitCode = "site_search_inspection_limit"

// AdminSiteSearchInspection is one Search Console URL Inspection and, once done, Google's answer.
// Every result field is Google's own value and is empty or nil when Google did not report it.
type AdminSiteSearchInspection struct {
	InspectionUUID         string                          `json:"inspectionUuid"`
	URL                    string                          `json:"url"`
	Status                 AdminSiteSearchInspectionStatus `json:"status"`
	RequestedAt            time.Time                       `json:"requestedAt"`
	InspectedAt            *time.Time                      `json:"inspectedAt,omitempty"`
	Failure                string                          `json:"failure,omitempty"`
	Verdict                string                          `json:"verdict,omitempty"`
	CoverageState          string                          `json:"coverageState,omitempty"`
	IndexingState          string                          `json:"indexingState,omitempty"`
	RobotsTxtState         string                          `json:"robotsTxtState,omitempty"`
	PageFetchState         string                          `json:"pageFetchState,omitempty"`
	LastCrawlAt            *time.Time                      `json:"lastCrawlAt,omitempty"`
	GoogleCanonical        string                          `json:"googleCanonical,omitempty"`
	UserCanonical          string                          `json:"userCanonical,omitempty"`
	CrawledAs              string                          `json:"crawledAs,omitempty"`
	MobileUsabilityVerdict string                          `json:"mobileUsabilityVerdict,omitempty"`
	ResultLink             string                          `json:"resultLink,omitempty"`
}

// AdminSiteSearchInspectionList is a domain's latest inspections and the tenant's daily use.
type AdminSiteSearchInspectionList struct {
	Inspections []AdminSiteSearchInspection `json:"inspections"`
	DailyLimit  int                         `json:"dailyLimit"`
	UsedToday   int                         `json:"usedToday"`
}

type adminSiteSearchInspectionRequest struct {
	URL string `json:"url"`
}

// ListSearchInspections returns the domain's latest URL inspections.
func (c *SiteDomainAdminClient) ListSearchInspections(
	ctx context.Context, domainUUID string,
) (*AdminSiteSearchInspectionList, error) {
	var list AdminSiteSearchInspectionList
	err := c.request(ctx, http.MethodGet, siteSearchInspectionsPath(domainUUID), nil, &list)
	return &list, err
}

// RequestSearchInspection queues a URL Inspection (202); the scheduler asks Google on its next
// tick. A URL already queued comes back as it is.
func (c *SiteDomainAdminClient) RequestSearchInspection(
	ctx context.Context, domainUUID, pageURL string,
) (*AdminSiteSearchInspection, error) {
	var inspection AdminSiteSearchInspection
	err := c.request(ctx, http.MethodPost, siteSearchInspectionsPath(domainUUID),
		adminSiteSearchInspectionRequest{URL: pageURL}, &inspection)
	return &inspection, err
}

func siteSearchInspectionsPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/search/inspections"
}
