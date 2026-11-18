package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// The values of AdminSitePageSearchQueries.State: why Queries is empty when it is.
const (
	AdminSitePageQueriesAvailable    = "available"
	AdminSitePageQueriesNotConnected = "not_connected"
	AdminSitePageQueriesNotImported  = "not_imported"
	AdminSitePageQueriesNone         = "no_queries"
)

// The templates of AdminSiteTitleOption.Template. Neither names a year.
const (
	AdminSiteTitleTemplateGeneral = "general"
	AdminSiteTitleTemplateGuide   = "guide"
)

// The values of AdminSiteLinkSource.Where: where the linking page mentions the query.
const (
	AdminSiteLinkSourceInTitle   = "title"
	AdminSiteLinkSourceInHeading = "heading"
)

// AdminSitePageQuery is one search the page showed for over the window, and whether the page's own
// title, description and main heading already use its words. Position is the impression-weighted
// average.
type AdminSitePageQuery struct {
	Query         string  `json:"query"`
	Clicks        int64   `json:"clicks"`
	Impressions   int64   `json:"impressions"`
	Position      float64 `json:"position"`
	InTitle       bool    `json:"inTitle"`
	InDescription bool    `json:"inDescription"`
	InH1          bool    `json:"inH1"`
}

// AdminSiteTitleOption is one title built from a ranking query and the Site's brand by a fixed
// template, within the title.length rule: text to copy, not AI output.
type AdminSiteTitleOption struct {
	Title       string `json:"title"`
	Template    string `json:"template"`
	Query       string `json:"query"`
	Impressions int64  `json:"impressions"`
}

// AdminSiteLinkSource is another crawled page whose title or a heading already uses one of this
// page's ranking queries, so a link from it with Query as the text reads naturally.
type AdminSiteLinkSource struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Query string `json:"query"`
	Where string `json:"where"`
}

// AdminSitePageQueryWindow is the one window of page-by-query numbers a domain keeps.
type AdminSitePageQueryWindow struct {
	Provider   string    `json:"provider"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	ImportedAt time.Time `json:"importedAt"`
}

// AdminSitePageSearchQueries is one page's ranking queries from Search Console for the newest
// window, with title options and linking pages built from them. Window is nil until an import has
// stored one; State says why Queries is empty.
type AdminSitePageSearchQueries struct {
	DomainUUID   string                    `json:"domainUuid"`
	URL          string                    `json:"url"`
	State        string                    `json:"state"`
	Window       *AdminSitePageQueryWindow `json:"window,omitempty"`
	MaxQueries   int                       `json:"maxQueries"`
	Queries      []AdminSitePageQuery      `json:"queries"`
	TitleOptions []AdminSiteTitleOption    `json:"titleOptions"`
	LinkSources  []AdminSiteLinkSource     `json:"linkSources"`
}

// GetPageSearchQueries returns the searches one page of the domain showed for, with title options
// and pages that could link to it.
func (c *SiteDomainAdminClient) GetPageSearchQueries(
	ctx context.Context, domainUUID, pageURL string,
) (*AdminSitePageSearchQueries, error) {
	var queries AdminSitePageSearchQueries
	path := siteDomainPath(domainUUID) + "/page-search-queries?" + url.Values{"url": {pageURL}}.Encode()
	err := c.request(ctx, http.MethodGet, path, nil, &queries)
	return &queries, err
}
